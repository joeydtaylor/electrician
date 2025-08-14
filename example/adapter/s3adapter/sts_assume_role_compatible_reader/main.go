package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

/* ------------ data model (matches writer) ------------ */

type Feedback struct {
	CustomerID string   `parquet:"name=customerId, type=BYTE_ARRAY, convertedtype=UTF8" json:"customerId"`
	Content    string   `parquet:"name=content, type=BYTE_ARRAY, convertedtype=UTF8" json:"content"`
	Category   string   `parquet:"name=category, type=BYTE_ARRAY, convertedtype=UTF8" json:"category,omitempty"`
	IsNegative bool     `parquet:"name=isNegative, type=BOOLEAN" json:"isNegative"`
	Tags       []string `parquet:"name=tags, type=LIST, valuetype=BYTE_ARRAY, valueconvertedtype=UTF8" json:"tags,omitempty"`
}

/* ------------ env helpers & org resolution ------------ */

const (
	defaultOrgID     = "4d948fa0-084e-490b-aad5-cfd01eeab79a"
	assertJWTEnvName = "ASSERT_JWT"
	bearerJWTEnvName = "BEARER_JWT"
	orgIDEnvName     = "ORG_ID"
)

func getenv(k string) string {
	v := strings.TrimSpace(strings.Trim(os.Getenv(k), `"`))
	if v != "" {
		return v
	}
	return ""
}

func b64UrlDecodeRaw(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

func orgFromJWT(tok string) (string, bool) {
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		return "", false
	}
	payload, err := b64UrlDecodeRaw(parts[1])
	if err != nil {
		return "", false
	}
	var claims map[string]string
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", false
	}
	for _, k := range []string{"org", "org_id", "orgId", "organization"} {
		if v := strings.TrimSpace(claims[k]); v != "" {
			return v, true
		}
	}
	return "", false
}

func resolveOrgID() string {
	if v := getenv(orgIDEnvName); v != "" {
		return v
	}
	if v := getenv(assertJWTEnvName); v != "" {
		if org, ok := orgFromJWT(v); ok {
			return org
		}
	}
	if v := getenv(bearerJWTEnvName); v != "" {
		if org, ok := orgFromJWT(v); ok {
			return org
		}
	}
	return defaultOrgID
}

/* ------------ simple content filter ------------ */

func containsFold(haystack, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}

/* ------------ main ------------ */

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// LocalStack / creds
	endpoint := getenv("S3_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:4566"
	}

	cli, err := builder.NewS3ClientAssumeRole(
		ctx,
		"us-east-1",
		"arn:aws:iam::000000000000:role/exodus-dev-role",
		"electrician-reader",
		15*time.Minute,
		"",
		aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("test", "test", "")),
		endpoint,
		true, // use_path_style
	)
	if err != nil {
		panic(err)
	}

	log := builder.NewLogger(builder.LoggerWithDevelopment(true))
	bucket := getenv("S3_BUCKET")
	if bucket == "" {
		bucket = "steeze-dev"
	}

	orgID := resolveOrgID()
	basePrefix := fmt.Sprintf("org=%s/feedback/demo/", orgID)
	filter := getenv("FILTER_CONTENT_SUBSTR") // e.g. "great", "slow and buggy", "works beautifully"

	fmt.Printf("scanning prefix: s3://%s/%s (endpoint=%s)\n", bucket, basePrefix, endpoint)
	if filter != "" {
		fmt.Printf("filter: content CONTAINS %q (case-insensitive)\n", filter)
	}

	// 1) Prove there are parquet objects with a raw List
	var (
		keys []string
		cont *string
	)
	for {
		out, err := cli.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(basePrefix),
			ContinuationToken: cont,
			MaxKeys:           aws.Int32(200),
		})
		if err != nil {
			panic(err)
		}
		for _, o := range out.Contents {
			k := aws.ToString(o.Key)
			if strings.HasSuffix(strings.ToLower(k), ".parquet") {
				keys = append(keys, k)
			}
		}
		if aws.ToBool(out.IsTruncated) {
			cont = out.NextContinuationToken
			continue
		}
		break
	}
	fmt.Printf("found %d parquet objects under prefix (via ListObjectsV2)\n", len(keys))

	// 2) Use builder’s reader to fetch ALL rows under the prefix in one shot.
	//    NOTE: we pass ".parquet" as suffix so the adapter only reads parquet keys.
	reader := builder.NewS3ClientAdapter[Feedback](
		ctx,
		builder.S3ClientAdapterWithClientAndBucket[Feedback](cli, bucket),
		builder.S3ClientAdapterWithReaderListSettings[Feedback](basePrefix, ".parquet", 5000, 0),
		// Let reader auto-detect parquet by extension. Optionally force:
		// builder.S3ClientAdapterWithReaderFormat[Feedback]("parquet", ""),
		builder.S3ClientAdapterWithReaderFormatOptions[Feedback](map[string]string{
			"spill_threshold_bytes": "134217728", // 128 MiB
		}),
		builder.S3ClientAdapterWithLogger[Feedback](log),
	)

	resp, err := reader.Fetch()
	if err != nil {
		panic(err)
	}
	if resp.StatusCode == 204 || len(resp.Body) == 0 {
		fmt.Println("no rows found under prefix (reader)")
		return
	}

	total := 0
	returned := 0
	samples := make([]Feedback, 0, 5)

	for _, r := range resp.Body {
		// quick sanity for empty rows
		if r.CustomerID == "" && r.Content == "" {
			continue
		}
		total++
		if filter == "" || containsFold(r.Content, filter) {
			returned++
			if len(samples) < 5 {
				samples = append(samples, r)
			}
		}
	}

	fmt.Printf("scanned=%d, returned=%d (filter=%q)\n", total, returned, filter)
	if len(samples) > 0 {
		out, _ := json.MarshalIndent(samples, "", "  ")
		fmt.Println("sample matches:")
		fmt.Println(string(out))
	} else {
		if filter == "" {
			fmt.Println("no rows returned (try setting FILTER_CONTENT_SUBSTR, e.g. FILTER_CONTENT_SUBSTR=great)")
		} else {
			fmt.Println("no rows matched the filter")
		}
	}
}

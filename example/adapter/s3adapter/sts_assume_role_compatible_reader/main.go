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

	"github.com/joeydtaylor/electrician/pkg/builder"
)

type Feedback struct {
	CustomerID string   `parquet:"name=customerId, type=BYTE_ARRAY, convertedtype=UTF8" json:"customerId"`
	Content    string   `parquet:"name=content, type=BYTE_ARRAY, convertedtype=UTF8" json:"content"`
	Category   string   `parquet:"name=category, type=BYTE_ARRAY, convertedtype=UTF8" json:"category,omitempty"`
	IsNegative bool     `parquet:"name=isNegative, type=BOOLEAN" json:"isNegative"`
	Tags       []string `parquet:"name=tags, type=LIST, valuetype=BYTE_ARRAY, valueconvertedtype=UTF8" json:"tags,omitempty"`
}

const (
	defaultOrgID     = "4d948fa0-084e-490b-aad5-cfd01eeab79a"
	assertJWTEnvName = "ASSERT_JWT"
	bearerJWTEnvName = "BEARER_JWT"
	orgIDEnvName     = "ORG_ID"
)

func hasTag(tags []string, want string) bool {
	for _, t := range tags {
		if strings.EqualFold(t, want) {
			return true
		}
	}
	return false
}

func isTheCurlRecord(r Feedback) bool {
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(r.CustomerID)), "cli-") {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(r.Content), "works beautifully") {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(r.Category), "ux") {
		return false
	}
	if r.IsNegative {
		return false
	}
	return hasTag(r.Tags, "hermes")
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

func getenv(k string) string {
	v := strings.TrimSpace(strings.Trim(os.Getenv(k), `"`))
	if v != "" {
		return v
	}
	return ""
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

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cli, err := builder.NewS3ClientAssumeRole(
		ctx,
		"us-east-1",
		"arn:aws:iam::000000000000:role/exodus-dev-role",
		"electrician-reader",
		15*time.Minute,
		"",
		aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("test", "test", "")),
		"http://localhost:4566",
		true,
	)
	if err != nil {
		panic(err)
	}

	log := builder.NewLogger()
	const bucket = "steeze-dev"

	orgID := resolveOrgID()
	prefix := fmt.Sprintf("org=%s/feedback/demo/", orgID)
	fmt.Printf("scanning prefix: s3://%s/%s\n", bucket, prefix)

	reader := builder.NewS3ClientAdapter[Feedback](
		ctx,
		builder.S3ClientAdapterWithClientAndBucket[Feedback](cli, bucket),
		// one-shot scan under the org-scoped prefix
		builder.S3ClientAdapterWithReaderListSettings[Feedback](prefix, "", 1000, 0),
		// DO NOT force reader format; let it autodetect by .parquet extension
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
		fmt.Println("no rows found under prefix")
		return
	}

	total := 0
	matches := make([]Feedback, 0, 8)
	for _, r := range resp.Body {
		if r.CustomerID == "" && r.Content == "" {
			continue
		}
		total++
		if isTheCurlRecord(r) {
			matches = append(matches, r)
		}
	}

	fmt.Printf("scanned=%d, matches=%d\n", total, len(matches))
	if len(matches) == 0 {
		fmt.Println("no hermes CLI record found")
		return
	}

	out, _ := json.MarshalIndent(matches, "", "  ")
	fmt.Println(string(out))
}

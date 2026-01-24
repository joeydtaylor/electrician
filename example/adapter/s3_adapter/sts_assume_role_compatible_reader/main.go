package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

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

/* ------------ org resolution ------------ */

const (
	defaultOrgID     = "4d948fa0-084e-490b-aad5-cfd01eeab79a"
	assertJWTEnvName = "ASSERT_JWT"
	bearerJWTEnvName = "BEARER_JWT"
	orgIDEnvName     = "ORG_ID"
)

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
	if v := builder.EnvOr(orgIDEnvName, ""); v != "" {
		return v
	}
	if v := builder.EnvOr(assertJWTEnvName, ""); v != "" {
		if org, ok := orgFromJWT(v); ok {
			return org
		}
	}
	if v := builder.EnvOr(bearerJWTEnvName, ""); v != "" {
		if org, ok := orgFromJWT(v); ok {
			return org
		}
	}
	return defaultOrgID
}

/* ------------ utilities ------------ */

func containsFold(haystack, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}

// parseTimeFromCustomerID expects: "C-<unix_nano>-NNNN"
func parseTimeFromCustomerID(id string) (time.Time, bool) {
	parts := strings.Split(id, "-")
	if len(parts) < 3 {
		return time.Time{}, false
	}
	nanoStr := parts[1]
	ns, err := strconv.ParseInt(nanoStr, 10, 64)
	if err != nil || ns <= 0 {
		return time.Time{}, false
	}
	return time.Unix(0, ns).UTC(), true
}

type agg struct {
	byCategory map[string]int
	tagFreq    map[string]int
	phraseFreq map[string]int
	negCount   int
	total      int
}

func newAgg() *agg {
	return &agg{
		byCategory: map[string]int{},
		tagFreq:    map[string]int{},
		phraseFreq: map[string]int{},
	}
}

func (a *agg) add(r Feedback) {
	a.total++
	if r.IsNegative {
		a.negCount++
	}
	if r.Category != "" {
		a.byCategory[r.Category]++
	}
	for _, t := range r.Tags {
		if t = strings.TrimSpace(t); t != "" {
			a.tagFreq[t]++
		}
	}
	if r.Content != "" {
		a.phraseFreq[strings.ToLower(r.Content)]++
	}
}

func topN(m map[string]int, n int) [][2]string {
	type kv struct {
		k string
		v int
	}
	out := make([]kv, 0, len(m))
	for k, v := range m {
		out = append(out, kv{k, v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].v == out[j].v {
			return out[i].k < out[j].k
		}
		return out[i].v > out[j].v
	})
	lim := n
	if lim > len(out) {
		lim = len(out)
	}
	res := make([][2]string, 0, lim)
	for i := 0; i < lim; i++ {
		res = append(res, [2]string{out[i].k, strconv.Itoa(out[i].v)})
	}
	return res
}

/* ------------ main ------------ */

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// LocalStack / creds
	endpoint := builder.EnvOr("S3_ENDPOINT", "http://localhost:4566")

	cli, err := builder.NewS3ClientAssumeRoleLocalstack(ctx, builder.LocalstackS3AssumeRoleConfig{
		RoleARN:     "arn:aws:iam::000000000000:role/exodus-dev-role",
		SessionName: "electrician-reader",
		Endpoint:    endpoint,
	})
	if err != nil {
		panic(err)
	}

	log := builder.NewLogger(builder.LoggerWithDevelopment(true))
	bucket := builder.EnvOr("S3_BUCKET", "steeze-dev")

	orgID := resolveOrgID()
	basePrefix := fmt.Sprintf("org=%s/feedback/demo/", orgID)
	filter := builder.EnvOr("FILTER_CONTENT_SUBSTR", "") // e.g. "great"
	windowMin := builder.EnvIntOr("WINDOW_MINUTES", 0)   // 0 = no time filter
	sampleN := builder.EnvIntOr("SAMPLE_N", 5)           // how many examples to print
	fmt.Printf("scanning prefix: s3://%s/%s (endpoint=%s)\n", bucket, basePrefix, endpoint)
	if filter != "" {
		fmt.Printf("filter: content CONTAINS %q (case-insensitive)\n", filter)
	}
	if windowMin > 0 {
		fmt.Printf("time window: last %d minute(s)\n", windowMin)
	}

	// 1) Quick inventory (optional but nice to print)
	keys, err := builder.S3ListKeys(ctx, cli, bucket, basePrefix, ".parquet")
	if err != nil {
		panic(err)
	}
	fmt.Printf("found %d parquet objects under prefix (via ListObjectsV2)\n", len(keys))

	// 2) Read all rows under prefix (suffix restricts to parquet keys)
	reader := builder.NewS3ClientAdapter[Feedback](
		ctx,
		builder.S3ClientAdapterWithClientAndBucket[Feedback](cli, bucket),
		builder.S3ClientAdapterWithReaderListSettings[Feedback](basePrefix, ".parquet", 5000, 0),
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

	// 3) Dedup + filter + (optional) time window
	now := time.Now().UTC()
	minCutoff := now.Add(-time.Duration(windowMin) * time.Minute)

	type keyed struct {
		Feedback
		When time.Time
	}
	seen := map[string]bool{} // key: CustomerID|Content|Category|IsNegative
	rows := make([]keyed, 0, len(resp.Body))

	for _, r := range resp.Body {
		if r.CustomerID == "" && r.Content == "" {
			continue
		}
		if filter != "" && !containsFold(r.Content, filter) {
			continue
		}
		ts, ok := parseTimeFromCustomerID(r.CustomerID)
		if windowMin > 0 && ok && ts.Before(minCutoff) {
			continue
		}
		key := r.CustomerID + "|" + r.Content + "|" + r.Category + "|" + strconv.FormatBool(r.IsNegative)
		if seen[key] {
			continue
		}
		seen[key] = true
		rows = append(rows, keyed{Feedback: r, When: ts})
	}

	if len(rows) == 0 {
		fmt.Println("no rows after filters/dedup (try relaxing FILTER_CONTENT_SUBSTR or WINDOW_MINUTES)")
		return
	}

	// 4) Aggregate
	a := newAgg()
	for _, r := range rows {
		a.add(r.Feedback)
	}

	negPct := float64(0)
	if a.total > 0 {
		negPct = (float64(a.negCount) * 100.0) / float64(a.total)
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("unique_events=%d  (from scanned=%d total rows)\n", len(rows), len(resp.Body))
	if windowMin > 0 {
		fmt.Printf("window: last %d min (cutoff %s)\n", windowMin, minCutoff.Format(time.RFC3339))
	}
	if filter != "" {
		fmt.Printf("content filter: %q\n", filter)
	}
	fmt.Printf("negative_rate=%.1f%%  negatives=%d  positives=%d\n", negPct, a.negCount, a.total-a.negCount)

	// by category (sorted)
	type cv struct {
		k string
		v int
	}
	cats := make([]cv, 0, len(a.byCategory))
	for k, v := range a.byCategory {
		cats = append(cats, cv{k, v})
	}
	sort.Slice(cats, func(i, j int) bool {
		if cats[i].v == cats[j].v {
			return cats[i].k < cats[j].k
		}
		return cats[i].v > cats[j].v
	})

	fmt.Printf("\nby_category:\n")
	for _, c := range cats {
		fmt.Printf("  - %-12s : %d\n", c.k, c.v)
	}

	fmt.Printf("\ntop_tags:\n")
	for _, kv := range topN(a.tagFreq, 8) {
		fmt.Printf("  - %-16s : %s\n", kv[0], kv[1])
	}

	fmt.Printf("\ntop_phrases:\n")
	for _, kv := range topN(a.phraseFreq, 8) {
		fmt.Printf("  - %-24s : %s\n", kv[0], kv[1])
	}

	// 5) Newest samples
	sort.Slice(rows, func(i, j int) bool {
		// if parse failed, zero time sorts first, so reverse to push unknowns last
		return rows[i].When.After(rows[j].When)
	})
	if sampleN > len(rows) {
		sampleN = len(rows)
	}
	sample := make([]Feedback, 0, sampleN)
	for i := 0; i < sampleN; i++ {
		sample = append(sample, rows[i].Feedback)
	}

	out, _ := json.MarshalIndent(sample, "", "  ")
	fmt.Printf("\nnewest %d sample(s):\n%s\n", sampleN, string(out))
}

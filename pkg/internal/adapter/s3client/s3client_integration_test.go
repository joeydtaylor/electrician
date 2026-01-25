//go:build integration
// +build integration

package s3client_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

type s3IntegrationRow struct {
	ID  string `json:"id"`
	Seq int    `json:"seq"`
}

func TestS3ClientLocalstackRoundTrip(t *testing.T) {
	if os.Getenv("SKIP_LOCALSTACK") == "1" {
		t.Skip("SKIP_LOCALSTACK=1")
	}

	endpoint := envOr("LOCALSTACK_ENDPOINT", envOr("S3_ENDPOINT", "http://localhost:4566"))
	if err := waitLocalstack(endpoint, 20*time.Second); err != nil {
		t.Fatalf("localstack not healthy (%s): %v", endpoint, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	roleARN := envOr("S3_ROLE_ARN", "arn:aws:iam::000000000000:role/exodus-dev-role")
	cli, err := builder.NewS3ClientAssumeRoleLocalstack(ctx, builder.LocalstackS3AssumeRoleConfig{
		RoleARN:     roleARN,
		SessionName: "electrician-integration",
		Endpoint:    endpoint,
	})
	if err != nil {
		t.Fatalf("create s3 client: %v", err)
	}

	bucket := envOr("S3_BUCKET", "steeze-dev")
	orgID := envOr("ORG_ID", "4d948fa0-084e-490b-aad5-cfd01eeab79a")
	testID := fmt.Sprintf("it-%d", time.Now().UnixNano())
	prefix := fmt.Sprintf("org=%s/integration/%s", orgID, testID)

	writer := builder.NewS3ClientAdapter[s3IntegrationRow](
		ctx,
		builder.S3ClientAdapterWithClientAndBucket[s3IntegrationRow](cli, bucket),
		builder.S3ClientAdapterWithWriterPrefixTemplate[s3IntegrationRow](prefix),
		builder.S3ClientAdapterWithBatchSettings[s3IntegrationRow](1, 1<<20, time.Second),
	)

	rows := []s3IntegrationRow{
		{ID: testID, Seq: 1},
		{ID: testID, Seq: 2},
		{ID: testID, Seq: 3},
	}

	in := make(chan s3IntegrationRow, len(rows))
	writeErr := make(chan error, 1)
	go func() { writeErr <- writer.ServeWriter(ctx, in) }()

	for _, r := range rows {
		in <- r
	}
	close(in)

	if err := <-writeErr; err != nil {
		t.Fatalf("write failed: %v", err)
	}

	reader := builder.NewS3ClientAdapter[s3IntegrationRow](
		ctx,
		builder.S3ClientAdapterWithClientAndBucket[s3IntegrationRow](cli, bucket),
		builder.S3ClientAdapterWithReaderListSettings[s3IntegrationRow](prefix, "", 1000, 0),
	)

	got, err := fetchUntil(ctx, reader, len(rows))
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}
	if len(got) < len(rows) {
		t.Fatalf("expected %d rows, got %d", len(rows), len(got))
	}

	want := map[string]map[int]bool{testID: map[int]bool{}}
	for _, r := range rows {
		want[r.ID][r.Seq] = true
	}
	for _, r := range got {
		if r.ID != testID {
			continue
		}
		if !want[r.ID][r.Seq] {
			t.Fatalf("unexpected row: %+v", r)
		}
		want[r.ID][r.Seq] = false
	}
	for id, seqs := range want {
		for seq, ok := range seqs {
			if ok {
				t.Fatalf("missing row id=%s seq=%d", id, seq)
			}
		}
	}
}

func fetchUntil[T any](ctx context.Context, reader interface {
	Fetch() (types.HttpResponse[[]T], error)
}, want int) ([]T, error) {
	deadline := time.Now().Add(15 * time.Second)
	for {
		resp, err := reader.Fetch()
		if err != nil {
			return nil, err
		}
		if len(resp.Body) >= want {
			return resp.Body, nil
		}
		if time.Now().After(deadline) {
			return resp.Body, nil
		}
		select {
		case <-ctx.Done():
			return resp.Body, ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func waitLocalstack(endpoint string, timeout time.Duration) error {
	endpoint = strings.TrimRight(endpoint, "/")
	if endpoint == "" {
		return fmt.Errorf("missing endpoint")
	}
	client := http.Client{Timeout: 2 * time.Second}
	healthURLs := []string{endpoint + "/health", endpoint + "/_localstack/health"}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var lastErr error
		for _, u := range healthURLs {
			resp, err := client.Get(u)
			if err != nil {
				lastErr = err
				continue
			}
			_ = resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return nil
			}
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
		}
		if lastErr != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
	}
	return fmt.Errorf("timeout waiting for localstack")
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

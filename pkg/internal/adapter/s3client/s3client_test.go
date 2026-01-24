package s3client

import (
	"fmt"
	"testing"
	"time"
)

func TestS3Client_RenderKey(t *testing.T) {
	client := &S3Client[int]{
		prefixTemplate: "demo/{yyyy}/{MM}/{dd}",
		fileNameTmpl:   "{ts}",
	}

	ts := time.Date(2024, 2, 3, 4, 5, 6, 0, time.UTC)
	got := client.renderKey(ts)

	expected := fmt.Sprintf("demo/2024/02/03/%d", ts.UnixMilli())
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestS3Client_ParquetCompressionName(t *testing.T) {
	client := &S3Client[int]{
		formatOpts: map[string]string{"compression": "zstd"},
	}

	if got := client.parquetCompressionNameFromOpts(); got != "zstd" {
		t.Fatalf("expected zstd, got %q", got)
	}

	client.formatOpts = map[string]string{}
	if got := client.parquetCompressionNameFromOpts(); got != "snappy" {
		t.Fatalf("expected snappy default, got %q", got)
	}
}

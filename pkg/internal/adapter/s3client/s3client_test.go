package s3client

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
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

func TestS3Client_SetComponentMetadataPreservesType(t *testing.T) {
	adapter := NewS3ClientAdapter[int](context.Background()).(*S3Client[int])
	adapter.SetComponentMetadata("demo", "id-1")
	if adapter.componentMetadata.Type != adapter.Name() {
		t.Fatalf("expected Type %q, got %q", adapter.Name(), adapter.componentMetadata.Type)
	}
}

func TestS3Client_SetWriterConfig(t *testing.T) {
	adapter := NewS3ClientAdapter[int](context.Background()).(*S3Client[int])
	cfg := types.S3WriterConfig{
		PrefixTemplate:  "prefix/{yyyy}/{MM}",
		FileNameTmpl:    "{ts}",
		Format:          "ndjson",
		FormatOptions:   map[string]string{"gzip": "false", "foo": "bar"},
		Compression:     "gzip",
		SSEMode:         "aws:kms",
		KMSKeyID:        "key",
		BatchMaxRecords: 10,
		BatchMaxBytes:   2048,
		BatchMaxAge:     2 * time.Second,
		RawExtension:    ".bin",
		RawContentType:  "application/octet-stream",
	}
	adapter.SetWriterConfig(cfg)

	if adapter.prefixTemplate != "prefix/{yyyy}/{MM}" {
		t.Fatalf("expected prefix template to be set")
	}
	if adapter.fileNameTmpl != "{ts}" {
		t.Fatalf("expected file name template to be set")
	}
	if adapter.formatOpts["foo"] != "bar" {
		t.Fatalf("expected format option to be merged")
	}
	if adapter.sseMode != "aws:kms" || adapter.kmsKey != "key" {
		t.Fatalf("expected SSE settings to be set")
	}
	if adapter.batchMaxRecords != 10 || adapter.batchMaxBytes != 2048 {
		t.Fatalf("expected batch settings to be set")
	}
}

func TestS3Client_SetReaderConfig(t *testing.T) {
	adapter := NewS3ClientAdapter[int](context.Background()).(*S3Client[int])
	cfg := types.S3ReaderConfig{
		Prefix:        "prefix/",
		StartAfterKey: "start",
		PageSize:      10,
		ListInterval:  2 * time.Second,
		Format:        "ndjson",
		FormatOptions: map[string]string{"gzip": "auto"},
		Compression:   "gzip",
	}
	adapter.SetReaderConfig(cfg)

	if adapter.listPrefix != "prefix/" {
		t.Fatalf("expected list prefix to be set")
	}
	if adapter.readerFormatOpts["gzip"] == "" {
		t.Fatalf("expected reader format options to be merged")
	}
}

func TestS3Client_WriteOneRejectsNonNDJSON(t *testing.T) {
	adapter := NewS3ClientAdapter[int](context.Background()).(*S3Client[int])
	adapter.formatName = "parquet"
	if err := adapter.writeOne(1); err == nil {
		t.Fatalf("expected writeOne to reject non-ndjson formats")
	}
}

func TestS3Client_RenderKeyAddsSlash(t *testing.T) {
	client := &S3Client[int]{
		prefixTemplate: "demo/path",
		fileNameTmpl:   "file",
	}
	got := client.renderKey(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	if got != "demo/path/file" {
		t.Fatalf("expected demo/path/file, got %q", got)
	}
}

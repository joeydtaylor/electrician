// pkg/internal/types/s3_adapter.go
package types

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Common S3 client wiring (no envs)
type S3ClientDeps struct {
	Client         *s3.Client // required; caller constructs (LocalStack, AWS, MinIO, etc.)
	Bucket         string     // required
	ForcePathStyle bool       // default true for emulators
}

// Writer config (format + batching + layout)
type S3WriterConfig struct {
	PrefixTemplate string // e.g. "organizations/{organizationId}/events/{yyyy}/{MM}/{dd}/{HH}/{mm}/"

	// Format controls write mode:
	//   - "ndjson": adapter batches records (T) and writes line-delimited JSON files
	//   - "parquet" or "raw": adapter expects complete file payloads ([]byte) via ServeWriterRaw
	Format string

	// Compression hint (used by NDJSON path; typically "" for parquet/raw)
	// ndjson: "gzip"|"" ; parquet: "lz4"|"" (future hint only)
	Compression string

	// Server-side encryption
	SSEMode  string // "" | "AES256" | "aws:kms"
	KMSKeyID string // used when SSEMode=="aws:kms"

	// NDJSON batching thresholds
	BatchMaxRecords int           // default 50_000
	BatchMaxBytes   int           // default 128<<20
	BatchMaxAge     time.Duration // default 60s

	// Filename template for NDJSON mode; parquet/raw will still use prefix layout
	// and may override extension (see RawExtension).
	// Default NDJSON: "{ts}-{ulid}.ndjson"
	FileNameTmpl string

	// ---- RAW / Parquet single-object write settings ----
	// When Format in {"parquet","raw"} and using ServeWriterRaw:
	// - RawContentType defaults to "application/x-parquet" for "parquet", otherwise "application/octet-stream"
	// - RawExtension   defaults to ".parquet" for "parquet", otherwise ".bin"
	RawContentType string
	RawExtension   string
}

// Reader config (selection + decode)
type S3ReaderConfig struct {
	Prefix        string        // list under this prefix
	StartAfterKey string        // simple cursor
	PageSize      int32         // default 1000
	Format        string        // "ndjson" | "parquet" (parquet typically handled outside adapter)
	Compression   string        // if needed to read (ndjson gzip)
	ListInterval  time.Duration // poll interval for tailing, optional
}

type S3WriterAdapter[T any] interface {
	// Dependency/config injection
	SetS3ClientDeps(S3ClientDeps)
	SetWriterConfig(S3WriterConfig)

	// Lifecycle
	Serve(ctx context.Context, in <-chan T) error // NDJSON batching writer
	Stop()

	// Introspection / hooks
	ConnectLogger(...Logger)
	ConnectSensor(...Sensor[T])
	GetComponentMetadata() ComponentMetadata
	SetComponentMetadata(name, id string)
	NotifyLoggers(level LogLevel, format string, args ...interface{})
	Name() string
}

type S3ReaderAdapter[T any] interface {
	SetS3ClientDeps(S3ClientDeps)
	SetReaderConfig(S3ReaderConfig)

	// Serve pulls S3 objects -> emits T (Decode is adapter-owned for NDJSON)
	Serve(ctx context.Context, submit func(context.Context, T) error) error
	Stop()

	ConnectLogger(...Logger)
	ConnectSensor(...Sensor[T])
	GetComponentMetadata() ComponentMetadata
	SetComponentMetadata(name, id string)
	NotifyLoggers(level LogLevel, format string, args ...interface{})
	Name() string
}

// Options to match your pattern
type S3WriterOption[T any] func(S3WriterAdapter[T])
type S3ReaderOption[T any] func(S3ReaderAdapter[T])

// ---- Unified S3 client adapter (read + write) ----
type S3ClientAdapter[T any] interface {
	// deps/config
	SetS3ClientDeps(S3ClientDeps)
	SetWriterConfig(S3WriterConfig)
	SetReaderConfig(S3ReaderConfig)

	// writer (NDJSON batching)
	ServeWriter(ctx context.Context, in <-chan T) error

	// writer (RAW single-object mode: one []byte -> one S3 object)
	// Use when Format is "parquet" or "raw".
	ServeWriterRaw(ctx context.Context, in <-chan []byte) error

	// reader (same signature style as HTTPClientAdapter)
	Serve(ctx context.Context, submit func(context.Context, T) error) error
	Fetch() (HttpResponse[[]T], error)

	// lifecycle/introspection
	Stop()
	ConnectLogger(...Logger)
	ConnectSensor(...Sensor[T])
	GetComponentMetadata() ComponentMetadata
	SetComponentMetadata(name, id string)
	NotifyLoggers(level LogLevel, format string, args ...interface{})
	Name() string
}

// Options to match your pattern
type S3ClientOption[T any] func(S3ClientAdapter[T])

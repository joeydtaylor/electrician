// pkg/internal/types/s3_adapter.go
package types

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

////////////////////////////////////////////////////////////////////////////////
// Common S3 client wiring (no envs)
////////////////////////////////////////////////////////////////////////////////

type S3ClientDeps struct {
	Client         *s3.Client // required; caller constructs (LocalStack, AWS, MinIO, etc.)
	Bucket         string     // required
	ForcePathStyle bool       // default true for emulators
}

////////////////////////////////////////////////////////////////////////////////
// Writer config (format + batching + layout)
////////////////////////////////////////////////////////////////////////////////

type S3WriterConfig struct {
	// Object key layout (prefix) for all writes.
	// Example: "organizations/{organizationId}/events/{yyyy}/{MM}/{dd}/{HH}/{mm}/"
	PrefixTemplate string

	// Pluggable format name. Examples: "ndjson", "parquet", "raw".
	// - "ndjson" (or any record-oriented encoder): ServeWriter(ctx, <-chan T)
	// - "parquet"/"raw" (bytes passthrough): ServeWriterRaw(ctx, <-chan []byte)
	Format string

	// Additional encoder knobs (preferred over Compression).
	// e.g. {"gzip":"true"} for ndjson, {"compression":"zstd","row_group_bytes":"134217728"} for parquet.
	FormatOptions map[string]string

	// Legacy compression hint (kept for back-compat):
	// ndjson: "gzip"|"" ; parquet: ignored (use FormatOptions["compression"]).
	Compression string

	// Server-side encryption
	SSEMode  string // "" | "AES256" | "aws:kms"
	KMSKeyID string // used when SSEMode=="aws:kms"

	// Enforce server-side encryption on writes.
	RequireSSE bool

	// Batching thresholds for record-oriented encoders.
	BatchMaxRecords int           // default 50_000
	BatchMaxBytes   int           // default 128<<20
	BatchMaxAge     time.Duration // default 60s

	// Basename template (extension comes from the format). Default: "{ts}-{ulid}"
	FileNameTmpl string

	// ---- BYTES passthrough (raw/parquet single-object) ----
	// If unset:
	// - RawContentType = "application/parquet" and RawExtension = ".parquet" when Format == "parquet"
	// - otherwise RawContentType = "application/octet-stream", RawExtension = ".bin"
	RawContentType string
	RawExtension   string

	// Client-side encryption (object-level) before upload.
	// - ClientSideEncryption: "AES-GCM" (case-insensitive) to enable AES-256-GCM.
	// - ClientSideKey: 32-byte hex string.
	// - RequireClientSideEncryption: fail if encryption is not configured.
	ClientSideEncryption        string
	ClientSideKey               string
	RequireClientSideEncryption bool
}

////////////////////////////////////////////////////////////////////////////////
// Reader config (selection + decode)
////////////////////////////////////////////////////////////////////////////////

type S3ReaderConfig struct {
	// Listing / pagination / tailing
	Prefix        string        // list under this prefix
	StartAfterKey string        // simple cursor
	PageSize      int32         // default 1000
	ListInterval  time.Duration // poll interval for tailing, optional

	// Pluggable format name for decoding ("ndjson","parquet",...).
	Format string

	// Decoder knobs (e.g., {"columns":"customerId,content"} for projection).
	// Use comma-separated lists for simple knobs, or format-specific keys.
	FormatOptions map[string]string

	// Legacy field; only used for gzip-compressed NDJSON.
	Compression string

	// Client-side decryption (object-level) before decode.
	// - ClientSideEncryption: "AES-GCM" (case-insensitive).
	// - ClientSideKey: 32-byte hex string.
	// - RequireClientSideEncryption: fail if object is not encrypted.
	ClientSideEncryption        string
	ClientSideKey               string
	RequireClientSideEncryption bool
}

////////////////////////////////////////////////////////////////////////////////
// Writer (record-oriented) adapter
////////////////////////////////////////////////////////////////////////////////

type S3WriterAdapter[T any] interface {
	// Dependency/config injection
	SetS3ClientDeps(S3ClientDeps)
	SetWriterConfig(S3WriterConfig)

	// NEW: directly attach one or more wires as inputs (reads each wire's OutputChan)
	ConnectInput(...Wire[T])

	// Lifecycle (record-oriented writer via format encoder)
	// 1) classic: provide a channel explicitly
	Serve(ctx context.Context, in <-chan T) error
	// 2) new: consume from previously connected wires
	StartWriter(ctx context.Context) error

	Stop()

	// Introspection / hooks
	ConnectLogger(...Logger)
	ConnectSensor(...Sensor[T])
	GetComponentMetadata() ComponentMetadata
	SetComponentMetadata(name, id string)
	NotifyLoggers(level LogLevel, msg string, keysAndValues ...interface{})
	Name() string
}

////////////////////////////////////////////////////////////////////////////////
// Reader (record-oriented) adapter
////////////////////////////////////////////////////////////////////////////////

type S3ReaderAdapter[T any] interface {
	SetS3ClientDeps(S3ClientDeps)
	SetReaderConfig(S3ReaderConfig)

	// Serve pulls S3 objects -> emits T (decode handled by selected format).
	Serve(ctx context.Context, submit func(context.Context, T) error) error
	Stop()

	ConnectLogger(...Logger)
	ConnectSensor(...Sensor[T])
	GetComponentMetadata() ComponentMetadata
	SetComponentMetadata(name, id string)
	NotifyLoggers(level LogLevel, msg string, keysAndValues ...interface{})
	Name() string
}

////////////////////////////////////////////////////////////////////////////////
// Unified S3 client adapter (read + write)
////////////////////////////////////////////////////////////////////////////////

type S3ClientAdapter[T any] interface {
	// deps/config
	SetS3ClientDeps(S3ClientDeps)
	SetWriterConfig(S3WriterConfig)
	SetReaderConfig(S3ReaderConfig)

	// --- WRITER ---
	// record-oriented via Encoder[T] (classic explicit channel)
	ServeWriter(ctx context.Context, in <-chan T) error
	// BYTES passthrough: one []byte -> one S3 object (parquet/raw externalized)
	ServeWriterRaw(ctx context.Context, in <-chan []byte) error
	// NEW: attach input wires and then start consuming from them
	ConnectInput(...Wire[T])
	StartWriter(ctx context.Context) error

	// --- READER ---
	Serve(ctx context.Context, submit func(context.Context, T) error) error
	Fetch() (HttpResponse[[]T], error)

	// lifecycle/introspection
	Stop()
	ConnectLogger(...Logger)
	ConnectSensor(...Sensor[T])
	GetComponentMetadata() ComponentMetadata
	SetComponentMetadata(name, id string)
	NotifyLoggers(level LogLevel, msg string, keysAndValues ...interface{})
	Name() string
}

////////////////////////////////////////////////////////////////////////////////
// Options (builder-style)
////////////////////////////////////////////////////////////////////////////////

type S3WriterOption[T any] func(S3WriterAdapter[T])
type S3ReaderOption[T any] func(S3ReaderAdapter[T])
type S3ClientOption[T any] func(S3ClientAdapter[T])

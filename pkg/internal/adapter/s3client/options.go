// pkg/internal/adapter/s3client/options.go
package s3client

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3" // for WithClientAndBucket convenience
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// WithS3ClientDeps injects the AWS S3 client dependencies.
func WithS3ClientDeps[T any](deps types.S3ClientDeps) types.S3ClientOption[T] {
	return func(adp types.S3ClientAdapter[T]) {
		adp.SetS3ClientDeps(deps)
	}
}

// Convenience: WithClientAndBucket (keeps ForcePathStyle default behavior up to implementation)
func WithClientAndBucket[T any](cli *s3.Client, bucket string) types.S3ClientOption[T] {
	return func(adp types.S3ClientAdapter[T]) {
		adp.SetS3ClientDeps(types.S3ClientDeps{
			Client: cli,
			Bucket: bucket,
			// ForcePathStyle left to the adapter default unless you expose another helper
		})
	}
}

// WithWriterConfig sets the S3 writer configuration.
func WithWriterConfig[T any](cfg types.S3WriterConfig) types.S3ClientOption[T] {
	return func(adp types.S3ClientAdapter[T]) {
		adp.SetWriterConfig(cfg)
	}
}

// WithReaderConfig sets the S3 reader configuration.
func WithReaderConfig[T any](cfg types.S3ReaderConfig) types.S3ClientOption[T] {
	return func(adp types.S3ClientAdapter[T]) {
		adp.SetReaderConfig(cfg)
	}
}

// WithSensor attaches sensors to the adapter.
func WithSensor[T any](sensor ...types.Sensor[T]) types.S3ClientOption[T] {
	return func(adp types.S3ClientAdapter[T]) {
		adp.ConnectSensor(sensor...)
	}
}

// WithLogger attaches loggers to the adapter.
func WithLogger[T any](l ...types.Logger) types.S3ClientOption[T] {
	return func(adp types.S3ClientAdapter[T]) {
		adp.ConnectLogger(l...)
	}
}

// WithBatchSettings overrides the batch thresholds for writer mode.
func WithBatchSettings[T any](maxRecords, maxBytes int, maxAge time.Duration) types.S3ClientOption[T] {
	return func(adp types.S3ClientAdapter[T]) {
		cfg := types.S3WriterConfig{
			BatchMaxRecords: maxRecords,
			BatchMaxBytes:   maxBytes,
			BatchMaxAge:     maxAge,
		}
		adp.SetWriterConfig(cfg)
	}
}

// WithFormat sets the WRITER format and legacy compression (ndjson).
// (Reader format has its own helper: WithReaderFormat)
func WithFormat[T any](format, compression string) types.S3ClientOption[T] {
	return func(adp types.S3ClientAdapter[T]) {
		cfg := types.S3WriterConfig{
			Format:      format,
			Compression: compression,
		}
		adp.SetWriterConfig(cfg)
	}
}

// WithSSE configures server-side encryption for writes.
func WithSSE[T any](mode, kmsKey string) types.S3ClientOption[T] {
	return func(adp types.S3ClientAdapter[T]) {
		cfg := types.S3WriterConfig{
			SSEMode:  mode,
			KMSKeyID: kmsKey,
		}
		adp.SetWriterConfig(cfg)
	}
}

// WithReaderListSettings configures list/polling settings for reader mode.
func WithReaderListSettings[T any](prefix, startAfter string, pageSize int32, pollEvery time.Duration) types.S3ClientOption[T] {
	return func(adp types.S3ClientAdapter[T]) {
		cfg := types.S3ReaderConfig{
			Prefix:        prefix,
			StartAfterKey: startAfter,
			PageSize:      pageSize,
			ListInterval:  pollEvery,
		}
		adp.SetReaderConfig(cfg)
	}
}

// -------------------------------
// NEW: Wire integration options
// -------------------------------

// WithWire connects one or more Wire[T] as inputs to the S3 adapter.
// You can later call ServeWriterFromWires(ctx) to start streaming.
func WithWire[T any](wires ...types.Wire[T]) types.S3ClientOption[T] {
	return func(adp types.S3ClientAdapter[T]) {
		// Requires the adapter to implement ConnectInputs(...types.Wire[T])
		adp.ConnectInput(wires...)
	}
}

// ---------------------------------------
// NEW: Reader/Writer format fine-tuners
// ---------------------------------------

// WithReaderFormat sets the READER format (e.g., "parquet", "ndjson").
// The compression parameter is only meaningful for ndjson (gzip); parquet ignores it.
func WithReaderFormat[T any](format, compression string) types.S3ClientOption[T] {
	return func(adp types.S3ClientAdapter[T]) {
		cfg := types.S3ReaderConfig{
			Format:      format,
			Compression: compression,
		}
		adp.SetReaderConfig(cfg)
	}
}

// WithWriterFormatOptions merges writer format-specific knobs.
// Example: map[string]string{"parquet_compression":"zstd","row_group_bytes":"134217728"}
func WithWriterFormatOptions[T any](opts map[string]string) types.S3ClientOption[T] {
	// Make a shallow copy to avoid caller mutation later
	cp := make(map[string]string, len(opts))
	for k, v := range opts {
		cp[k] = v
	}
	return func(adp types.S3ClientAdapter[T]) {
		adp.SetWriterConfig(types.S3WriterConfig{
			FormatOptions: cp,
		})
	}
}

// WithReaderFormatOptions merges reader format-specific knobs.
// Example: map[string]string{"spill_threshold_bytes":"134217728"}
func WithReaderFormatOptions[T any](opts map[string]string) types.S3ClientOption[T] {
	cp := make(map[string]string, len(opts))
	for k, v := range opts {
		cp[k] = v
	}
	return func(adp types.S3ClientAdapter[T]) {
		adp.SetReaderConfig(types.S3ReaderConfig{
			FormatOptions: cp,
		})
	}
}

// ---------------------------------------
// NEW: Writer naming/layout helpers
// ---------------------------------------

// WithWriterPrefixTemplate sets the object key prefix pattern used by the writer.
func WithWriterPrefixTemplate[T any](prefix string) types.S3ClientOption[T] {
	return func(adp types.S3ClientAdapter[T]) {
		adp.SetWriterConfig(types.S3WriterConfig{
			PrefixTemplate: prefix,
		})
	}
}

// WithWriterFileNameTemplate sets the basename template (extension is derived from format).
func WithWriterFileNameTemplate[T any](tmpl string) types.S3ClientOption[T] {
	return func(adp types.S3ClientAdapter[T]) {
		adp.SetWriterConfig(types.S3WriterConfig{
			FileNameTmpl: tmpl,
		})
	}
}

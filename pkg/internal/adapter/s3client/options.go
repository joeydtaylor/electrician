package s3client

import (
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// WithS3ClientDeps injects the AWS S3 client dependencies.
func WithS3ClientDeps[T any](deps types.S3ClientDeps) types.S3ClientOption[T] {
	return func(adp types.S3ClientAdapter[T]) {
		adp.SetS3ClientDeps(deps)
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

// WithFormat sets the format and compression for reading/writing.
func WithFormat[T any](format, compression string) types.S3ClientOption[T] {
	return func(adp types.S3ClientAdapter[T]) {
		cfg := types.S3WriterConfig{
			Format:      format,
			Compression: compression,
		}
		adp.SetWriterConfig(cfg)
	}
}

// WithSSE configures server-side encryption.
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

// pkg/internal/adapter/s3client/s3client.go
package s3client

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// S3Client[T] concrete implementation (read + write in one component).
// Back-compat: preserves NDJSON defaults while adding pluggable format knobs and wire inputs.
type S3Client[T any] struct {
	// --- meta / lifecycle ---
	componentMetadata types.ComponentMetadata
	ctx               context.Context
	cancel            context.CancelFunc
	isServing         int32 // atomic

	// --- logging / sensors ---
	loggers     []types.Logger
	loggersLock sync.Mutex
	sensors     []types.Sensor[T]
	sensorLock  sync.Mutex

	// --- deps ---
	cli    *s3.Client
	bucket string

	// --- writer naming/layout ---
	prefixTemplate string // e.g. "organizations/{organizationId}/events/{yyyy}/{MM}/{dd}/{HH}/{mm}/"
	fileNameTmpl   string // basename; extension derived from format (default "{ts}-{ulid}")

	// --- format (pluggable via name + options) ---
	// Selection by name (e.g., "ndjson", "parquet", "raw")
	formatName  string
	formatOpts  map[string]string // encoder/format knobs
	ndjsonMime  string            // legacy defaults for NDJSON
	ndjsonEncGz bool              // legacy gzip toggle

	// --- BYTES passthrough (raw/parquet single-object) ---
	rawWriterExt         string // default ".parquet" when formatName == "parquet"
	rawWriterContentType string // default "application/octet-stream" or "application/parquet" for parquet

	// --- SSE ---
	sseMode string // "" | "AES256" | "aws:kms"
	kmsKey  string

	// --- writer batching (record-oriented encoders) ---
	batchMaxRecords int
	batchMaxBytes   int
	batchMaxAge     time.Duration

	// --- writer state ---
	buf       *bytes.Buffer
	count     int
	lastFlush time.Time

	// --- reader config/state ---
	listPrefix       string
	listStartAfter   string
	listPageSize     int32
	listPollInterval time.Duration

	// reader-side format selection
	readerFormatName string
	readerFormatOpts map[string]string

	// --- inputs from Wire(s) (fan-in) ---
	inputWires []types.Wire[T] // wires connected via ConnectInput(...)
	mergedIn   chan T          // internal merged channel fed by fan-in goroutines
}

// NewS3ClientAdapter mirrors the public constructor; options mutate the concrete impl.
func NewS3ClientAdapter[T any](ctx context.Context, options ...types.S3ClientOption[T]) types.S3ClientAdapter[T] {
	ctx, cancel := context.WithCancel(ctx)

	a := &S3Client[T]{
		ctx:    ctx,
		cancel: cancel,
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "S3_CLIENT",
		},
		loggers: make([]types.Logger, 0),
		sensors: make([]types.Sensor[T], 0),

		// --- writer defaults ---
		prefixTemplate: "organizations/{organizationId}/events/{yyyy}/{MM}/{dd}/{HH}/{mm}/",
		fileNameTmpl:   "{ts}-{ulid}", // extension added by format

		// default to NDJSON (back-compat path)
		formatName: "ndjson",
		formatOpts: map[string]string{
			"gzip": "true",
		},
		ndjsonMime:  "application/x-ndjson",
		ndjsonEncGz: true,

		// raw write defaults (Parquet-friendly)
		rawWriterExt:         ".parquet",
		rawWriterContentType: "application/octet-stream",

		// batching defaults
		batchMaxRecords: 50_000,
		batchMaxBytes:   128 << 20, // 128 MiB
		batchMaxAge:     60 * time.Second,

		// --- reader defaults ---
		listPageSize:     1000,
		listPollInterval: 5 * time.Second,
		readerFormatName: "ndjson",
		readerFormatOpts: map[string]string{"gzip": "auto"},
		listPrefix:       "",
		listStartAfter:   "",

		// --- buffer ---
		buf: bytes.NewBuffer(make([]byte, 0, 1<<20)), // 1 MiB initial cap

		// --- wire fan-in ---
		inputWires: make([]types.Wire[T], 0),
		mergedIn:   nil, // allocated when ConnectInput or StartWriter is used
	}

	for _, opt := range options {
		opt(a)
	}

	// Derive implicit content-type/extension defaults for raw/parquet if caller set formatName.
	switch a.formatName {
	case "parquet":
		if a.rawWriterExt == "" {
			a.rawWriterExt = ".parquet"
		}
		if a.rawWriterContentType == "" {
			// Commonly preferred for parquet payloads
			a.rawWriterContentType = "application/parquet"
		}
	}

	return a
}

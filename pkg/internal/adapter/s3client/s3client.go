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

// S3ClientAdapter[T] concrete implementation (read + write in one component).
type S3Client[T any] struct {
	// meta / lifecycle
	componentMetadata types.ComponentMetadata
	ctx               context.Context
	cancel            context.CancelFunc
	isServing         int32 // atomic

	// logging / sensors
	loggers     []types.Logger
	loggersLock sync.Mutex
	sensors     []types.Sensor[T]
	sensorLock  sync.Mutex

	// deps
	cli    *s3.Client
	bucket string

	// writer naming/layout
	prefixTemplate string // e.g. "organizations/{organizationId}/events/{yyyy}/{MM}/{dd}/{HH}/{mm}/"
	fileNameTmpl   string // e.g. "{ts}-{ulid}.ndjson"

	// format
	format      string // "ndjson" (MVP) | "parquet" (later)
	compression string // ndjson: "gzip" | ""

	// SSE
	sseMode string // "" | "AES256" | "aws:kms"
	kmsKey  string

	// writer batching
	batchMaxRecords int
	batchMaxBytes   int
	batchMaxAge     time.Duration

	// writer state
	buf       *bytes.Buffer
	count     int
	lastFlush time.Time

	// reader config/state
	listPrefix       string
	listStartAfter   string
	listPageSize     int32
	listPollInterval time.Duration
}

// NewS3ClientAdapter mirrors your constructor pattern; options mutate the concrete impl.
func NewS3ClientAdapter[T any](ctx context.Context, options ...types.S3ClientOption[T]) types.S3ClientAdapter[T] {
	ctx, cancel := context.WithCancel(ctx)

	a := &S3Client[T]{
		ctx:    ctx,
		cancel: cancel,
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "S3_CLIENT",
		},
		loggers:          make([]types.Logger, 0),
		sensors:          make([]types.Sensor[T], 0),
		prefixTemplate:   "organizations/{organizationId}/events/{yyyy}/{MM}/{dd}/{HH}/{mm}/",
		fileNameTmpl:     "{ts}-{ulid}.ndjson",
		format:           "ndjson",
		compression:      "gzip",
		batchMaxRecords:  50_000,
		batchMaxBytes:    128 << 20, // 128 MiB
		batchMaxAge:      60 * time.Second,
		listPageSize:     1000,
		listPollInterval: 5 * time.Second,
		buf:              bytes.NewBuffer(make([]byte, 0, 1<<20)), // 1 MiB
	}

	for _, opt := range options {
		opt(a)
	}
	return a
}

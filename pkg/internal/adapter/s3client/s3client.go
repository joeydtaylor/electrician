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

// S3Client implements the read/write S3 adapter.
type S3Client[T any] struct {
	componentMetadata types.ComponentMetadata
	ctx               context.Context
	cancel            context.CancelFunc
	isServing         int32

	loggers     []types.Logger
	loggersLock sync.Mutex
	sensors     []types.Sensor[T]
	sensorLock  sync.Mutex

	cli    *s3.Client
	bucket string

	prefixTemplate string
	fileNameTmpl   string

	formatName  string
	formatOpts  map[string]string
	ndjsonMime  string
	ndjsonEncGz bool

	rawWriterExt         string
	rawWriterContentType string

	sseMode string
	kmsKey  string

	batchMaxRecords int
	batchMaxBytes   int
	batchMaxAge     time.Duration

	buf       *bytes.Buffer
	count     int
	lastFlush time.Time

	listPrefix       string
	listStartAfter   string
	listPageSize     int32
	listPollInterval time.Duration

	readerFormatName string
	readerFormatOpts map[string]string

	inputWires []types.Wire[T]
	mergedIn   chan T
}

// NewS3ClientAdapter constructs an S3 adapter with defaults and applies options.
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

		prefixTemplate: "organizations/{organizationId}/events/{yyyy}/{MM}/{dd}/{HH}/{mm}/",
		fileNameTmpl:   "{ts}-{ulid}",

		formatName: "ndjson",
		formatOpts: map[string]string{
			"gzip": "true",
		},
		ndjsonMime:  "application/x-ndjson",
		ndjsonEncGz: true,

		rawWriterExt:         ".parquet",
		rawWriterContentType: "application/octet-stream",

		batchMaxRecords: 50_000,
		batchMaxBytes:   128 << 20,
		batchMaxAge:     60 * time.Second,

		listPageSize:     1000,
		listPollInterval: 5 * time.Second,
		readerFormatName: "ndjson",
		readerFormatOpts: map[string]string{"gzip": "auto"},
		listPrefix:       "",
		listStartAfter:   "",

		buf: bytes.NewBuffer(make([]byte, 0, 1<<20)),

		inputWires: make([]types.Wire[T], 0),
		mergedIn:   nil,
	}

	for _, opt := range options {
		opt(a)
	}

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

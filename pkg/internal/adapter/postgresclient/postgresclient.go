package postgresclient

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// PostgresClient implements the read/write Postgres adapter.
type PostgresClient[T any] struct {
	componentMetadata types.ComponentMetadata
	ctx               context.Context
	cancel            context.CancelFunc
	isServing         int32

	loggers     []types.Logger
	loggersLock sync.Mutex
	sensors     []types.Sensor[T]
	sensorLock  sync.Mutex

	connString string
	driverName string
	security   types.PostgresSecurity
	poolCfg    types.PostgresPoolSettings
	db         *sql.DB

	// writer config
	writerCfg types.PostgresWriterConfig

	// reader config
	readerCfg types.PostgresReaderConfig

	inputWires []types.Wire[T]
	mergedIn   chan T
}

// NewPostgresClientAdapter constructs a Postgres adapter with defaults and applies options.
func NewPostgresClientAdapter[T any](ctx context.Context, options ...types.PostgresClientOption[T]) types.PostgresClientAdapter[T] {
	ctx, cancel := context.WithCancel(ctx)

	a := &PostgresClient[T]{
		ctx:    ctx,
		cancel: cancel,
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "POSTGRES_CLIENT",
		},
		loggers: make([]types.Logger, 0),
		sensors: make([]types.Sensor[T], 0),
		security: types.PostgresSecurity{
			RequireTLS: true,
		},
		writerCfg:  defaultWriterConfig(),
		readerCfg:  defaultReaderConfig(),
		inputWires: make([]types.Wire[T], 0),
		mergedIn:   nil,
	}

	for _, opt := range options {
		opt(a)
	}

	return a
}

func defaultWriterConfig() types.PostgresWriterConfig {
	return types.PostgresWriterConfig{
		Table:                  "electrician_events",
		ColumnID:               "id",
		ColumnCreatedAt:        "created_at",
		ColumnTraceID:          "trace_id",
		ColumnPayload:          "payload",
		ColumnPayloadEncrypted: "payload_encrypted",
		ColumnMetadata:         "metadata",
		ColumnContentType:      "content_type",
		ColumnPayloadType:      "payload_type",
		ColumnPayloadEncoding:  "payload_encoding",
		PayloadFormat:          "json",
		BatchMaxRecords:        1000,
		BatchMaxBytes:          8 << 20,
		BatchMaxAge:            2 * time.Second,
		UseCopy:                true,
		CopyThreshold:          1000,
		AutoCreateTable:        false,
	}
}

func defaultReaderConfig() types.PostgresReaderConfig {
	return types.PostgresReaderConfig{
		Table:                  "electrician_events",
		ColumnTraceID:          "trace_id",
		ColumnPayload:          "payload",
		ColumnPayloadEncrypted: "payload_encrypted",
		ColumnMetadata:         "metadata",
		PayloadFormat:          "json",
	}
}

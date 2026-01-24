package kafkaclient

import (
	"context"
	"sync"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// KafkaClient implements the read/write Kafka adapter using injected driver handles.
type KafkaClient[T any] struct {
	componentMetadata types.ComponentMetadata
	ctx               context.Context
	cancel            context.CancelFunc
	isServingWriter   int32
	isServingReader   int32

	loggers     []types.Logger
	loggersLock sync.Mutex
	sensors     []types.Sensor[T]
	sensorLock  sync.Mutex

	brokers  []string
	producer any
	consumer any
	dlqTopic string

	wTopic        string
	wFormat       string
	wFormatOpts   map[string]string
	wCompression  string
	wKeyTemplate  string
	wHdrTemplates map[string]string

	wBatchMaxRecords int
	wBatchMaxBytes   int
	wBatchMaxAge     time.Duration

	wAcks            string
	wReqTimeout      time.Duration
	wPartitionStrat  string
	wManualPartition *int
	wEnableDLQ       bool

	rGroupID      string
	rTopics       []string
	rStartAt      string
	rStartAtTime  time.Time
	rPollInterval time.Duration
	rMaxPollRecs  int
	rMaxPollBytes int
	rFormat       string
	rFormatOpts   map[string]string
	rCommitMode   string
	rCommitPolicy string
	rCommitEvery  time.Duration

	inputWires []types.Wire[T]
	mergedIn   chan T
}

// NewKafkaClientAdapter constructs a Kafka adapter with defaults and applies options.
func NewKafkaClientAdapter[T any](ctx context.Context, options ...types.KafkaClientOption[T]) types.KafkaClientAdapter[T] {
	ctx, cancel := context.WithCancel(ctx)

	a := &KafkaClient[T]{
		ctx:    ctx,
		cancel: cancel,
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "KAFKA_CLIENT",
		},
		loggers: make([]types.Logger, 0),
		sensors: make([]types.Sensor[T], 0),

		wFormat:          "ndjson",
		wFormatOpts:      map[string]string{},
		wCompression:     "",
		wBatchMaxRecords: 50_000,
		wBatchMaxBytes:   4 << 20,
		wBatchMaxAge:     1 * time.Second,
		wAcks:            "all",
		wReqTimeout:      30 * time.Second,
		wPartitionStrat:  "hash",
		wManualPartition: nil,
		wEnableDLQ:       false,
		wHdrTemplates:    map[string]string{},

		rStartAt:      "latest",
		rPollInterval: 1 * time.Second,
		rMaxPollRecs:  10_000,
		rMaxPollBytes: 1 << 20,
		rFormat:       "ndjson",
		rFormatOpts:   map[string]string{},
		rCommitMode:   "auto",
		rCommitPolicy: "after-batch",
		rCommitEvery:  5 * time.Second,

		inputWires: make([]types.Wire[T], 0),
		mergedIn:   nil,
	}

	for _, opt := range options {
		opt(a)
	}

	return a
}

package kafkaclient

import (
	"context"
	"sync"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// KafkaClient[T] concrete implementation (read + write in one component).
// Driver-agnostic: callers inject concrete producer/consumer via types.KafkaClientDeps.
type KafkaClient[T any] struct {
	// --- meta / lifecycle ---
	componentMetadata types.ComponentMetadata
	ctx               context.Context
	cancel            context.CancelFunc
	isServingWriter   int32 // atomic
	isServingReader   int32 // atomic

	// --- logging / sensors ---
	loggers     []types.Logger
	loggersLock sync.Mutex
	sensors     []types.Sensor[T]
	sensorLock  sync.Mutex

	// --- deps (driver-agnostic) ---
	brokers  []string
	producer any
	consumer any
	dlqTopic string

	// --- writer config/state ---
	// destination / format
	wTopic        string
	wFormat       string
	wFormatOpts   map[string]string
	wCompression  string
	wKeyTemplate  string
	wHdrTemplates map[string]string

	// batching
	wBatchMaxRecords int
	wBatchMaxBytes   int
	wBatchMaxAge     time.Duration

	// producer semantics
	wAcks            string
	wReqTimeout      time.Duration
	wPartitionStrat  string
	wManualPartition *int
	wEnableDLQ       bool

	// --- reader config/state ---
	rGroupID      string
	rTopics       []string
	rStartAt      string // "latest"|"earliest"|"timestamp"
	rStartAtTime  time.Time
	rPollInterval time.Duration
	rMaxPollRecs  int
	rMaxPollBytes int
	rFormat       string
	rFormatOpts   map[string]string
	rCommitMode   string // "auto"|"manual"
	rCommitPolicy string // "after-each"|"after-batch"|"time"
	rCommitEvery  time.Duration

	// --- inputs from Wire(s) (fan-in) ---
	inputWires []types.Wire[T] // wires connected via ConnectInput(...)
	mergedIn   chan T          // internal merged channel fed by fan-in goroutines
}

// NewKafkaClientAdapter mirrors the public constructor; options mutate the concrete impl.
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

		// writer defaults
		wFormat:          "ndjson",
		wFormatOpts:      map[string]string{},
		wCompression:     "", // driver default
		wBatchMaxRecords: 50_000,
		wBatchMaxBytes:   4 << 20, // 4 MiB conservative default
		wBatchMaxAge:     1 * time.Second,
		wAcks:            "all",
		wReqTimeout:      30 * time.Second,
		wPartitionStrat:  "hash",
		wManualPartition: nil,
		wEnableDLQ:       false,
		wHdrTemplates:    map[string]string{},

		// reader defaults
		rStartAt:      "latest",
		rPollInterval: 1 * time.Second,
		rMaxPollRecs:  10_000,
		rMaxPollBytes: 1 << 20, // 1 MiB
		rFormat:       "ndjson",
		rFormatOpts:   map[string]string{},
		rCommitMode:   "auto",
		rCommitPolicy: "after-batch",
		rCommitEvery:  5 * time.Second,

		// wire fan-in
		inputWires: make([]types.Wire[T], 0),
		mergedIn:   nil,
	}

	for _, opt := range options {
		opt(a)
	}

	return a
}

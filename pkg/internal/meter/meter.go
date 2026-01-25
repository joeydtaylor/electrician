package meter

import (
	"context"
	"sync"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

const (
	defaultIdleTimeout    = 60 * time.Second
	defaultUpdateInterval = 100 * time.Millisecond
)

// Meter aggregates metrics and emits periodic summaries for a running pipeline.
type Meter[T any] struct {
	ctx               context.Context
	cancel            context.CancelFunc
	componentMetadata types.ComponentMetadata

	mu               sync.Mutex
	counts           map[string]*uint64
	totals           map[string]uint64
	thresholds       map[string]float64
	metrics          map[string]*types.MetricInfo
	startTimes       map[string]time.Time
	monitoredMetrics []*types.MetricInfo
	metricNames      []string
	totalItems       uint64
	idleTimeout      time.Duration
	idleTimer        *time.Timer
	dataCh           chan struct{}
	ticker           *time.Ticker

	loggersMu sync.Mutex
	loggers   []types.Logger

	pauseCh   chan struct{}
	startTime time.Time
	endTime   time.Time
}

// NewMeter constructs a Meter with defaults and applies any provided options.
func NewMeter[T any](ctx context.Context, options ...types.Option[types.Meter[T]]) types.Meter[T] {
	ctx, cancel := context.WithCancel(ctx)
	m := &Meter[T]{
		ctx:              ctx,
		cancel:           cancel,
		idleTimeout:      defaultIdleTimeout,
		idleTimer:        time.NewTimer(defaultIdleTimeout),
		dataCh:           make(chan struct{}, 1),
		counts:           make(map[string]*uint64),
		totals:           make(map[string]uint64),
		metrics:          make(map[string]*types.MetricInfo),
		startTimes:       make(map[string]time.Time),
		thresholds:       make(map[string]float64),
		loggers:          make([]types.Logger, 0),
		pauseCh:          make(chan struct{}, 1),
		ticker:           time.NewTicker(defaultUpdateInterval),
		metricNames:      make([]string, 0, len(defaultMetricNames)),
		monitoredMetrics: make([]*types.MetricInfo, 0),
	}

	m.initializeMetrics()

	for _, opt := range options {
		opt(m)
	}

	go m.monitorIdleTime(ctx)

	return m
}

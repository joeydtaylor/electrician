package meter

import (
	"context"
	"sync"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

type Meter[T any] struct {
	ctx               context.Context
	cancel            context.CancelFunc
	componentMetadata types.ComponentMetadata
	counts            map[string]*uint64
	totals            map[string]uint64
	metrics           map[string]*types.MetricInfo

	startTimes       map[string]time.Time
	mutex            sync.Mutex
	loggers          []types.Logger
	loggersLock      sync.Mutex
	sensors          []types.Sensor[T]
	pauseChan        chan bool
	totalItems       uint64
	thresholds       map[string]float64 // Tracks thresholds for various metrics
	ticker           *time.Ticker
	startTime        time.Time
	endTime          time.Time
	idleTimer        *time.Timer
	idleTimeout      time.Duration
	dataChan         chan bool // Channel to signal new data
	monitoredMetrics []*types.MetricInfo
	metricNames      []string
}

func NewMeter[T any](ctx context.Context, options ...types.Option[types.Meter[T]]) types.Meter[T] {
	ctx, cancel := context.WithCancel(ctx)
	m := &Meter[T]{
		ctx:              ctx,
		cancel:           cancel,
		idleTimeout:      60 * time.Second,
		idleTimer:        time.NewTimer(60 * time.Second),
		dataChan:         make(chan bool),
		counts:           make(map[string]*uint64),
		totals:           make(map[string]uint64),
		monitoredMetrics: make([]*types.MetricInfo, 0),
		metrics:          make(map[string]*types.MetricInfo), // Initialize the metrics map
		startTimes:       make(map[string]time.Time),
		thresholds:       make(map[string]float64),
		loggers:          make([]types.Logger, 0),
		sensors:          make([]types.Sensor[T], 0),
		pauseChan:        make(chan bool, 1), // Ensure non-blocking sends
		ticker:           time.NewTicker(100 * time.Millisecond),
	}

	m.initializeMetrics() // Make sure the counts are initialized

	for _, opt := range options {
		opt(m)
	}

	go m.monitorIdleTime(ctx)

	return m
}

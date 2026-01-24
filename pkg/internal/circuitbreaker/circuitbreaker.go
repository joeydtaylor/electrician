// Package circuitbreaker provides a simple, in-process circuit breaker implementation.
package circuitbreaker

import (
	"context"
	"sync"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// CircuitBreaker guards a pipeline by stopping work when errors cross a threshold.
type CircuitBreaker[T any] struct {
	componentMetadata types.ComponentMetadata

	ctx    context.Context
	cancel context.CancelFunc

	errorThreshold int
	timeWindow     time.Duration
	debounceNanos  int64

	stateLock     sync.Mutex
	allowed       bool
	errorCount    int
	lastErrorTime time.Time
	lastTripped   time.Time

	configLock   sync.Mutex
	loggers      []types.Logger
	sensors      []types.Sensor[T]
	neutralWires []types.Wire[T]

	resetNotifyChan chan struct{}
}

// NewCircuitBreaker constructs a CircuitBreaker with the provided policy.
func NewCircuitBreaker[T any](ctx context.Context, errorThreshold int, timeWindow time.Duration, options ...types.Option[types.CircuitBreaker[T]]) types.CircuitBreaker[T] {
	if ctx == nil {
		ctx = context.Background()
	}

	if timeWindow < 0 {
		timeWindow = 0
	}

	ctx, cancel := context.WithCancel(ctx)
	cb := &CircuitBreaker[T]{
		ctx:    ctx,
		cancel: cancel,
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "CIRCUIT_BREAKER",
		},
		errorThreshold:  errorThreshold,
		timeWindow:      timeWindow,
		allowed:         true,
		loggers:         make([]types.Logger, 0),
		sensors:         make([]types.Sensor[T], 0),
		neutralWires:    make([]types.Wire[T], 0),
		resetNotifyChan: make(chan struct{}, 1),
	}

	for _, option := range options {
		if option == nil {
			continue
		}
		option(cb)
	}

	cb.notifyCircuitBreakerCreation(cb.snapshotSensors())

	return cb
}

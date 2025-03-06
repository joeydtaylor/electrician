package generator

import (
	"context"
	"sync"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

type Generator[T any] struct {
	wg                  sync.WaitGroup
	ctx                 context.Context
	cancel              context.CancelFunc
	configLock          sync.Mutex
	loggers             []types.Logger
	loggersLock         sync.Mutex
	componentMetadata   types.ComponentMetadata
	connectedComponents []types.Submitter[T]
	plugs               []types.Plug[T]
	started             int32
	sensors             []types.Sensor[T]
	CircuitBreaker      types.CircuitBreaker[T] // Circuit breaker for handling errors.
	controlChan         chan bool               // Channel for controlling circuit breaker.
	cbLock              sync.Mutex              // Protects access to the circuitBreaker.
	stopOnce            sync.Once
}

func NewGenerator[T any](ctx context.Context, options ...types.Option[types.Generator[T]]) types.Generator[T] {

	ctx, cancel := context.WithCancel(ctx)

	g := &Generator[T]{
		ctx:    ctx,
		cancel: cancel,
		componentMetadata: types.ComponentMetadata{
			Type: "GENERATOR",
			ID:   utils.GenerateUniqueHash(),
		},
		sensors:             []types.Sensor[T]{}, // Initialize empty list of sensors.
		loggers:             make([]types.Logger, 0),
		connectedComponents: make([]types.Submitter[T], 0),
		plugs:               make([]types.Plug[T], 0),
		controlChan:         make(chan bool, 1), // Ensure non-blocking sends with buffered channel
	}

	for _, opt := range options {
		opt(g)
	}

	if len(g.sensors) != 0 {
		for _, p := range g.plugs {
			p.ConnectSensor(g.sensors...)
		}
	}

	return g
}

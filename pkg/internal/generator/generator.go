package generator

import (
	"context"
	"sync"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// Generator produces elements from plugs and submits them downstream.
// Configuration is immutable after Start; stop the generator before mutating settings.
type Generator[T any] struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	configLock  sync.Mutex
	loggers     []types.Logger
	loggersLock sync.Mutex
	sensors     []types.Sensor[T]

	componentMetadata   types.ComponentMetadata
	connectedComponents []types.Submitter[T]
	plugs               []types.Plug[T]

	started        int32
	CircuitBreaker types.CircuitBreaker[T]
	controlChan    chan bool
	cbLock         sync.Mutex
	stopLock       sync.Mutex
	stopOnce       sync.Once
}

// NewGenerator constructs a Generator with defaults and applies options.
func NewGenerator[T any](ctx context.Context, options ...types.Option[types.Generator[T]]) types.Generator[T] {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	g := &Generator[T]{
		ctx:    ctx,
		cancel: cancel,
		componentMetadata: types.ComponentMetadata{
			Type: "GENERATOR",
			ID:   utils.GenerateUniqueHash(),
		},
		sensors:             make([]types.Sensor[T], 0),
		loggers:             make([]types.Logger, 0),
		connectedComponents: make([]types.Submitter[T], 0),
		plugs:               make([]types.Plug[T], 0),
		controlChan:         make(chan bool, 1),
	}

	for _, opt := range options {
		if opt == nil {
			continue
		}
		opt(g)
	}

	g.attachSensorsToPlugs(g.plugs)

	return g
}

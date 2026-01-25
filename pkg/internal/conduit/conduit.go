package conduit

import (
	"bytes"
	"context"
	"runtime"
	"sync"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// Conduit orchestrates the flow of data through a series of wires.
type Conduit[T any] struct {
	componentMetadata types.ComponentMetadata

	wires []types.Wire[T]

	loggers     []types.Logger
	loggersLock sync.Mutex

	sensors []types.Sensor[T]

	ctx    context.Context
	cancel context.CancelFunc

	Outputs []*bytes.Buffer

	generators []types.Generator[T]

	NextConduit types.Conduit[T]

	OutputChan chan T
	InputChan  chan T

	completeSignal sync.WaitGroup
	terminateOnce  sync.Once

	CircuitBreaker types.CircuitBreaker[T]
	cbLock         sync.Mutex

	started int32

	MaxBufferSize  int
	MaxConcurrency int

	surgeProtector     types.SurgeProtector[T]
	configLock         sync.Mutex
	surgeProtectorLock sync.Mutex
}

// NewConduit constructs a conduit with defaults and applies options.
func NewConduit[T any](ctx context.Context, options ...types.Option[types.Conduit[T]]) types.Conduit[T] {
	if ctx == nil {
		ctx = context.Background()
	}

	childCtx, cancel := context.WithCancel(ctx)

	workers := runtime.GOMAXPROCS(0)
	if workers < 1 {
		workers = 1
	}

	c := &Conduit[T]{
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "CONDUIT",
		},
		wires:          make([]types.Wire[T], 0),
		generators:     make([]types.Generator[T], 0),
		ctx:            childCtx,
		cancel:         cancel,
		completeSignal: sync.WaitGroup{},
		started:        0,
		MaxBufferSize:  1024,
		MaxConcurrency: workers,
	}

	for _, option := range options {
		if option == nil {
			continue
		}
		option(c)
	}

	c.ensureDefaults()
	c.chainWires()

	return c
}

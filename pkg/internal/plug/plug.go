package plug

import (
	"context"
	"sync"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// Plug binds adapters and adapter funcs for generators to consume.
// Configuration is immutable after Freeze; avoid mutating once a generator starts.
type Plug[T any] struct {
	componentMetadata types.ComponentMetadata
	ctx               context.Context
	cancel            context.CancelFunc

	loggers     []types.Logger
	loggersLock sync.Mutex
	configLock  sync.Mutex
	sensors     []types.Sensor[T]
	adapterFns  []types.AdapterFunc[T]
	adapters    []types.Adapter[T]

	frozen int32
}

// NewPlug constructs a Plug with defaults and applies options.
func NewPlug[T any](ctx context.Context, options ...types.Option[types.Plug[T]]) types.Plug[T] {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	p := &Plug[T]{
		ctx:        ctx,
		cancel:     cancel,
		loggers:    make([]types.Logger, 0),
		sensors:    make([]types.Sensor[T], 0),
		adapterFns: make([]types.AdapterFunc[T], 0),
		adapters:   make([]types.Adapter[T], 0),
		componentMetadata: types.ComponentMetadata{
			Type: "PLUG",
			ID:   utils.GenerateUniqueHash(),
		},
	}

	for _, opt := range options {
		if opt == nil {
			continue
		}
		opt(p)
	}

	return p
}

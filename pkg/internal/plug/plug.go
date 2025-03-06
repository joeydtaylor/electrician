package plug

import (
	"context"
	"sync"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

type Plug[T any] struct {
	componentMetadata types.ComponentMetadata
	ctx               context.Context
	cancel            context.CancelFunc
	loggers           []types.Logger
	loggersLock       sync.Mutex
	configLock        sync.Mutex
	sensors           []types.Sensor[T]
	plugFunc          []types.AdapterFunc[T]
	clients           []types.Adapter[T]
}

func NewPlug[T any](ctx context.Context, options ...types.Option[types.Plug[T]]) types.Plug[T] {
	ctx, cancel := context.WithCancel(ctx)

	p := &Plug[T]{
		ctx:      ctx,
		cancel:   cancel,
		loggers:  make([]types.Logger, 0),
		sensors:  make([]types.Sensor[T], 0),
		plugFunc: make([]types.AdapterFunc[T], 0),
		clients:  make([]types.Adapter[T], 0),
	}

	for _, opt := range options {
		opt(p)
	}

	return p
}

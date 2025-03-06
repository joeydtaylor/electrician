package resister

import (
	"context"
	"sync"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue[T any] []*types.Element[T]

// Resister is a thread-safe wrapper around PriorityQueue.
type Resister[T any] struct {
	ctx               context.Context
	cancel            context.CancelFunc
	pq                PriorityQueue[T]
	indexMap          map[string]*types.Element[T]
	lock              sync.Mutex
	loggersLock       sync.Mutex
	loggers           []types.Logger
	componentMetadata types.ComponentMetadata
	sensors           []types.Sensor[T]
	sensorLock        sync.Mutex
}

func NewResister[T any](ctx context.Context, options ...types.Option[types.Resister[T]]) types.Resister[T] {
	ctx, cancel := context.WithCancel(ctx)

	r := &Resister[T]{
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "RESISTER",
		},
		ctx:      ctx,
		cancel:   cancel,
		pq:       make(PriorityQueue[T], 0),
		indexMap: make(map[string]*types.Element[T]),
		loggers:  make([]types.Logger, 0),
	}

	for _, opt := range options {
		opt(r)
	}

	return r
}

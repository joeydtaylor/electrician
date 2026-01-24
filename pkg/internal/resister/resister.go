package resister

import (
	"context"
	"sync"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// Resister is an in-memory, concurrency-safe queue used by surge protection logic.
type Resister[T any] struct {
	componentMetadata types.ComponentMetadata

	queue     priorityQueue[T]
	indexByID map[string]*types.Element[T]
	queueLock sync.Mutex

	loggers     []types.Logger
	loggersLock sync.Mutex

	sensors    []types.Sensor[T]
	sensorLock sync.Mutex

	metadataLock sync.Mutex
}

// NewResister constructs a new resister instance.
func NewResister[T any](ctx context.Context, options ...types.Option[types.Resister[T]]) types.Resister[T] {
	if ctx == nil {
		ctx = context.Background()
	}

	r := &Resister[T]{
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "RESISTER",
		},
		queue:     make(priorityQueue[T], 0),
		indexByID: make(map[string]*types.Element[T]),
		loggers:   make([]types.Logger, 0),
		sensors:   make([]types.Sensor[T], 0),
	}

	for _, opt := range options {
		if opt == nil {
			continue
		}
		opt(r)
	}

	return r
}

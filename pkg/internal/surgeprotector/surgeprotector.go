package surgeprotector

import (
	"context"
	"sync"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// SurgeProtector rate-limits submissions and optionally buffers work in a resister queue.
type SurgeProtector[T any] struct {
	componentMetadata types.ComponentMetadata
	metadataLock      sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc

	active int32

	managedLock       sync.Mutex
	managedComponents []types.Submitter[T]

	backupLock    sync.Mutex
	backupSystems []types.Submitter[T]

	loggersLock sync.Mutex
	loggers     []types.Logger

	sensorLock sync.Mutex
	sensors    []types.Sensor[T]

	resisterLock    sync.Mutex
	resister        types.Resister[T]
	resisterEnabled int32

	rateLock         sync.Mutex
	tokens           int32
	capacity         int32
	refillRate       time.Duration
	nextRefill       time.Time
	backoffJitter    int
	maxRetryAttempts int
	rateLimited      int32
}

// NewSurgeProtector constructs a surge protector with the provided options.
func NewSurgeProtector[T any](ctx context.Context, options ...types.Option[types.SurgeProtector[T]]) types.SurgeProtector[T] {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	sp := &SurgeProtector[T]{
		ctx:    ctx,
		cancel: cancel,
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "SURGE_PROTECTOR",
		},
		managedComponents: make([]types.Submitter[T], 0),
		backupSystems:     make([]types.Submitter[T], 0),
		loggers:           make([]types.Logger, 0),
		sensors:           make([]types.Sensor[T], 0),
	}

	for _, option := range options {
		if option == nil {
			continue
		}
		option(sp)
	}

	sp.notifySurgeProtectorCreation(sp.snapshotSensors())

	return sp
}

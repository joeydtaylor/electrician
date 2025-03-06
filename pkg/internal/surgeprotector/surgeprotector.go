package surgeprotector

import (
	"context"
	"sync"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types" // Import your queue package.
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

type SurgeProtector[T any] struct {
	managedComponents []types.Submitter[T]
	backupSystems     []types.Submitter[T]
	ctx               context.Context
	cancel            context.CancelFunc
	active            int32 // Use int32 for atomic operations
	loggers           []types.Logger
	loggersLock       sync.Mutex
	componentMetadata types.ComponentMetadata
	sensors           []types.Sensor[T] // List of sensors attached to the wire.
	sensorLock        sync.Mutex
	mu                sync.Mutex
	backoffJitter     int
	maxRetryAttempts  int
	tokens            int32
	capacity          int32
	refillRate        time.Duration
	nextRefill        time.Time
	resister          types.Resister[T] // Updated to use Resister from the queue package
	resisterEnabled   int32             // 0 = inactive, 1 = active
	rateLimited       int32             // 0 = inactive, 1 = active
	tokenLock         sync.Mutex
}

// NewSurgeProtector creates a new instance of a SurgeProtector.
func NewSurgeProtector[T any](ctx context.Context, options ...types.Option[types.SurgeProtector[T]]) types.SurgeProtector[T] {
	ctx, cancel := context.WithCancel(ctx)

	sp := &SurgeProtector[T]{
		ctx:     ctx,
		cancel:  cancel,
		sensors: make([]types.Sensor[T], 0),
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(), // Generate unique ID for the wire.
			Type: "SURGE_PROTECTOR",
		},
		managedComponents: make([]types.Submitter[T], 0),
		backupSystems:     make([]types.Submitter[T], 0),
	}

	// Apply all provided options to configure the SurgeProtector.
	for _, option := range options {
		option(sp) // Apply each option directly to the SurgeProtector instance.
	}

	sp.notifySurgeProtectorCreation()

	return sp
}

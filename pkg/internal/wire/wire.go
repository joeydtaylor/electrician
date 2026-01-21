package wire

import (
	"bytes"
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// Wire is the core in-process pipeline primitive.
// Contract: configure (connect components / set options) before Start.
// Mutating configuration while running is not supported.
type Wire[T any] struct {
	// Identity
	componentMetadata types.ComponentMetadata

	// Primary data path
	inChan     chan T
	OutputChan chan T

	// Optional encoded output
	bufferMutex  sync.Mutex
	OutputBuffer *bytes.Buffer
	encoder      types.Encoder[T]

	// Telemetry
	loggers     []types.Logger
	loggersLock sync.Mutex
	sensors     []types.Sensor[T]
	sensorLock  sync.Mutex

	// Generators and transforms
	generators      []types.Generator[T]
	transformations []types.Transformer[T]

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Completion / termination
	completeSignal      chan struct{}
	errorChan           chan types.ElementError[T]
	closeErrorChanOnce  sync.Once
	closeOutputChanOnce sync.Once
	closeInputChanOnce  sync.Once
	terminateOnce       sync.Once

	// Circuit breaker
	cbLock         sync.Mutex
	CircuitBreaker types.CircuitBreaker[T]
	controlChan    chan bool

	// Concurrency
	maxBufferSize   int
	maxConcurrency  int
	started         int32
	loggerCount     int32
	sensorCount     int32
	fastPathEnabled bool
	fastTransform   types.Transformer[T]

	// Retry / recovery
	insulatorFunc  func(ctx context.Context, elem T, err error) (T, error)
	retryThreshold int
	retryInterval  time.Duration

	// Rate limiting / queueing
	surgeProtector       types.SurgeProtector[T]
	queueRetryMaxAttempt int

	// State guard (used by Stop/Restart)
	isClosed  bool
	closeLock sync.Mutex

	transformerFactory func() types.Transformer[T]
}

// NewWire constructs a wire with sane defaults and applies any options.
// Channels are allocated after options run so WithConcurrencyControl can affect capacity.
func NewWire[T any](ctx context.Context, options ...types.Option[types.Wire[T]]) types.Wire[T] {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	workers := runtime.GOMAXPROCS(0)
	if workers < 1 {
		workers = 1
	}

	w := &Wire[T]{
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "WIRE",
		},
		ctx:            ctx,
		cancel:         cancel,
		completeSignal: make(chan struct{}),
		controlChan:    make(chan bool, 1),

		maxBufferSize:  1024,
		maxConcurrency: workers,

		// Leave slices nil by default; append handles nil efficiently.
	}

	for _, opt := range options {
		if opt == nil {
			continue
		}
		opt(w)
	}

	// Allocate channels/buffers after options.
	if w.maxConcurrency < 1 {
		w.maxConcurrency = 1
	}
	if w.maxBufferSize < 0 {
		w.maxBufferSize = 0
	}

	if w.inChan == nil {
		w.inChan = make(chan T, w.maxBufferSize)
	}
	if w.OutputChan == nil {
		w.OutputChan = make(chan T, w.maxBufferSize)
	}
	if w.errorChan == nil {
		w.errorChan = make(chan types.ElementError[T], w.maxBufferSize)
	}
	if w.OutputBuffer == nil {
		w.OutputBuffer = &bytes.Buffer{}
	}

	if w.CircuitBreaker != nil {
		go w.startCircuitBreakerTicker()
	}

	return w
}

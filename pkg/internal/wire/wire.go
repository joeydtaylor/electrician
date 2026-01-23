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

// Wire is a concurrent in-process pipeline stage.
// Configuration must not be mutated after Start; stop the wire before changing settings.
type Wire[T any] struct {
	componentMetadata types.ComponentMetadata

	inChan     chan T
	OutputChan chan T

	bufferMutex  sync.Mutex
	OutputBuffer *bytes.Buffer
	encoder      types.Encoder[T]

	loggers     []types.Logger
	loggersLock sync.Mutex
	sensors     []types.Sensor[T]
	sensorLock  sync.Mutex

	generators      []types.Generator[T]
	transformations []types.Transformer[T]

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	completeSignal      chan struct{}
	errorChan           chan types.ElementError[T]
	closeErrorChanOnce  sync.Once
	closeOutputChanOnce sync.Once
	closeInputChanOnce  sync.Once
	terminateOnce       sync.Once

	cbLock         sync.Mutex
	CircuitBreaker types.CircuitBreaker[T]
	controlChan    chan bool

	maxBufferSize   int
	maxConcurrency  int
	started         int32
	loggerCount     int32
	sensorCount     int32
	fastPathEnabled bool
	fastTransform   types.Transformer[T]

	insulatorFunc  func(ctx context.Context, elem T, err error) (T, error)
	retryThreshold int
	retryInterval  time.Duration

	surgeProtector       types.SurgeProtector[T]
	queueRetryMaxAttempt int

	isClosed  bool
	closeLock sync.Mutex

	transformerFactory func() types.Transformer[T]
}

// NewWire constructs a Wire with defaults and applies options.
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
	}

	for _, opt := range options {
		if opt == nil {
			continue
		}
		opt(w)
	}

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

//go:build nats

package natsrelay

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"

	"github.com/joeydtaylor/electrician/pkg/internal/receivingrelay"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// ReceivingRelay subscribes to a NATS subject and forwards decoded items to submitters.
type ReceivingRelay[T any] struct {
	URL     string
	Subject string
	Queue   string
	Conn    *nats.Conn

	ctx    context.Context
	cancel context.CancelFunc

	Outputs []types.Submitter[T]

	componentMetadata types.ComponentMetadata

	DataCh chan T

	Loggers     []types.Logger
	loggersLock sync.Mutex

	DecryptionKey string

	staticHeaders map[string]string
	passthrough   bool

	isRunning    int32
	configFrozen int32

	outputsOnce sync.Once
	sub         *nats.Subscription
}

func NewReceivingRelay[T any](ctx context.Context, options ...types.Option[*ReceivingRelay[T]]) *ReceivingRelay[T] {
	ctx, cancel := context.WithCancel(ctx)
	rr := &ReceivingRelay[T]{
		ctx:    ctx,
		cancel: cancel,
		DataCh: make(chan T, 1024),
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "NATS_RECEIVING_RELAY",
		},
		Loggers:       make([]types.Logger, 0),
		Subject:       "relay",
		staticHeaders: make(map[string]string),
	}
	for _, option := range options {
		option(rr)
	}
	return rr
}

func (rr *ReceivingRelay[T]) IsRunning() bool {
	return atomic.LoadInt32(&rr.isRunning) == 1
}

func (rr *ReceivingRelay[T]) Start(ctx context.Context) error {
	atomic.StoreInt32(&rr.configFrozen, 1)
	rr.startOutputFanout()
	if rr.Conn == nil {
		conn, err := nats.Connect(rr.URL)
		if err != nil {
			return err
		}
		rr.Conn = conn
	}

	subj := normalizeSubject(rr.Subject)
	var err error
	if rr.Queue != "" {
		rr.sub, err = rr.Conn.QueueSubscribe(subj, rr.Queue, rr.onMsg)
	} else {
		rr.sub, err = rr.Conn.Subscribe(subj, rr.onMsg)
	}
	if err != nil {
		return err
	}

	atomic.StoreInt32(&rr.isRunning, 1)
	return nil
}

func (rr *ReceivingRelay[T]) Stop() {
	rr.cancel()
	if rr.sub != nil {
		_ = rr.sub.Unsubscribe()
	}
	if rr.Conn != nil {
		rr.Conn.Close()
	}
	atomic.StoreInt32(&rr.isRunning, 0)
}

func (rr *ReceivingRelay[T]) startOutputFanout() {
	if len(rr.Outputs) == 0 {
		return
	}
	rr.outputsOnce.Do(func() {
		go func() {
			for data := range rr.DataCh {
				for _, out := range rr.Outputs {
					_ = out.Submit(rr.ctx, data)
				}
			}
		}()
	})
}

func (rr *ReceivingRelay[T]) onMsg(msg *nats.Msg) {
	wp := &relay.WrappedPayload{}
	if err := proto.Unmarshal(msg.Data, wp); err != nil {
		return
	}
	if rr.passthrough {
		if cast, ok := any(wp).(T); ok {
			rr.DataCh <- cast
		}
		return
	}

	var out T
	if err := receivingrelay.UnwrapPayload(wp, rr.DecryptionKey, &out); err != nil {
		return
	}
	rr.DataCh <- out
}

// Config helpers

func (rr *ReceivingRelay[T]) ConnectOutput(outputs ...types.Submitter[T]) {
	rr.requireNotFrozen("ConnectOutput")
	rr.Outputs = append(rr.Outputs, outputs...)
}

func (rr *ReceivingRelay[T]) ConnectLogger(loggers ...types.Logger) {
	rr.requireNotFrozen("ConnectLogger")
	rr.Loggers = append(rr.Loggers, loggers...)
}

func (rr *ReceivingRelay[T]) SetDecryptionKey(key string) {
	rr.requireNotFrozen("SetDecryptionKey")
	rr.DecryptionKey = key
}

func (rr *ReceivingRelay[T]) SetStaticHeaders(h map[string]string) {
	rr.requireNotFrozen("SetStaticHeaders")
	rr.staticHeaders = cloneStringMap(h)
}

func (rr *ReceivingRelay[T]) SetPassthrough(enabled bool) {
	rr.requireNotFrozen("SetPassthrough")
	rr.passthrough = enabled
}

func (rr *ReceivingRelay[T]) SetURL(url string) {
	rr.requireNotFrozen("SetURL")
	rr.URL = url
}

func (rr *ReceivingRelay[T]) SetSubject(subject string) {
	rr.requireNotFrozen("SetSubject")
	rr.Subject = subject
}

func (rr *ReceivingRelay[T]) SetQueue(queue string) {
	rr.requireNotFrozen("SetQueue")
	rr.Queue = queue
}

func (rr *ReceivingRelay[T]) SetBufferSize(n uint32) {
	rr.requireNotFrozen("SetBufferSize")
	if n == 0 {
		return
	}
	rr.DataCh = make(chan T, n)
}

//go:build nats

package natsrelay

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/relaycodec"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// ForwardRelay publishes WrappedPayloads to NATS subjects.
type ForwardRelay[T any] struct {
	URL     string
	Subject string
	Conn    *nats.Conn

	ctx    context.Context
	cancel context.CancelFunc

	componentMetadata types.ComponentMetadata

	Loggers     []types.Logger
	loggersLock sync.Mutex

	PerformanceOptions *relay.PerformanceOptions
	SecurityOptions    *relay.SecurityOptions
	EncryptionKey      string

	staticHeaders  map[string]string
	dynamicHeaders func(ctx context.Context) map[string]string

	passthrough   bool
	payloadFormat string
	payloadType   string

	seq uint64

	isRunning    int32
	configFrozen int32
}

func NewForwardRelay[T any](ctx context.Context, options ...types.Option[*ForwardRelay[T]]) *ForwardRelay[T] {
	ctx, cancel := context.WithCancel(ctx)
	fr := &ForwardRelay[T]{
		ctx:    ctx,
		cancel: cancel,
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "NATS_FORWARD_RELAY",
		},
		Loggers:            make([]types.Logger, 0),
		Subject:            "relay",
		payloadFormat:      "json",
		PerformanceOptions: &relay.PerformanceOptions{UseCompression: false, CompressionAlgorithm: relay.CompressionAlgorithm_COMPRESS_NONE},
		staticHeaders:      make(map[string]string),
	}
	for _, option := range options {
		option(fr)
	}
	return fr
}

func (fr *ForwardRelay[T]) IsRunning() bool {
	return atomic.LoadInt32(&fr.isRunning) == 1
}

func (fr *ForwardRelay[T]) Start(ctx context.Context) error {
	atomic.StoreInt32(&fr.configFrozen, 1)
	atomic.StoreInt32(&fr.isRunning, 1)
	if fr.Conn == nil {
		conn, err := nats.Connect(fr.URL)
		if err != nil {
			return err
		}
		fr.Conn = conn
	}
	return nil
}

func (fr *ForwardRelay[T]) Stop() {
	fr.cancel()
	if fr.Conn != nil {
		fr.Conn.Close()
	}
	atomic.StoreInt32(&fr.isRunning, 0)
}

func (fr *ForwardRelay[T]) Submit(ctx context.Context, item T) error {
	if fr.Conn == nil {
		if err := fr.Start(ctx); err != nil {
			return err
		}
	}
	wp, err := fr.wrapItem(ctx, item)
	if err != nil {
		return err
	}
	b, err := proto.Marshal(wp)
	if err != nil {
		return err
	}
	return fr.Conn.Publish(fr.Subject, b)
}

func (fr *ForwardRelay[T]) wrapItem(ctx context.Context, item T) (*relay.WrappedPayload, error) {
	if fr.passthrough {
		if wp, ok := any(item).(*relay.WrappedPayload); ok && wp != nil {
			return wp, nil
		}
		return nil, fmt.Errorf("passthrough enabled but item is not *relay.WrappedPayload")
	}
	headers := cloneStringMap(fr.staticHeaders)
	if fr.dynamicHeaders != nil {
		for k, v := range fr.dynamicHeaders(ctx) {
			headers[k] = v
		}
	}
	wp, err := relaycodec.WrapPayload(item, fr.payloadFormat, fr.payloadType, fr.PerformanceOptions, fr.SecurityOptions, fr.EncryptionKey, headers)
	if err != nil {
		return nil, err
	}
	wp.Seq = atomic.AddUint64(&fr.seq, 1)
	return wp, nil
}

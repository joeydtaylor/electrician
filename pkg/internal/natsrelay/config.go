//go:build nats

package natsrelay

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// Forward relay config helpers

func (fr *ForwardRelay[T]) ConnectLogger(loggers ...types.Logger) {
	fr.requireNotFrozen("ConnectLogger")
	fr.Loggers = append(fr.Loggers, loggers...)
}

func (fr *ForwardRelay[T]) SetPerformanceOptions(opts *relay.PerformanceOptions) {
	fr.requireNotFrozen("SetPerformanceOptions")
	fr.PerformanceOptions = opts
}

func (fr *ForwardRelay[T]) SetSecurityOptions(opts *relay.SecurityOptions, key string) {
	fr.requireNotFrozen("SetSecurityOptions")
	fr.SecurityOptions = opts
	fr.EncryptionKey = key
}

func (fr *ForwardRelay[T]) SetStaticHeaders(h map[string]string) {
	fr.requireNotFrozen("SetStaticHeaders")
	fr.staticHeaders = cloneStringMap(h)
}

func (fr *ForwardRelay[T]) SetDynamicHeaders(fn func(ctx context.Context) map[string]string) {
	fr.requireNotFrozen("SetDynamicHeaders")
	fr.dynamicHeaders = fn
}

func (fr *ForwardRelay[T]) SetPayloadFormat(format string) {
	fr.requireNotFrozen("SetPayloadFormat")
	if format != "" {
		fr.payloadFormat = format
	}
}

func (fr *ForwardRelay[T]) SetPayloadType(t string) {
	fr.requireNotFrozen("SetPayloadType")
	fr.payloadType = t
}

func (fr *ForwardRelay[T]) SetPassthrough(enabled bool) {
	fr.requireNotFrozen("SetPassthrough")
	fr.passthrough = enabled
}

// Receiving relay config helpers

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

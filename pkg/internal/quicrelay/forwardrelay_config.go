package quicrelay

import (
	"context"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// ConnectInput registers one or more receivers as inputs for the relay.
func (fr *ForwardRelay[T]) ConnectInput(inputs ...types.Receiver[T]) {
	fr.requireNotFrozen("ConnectInput")
	if len(inputs) == 0 {
		return
	}
	out := inputs[:0]
	for _, input := range inputs {
		if input != nil {
			out = append(out, input)
		}
	}
	if len(out) == 0 {
		return
	}
	fr.Input = append(fr.Input, out...)
	for _, input := range out {
		fr.logKV(types.DebugLevel, "Input connected",
			"event", "ConnectInput",
			"result", "SUCCESS",
			"input", input.GetComponentMetadata(),
		)
	}
}

// ConnectLogger registers loggers for the relay.
func (fr *ForwardRelay[T]) ConnectLogger(loggers ...types.Logger) {
	fr.requireNotFrozen("ConnectLogger")
	if len(loggers) == 0 {
		return
	}
	out := loggers[:0]
	for _, logger := range loggers {
		if logger != nil {
			out = append(out, logger)
		}
	}
	if len(out) == 0 {
		return
	}
	fr.loggersLock.Lock()
	fr.Loggers = append(fr.Loggers, out...)
	fr.loggersLock.Unlock()
}

// GetTargets returns current targets.
func (fr *ForwardRelay[T]) GetTargets() []string {
	return fr.Targets
}

// GetComponentMetadata returns relay metadata.
func (fr *ForwardRelay[T]) GetComponentMetadata() types.ComponentMetadata {
	return fr.componentMetadata
}

// GetInput returns configured input receivers.
func (fr *ForwardRelay[T]) GetInput() []types.Receiver[T] {
	return fr.Input
}

// SetTargets sets outbound targets.
func (fr *ForwardRelay[T]) SetTargets(targets ...string) {
	fr.requireNotFrozen("SetTargets")
	fr.Targets = normalizeTargets(targets)
}

// SetComponentMetadata sets name and id metadata.
func (fr *ForwardRelay[T]) SetComponentMetadata(name string, id string) {
	fr.requireNotFrozen("SetComponentMetadata")
	fr.componentMetadata.Name = name
	fr.componentMetadata.ID = id
}

// SetPerformanceOptions sets performance options.
func (fr *ForwardRelay[T]) SetPerformanceOptions(perf *relay.PerformanceOptions) {
	fr.requireNotFrozen("SetPerformanceOptions")
	fr.PerformanceOptions = perf
}

// SetSecurityOptions sets security options and encryption key.
func (fr *ForwardRelay[T]) SetSecurityOptions(sec *relay.SecurityOptions, key string) {
	fr.requireNotFrozen("SetSecurityOptions")
	fr.SecurityOptions = sec
	fr.EncryptionKey = key
}

// SetTLSConfig sets TLS configuration.
func (fr *ForwardRelay[T]) SetTLSConfig(cfg *types.TLSConfig) {
	fr.requireNotFrozen("SetTLSConfig")
	fr.TlsConfig = cfg
}

// SetPassthrough enables forwarding pre-wrapped payloads without modification.
func (fr *ForwardRelay[T]) SetPassthrough(enabled bool) {
	fr.requireNotFrozen("SetPassthrough")
	fr.passthrough = enabled
}

// SetAuthenticationOptions sets auth hints.
func (fr *ForwardRelay[T]) SetAuthenticationOptions(opts *relay.AuthenticationOptions) {
	fr.requireNotFrozen("SetAuthenticationOptions")
	fr.authOptions = opts
}

// SetOAuth2 sets the token source for bearer auth.
func (fr *ForwardRelay[T]) SetOAuth2(ts types.OAuth2TokenSource) {
	fr.requireNotFrozen("SetOAuth2")
	fr.tokenSource = ts
}

// SetStaticHeaders sets headers added to StreamOpen defaults.
func (fr *ForwardRelay[T]) SetStaticHeaders(headers map[string]string) {
	fr.requireNotFrozen("SetStaticHeaders")
	fr.staticHeaders = cloneStringMap(headers)
}

// SetDynamicHeaders sets a per-request header builder.
func (fr *ForwardRelay[T]) SetDynamicHeaders(fn func(ctx context.Context) map[string]string) {
	fr.requireNotFrozen("SetDynamicHeaders")
	fr.dynamicHeaders = fn
}

// SetAuthRequired toggles strict auth enforcement.
func (fr *ForwardRelay[T]) SetAuthRequired(required bool) {
	fr.requireNotFrozen("SetAuthRequired")
	fr.authRequired = required
}

// GetAuthRequired returns auth-required setting.
func (fr *ForwardRelay[T]) GetAuthRequired() bool {
	return fr.authRequired
}

// SetAckMode sets the stream ack mode.
func (fr *ForwardRelay[T]) SetAckMode(mode relay.AckMode) {
	fr.requireNotFrozen("SetAckMode")
	fr.ackMode = mode
}

// SetAckEveryN sets batch ack frequency.
func (fr *ForwardRelay[T]) SetAckEveryN(n uint32) {
	fr.requireNotFrozen("SetAckEveryN")
	fr.ackEveryN = n
}

// SetMaxInFlight sets the max in-flight hint for StreamOpen.
func (fr *ForwardRelay[T]) SetMaxInFlight(n uint32) {
	fr.requireNotFrozen("SetMaxInFlight")
	fr.maxInFlight = n
}

// SetOmitPayloadMetadata toggles payload metadata omission when defaults are set.
func (fr *ForwardRelay[T]) SetOmitPayloadMetadata(omit bool) {
	fr.requireNotFrozen("SetOmitPayloadMetadata")
	fr.omitPayloadMetadata = omit
}

// SetStreamSendBuffer sets stream send buffer size.
func (fr *ForwardRelay[T]) SetStreamSendBuffer(n int) {
	fr.requireNotFrozen("SetStreamSendBuffer")
	if n > 0 {
		fr.streamSendBuf = n
	}
}

// SetMaxFrameBytes sets the max frame size for reads.
func (fr *ForwardRelay[T]) SetMaxFrameBytes(n int) {
	fr.requireNotFrozen("SetMaxFrameBytes")
	if n > 0 {
		fr.maxFrameBytes = n
	}
}

// SetDropOnFull toggles dropping when send buffers are full.
func (fr *ForwardRelay[T]) SetDropOnFull(drop bool) {
	fr.requireNotFrozen("SetDropOnFull")
	fr.dropOnFull = drop
}

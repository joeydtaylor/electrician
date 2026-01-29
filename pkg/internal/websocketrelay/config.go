package websocketrelay

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// SetTargets sets target WebSocket URLs.
func (fr *ForwardRelay[T]) SetTargets(targets ...string) {
	fr.requireNotFrozen("SetTargets")
	fr.Targets = normalizeTargets(targets)
}

// ConnectInput wires receivers into the relay.
func (fr *ForwardRelay[T]) ConnectInput(input ...types.Receiver[T]) {
	fr.requireNotFrozen("ConnectInput")
	fr.Input = append(fr.Input, input...)
}

// ConnectLogger attaches loggers.
func (fr *ForwardRelay[T]) ConnectLogger(loggers ...types.Logger) {
	fr.requireNotFrozen("ConnectLogger")
	fr.Loggers = append(fr.Loggers, loggers...)
}

// SetComponentMetadata updates the relay's metadata.
func (fr *ForwardRelay[T]) SetComponentMetadata(name string, id string) {
	fr.requireNotFrozen("SetComponentMetadata")
	fr.componentMetadata.Name = name
	if id != "" {
		fr.componentMetadata.ID = id
	}
}

// GetComponentMetadata returns component metadata.
func (fr *ForwardRelay[T]) GetComponentMetadata() types.ComponentMetadata {
	return fr.componentMetadata
}

// SetTLSConfig configures TLS for WSS.
func (fr *ForwardRelay[T]) SetTLSConfig(cfg *types.TLSConfig) {
	fr.requireNotFrozen("SetTLSConfig")
	fr.TlsConfig = cfg
}

// SetPerformanceOptions sets compression config.
func (fr *ForwardRelay[T]) SetPerformanceOptions(opts *relay.PerformanceOptions) {
	fr.requireNotFrozen("SetPerformanceOptions")
	fr.PerformanceOptions = opts
}

// SetSecurityOptions sets encryption config.
func (fr *ForwardRelay[T]) SetSecurityOptions(opts *relay.SecurityOptions, key string) {
	fr.requireNotFrozen("SetSecurityOptions")
	fr.SecurityOptions = opts
	fr.EncryptionKey = key
}

// SetAuthenticationOptions sets OAuth2 config (optional).
func (fr *ForwardRelay[T]) SetAuthenticationOptions(opts *relay.AuthenticationOptions) {
	fr.requireNotFrozen("SetAuthenticationOptions")
	fr.authOptions = opts
}

// SetOAuth2 sets a token source.
func (fr *ForwardRelay[T]) SetOAuth2(ts types.OAuth2TokenSource) {
	fr.requireNotFrozen("SetOAuth2")
	fr.tokenSource = ts
}

// SetStaticHeaders defines fixed metadata headers.
func (fr *ForwardRelay[T]) SetStaticHeaders(h map[string]string) {
	fr.requireNotFrozen("SetStaticHeaders")
	fr.staticHeaders = cloneStringMap(h)
}

// SetDynamicHeaders defines a callback for per-send headers.
func (fr *ForwardRelay[T]) SetDynamicHeaders(fn func(ctx context.Context) map[string]string) {
	fr.requireNotFrozen("SetDynamicHeaders")
	fr.dynamicHeaders = fn
}

// SetAuthRequired toggles auth enforcement when OAuth2 configured.
func (fr *ForwardRelay[T]) SetAuthRequired(required bool) {
	fr.requireNotFrozen("SetAuthRequired")
	fr.authRequired = required
}

// GetAuthRequired returns auth requirement setting.
func (fr *ForwardRelay[T]) GetAuthRequired() bool {
	return fr.authRequired
}

// SetPassthrough enables forwarding pre-wrapped payloads without modification.
func (fr *ForwardRelay[T]) SetPassthrough(enabled bool) {
	fr.requireNotFrozen("SetPassthrough")
	fr.passthrough = enabled
}

// SetPayloadFormat sets encoding format: json | gob | proto.
func (fr *ForwardRelay[T]) SetPayloadFormat(format string) {
	fr.requireNotFrozen("SetPayloadFormat")
	if format != "" {
		fr.payloadFormat = format
	}
}

// SetPayloadType overrides payload type label.
func (fr *ForwardRelay[T]) SetPayloadType(t string) {
	fr.requireNotFrozen("SetPayloadType")
	fr.payloadType = t
}

// SetAckMode configures ack behavior.
func (fr *ForwardRelay[T]) SetAckMode(mode relay.AckMode) {
	fr.requireNotFrozen("SetAckMode")
	fr.ackMode = mode
}

// SetAckEveryN configures batch ack frequency.
func (fr *ForwardRelay[T]) SetAckEveryN(n uint32) {
	fr.requireNotFrozen("SetAckEveryN")
	fr.ackEveryN = n
}

// SetMaxInFlight sets max in-flight setting.
func (fr *ForwardRelay[T]) SetMaxInFlight(n uint32) {
	fr.requireNotFrozen("SetMaxInFlight")
	fr.maxInFlight = n
}

// SetOmitPayloadMetadata toggles metadata elision.
func (fr *ForwardRelay[T]) SetOmitPayloadMetadata(omit bool) {
	fr.requireNotFrozen("SetOmitPayloadMetadata")
	fr.omitPayloadMetadata = omit
}

// SetStreamSendBuffer sets internal send channel size.
func (fr *ForwardRelay[T]) SetStreamSendBuffer(n int) {
	fr.requireNotFrozen("SetStreamSendBuffer")
	fr.streamSendBuf = n
}

// SetMaxMessageBytes caps inbound message size.
func (fr *ForwardRelay[T]) SetMaxMessageBytes(n int) {
	fr.requireNotFrozen("SetMaxMessageBytes")
	fr.maxMessageBytes = n
}

// SetDropOnFull sets drop policy when send channel is full.
func (fr *ForwardRelay[T]) SetDropOnFull(drop bool) {
	fr.requireNotFrozen("SetDropOnFull")
	fr.dropOnFull = drop
}

// Receiving side

// ConnectOutput attaches outputs.
func (rr *ReceivingRelay[T]) ConnectOutput(outputs ...types.Submitter[T]) {
	rr.requireNotFrozen("ConnectOutput")
	rr.Outputs = append(rr.Outputs, outputs...)
}

// ConnectLogger attaches loggers.
func (rr *ReceivingRelay[T]) ConnectLogger(loggers ...types.Logger) {
	rr.requireNotFrozen("ConnectLogger")
	rr.Loggers = append(rr.Loggers, loggers...)
}

// SetComponentMetadata updates metadata.
func (rr *ReceivingRelay[T]) SetComponentMetadata(name string, id string) {
	rr.requireNotFrozen("SetComponentMetadata")
	rr.componentMetadata.Name = name
	if id != "" {
		rr.componentMetadata.ID = id
	}
}

// GetComponentMetadata returns metadata.
func (rr *ReceivingRelay[T]) GetComponentMetadata() types.ComponentMetadata {
	return rr.componentMetadata
}

// SetAddress sets HTTP address.
func (rr *ReceivingRelay[T]) SetAddress(addr string) {
	rr.requireNotFrozen("SetAddress")
	rr.Address = addr
}

// SetPath sets WebSocket path.
func (rr *ReceivingRelay[T]) SetPath(path string) {
	rr.requireNotFrozen("SetPath")
	if path != "" {
		rr.Path = path
	}
}

// SetTLSConfig configures WSS.
func (rr *ReceivingRelay[T]) SetTLSConfig(cfg *types.TLSConfig) {
	rr.requireNotFrozen("SetTLSConfig")
	rr.TlsConfig = cfg
}

// SetDecryptionKey sets AES-GCM key.
func (rr *ReceivingRelay[T]) SetDecryptionKey(key string) {
	rr.requireNotFrozen("SetDecryptionKey")
	rr.DecryptionKey = key
}

// SetAuthenticationOptions configures auth expectations.
func (rr *ReceivingRelay[T]) SetAuthenticationOptions(opts *relay.AuthenticationOptions) {
	rr.requireNotFrozen("SetAuthenticationOptions")
	rr.authOptions = opts
}

// SetStaticHeaders defines required headers.
func (rr *ReceivingRelay[T]) SetStaticHeaders(h map[string]string) {
	rr.requireNotFrozen("SetStaticHeaders")
	rr.staticHeaders = cloneStringMap(h)
}

// SetDynamicAuthValidator sets custom auth validator.
func (rr *ReceivingRelay[T]) SetDynamicAuthValidator(fn func(ctx context.Context, md map[string]string) error) {
	rr.requireNotFrozen("SetDynamicAuthValidator")
	rr.dynamicAuthValidator = fn
}

// SetAuthRequired toggles auth enforcement.
func (rr *ReceivingRelay[T]) SetAuthRequired(required bool) {
	rr.requireNotFrozen("SetAuthRequired")
	rr.authRequired = required
}

// SetPassthrough toggles raw payload passthrough.
func (rr *ReceivingRelay[T]) SetPassthrough(enabled bool) {
	rr.requireNotFrozen("SetPassthrough")
	rr.passthrough = enabled
}

// SetBufferSize sets channel buffer size.
func (rr *ReceivingRelay[T]) SetBufferSize(n uint32) {
	rr.requireNotFrozen("SetBufferSize")
	if n == 0 {
		return
	}
	rr.DataCh = make(chan T, n)
}

// SetMaxMessageBytes caps inbound size.
func (rr *ReceivingRelay[T]) SetMaxMessageBytes(n int) {
	rr.requireNotFrozen("SetMaxMessageBytes")
	rr.maxMessageBytes = n
}

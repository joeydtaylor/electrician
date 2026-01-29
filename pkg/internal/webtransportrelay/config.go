//go:build webtransport

package webtransportrelay

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// Forward relay config

func (fr *ForwardRelay[T]) SetTargets(targets ...string) {
	fr.requireNotFrozen("SetTargets")
	fr.Targets = normalizeTargets(targets)
}

func (fr *ForwardRelay[T]) ConnectInput(input ...types.Receiver[T]) {
	fr.requireNotFrozen("ConnectInput")
	fr.Input = append(fr.Input, input...)
}

func (fr *ForwardRelay[T]) ConnectLogger(loggers ...types.Logger) {
	fr.requireNotFrozen("ConnectLogger")
	fr.Loggers = append(fr.Loggers, loggers...)
}

func (fr *ForwardRelay[T]) SetComponentMetadata(name string, id string) {
	fr.requireNotFrozen("SetComponentMetadata")
	fr.componentMetadata.Name = name
	if id != "" {
		fr.componentMetadata.ID = id
	}
}

func (fr *ForwardRelay[T]) GetComponentMetadata() types.ComponentMetadata {
	return fr.componentMetadata
}

func (fr *ForwardRelay[T]) SetTLSConfig(cfg *types.TLSConfig) {
	fr.requireNotFrozen("SetTLSConfig")
	fr.TlsConfig = cfg
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

func (fr *ForwardRelay[T]) SetAuthenticationOptions(opts *relay.AuthenticationOptions) {
	fr.requireNotFrozen("SetAuthenticationOptions")
	fr.authOptions = opts
}

func (fr *ForwardRelay[T]) SetOAuth2(ts types.OAuth2TokenSource) {
	fr.requireNotFrozen("SetOAuth2")
	fr.tokenSource = ts
}

func (fr *ForwardRelay[T]) SetStaticHeaders(h map[string]string) {
	fr.requireNotFrozen("SetStaticHeaders")
	fr.staticHeaders = cloneStringMap(h)
}

func (fr *ForwardRelay[T]) SetDynamicHeaders(fn func(ctx context.Context) map[string]string) {
	fr.requireNotFrozen("SetDynamicHeaders")
	fr.dynamicHeaders = fn
}

func (fr *ForwardRelay[T]) SetAuthRequired(required bool) {
	fr.requireNotFrozen("SetAuthRequired")
	fr.authRequired = required
}

func (fr *ForwardRelay[T]) GetAuthRequired() bool {
	return fr.authRequired
}

func (fr *ForwardRelay[T]) SetPassthrough(enabled bool) {
	fr.requireNotFrozen("SetPassthrough")
	fr.passthrough = enabled
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

func (fr *ForwardRelay[T]) SetAckMode(mode relay.AckMode) {
	fr.requireNotFrozen("SetAckMode")
	fr.ackMode = mode
}

func (fr *ForwardRelay[T]) SetAckEveryN(n uint32) {
	fr.requireNotFrozen("SetAckEveryN")
	fr.ackEveryN = n
}

func (fr *ForwardRelay[T]) SetMaxInFlight(n uint32) {
	fr.requireNotFrozen("SetMaxInFlight")
	fr.maxInFlight = n
}

func (fr *ForwardRelay[T]) SetOmitPayloadMetadata(omit bool) {
	fr.requireNotFrozen("SetOmitPayloadMetadata")
	fr.omitPayloadMetadata = omit
}

func (fr *ForwardRelay[T]) SetStreamSendBuffer(n int) {
	fr.requireNotFrozen("SetStreamSendBuffer")
	fr.streamSendBuf = n
}

func (fr *ForwardRelay[T]) SetMaxFrameBytes(n int) {
	fr.requireNotFrozen("SetMaxFrameBytes")
	fr.maxFrameBytes = n
}

func (fr *ForwardRelay[T]) SetDropOnFull(drop bool) {
	fr.requireNotFrozen("SetDropOnFull")
	fr.dropOnFull = drop
}

func (fr *ForwardRelay[T]) SetUseDatagrams(use bool) {
	fr.requireNotFrozen("SetUseDatagrams")
	fr.useDatagrams = use
}

// Receiving relay config

func (rr *ReceivingRelay[T]) ConnectOutput(outputs ...types.Submitter[T]) {
	rr.requireNotFrozen("ConnectOutput")
	rr.Outputs = append(rr.Outputs, outputs...)
}

func (rr *ReceivingRelay[T]) ConnectLogger(loggers ...types.Logger) {
	rr.requireNotFrozen("ConnectLogger")
	rr.Loggers = append(rr.Loggers, loggers...)
}

func (rr *ReceivingRelay[T]) SetComponentMetadata(name string, id string) {
	rr.requireNotFrozen("SetComponentMetadata")
	rr.componentMetadata.Name = name
	if id != "" {
		rr.componentMetadata.ID = id
	}
}

func (rr *ReceivingRelay[T]) GetComponentMetadata() types.ComponentMetadata {
	return rr.componentMetadata
}

func (rr *ReceivingRelay[T]) SetAddress(addr string) {
	rr.requireNotFrozen("SetAddress")
	rr.Address = addr
}

func (rr *ReceivingRelay[T]) SetPath(path string) {
	rr.requireNotFrozen("SetPath")
	if path != "" {
		rr.Path = path
	}
}

func (rr *ReceivingRelay[T]) SetTLSConfig(cfg *types.TLSConfig) {
	rr.requireNotFrozen("SetTLSConfig")
	rr.TlsConfig = cfg
}

func (rr *ReceivingRelay[T]) SetDecryptionKey(key string) {
	rr.requireNotFrozen("SetDecryptionKey")
	rr.DecryptionKey = key
}

func (rr *ReceivingRelay[T]) SetAuthenticationOptions(opts *relay.AuthenticationOptions) {
	rr.requireNotFrozen("SetAuthenticationOptions")
	rr.authOptions = opts
}

func (rr *ReceivingRelay[T]) SetStaticHeaders(h map[string]string) {
	rr.requireNotFrozen("SetStaticHeaders")
	rr.staticHeaders = cloneStringMap(h)
}

func (rr *ReceivingRelay[T]) SetDynamicAuthValidator(fn func(ctx context.Context, md map[string]string) error) {
	rr.requireNotFrozen("SetDynamicAuthValidator")
	rr.dynamicAuthValidator = fn
}

func (rr *ReceivingRelay[T]) SetAuthRequired(required bool) {
	rr.requireNotFrozen("SetAuthRequired")
	rr.authRequired = required
}

func (rr *ReceivingRelay[T]) SetPassthrough(enabled bool) {
	rr.requireNotFrozen("SetPassthrough")
	rr.passthrough = enabled
}

func (rr *ReceivingRelay[T]) SetBufferSize(n uint32) {
	rr.requireNotFrozen("SetBufferSize")
	if n == 0 {
		return
	}
	rr.DataCh = make(chan T, n)
}

func (rr *ReceivingRelay[T]) SetMaxFrameBytes(n int) {
	rr.requireNotFrozen("SetMaxFrameBytes")
	rr.maxFrameBytes = n
}

func (rr *ReceivingRelay[T]) SetEnableDatagrams(enabled bool) {
	rr.requireNotFrozen("SetEnableDatagrams")
	rr.enableDatagrams = enabled
}

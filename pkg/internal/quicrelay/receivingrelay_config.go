package quicrelay

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// ConnectLogger registers loggers for the relay.
func (rr *ReceivingRelay[T]) ConnectLogger(loggers ...types.Logger) {
	rr.requireNotFrozen("ConnectLogger")
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
	rr.loggersLock.Lock()
	rr.Loggers = append(rr.Loggers, out...)
	rr.loggersLock.Unlock()
}

// ConnectOutput registers downstream submitters.
func (rr *ReceivingRelay[T]) ConnectOutput(outputs ...types.Submitter[T]) {
	rr.requireNotFrozen("ConnectOutput")
	if len(outputs) == 0 {
		return
	}
	out := outputs[:0]
	for _, output := range outputs {
		if output != nil {
			out = append(out, output)
		}
	}
	if len(out) == 0 {
		return
	}
	rr.Outputs = append(rr.Outputs, out...)
	for _, output := range out {
		rr.NotifyLoggers(types.DebugLevel, "ConnectOutput: added %s", output.GetComponentMetadata())
	}
}

// GetAddress returns listening address.
func (rr *ReceivingRelay[T]) GetAddress() string {
	return rr.Address
}

// GetComponentMetadata returns relay metadata.
func (rr *ReceivingRelay[T]) GetComponentMetadata() types.ComponentMetadata {
	return rr.componentMetadata
}

// SetComponentMetadata sets name and ID metadata.
func (rr *ReceivingRelay[T]) SetComponentMetadata(name string, id string) {
	rr.requireNotFrozen("SetComponentMetadata")
	rr.componentMetadata.Name = name
	rr.componentMetadata.ID = id
}

// SetAddress sets the listening address.
func (rr *ReceivingRelay[T]) SetAddress(addr string) {
	rr.requireNotFrozen("SetAddress")
	rr.Address = addr
}

// SetDataChannel sets the internal data channel buffer size.
func (rr *ReceivingRelay[T]) SetDataChannel(bufferSize uint32) {
	rr.requireNotFrozen("SetDataChannel")
	rr.DataCh = make(chan T, bufferSize)
}

// SetTLSConfig configures TLS for incoming connections.
func (rr *ReceivingRelay[T]) SetTLSConfig(cfg *types.TLSConfig) {
	rr.requireNotFrozen("SetTLSConfig")
	rr.TlsConfig = cfg
}

// SetPassthrough enables forwarding raw WrappedPayload values without unwrap/decrypt.
func (rr *ReceivingRelay[T]) SetPassthrough(enabled bool) {
	rr.requireNotFrozen("SetPassthrough")
	rr.passthrough = enabled
}

// SetDecryptionKey sets the AES-GCM decryption key.
func (rr *ReceivingRelay[T]) SetDecryptionKey(key string) {
	rr.requireNotFrozen("SetDecryptionKey")
	rr.DecryptionKey = key
}

// SetAuthenticationOptions configures expected auth mode/parameters.
func (rr *ReceivingRelay[T]) SetAuthenticationOptions(opts *relay.AuthenticationOptions) {
	rr.requireNotFrozen("SetAuthenticationOptions")
	rr.authOptions = opts
}

// SetStaticHeaders enforces constant metadata on incoming requests.
func (rr *ReceivingRelay[T]) SetStaticHeaders(headers map[string]string) {
	rr.requireNotFrozen("SetStaticHeaders")
	rr.staticHeaders = cloneStringMap(headers)
}

// SetDynamicAuthValidator registers a per-request validation callback.
func (rr *ReceivingRelay[T]) SetDynamicAuthValidator(fn func(ctx context.Context, md map[string]string) error) {
	rr.requireNotFrozen("SetDynamicAuthValidator")
	rr.dynamicAuthValidator = fn
}

// SetAuthRequired toggles strict auth enforcement.
func (rr *ReceivingRelay[T]) SetAuthRequired(required bool) {
	rr.requireNotFrozen("SetAuthRequired")
	rr.authRequired = required
}

// SetMaxFrameBytes sets the max frame size for reads.
func (rr *ReceivingRelay[T]) SetMaxFrameBytes(n int) {
	rr.requireNotFrozen("SetMaxFrameBytes")
	if n > 0 {
		rr.maxFrameBytes = n
	}
}

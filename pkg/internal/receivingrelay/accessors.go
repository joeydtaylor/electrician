package receivingrelay

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// GetAddress returns the configured address.
func (rr *ReceivingRelay[T]) GetAddress() string {
	rr.NotifyLoggers(types.DebugLevel, "GetAddress: %s", rr.Address)
	return rr.Address
}

// GetComponentMetadata returns relay metadata.
func (rr *ReceivingRelay[T]) GetComponentMetadata() types.ComponentMetadata {
	rr.NotifyLoggers(types.DebugLevel, "GetComponentMetadata: %v", rr.componentMetadata)
	return rr.componentMetadata
}

// GetDataChannel returns the internal data channel.
func (rr *ReceivingRelay[T]) GetDataChannel() chan T {
	rr.NotifyLoggers(types.DebugLevel, "GetDataChannel: %v", rr.DataCh)
	return rr.DataCh
}

// GetOutputs returns configured output submitters.
func (rr *ReceivingRelay[T]) GetOutputs() []types.Submitter[T] {
	rr.NotifyLoggers(types.DebugLevel, "GetOutputs: %v", rr.Outputs)
	return rr.Outputs
}

// GetOutputChannel returns the output channel for consumers.
func (rr *ReceivingRelay[T]) GetOutputChannel() chan T {
	rr.NotifyLoggers(types.DebugLevel, "GetOutputChannel: %v", rr.DataCh)
	return rr.DataCh
}

// GetTLSConfig returns the TLS configuration.
func (rr *ReceivingRelay[T]) GetTLSConfig() *types.TLSConfig {
	rr.NotifyLoggers(types.DebugLevel, "GetTLSConfig: %v", rr.TlsConfig)
	return rr.TlsConfig
}

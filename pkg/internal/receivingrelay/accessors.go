package receivingrelay

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// GetAddress returns the configured address.
func (rr *ReceivingRelay[T]) GetAddress() string {
	rr.logKV(
		types.DebugLevel,
		"GetAddress",
		"event", "GetAddress",
		"address", rr.Address,
	)
	return rr.Address
}

// GetComponentMetadata returns relay metadata.
func (rr *ReceivingRelay[T]) GetComponentMetadata() types.ComponentMetadata {
	rr.logKV(
		types.DebugLevel,
		"GetComponentMetadata",
		"event", "GetComponentMetadata",
	)
	return rr.componentMetadata
}

// GetDataChannel returns the internal data channel.
func (rr *ReceivingRelay[T]) GetDataChannel() chan T {
	rr.logKV(
		types.DebugLevel,
		"GetDataChannel",
		"event", "GetDataChannel",
		"data_channel", rr.DataCh,
	)
	return rr.DataCh
}

// GetOutputs returns configured output submitters.
func (rr *ReceivingRelay[T]) GetOutputs() []types.Submitter[T] {
	rr.logKV(
		types.DebugLevel,
		"GetOutputs",
		"event", "GetOutputs",
		"outputs", rr.Outputs,
	)
	return rr.Outputs
}

// GetOutputChannel returns the output channel for consumers.
func (rr *ReceivingRelay[T]) GetOutputChannel() chan T {
	rr.logKV(
		types.DebugLevel,
		"GetOutputChannel",
		"event", "GetOutputChannel",
		"output_channel", rr.DataCh,
	)
	return rr.DataCh
}

// GetTLSConfig returns the TLS configuration.
func (rr *ReceivingRelay[T]) GetTLSConfig() *types.TLSConfig {
	rr.logKV(
		types.DebugLevel,
		"GetTLSConfig",
		"event", "GetTLSConfig",
		"tls", rr.TlsConfig,
	)
	return rr.TlsConfig
}

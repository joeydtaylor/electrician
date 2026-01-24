package forwardrelay

import (
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// GetTargets returns the configured relay targets.
func (fr *ForwardRelay[T]) GetTargets() []string {
	fr.NotifyLoggers(types.DebugLevel, "GetTargets: %v", fr.Targets)
	return fr.Targets
}

// GetComponentMetadata returns the relay metadata.
func (fr *ForwardRelay[T]) GetComponentMetadata() types.ComponentMetadata {
	fr.NotifyLoggers(types.DebugLevel, "GetComponentMetadata: %v", fr.componentMetadata)
	return fr.componentMetadata
}

// GetInput returns configured input receivers.
func (fr *ForwardRelay[T]) GetInput() []types.Receiver[T] {
	fr.NotifyLoggers(types.DebugLevel, "GetInput: %v", fr.Input)
	return fr.Input
}

// GetPerformanceOptions returns the performance options.
func (fr *ForwardRelay[T]) GetPerformanceOptions() *relay.PerformanceOptions {
	fr.NotifyLoggers(types.DebugLevel, "GetPerformanceOptions: %v", fr.PerformanceOptions)
	return fr.PerformanceOptions
}

// GetTLSConfig returns the TLS configuration.
func (fr *ForwardRelay[T]) GetTLSConfig() *types.TLSConfig {
	fr.NotifyLoggers(types.DebugLevel, "GetTLSConfig: %v", fr.TlsConfig)
	return fr.TlsConfig
}

// GetAuthRequired reports whether OAuth2 tokens are required for RPCs.
func (fr *ForwardRelay[T]) GetAuthRequired() bool {
	return fr.authRequired
}

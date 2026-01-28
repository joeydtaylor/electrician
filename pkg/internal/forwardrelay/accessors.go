package forwardrelay

import (
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// GetTargets returns the configured relay targets.
func (fr *ForwardRelay[T]) GetTargets() []string {
	fr.logKV(types.DebugLevel, "GetTargets",
		"event", "GetTargets",
		"targets", fr.Targets,
	)
	return fr.Targets
}

// GetComponentMetadata returns the relay metadata.
func (fr *ForwardRelay[T]) GetComponentMetadata() types.ComponentMetadata {
	fr.logKV(types.DebugLevel, "GetComponentMetadata",
		"event", "GetComponentMetadata",
	)
	return fr.componentMetadata
}

// GetInput returns configured input receivers.
func (fr *ForwardRelay[T]) GetInput() []types.Receiver[T] {
	fr.logKV(types.DebugLevel, "GetInput",
		"event", "GetInput",
		"inputs", fr.Input,
	)
	return fr.Input
}

// GetPerformanceOptions returns the performance options.
func (fr *ForwardRelay[T]) GetPerformanceOptions() *relay.PerformanceOptions {
	fr.logKV(types.DebugLevel, "GetPerformanceOptions",
		"event", "GetPerformanceOptions",
		"performance", fr.PerformanceOptions,
	)
	return fr.PerformanceOptions
}

// GetTLSConfig returns the TLS configuration.
func (fr *ForwardRelay[T]) GetTLSConfig() *types.TLSConfig {
	fr.logKV(types.DebugLevel, "GetTLSConfig",
		"event", "GetTLSConfig",
		"tls", fr.TlsConfig,
	)
	return fr.TlsConfig
}

// GetAuthRequired reports whether OAuth2 tokens are required for RPCs.
func (fr *ForwardRelay[T]) GetAuthRequired() bool {
	return fr.authRequired
}

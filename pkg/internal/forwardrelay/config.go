package forwardrelay

import (
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// SetTargets sets one or more relay targets.
func (fr *ForwardRelay[T]) SetTargets(targets ...string) {
	fr.requireNotFrozen("SetTargets")
	if len(targets) == 0 {
		return
	}
	fr.Targets = append(fr.Targets, targets...)
	fr.logKV(types.DebugLevel, "Targets updated",
		"event", "SetTargets",
		"result", "SUCCESS",
		"targets", fr.Targets,
	)
}

// SetComponentMetadata updates the relay metadata.
func (fr *ForwardRelay[T]) SetComponentMetadata(name string, id string) {
	fr.requireNotFrozen("SetComponentMetadata")
	old := fr.componentMetadata
	fr.componentMetadata.Name = name
	fr.componentMetadata.ID = id
	fr.logKV(types.DebugLevel, "Component metadata updated",
		"event", "SetComponentMetadata",
		"result", "SUCCESS",
		"before", old,
		"after", fr.componentMetadata,
	)
}

// SetPerformanceOptions sets compression-related options.
func (fr *ForwardRelay[T]) SetPerformanceOptions(perfOptions *relay.PerformanceOptions) {
	fr.requireNotFrozen("SetPerformanceOptions")
	fr.PerformanceOptions = perfOptions
	fr.logKV(types.DebugLevel, "Performance options updated",
		"event", "SetPerformanceOptions",
		"result", "SUCCESS",
		"performance", fr.PerformanceOptions,
	)
}

// SetSecurityOptions configures payload encryption settings and key.
func (fr *ForwardRelay[T]) SetSecurityOptions(secOptions *relay.SecurityOptions, encryptionKey string) {
	fr.requireNotFrozen("SetSecurityOptions")
	fr.SecurityOptions = secOptions
	fr.EncryptionKey = encryptionKey
	fr.logKV(types.DebugLevel, "Security options updated",
		"event", "SetSecurityOptions",
		"result", "SUCCESS",
		"security", fr.SecurityOptions,
	)
}

// SetTLSConfig sets the TLS configuration for outbound connections.
func (fr *ForwardRelay[T]) SetTLSConfig(config *types.TLSConfig) {
	fr.requireNotFrozen("SetTLSConfig")
	fr.TlsConfig = config
	fr.logKV(types.DebugLevel, "TLS config updated",
		"event", "SetTLSConfig",
		"result", "SUCCESS",
		"tls", fr.TlsConfig,
	)
}

// SetPassthrough enables forwarding pre-wrapped payloads without modification.
func (fr *ForwardRelay[T]) SetPassthrough(enabled bool) {
	fr.requireNotFrozen("SetPassthrough")
	fr.passthrough = enabled
	fr.logKV(types.InfoLevel, "Passthrough updated",
		"event", "SetPassthrough",
		"result", "SUCCESS",
		"enabled", enabled,
	)
}

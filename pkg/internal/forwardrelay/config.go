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
	fr.NotifyLoggers(types.DebugLevel, "SetTargets: %v", fr.Targets)
}

// SetComponentMetadata updates the relay metadata.
func (fr *ForwardRelay[T]) SetComponentMetadata(name string, id string) {
	fr.requireNotFrozen("SetComponentMetadata")
	old := fr.componentMetadata
	fr.componentMetadata.Name = name
	fr.componentMetadata.ID = id
	fr.NotifyLoggers(types.DebugLevel, "SetComponentMetadata: %v -> %v", old, fr.componentMetadata)
}

// SetPerformanceOptions sets compression-related options.
func (fr *ForwardRelay[T]) SetPerformanceOptions(perfOptions *relay.PerformanceOptions) {
	fr.requireNotFrozen("SetPerformanceOptions")
	fr.PerformanceOptions = perfOptions
	fr.NotifyLoggers(types.DebugLevel, "SetPerformanceOptions: %v", fr.PerformanceOptions)
}

// SetSecurityOptions configures payload encryption settings and key.
func (fr *ForwardRelay[T]) SetSecurityOptions(secOptions *relay.SecurityOptions, encryptionKey string) {
	fr.requireNotFrozen("SetSecurityOptions")
	fr.SecurityOptions = secOptions
	fr.EncryptionKey = encryptionKey
	fr.NotifyLoggers(types.DebugLevel, "SetSecurityOptions: %v", fr.SecurityOptions)
}

// SetTLSConfig sets the TLS configuration for outbound connections.
func (fr *ForwardRelay[T]) SetTLSConfig(config *types.TLSConfig) {
	fr.requireNotFrozen("SetTLSConfig")
	fr.TlsConfig = config
	fr.NotifyLoggers(types.DebugLevel, "SetTLSConfig: %v", fr.TlsConfig)
}

// SetPassthrough enables forwarding pre-wrapped payloads without modification.
func (fr *ForwardRelay[T]) SetPassthrough(enabled bool) {
	fr.requireNotFrozen("SetPassthrough")
	fr.passthrough = enabled
	fr.NotifyLoggers(types.InfoLevel, "SetPassthrough: %t", enabled)
}

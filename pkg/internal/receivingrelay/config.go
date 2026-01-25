package receivingrelay

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// SetAddress configures the listen address.
func (rr *ReceivingRelay[T]) SetAddress(address string) {
	rr.requireNotFrozen("SetAddress")
	rr.Address = address
	rr.NotifyLoggers(types.DebugLevel, "SetAddress: %s", address)
}

// SetDataChannel configures the data channel buffer size.
func (rr *ReceivingRelay[T]) SetDataChannel(bufferSize uint32) {
	rr.requireNotFrozen("SetDataChannel")
	rr.DataCh = make(chan T, bufferSize)
	rr.NotifyLoggers(types.DebugLevel, "SetDataChannel: size=%d", bufferSize)
}

// SetDecryptionKey sets the AES-GCM decryption key.
func (rr *ReceivingRelay[T]) SetDecryptionKey(decryptionKey string) {
	rr.requireNotFrozen("SetDecryptionKey")
	rr.DecryptionKey = decryptionKey
	rr.NotifyLoggers(types.InfoLevel, "SetDecryptionKey: updated")
	rr.NotifyLoggers(types.DebugLevel, "SetDecryptionKey: key updated")
}

// SetComponentMetadata updates the relay metadata.
func (rr *ReceivingRelay[T]) SetComponentMetadata(name string, id string) {
	rr.requireNotFrozen("SetComponentMetadata")
	old := rr.componentMetadata
	rr.componentMetadata.Name = name
	rr.componentMetadata.ID = id
	rr.NotifyLoggers(types.DebugLevel, "SetComponentMetadata: %v -> %v", old, rr.componentMetadata)
}

// SetTLSConfig configures TLS for incoming connections.
func (rr *ReceivingRelay[T]) SetTLSConfig(config *types.TLSConfig) {
	rr.requireNotFrozen("SetTLSConfig")
	rr.TlsConfig = config
	rr.NotifyLoggers(types.DebugLevel, "SetTLSConfig: %v", rr.TlsConfig)
}

// SetPassthrough enables forwarding raw WrappedPayload values without unwrap/decrypt.
func (rr *ReceivingRelay[T]) SetPassthrough(enabled bool) {
	rr.requireNotFrozen("SetPassthrough")
	rr.passthrough = enabled
	rr.NotifyLoggers(types.InfoLevel, "SetPassthrough: %t", enabled)
}

// SetGRPCWebConfig configures gRPC-Web CORS and transport behavior.
func (rr *ReceivingRelay[T]) SetGRPCWebConfig(config *types.GRPCWebConfig) {
	rr.requireNotFrozen("SetGRPCWebConfig")
	rr.grpcWebConfig = cloneGRPCWebConfig(config)
	rr.NotifyLoggers(types.DebugLevel, "SetGRPCWebConfig: %v", rr.grpcWebConfig)
}

package receivingrelay

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// SetAddress configures the listen address.
func (rr *ReceivingRelay[T]) SetAddress(address string) {
	rr.requireNotFrozen("SetAddress")
	rr.Address = address
	rr.logKV(
		types.DebugLevel,
		"Address updated",
		"event", "SetAddress",
		"result", "SUCCESS",
		"address", address,
	)
}

// SetDataChannel configures the data channel buffer size.
func (rr *ReceivingRelay[T]) SetDataChannel(bufferSize uint32) {
	rr.requireNotFrozen("SetDataChannel")
	rr.DataCh = make(chan T, bufferSize)
	rr.logKV(
		types.DebugLevel,
		"Data channel updated",
		"event", "SetDataChannel",
		"result", "SUCCESS",
		"size", bufferSize,
	)
}

// SetDecryptionKey sets the AES-GCM decryption key.
func (rr *ReceivingRelay[T]) SetDecryptionKey(decryptionKey string) {
	rr.requireNotFrozen("SetDecryptionKey")
	rr.DecryptionKey = decryptionKey
	rr.logKV(
		types.InfoLevel,
		"Decryption key updated",
		"event", "SetDecryptionKey",
		"result", "SUCCESS",
	)
	rr.logKV(
		types.DebugLevel,
		"Decryption key details",
		"event", "SetDecryptionKey",
		"result", "DETAIL",
		"key_len", len(decryptionKey),
	)
}

// SetComponentMetadata updates the relay metadata.
func (rr *ReceivingRelay[T]) SetComponentMetadata(name string, id string) {
	rr.requireNotFrozen("SetComponentMetadata")
	old := rr.componentMetadata
	rr.componentMetadata.Name = name
	rr.componentMetadata.ID = id
	rr.logKV(
		types.DebugLevel,
		"Component metadata updated",
		"event", "SetComponentMetadata",
		"result", "SUCCESS",
		"previous", old,
		"current", rr.componentMetadata,
	)
}

// SetTLSConfig configures TLS for incoming connections.
func (rr *ReceivingRelay[T]) SetTLSConfig(config *types.TLSConfig) {
	rr.requireNotFrozen("SetTLSConfig")
	rr.TlsConfig = config
	rr.logKV(
		types.DebugLevel,
		"TLS config updated",
		"event", "SetTLSConfig",
		"result", "SUCCESS",
		"tls", rr.TlsConfig,
	)
}

// SetPassthrough enables forwarding raw WrappedPayload values without unwrap/decrypt.
func (rr *ReceivingRelay[T]) SetPassthrough(enabled bool) {
	rr.requireNotFrozen("SetPassthrough")
	rr.passthrough = enabled
	rr.logKV(
		types.InfoLevel,
		"Passthrough updated",
		"event", "SetPassthrough",
		"result", "SUCCESS",
		"enabled", enabled,
	)
}

// SetGRPCWebConfig configures gRPC-Web CORS and transport behavior.
func (rr *ReceivingRelay[T]) SetGRPCWebConfig(config *types.GRPCWebConfig) {
	rr.requireNotFrozen("SetGRPCWebConfig")
	rr.grpcWebConfig = cloneGRPCWebConfig(config)
	rr.logKV(
		types.DebugLevel,
		"gRPC-Web config updated",
		"event", "SetGRPCWebConfig",
		"result", "SUCCESS",
		"config", rr.grpcWebConfig,
	)
}

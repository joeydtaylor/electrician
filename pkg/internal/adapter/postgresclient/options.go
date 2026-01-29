package postgresclient

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// WithPostgresClientDeps injects Postgres client dependencies.
func WithPostgresClientDeps[T any](deps types.PostgresClientDeps) types.PostgresClientOption[T] {
	return func(adp types.PostgresClientAdapter[T]) {
		adp.SetPostgresClientDeps(deps)
	}
}

// WithWriterConfig sets Postgres writer configuration.
func WithWriterConfig[T any](cfg types.PostgresWriterConfig) types.PostgresClientOption[T] {
	return func(adp types.PostgresClientAdapter[T]) {
		adp.SetWriterConfig(cfg)
	}
}

// WithReaderConfig sets Postgres reader configuration.
func WithReaderConfig[T any](cfg types.PostgresReaderConfig) types.PostgresClientOption[T] {
	return func(adp types.PostgresClientAdapter[T]) {
		adp.SetReaderConfig(cfg)
	}
}

// WithLogger attaches loggers to the adapter.
func WithLogger[T any](l ...types.Logger) types.PostgresClientOption[T] {
	return func(adp types.PostgresClientAdapter[T]) {
		adp.ConnectLogger(l...)
	}
}

// WithSensor attaches sensors to the adapter.
func WithSensor[T any](s ...types.Sensor[T]) types.PostgresClientOption[T] {
	return func(adp types.PostgresClientAdapter[T]) {
		adp.ConnectSensor(s...)
	}
}

// WithWire connects one or more wires as inputs to the adapter.
func WithWire[T any](wires ...types.Wire[T]) types.PostgresClientOption[T] {
	return func(adp types.PostgresClientAdapter[T]) {
		adp.ConnectInput(wires...)
	}
}

// WithSecureDefaults enforces TLS and client-side encryption defaults.
func WithSecureDefaults[T any](keyHex string) types.PostgresClientOption[T] {
	return func(adp types.PostgresClientAdapter[T]) {
		adp.SetPostgresClientDeps(types.PostgresClientDeps{
			Security: &types.PostgresSecurity{
				RequireTLS: true,
			},
		})
		adp.SetWriterConfig(types.PostgresWriterConfig{
			ClientSideEncryption:        "AES-GCM",
			ClientSideKey:               keyHex,
			RequireClientSideEncryption: true,
		})
		adp.SetReaderConfig(types.PostgresReaderConfig{
			ClientSideEncryption:        "AES-GCM",
			ClientSideKey:               keyHex,
			RequireClientSideEncryption: true,
		})
	}
}

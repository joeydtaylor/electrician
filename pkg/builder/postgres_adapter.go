package builder

import (
	"context"
	"crypto/tls"
	"database/sql"
	"time"

	pgClientAdapter "github.com/joeydtaylor/electrician/pkg/internal/adapter/postgresclient"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// NewPostgresClientAdapter creates a new Postgres adapter (read + write capable).
func NewPostgresClientAdapter[T any](ctx context.Context, options ...types.PostgresClientOption[T]) types.PostgresClientAdapter[T] {
	return pgClientAdapter.NewPostgresClientAdapter[T](ctx, options...)
}

// PostgresAdapterWithConnString configures the connection string and driver name.
func PostgresAdapterWithConnString[T any](connString string, driver string) types.PostgresClientOption[T] {
	return pgClientAdapter.WithPostgresClientDeps[T](types.PostgresClientDeps{
		ConnString: connString,
		DriverName: driver,
	})
}

// PostgresAdapterWithDB injects a pre-built sql.DB.
func PostgresAdapterWithDB[T any](db *sql.DB) types.PostgresClientOption[T] {
	return pgClientAdapter.WithPostgresClientDeps[T](types.PostgresClientDeps{
		DB: db,
	})
}

// PostgresAdapterWithPoolSettings configures pool behavior.
func PostgresAdapterWithPoolSettings[T any](cfg types.PostgresPoolSettings) types.PostgresClientOption[T] {
	return pgClientAdapter.WithPostgresClientDeps[T](types.PostgresClientDeps{
		PoolConfig: cfg,
	})
}

// PostgresAdapterWithRequireTLS enforces TLS usage.
func PostgresAdapterWithRequireTLS[T any](required bool) types.PostgresClientOption[T] {
	return pgClientAdapter.WithPostgresClientDeps[T](types.PostgresClientDeps{
		Security: &types.PostgresSecurity{
			RequireTLS: required,
		},
	})
}

// PostgresAdapterWithAllowInsecure allows plaintext connections (not recommended).
func PostgresAdapterWithAllowInsecure[T any](allow bool) types.PostgresClientOption[T] {
	return pgClientAdapter.WithPostgresClientDeps[T](types.PostgresClientDeps{
		Security: &types.PostgresSecurity{
			AllowInsecure: allow,
		},
	})
}

// PostgresAdapterWithTLSConfig sets a custom TLS config.
func PostgresAdapterWithTLSConfig[T any](cfg *tls.Config) types.PostgresClientOption[T] {
	return pgClientAdapter.WithPostgresClientDeps[T](types.PostgresClientDeps{
		Security: &types.PostgresSecurity{
			RequireTLS: true,
			TLSConfig:  cfg,
		},
	})
}

// PostgresAdapterWithWriterConfig sets writer config.
func PostgresAdapterWithWriterConfig[T any](cfg types.PostgresWriterConfig) types.PostgresClientOption[T] {
	return pgClientAdapter.WithWriterConfig[T](cfg)
}

// PostgresAdapterWithReaderConfig sets reader config.
func PostgresAdapterWithReaderConfig[T any](cfg types.PostgresReaderConfig) types.PostgresClientOption[T] {
	return pgClientAdapter.WithReaderConfig[T](cfg)
}

// PostgresAdapterWithTable sets the table name for both reader and writer.
func PostgresAdapterWithTable[T any](table string) types.PostgresClientOption[T] {
	return func(adp types.PostgresClientAdapter[T]) {
		adp.SetWriterConfig(types.PostgresWriterConfig{Table: table})
		adp.SetReaderConfig(types.PostgresReaderConfig{Table: table})
	}
}

// PostgresAdapterWithAutoCreateTable enables auto-create for the writer.
func PostgresAdapterWithAutoCreateTable[T any](enabled bool) types.PostgresClientOption[T] {
	return pgClientAdapter.WithWriterConfig[T](types.PostgresWriterConfig{AutoCreateTable: enabled})
}

// PostgresAdapterWithCreateTableDDL provides custom DDL for auto-create.
func PostgresAdapterWithCreateTableDDL[T any](ddl string) types.PostgresClientOption[T] {
	return pgClientAdapter.WithWriterConfig[T](types.PostgresWriterConfig{CreateTableDDL: ddl})
}

// PostgresAdapterWithBatchSettings controls insert batching.
func PostgresAdapterWithBatchSettings[T any](maxRecords, maxBytes int, maxAge time.Duration) types.PostgresClientOption[T] {
	return pgClientAdapter.WithWriterConfig[T](types.PostgresWriterConfig{
		BatchMaxRecords: maxRecords,
		BatchMaxBytes:   maxBytes,
		BatchMaxAge:     maxAge,
	})
}

// PostgresAdapterWithUpsert enables upsert behavior.
func PostgresAdapterWithUpsert[T any](conflictColumns, updateColumns []string) types.PostgresClientOption[T] {
	return pgClientAdapter.WithWriterConfig[T](types.PostgresWriterConfig{
		Upsert:                true,
		UpsertConflictColumns: conflictColumns,
		UpsertUpdateColumns:   updateColumns,
	})
}

// PostgresAdapterWithClientSideEncryptionAESGCM enables AES-GCM payload encryption (hex key).
func PostgresAdapterWithClientSideEncryptionAESGCM[T any](keyHex string, required bool) types.PostgresClientOption[T] {
	return func(adp types.PostgresClientAdapter[T]) {
		adp.SetWriterConfig(types.PostgresWriterConfig{
			ClientSideEncryption:        "AES-GCM",
			ClientSideKey:               keyHex,
			RequireClientSideEncryption: required,
		})
		adp.SetReaderConfig(types.PostgresReaderConfig{
			ClientSideEncryption:        "AES-GCM",
			ClientSideKey:               keyHex,
			RequireClientSideEncryption: required,
		})
	}
}

// PostgresAdapterWithSecureDefaults enforces TLS + client-side encryption defaults.
func PostgresAdapterWithSecureDefaults[T any](keyHex string) types.PostgresClientOption[T] {
	return pgClientAdapter.WithSecureDefaults[T](keyHex)
}

// PostgresAdapterWithLogger attaches loggers.
func PostgresAdapterWithLogger[T any](l ...types.Logger) types.PostgresClientOption[T] {
	return pgClientAdapter.WithLogger[T](l...)
}

// PostgresAdapterWithSensor attaches sensors.
func PostgresAdapterWithSensor[T any](s ...types.Sensor[T]) types.PostgresClientOption[T] {
	return pgClientAdapter.WithSensor[T](s...)
}

// PostgresAdapterWithWire connects input wires.
func PostgresAdapterWithWire[T any](wires ...types.Wire[T]) types.PostgresClientOption[T] {
	return pgClientAdapter.WithWire[T](wires...)
}

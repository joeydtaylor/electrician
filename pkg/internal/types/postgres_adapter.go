package types

import (
	"context"
	"crypto/tls"
	"database/sql"
	"time"
)

// PostgresSecurity configures TLS requirements for Postgres connections.
type PostgresSecurity struct {
	RequireTLS    bool
	AllowInsecure bool
	TLSConfig     *tls.Config
}

// PostgresPoolSettings controls pgxpool behavior.
type PostgresPoolSettings struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// PostgresClientDeps wires Postgres connection dependencies (no envs).
type PostgresClientDeps struct {
	ConnString string // required unless DB is provided
	DriverName string // e.g. "pgx" or "postgres"
	DB         *sql.DB
	Security   *PostgresSecurity
	PoolConfig PostgresPoolSettings
}

////////////////////////////////////////////////////////////////////////////////
// Writer config (format + batching + schema)
////////////////////////////////////////////////////////////////////////////////

type PostgresWriterConfig struct {
	Table string

	ColumnID               string
	ColumnCreatedAt        string
	ColumnTraceID          string
	ColumnPayload          string
	ColumnPayloadEncrypted string
	ColumnMetadata         string
	ColumnContentType      string
	ColumnPayloadType      string
	ColumnPayloadEncoding  string

	PayloadFormat string // "json" | "bytes"

	// Static metadata merged into the metadata column (JSONB).
	Metadata map[string]any

	// Client-side encryption (AES-GCM) before insert.
	ClientSideEncryption        string
	ClientSideKey               string
	RequireClientSideEncryption bool

	// Upsert behavior (INSERT ... ON CONFLICT ...)
	Upsert                bool
	UpsertConflictColumns []string
	UpsertUpdateColumns   []string

	// Batching thresholds for inserts.
	BatchMaxRecords int
	BatchMaxBytes   int
	BatchMaxAge     time.Duration

	// COPY optimization.
	UseCopy       bool
	CopyThreshold int

	// Optional table creation.
	AutoCreateTable bool
	CreateTableDDL  string
}

////////////////////////////////////////////////////////////////////////////////
// Reader config (selection + decode)
////////////////////////////////////////////////////////////////////////////////

type PostgresReaderConfig struct {
	Table string

	ColumnTraceID          string
	ColumnPayload          string
	ColumnPayloadEncrypted string
	ColumnMetadata         string

	PayloadFormat string // "json" | "bytes"

	// Optional query overrides.
	Query     string
	QueryArgs []any

	// Table-mode query helpers.
	WhereClause string
	OrderBy     string
	Limit       int

	// Polling behavior (Serve). 0 => one-shot.
	PollInterval time.Duration

	// Client-side decryption before decode.
	ClientSideEncryption        string
	ClientSideKey               string
	RequireClientSideEncryption bool
}

////////////////////////////////////////////////////////////////////////////////
// Writer (record-oriented) adapter
////////////////////////////////////////////////////////////////////////////////

type PostgresWriterAdapter[T any] interface {
	SetPostgresClientDeps(PostgresClientDeps)
	SetWriterConfig(PostgresWriterConfig)
	ConnectInput(...Wire[T])

	Serve(ctx context.Context, in <-chan T) error
	StartWriter(ctx context.Context) error

	Stop()

	ConnectLogger(...Logger)
	ConnectSensor(...Sensor[T])
	GetComponentMetadata() ComponentMetadata
	SetComponentMetadata(name, id string)
	NotifyLoggers(level LogLevel, msg string, keysAndValues ...interface{})
	Name() string
}

////////////////////////////////////////////////////////////////////////////////
// Reader (record-oriented) adapter
////////////////////////////////////////////////////////////////////////////////

type PostgresReaderAdapter[T any] interface {
	SetPostgresClientDeps(PostgresClientDeps)
	SetReaderConfig(PostgresReaderConfig)

	Serve(ctx context.Context, submit func(context.Context, T) error) error
	Fetch() (HttpResponse[[]T], error)

	Stop()

	ConnectLogger(...Logger)
	ConnectSensor(...Sensor[T])
	GetComponentMetadata() ComponentMetadata
	SetComponentMetadata(name, id string)
	NotifyLoggers(level LogLevel, msg string, keysAndValues ...interface{})
	Name() string
}

////////////////////////////////////////////////////////////////////////////////
// Unified Postgres client adapter (read + write)
////////////////////////////////////////////////////////////////////////////////

type PostgresClientAdapter[T any] interface {
	SetPostgresClientDeps(PostgresClientDeps)
	SetWriterConfig(PostgresWriterConfig)
	SetReaderConfig(PostgresReaderConfig)

	ServeWriter(ctx context.Context, in <-chan T) error
	ConnectInput(...Wire[T])
	StartWriter(ctx context.Context) error

	Serve(ctx context.Context, submit func(context.Context, T) error) error
	Fetch() (HttpResponse[[]T], error)

	Stop()

	ConnectLogger(...Logger)
	ConnectSensor(...Sensor[T])
	GetComponentMetadata() ComponentMetadata
	SetComponentMetadata(name, id string)
	NotifyLoggers(level LogLevel, msg string, keysAndValues ...interface{})
	Name() string
}

////////////////////////////////////////////////////////////////////////////////
// Options (builder-style)
////////////////////////////////////////////////////////////////////////////////

type PostgresWriterOption[T any] func(PostgresWriterAdapter[T])
type PostgresReaderOption[T any] func(PostgresReaderAdapter[T])
type PostgresClientOption[T any] func(PostgresClientAdapter[T])

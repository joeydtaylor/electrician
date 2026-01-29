package postgresclient

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// SetPostgresClientDeps wires Postgres connection dependencies.
func (p *PostgresClient[T]) SetPostgresClientDeps(d types.PostgresClientDeps) {
	if d.ConnString != "" {
		p.connString = d.ConnString
	}
	if d.DriverName != "" {
		p.driverName = d.DriverName
	}
	if d.DB != nil {
		p.db = d.DB
	}
	if d.Security != nil {
		p.security = *d.Security
	}
	p.poolCfg = mergePoolSettings(p.poolCfg, d.PoolConfig)
}

// SetWriterConfig applies writer configuration fields that are explicitly set.
func (p *PostgresClient[T]) SetWriterConfig(c types.PostgresWriterConfig) {
	if c.Table != "" {
		p.writerCfg.Table = c.Table
	}
	if c.ColumnID != "" {
		p.writerCfg.ColumnID = c.ColumnID
	}
	if c.ColumnCreatedAt != "" {
		p.writerCfg.ColumnCreatedAt = c.ColumnCreatedAt
	}
	if c.ColumnTraceID != "" {
		p.writerCfg.ColumnTraceID = c.ColumnTraceID
	}
	if c.ColumnPayload != "" {
		p.writerCfg.ColumnPayload = c.ColumnPayload
	}
	if c.ColumnPayloadEncrypted != "" {
		p.writerCfg.ColumnPayloadEncrypted = c.ColumnPayloadEncrypted
	}
	if c.ColumnMetadata != "" {
		p.writerCfg.ColumnMetadata = c.ColumnMetadata
	}
	if c.ColumnContentType != "" {
		p.writerCfg.ColumnContentType = c.ColumnContentType
	}
	if c.ColumnPayloadType != "" {
		p.writerCfg.ColumnPayloadType = c.ColumnPayloadType
	}
	if c.ColumnPayloadEncoding != "" {
		p.writerCfg.ColumnPayloadEncoding = c.ColumnPayloadEncoding
	}
	if c.PayloadFormat != "" {
		p.writerCfg.PayloadFormat = strings.ToLower(c.PayloadFormat)
	}
	if c.Metadata != nil {
		if p.writerCfg.Metadata == nil {
			p.writerCfg.Metadata = map[string]any{}
		}
		for k, v := range c.Metadata {
			p.writerCfg.Metadata[k] = v
		}
	}
	if c.ClientSideEncryption != "" {
		p.writerCfg.ClientSideEncryption = c.ClientSideEncryption
	}
	if c.ClientSideKey != "" {
		p.writerCfg.ClientSideKey = c.ClientSideKey
	}
	if c.RequireClientSideEncryption {
		p.writerCfg.RequireClientSideEncryption = true
	}
	if c.Upsert {
		p.writerCfg.Upsert = true
	}
	if len(c.UpsertConflictColumns) > 0 {
		p.writerCfg.UpsertConflictColumns = c.UpsertConflictColumns
	}
	if len(c.UpsertUpdateColumns) > 0 {
		p.writerCfg.UpsertUpdateColumns = c.UpsertUpdateColumns
	}
	if c.BatchMaxRecords > 0 {
		p.writerCfg.BatchMaxRecords = c.BatchMaxRecords
	}
	if c.BatchMaxBytes > 0 {
		p.writerCfg.BatchMaxBytes = c.BatchMaxBytes
	}
	if c.BatchMaxAge > 0 {
		p.writerCfg.BatchMaxAge = c.BatchMaxAge
	}
	if c.UseCopy {
		p.writerCfg.UseCopy = true
	}
	if c.CopyThreshold > 0 {
		p.writerCfg.CopyThreshold = c.CopyThreshold
	}
	if c.AutoCreateTable {
		p.writerCfg.AutoCreateTable = true
	}
	if c.CreateTableDDL != "" {
		p.writerCfg.CreateTableDDL = c.CreateTableDDL
	}
}

// SetReaderConfig applies reader configuration fields that are explicitly set.
func (p *PostgresClient[T]) SetReaderConfig(c types.PostgresReaderConfig) {
	if c.Table != "" {
		p.readerCfg.Table = c.Table
	}
	if c.ColumnTraceID != "" {
		p.readerCfg.ColumnTraceID = c.ColumnTraceID
	}
	if c.ColumnPayload != "" {
		p.readerCfg.ColumnPayload = c.ColumnPayload
	}
	if c.ColumnPayloadEncrypted != "" {
		p.readerCfg.ColumnPayloadEncrypted = c.ColumnPayloadEncrypted
	}
	if c.ColumnMetadata != "" {
		p.readerCfg.ColumnMetadata = c.ColumnMetadata
	}
	if c.PayloadFormat != "" {
		p.readerCfg.PayloadFormat = strings.ToLower(c.PayloadFormat)
	}
	if c.Query != "" {
		p.readerCfg.Query = c.Query
	}
	if len(c.QueryArgs) > 0 {
		p.readerCfg.QueryArgs = c.QueryArgs
	}
	if c.WhereClause != "" {
		p.readerCfg.WhereClause = c.WhereClause
	}
	if c.OrderBy != "" {
		p.readerCfg.OrderBy = c.OrderBy
	}
	if c.Limit > 0 {
		p.readerCfg.Limit = c.Limit
	}
	if c.PollInterval > 0 {
		p.readerCfg.PollInterval = c.PollInterval
	}
	if c.ClientSideEncryption != "" {
		p.readerCfg.ClientSideEncryption = c.ClientSideEncryption
	}
	if c.ClientSideKey != "" {
		p.readerCfg.ClientSideKey = c.ClientSideKey
	}
	if c.RequireClientSideEncryption {
		p.readerCfg.RequireClientSideEncryption = true
	}
}

func mergePoolSettings(base types.PostgresPoolSettings, in types.PostgresPoolSettings) types.PostgresPoolSettings {
	if in.MaxOpenConns > 0 {
		base.MaxOpenConns = in.MaxOpenConns
	}
	if in.MaxIdleConns > 0 {
		base.MaxIdleConns = in.MaxIdleConns
	}
	if in.ConnMaxLifetime > 0 {
		base.ConnMaxLifetime = in.ConnMaxLifetime
	}
	if in.ConnMaxIdleTime > 0 {
		base.ConnMaxIdleTime = in.ConnMaxIdleTime
	}
	return base
}

func (p *PostgresClient[T]) ensureDB(ctx context.Context) (*sql.DB, error) {
	if p.db != nil {
		return p.db, nil
	}
	if p.connString == "" {
		return nil, fmt.Errorf("postgres: conn string is required")
	}
	if p.driverName == "" {
		return nil, fmt.Errorf("postgres: driver name is required (e.g. \"pgx\" or \"postgres\")")
	}
	if err := p.validateTLSRequirement(); err != nil {
		return nil, err
	}
	db, err := sql.Open(p.driverName, p.connString)
	if err != nil {
		return nil, err
	}
	if p.poolCfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(p.poolCfg.MaxOpenConns)
	}
	if p.poolCfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(p.poolCfg.MaxIdleConns)
	}
	if p.poolCfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(p.poolCfg.ConnMaxLifetime)
	}
	if p.poolCfg.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(p.poolCfg.ConnMaxIdleTime)
	}
	if p.security.RequireTLS && p.security.TLSConfig != nil {
		_ = p.security.TLSConfig
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	p.db = db
	return db, nil
}

func (p *PostgresClient[T]) validateTLSRequirement() error {
	if p.security.AllowInsecure || !p.security.RequireTLS {
		return nil
	}
	mode := extractSSLMode(p.connString)
	if mode == "" {
		return fmt.Errorf("postgres: sslmode must be set to require/verify-full (TLS required)")
	}
	switch strings.ToLower(mode) {
	case "require", "verify-full", "verify-ca":
		return nil
	default:
		return fmt.Errorf("postgres: insecure sslmode=%s (TLS required)", mode)
	}
}

func extractSSLMode(conn string) string {
	if strings.HasPrefix(conn, "postgres://") || strings.HasPrefix(conn, "postgresql://") {
		u, err := url.Parse(conn)
		if err != nil {
			return ""
		}
		return u.Query().Get("sslmode")
	}
	// keyword/value form: "host=... sslmode=require"
	for _, part := range strings.Fields(conn) {
		if strings.HasPrefix(part, "sslmode=") {
			return strings.TrimPrefix(part, "sslmode=")
		}
	}
	return ""
}

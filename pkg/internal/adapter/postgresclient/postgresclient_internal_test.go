package postgresclient

import (
	"testing"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func TestBuildInsertBatchSQL(t *testing.T) {
	cfg := defaultWriterConfig()
	cfg.Table = "events"
	cfg.Upsert = true
	cfg.UpsertConflictColumns = []string{"id"}

	rows := []rowData{
		{id: "a", payload: []byte("one")},
		{id: "b", payload: []byte("two")},
	}

	sql, args := buildInsertBatchSQL(cfg, rows)
	if sql == "" {
		t.Fatalf("expected sql")
	}
	if len(args) != len(rows)*len(buildInsertColumns(cfg)) {
		t.Fatalf("unexpected args length: got %d", len(args))
	}
}

func TestValidateTLSRequirement(t *testing.T) {
	p := &PostgresClient[any]{
		connString: "postgres://user:pass@localhost:5432/db?sslmode=disable",
		security: types.PostgresSecurity{
			RequireTLS: true,
		},
		driverName: "pgx",
	}
	if err := p.validateTLSRequirement(); err == nil {
		t.Fatalf("expected TLS requirement error")
	}
}

func TestResolveWriterEncryption(t *testing.T) {
	p := &PostgresClient[any]{
		writerCfg: types.PostgresWriterConfig{
			RequireClientSideEncryption: true,
		},
	}
	if _, _, err := p.resolveWriterEncryption(); err == nil {
		t.Fatalf("expected encryption requirement error")
	}
}

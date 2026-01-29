package postgresclient

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func (p *PostgresClient[T]) ServeWriter(ctx context.Context, in <-chan T) error {
	return p.serveWriter(ctx, in)
}

func (p *PostgresClient[T]) StartWriter(ctx context.Context) error {
	in := p.mergeInputs()
	if in == nil {
		return fmt.Errorf("postgres: no input wires connected")
	}
	return p.serveWriter(ctx, in)
}

// serveWriter writes records into Postgres using batch inserts.
func (p *PostgresClient[T]) serveWriter(ctx context.Context, in <-chan T) error {
	db, err := p.ensureDB(ctx)
	if err != nil {
		return err
	}
	if err := p.ensureTable(ctx); err != nil {
		return err
	}

	encMode, encKey, err := p.resolveWriterEncryption()
	if err != nil {
		return err
	}

	batch := make([]rowData, 0, p.writerCfg.BatchMaxRecords)
	byteCount := 0
	lastFlush := time.Now()

	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		defer func() {
			batch = batch[:0]
			byteCount = 0
			lastFlush = time.Now()
		}()

		return p.insertRows(ctx, db, batch)
	}

	for {
		select {
		case <-ctx.Done():
			return flush()
		case item, ok := <-in:
			if !ok {
				return flush()
			}
			payload, err := encodePayload(item, p.writerCfg.PayloadFormat)
			if err != nil {
				return err
			}
			encrypted := false
			if encMode == cseModeAESGCM {
				payload, err = encryptAESGCM(payload, encKey)
				if err != nil {
					return err
				}
				encrypted = true
			}
			row := newRowData(p.writerCfg, payload, encrypted)
			meta, err := buildMetadata(p.writerCfg, row.traceID)
			if err != nil {
				return err
			}
			row.metadata = meta

			byteCount += len(payload)
			batch = append(batch, row)

			if len(batch) >= p.writerCfg.BatchMaxRecords || byteCount >= p.writerCfg.BatchMaxBytes || time.Since(lastFlush) >= p.writerCfg.BatchMaxAge {
				if err := flush(); err != nil {
					return err
				}
			}
		}
	}
}

func (p *PostgresClient[T]) resolveWriterEncryption() (string, []byte, error) {
	mode := strings.ToLower(strings.TrimSpace(p.writerCfg.ClientSideEncryption))
	keyHex := strings.TrimSpace(p.writerCfg.ClientSideKey)
	if mode == "" && keyHex != "" {
		mode = cseModeAESGCM
	}
	if p.writerCfg.RequireClientSideEncryption && (mode == "" || keyHex == "") {
		return "", nil, fmt.Errorf("postgres: client-side encryption is required")
	}
	if mode == "" {
		return "", nil, nil
	}
	if mode != cseModeAESGCM {
		return "", nil, fmt.Errorf("postgres: unsupported client-side encryption mode: %s", mode)
	}
	key, err := parseAESGCMKeyHex(keyHex)
	if err != nil {
		return "", nil, err
	}
	return mode, key, nil
}

func (p *PostgresClient[T]) insertRows(ctx context.Context, db *sql.DB, rows []rowData) error {
	query, args := buildInsertBatchSQL(p.writerCfg, rows)
	if query == "" {
		return fmt.Errorf("postgres: empty insert sql")
	}
	_, err := db.ExecContext(ctx, query, args...)
	return err
}

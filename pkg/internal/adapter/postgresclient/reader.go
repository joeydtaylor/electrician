package postgresclient

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func (p *PostgresClient[T]) Serve(ctx context.Context, submit func(context.Context, T) error) error {
	if p.readerCfg.PollInterval <= 0 {
		resp, err := p.Fetch()
		if err != nil {
			return err
		}
		for _, item := range resp.Body {
			if err := submit(ctx, item); err != nil {
				return err
			}
		}
		return nil
	}

	ticker := time.NewTicker(p.readerCfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			resp, err := p.Fetch()
			if err != nil {
				return err
			}
			for _, item := range resp.Body {
				if err := submit(ctx, item); err != nil {
					return err
				}
			}
		}
	}
}

func (p *PostgresClient[T]) Fetch() (types.HttpResponse[[]T], error) {
	db, err := p.ensureDB(p.ctx)
	if err != nil {
		return types.HttpResponse[[]T]{StatusCode: 500}, err
	}

	query, args := p.buildSelectQuery()
	rows, err := db.QueryContext(p.ctx, query, args...)
	if err != nil {
		return types.HttpResponse[[]T]{StatusCode: 500}, err
	}
	defer rows.Close()

	mode, key, err := p.resolveReaderEncryption()
	if err != nil {
		return types.HttpResponse[[]T]{StatusCode: 500}, err
	}

	var out []T
	for rows.Next() {
		var payload []byte
		var encrypted bool
		if p.readerCfg.ColumnPayloadEncrypted != "" {
			if err := rows.Scan(&payload, &encrypted); err != nil {
				return types.HttpResponse[[]T]{StatusCode: 500}, err
			}
		} else {
			if err := rows.Scan(&payload); err != nil {
				return types.HttpResponse[[]T]{StatusCode: 500}, err
			}
		}
		if encrypted {
			if mode == "" {
				return types.HttpResponse[[]T]{StatusCode: 500}, fmt.Errorf("postgres: encrypted payload without decryption key")
			}
			payload, err = decryptAESGCM(payload, key)
			if err != nil {
				return types.HttpResponse[[]T]{StatusCode: 500}, err
			}
		} else if p.readerCfg.RequireClientSideEncryption {
			return types.HttpResponse[[]T]{StatusCode: 500}, fmt.Errorf("postgres: payload is not encrypted")
		}

		item, err := decodePayload[T](payload, p.readerCfg.PayloadFormat)
		if err != nil {
			return types.HttpResponse[[]T]{StatusCode: 500}, err
		}
		out = append(out, item)
	}
	if rows.Err() != nil {
		return types.HttpResponse[[]T]{StatusCode: 500}, rows.Err()
	}
	return types.HttpResponse[[]T]{StatusCode: 200, Body: out}, nil
}

func (p *PostgresClient[T]) buildSelectQuery() (string, []any) {
	if p.readerCfg.Query != "" {
		return p.readerCfg.Query, p.readerCfg.QueryArgs
	}
	cols := []string{p.readerCfg.ColumnPayload}
	if p.readerCfg.ColumnPayloadEncrypted != "" {
		cols = append(cols, p.readerCfg.ColumnPayloadEncrypted)
	}

	var sb strings.Builder
	sb.WriteString("SELECT ")
	sb.WriteString(strings.Join(cols, ","))
	sb.WriteString(" FROM ")
	sb.WriteString(p.readerCfg.Table)
	if p.readerCfg.WhereClause != "" {
		sb.WriteString(" WHERE ")
		sb.WriteString(p.readerCfg.WhereClause)
	}
	if p.readerCfg.OrderBy != "" {
		sb.WriteString(" ORDER BY ")
		sb.WriteString(p.readerCfg.OrderBy)
	}
	if p.readerCfg.Limit > 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", p.readerCfg.Limit))
	}
	return sb.String(), nil
}

func (p *PostgresClient[T]) resolveReaderEncryption() (string, []byte, error) {
	mode := strings.ToLower(strings.TrimSpace(p.readerCfg.ClientSideEncryption))
	keyHex := strings.TrimSpace(p.readerCfg.ClientSideKey)
	if mode == "" && keyHex != "" {
		mode = cseModeAESGCM
	}
	if p.readerCfg.RequireClientSideEncryption && (mode == "" || keyHex == "") {
		return "", nil, fmt.Errorf("postgres: client-side encryption is required for reads")
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

func decodePayload[T any](payload []byte, format string) (T, error) {
	var out T
	switch strings.ToLower(format) {
	case "", payloadFormatJSON:
		if err := json.Unmarshal(payload, &out); err != nil {
			return out, err
		}
		return out, nil
	case payloadFormatBytes:
		if b, ok := any(&out).(*[]byte); ok {
			*b = payload
			return out, nil
		}
		return out, fmt.Errorf("payload format bytes requires T to be []byte")
	default:
		return out, fmt.Errorf("unsupported payload format: %s", format)
	}
}

package postgresclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

const (
	payloadFormatJSON  = "json"
	payloadFormatBytes = "bytes"
)

func defaultTableDDL(cfg types.PostgresWriterConfig) string {
	table := cfg.Table
	if table == "" {
		return ""
	}
	idCol := cfg.ColumnID
	createdCol := cfg.ColumnCreatedAt
	traceCol := cfg.ColumnTraceID
	payloadCol := cfg.ColumnPayload
	encryptedCol := cfg.ColumnPayloadEncrypted
	metaCol := cfg.ColumnMetadata
	ctCol := cfg.ColumnContentType
	ptCol := cfg.ColumnPayloadType
	peCol := cfg.ColumnPayloadEncoding

	var buf strings.Builder
	buf.WriteString("CREATE TABLE IF NOT EXISTS ")
	buf.WriteString(table)
	buf.WriteString(" (")
	buf.WriteString(idCol)
	buf.WriteString(" TEXT PRIMARY KEY, ")
	buf.WriteString(createdCol)
	buf.WriteString(" TIMESTAMPTZ NOT NULL, ")
	if traceCol != "" {
		buf.WriteString(traceCol)
		buf.WriteString(" TEXT, ")
	}
	buf.WriteString(payloadCol)
	buf.WriteString(" BYTEA NOT NULL, ")
	if encryptedCol != "" {
		buf.WriteString(encryptedCol)
		buf.WriteString(" BOOLEAN NOT NULL DEFAULT false, ")
	}
	if metaCol != "" {
		buf.WriteString(metaCol)
		buf.WriteString(" JSONB, ")
	}
	if ctCol != "" {
		buf.WriteString(ctCol)
		buf.WriteString(" TEXT, ")
	}
	if ptCol != "" {
		buf.WriteString(ptCol)
		buf.WriteString(" TEXT, ")
	}
	if peCol != "" {
		buf.WriteString(peCol)
		buf.WriteString(" TEXT, ")
	}
	sql := buf.String()
	sql = strings.TrimSuffix(sql, ", ")
	sql += ")"
	return sql
}

func encodePayload[T any](item T, format string) ([]byte, error) {
	switch strings.ToLower(format) {
	case "", payloadFormatJSON:
		return json.Marshal(item)
	case payloadFormatBytes:
		if b, ok := any(item).([]byte); ok {
			return b, nil
		}
		return nil, fmt.Errorf("payload format bytes requires []byte")
	default:
		return nil, fmt.Errorf("unsupported payload format: %s", format)
	}
}

func buildMetadata(cfg types.PostgresWriterConfig, traceID string) ([]byte, error) {
	if cfg.ColumnMetadata == "" {
		return nil, nil
	}
	meta := map[string]any{}
	for k, v := range cfg.Metadata {
		meta[k] = v
	}
	if traceID != "" {
		meta["trace_id"] = traceID
	}
	if len(meta) == 0 {
		return nil, nil
	}
	return json.Marshal(meta)
}

func extractTraceID(payload []byte) string {
	var m map[string]any
	if err := json.Unmarshal(payload, &m); err != nil {
		return ""
	}
	for _, key := range []string{"trace_id", "traceId", "trace-id"} {
		if v, ok := m[key]; ok {
			switch t := v.(type) {
			case string:
				return t
			}
		}
	}
	return ""
}

type rowData struct {
	id        string
	createdAt time.Time
	traceID   string
	payload   []byte
	encrypted bool
	metadata  []byte
}

func newRowData(cfg types.PostgresWriterConfig, payload []byte, encrypted bool) rowData {
	traceID := ""
	if cfg.ColumnTraceID != "" {
		traceID = extractTraceID(payload)
	}
	return rowData{
		id:        utils.GenerateUniqueHash(),
		createdAt: time.Now().UTC(),
		traceID:   traceID,
		payload:   payload,
		encrypted: encrypted,
	}
}

func buildInsertColumns(cfg types.PostgresWriterConfig) []string {
	cols := []string{cfg.ColumnID, cfg.ColumnCreatedAt, cfg.ColumnPayload}
	if cfg.ColumnTraceID != "" {
		cols = append(cols, cfg.ColumnTraceID)
	}
	if cfg.ColumnPayloadEncrypted != "" {
		cols = append(cols, cfg.ColumnPayloadEncrypted)
	}
	if cfg.ColumnMetadata != "" {
		cols = append(cols, cfg.ColumnMetadata)
	}
	if cfg.ColumnContentType != "" {
		cols = append(cols, cfg.ColumnContentType)
	}
	if cfg.ColumnPayloadType != "" {
		cols = append(cols, cfg.ColumnPayloadType)
	}
	if cfg.ColumnPayloadEncoding != "" {
		cols = append(cols, cfg.ColumnPayloadEncoding)
	}
	return cols
}

func buildInsertSQL(cfg types.PostgresWriterConfig) string {
	cols := buildInsertColumns(cfg)
	if len(cols) == 0 {
		return ""
	}
	var buf bytes.Buffer
	buf.WriteString("INSERT INTO ")
	buf.WriteString(cfg.Table)
	buf.WriteString(" (")
	buf.WriteString(strings.Join(cols, ","))
	buf.WriteString(") VALUES (")
	for i := range cols {
		if i > 0 {
			buf.WriteString(",")
		}
		buf.WriteString(fmt.Sprintf("$%d", i+1))
	}
	buf.WriteString(")")

	if cfg.Upsert && len(cfg.UpsertConflictColumns) > 0 {
		buf.WriteString(" ON CONFLICT (")
		buf.WriteString(strings.Join(cfg.UpsertConflictColumns, ","))
		buf.WriteString(") DO UPDATE SET ")
		updates := cfg.UpsertUpdateColumns
		if len(updates) == 0 {
			for _, col := range cols {
				if col == cfg.ColumnID || col == cfg.ColumnCreatedAt {
					continue
				}
				updates = append(updates, col)
			}
		}
		for i, col := range updates {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(col)
			buf.WriteString("=EXCLUDED.")
			buf.WriteString(col)
		}
	}
	return buf.String()
}

func buildInsertBatchSQL(cfg types.PostgresWriterConfig, rows []rowData) (string, []any) {
	if len(rows) == 0 {
		return "", nil
	}
	cols := buildInsertColumns(cfg)
	if len(cols) == 0 {
		return "", nil
	}
	var buf bytes.Buffer
	args := make([]any, 0, len(rows)*len(cols))
	buf.WriteString("INSERT INTO ")
	buf.WriteString(cfg.Table)
	buf.WriteString(" (")
	buf.WriteString(strings.Join(cols, ","))
	buf.WriteString(") VALUES ")

	argIndex := 1
	for i, r := range rows {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString("(")
		vals := rowValuesFor(cfg, r)
		for j := range cols {
			if j > 0 {
				buf.WriteString(",")
			}
			buf.WriteString(fmt.Sprintf("$%d", argIndex))
			argIndex++
		}
		buf.WriteString(")")
		args = append(args, vals...)
	}

	if cfg.Upsert && len(cfg.UpsertConflictColumns) > 0 {
		buf.WriteString(" ON CONFLICT (")
		buf.WriteString(strings.Join(cfg.UpsertConflictColumns, ","))
		buf.WriteString(") DO UPDATE SET ")
		updates := cfg.UpsertUpdateColumns
		if len(updates) == 0 {
			for _, col := range cols {
				if col == cfg.ColumnID || col == cfg.ColumnCreatedAt {
					continue
				}
				updates = append(updates, col)
			}
		}
		for i, col := range updates {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(col)
			buf.WriteString("=EXCLUDED.")
			buf.WriteString(col)
		}
	}
	return buf.String(), args
}

func rowValuesFor(cfg types.PostgresWriterConfig, r rowData) []any {
	values := []any{r.id, r.createdAt, r.payload}
	if cfg.ColumnTraceID != "" {
		values = append(values, r.traceID)
	}
	if cfg.ColumnPayloadEncrypted != "" {
		values = append(values, r.encrypted)
	}
	if cfg.ColumnMetadata != "" {
		values = append(values, r.metadata)
	}
	if cfg.ColumnContentType != "" {
		values = append(values, "application/json")
	}
	if cfg.ColumnPayloadType != "" {
		values = append(values, "")
	}
	if cfg.ColumnPayloadEncoding != "" {
		values = append(values, cfg.PayloadFormat)
	}
	return values
}

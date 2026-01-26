package logschema

// Log schema constants for Electrician structured logs.
const (
	SchemaID    = "electrician.log.v1"
	FieldSchema = "log_schema"

	FieldTimestamp = "ts"
	FieldLevel     = "level"
	FieldMessage   = "msg"
	FieldLogger    = "logger"
	FieldCaller    = "caller"
	FieldStack     = "stack"

	FieldComponent = "component"
	FieldEvent     = "event"
	FieldResult    = "result"
	FieldError     = "error"
	FieldTraceID   = "trace_id"
	FieldSpanID    = "span_id"
)

// LogRecord is a generic map representation of a log entry.
type LogRecord map[string]interface{}

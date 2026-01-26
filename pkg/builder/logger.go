package builder

import (
	internalLogger "github.com/joeydtaylor/electrician/pkg/internal/internallogger"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/logschema"
)

type LoggerOption = internalLogger.LoggerOption

type SinkConfig = types.SinkConfig

type SinkType = types.SinkType

type HTTPServerResponse = types.HTTPServerResponse

const (
	FileSink    SinkType = "file"
	StdoutSink  SinkType = "stdout"
	NetworkSink SinkType = "network"
	RelaySink   SinkType = "relay"
)

func NewLogger(options ...internalLogger.LoggerOption) types.Logger {
	return internalLogger.NewLogger(options...)
}

// WithLevel configures the logger to use the specified log level
func LoggerWithLevel(levelStr string) LoggerOption {
	return internalLogger.LoggerWithLevel(levelStr)
}

// WithDevelopment enables or disables development mode
func LoggerWithDevelopment(dev bool) LoggerOption {
	return internalLogger.LoggerWithDevelopment(dev)
}

// LoggerWithFields attaches fields to every log line.
func LoggerWithFields(fields map[string]interface{}) LoggerOption {
	return internalLogger.LoggerWithFields(fields)
}

// LoggerWithSchema overrides the log schema identifier field.
func LoggerWithSchema(schema string) LoggerOption {
	return internalLogger.LoggerWithSchema(schema)
}

// Log schema constants for the standard Electrician log format.
const (
	LogSchemaID    = logschema.SchemaID
	LogSchemaField = logschema.FieldSchema
)

// LogLevel is exported from the internal types package.
type LogLevel = types.LogLevel

// Export log levels to be accessible under the builder package
const (
	DebugLevel  = types.DebugLevel
	InfoLevel   = types.InfoLevel
	WarnLevel   = types.WarnLevel
	ErrorLevel  = types.ErrorLevel
	DPanicLevel = types.DPanicLevel
	PanicLevel  = types.PanicLevel
	FatalLevel  = types.FatalLevel
)

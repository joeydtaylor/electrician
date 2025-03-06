package builder

import "github.com/joeydtaylor/electrician/pkg/internal/types"

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

package internallogger

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// parseLogLevel converts a string representation of the log level to types.LogLevel
func parseLogLevel(levelStr string) types.LogLevel {
	switch levelStr {
	case "debug":
		return types.DebugLevel
	case "info":
		return types.InfoLevel
	case "warn":
		return types.WarnLevel
	case "error":
		return types.ErrorLevel
	case "dpanic":
		return types.DPanicLevel
	case "panic":
		return types.PanicLevel
	case "fatal":
		return types.FatalLevel
	default:
		return types.InfoLevel // Default level is Info if unspecified
	}
}

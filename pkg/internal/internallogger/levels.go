package internallogger

import (
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"go.uber.org/zap/zapcore"
)

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
		return types.InfoLevel
	}
}

// ConvertLevel converts a types.LogLevel to a zap level.
func ConvertLevel(level types.LogLevel) zapcore.Level {
	switch level {
	case types.DebugLevel:
		return zapcore.DebugLevel
	case types.InfoLevel:
		return zapcore.InfoLevel
	case types.WarnLevel:
		return zapcore.WarnLevel
	case types.ErrorLevel:
		return zapcore.ErrorLevel
	case types.DPanicLevel:
		return zapcore.DPanicLevel
	case types.PanicLevel:
		return zapcore.PanicLevel
	case types.FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func convertZapLevel(level zapcore.Level) types.LogLevel {
	switch level {
	case zapcore.DebugLevel:
		return types.DebugLevel
	case zapcore.InfoLevel:
		return types.InfoLevel
	case zapcore.WarnLevel:
		return types.WarnLevel
	case zapcore.ErrorLevel:
		return types.ErrorLevel
	case zapcore.DPanicLevel:
		return types.DPanicLevel
	case zapcore.PanicLevel:
		return types.PanicLevel
	case zapcore.FatalLevel:
		return types.FatalLevel
	default:
		return types.InfoLevel
	}
}

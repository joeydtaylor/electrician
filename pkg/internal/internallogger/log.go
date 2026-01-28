package internallogger

import (
	"fmt"
	"strings"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"go.uber.org/zap"
)

// Log emits a log entry at the requested level with structured fields.
func (z *ZapLoggerAdapter) Log(level types.LogLevel, msg string, keysAndValues ...interface{}) {
	z.mu.Lock()
	logger := z.logger
	z.mu.Unlock()

	if logger == nil || logger.Core() == nil {
		fmt.Println("Logger or logger core is not initialized.")
		return
	}

	zapLevel := ConvertLevel(level)
	if !logger.Core().Enabled(zapLevel) {
		return
	}

	limit := len(keysAndValues)
	if limit%2 != 0 {
		limit--
	}

	fields := make([]zap.Field, 0, limit/2)
	for i := 0; i < limit; i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			continue
		}
		value := keysAndValues[i+1]
		switch v := value.(type) {
		case types.ComponentMetadata:
			fields = append(fields, zap.Any(key, componentToLogMap(v)))
			continue
		case *types.ComponentMetadata:
			if v == nil {
				fields = append(fields, zap.Any(key, nil))
				continue
			}
			fields = append(fields, zap.Any(key, componentToLogMap(*v)))
			continue
		case error:
			fields = append(fields, zap.NamedError(key, v))
			continue
		}
		fields = append(fields, zap.Any(key, value))
	}

	logger.Check(zapLevel, msg).Write(fields...)
}

func componentToLogMap(meta types.ComponentMetadata) map[string]string {
	return map[string]string{
		"id":   meta.ID,
		"type": meta.Type,
		"name": meta.Name,
	}
}

// Debug logs a debug message.
func (z *ZapLoggerAdapter) Debug(msg string, keysAndValues ...interface{}) {
	z.Log(types.DebugLevel, msg, keysAndValues...)
}

// Info logs an informational message.
func (z *ZapLoggerAdapter) Info(msg string, keysAndValues ...interface{}) {
	z.Log(types.InfoLevel, msg, keysAndValues...)
}

// Warn logs a warning message.
func (z *ZapLoggerAdapter) Warn(msg string, keysAndValues ...interface{}) {
	z.Log(types.WarnLevel, msg, keysAndValues...)
}

// Error logs an error message.
func (z *ZapLoggerAdapter) Error(msg string, keysAndValues ...interface{}) {
	z.Log(types.ErrorLevel, msg, keysAndValues...)
}

// DPanic logs a critical message.
func (z *ZapLoggerAdapter) DPanic(msg string, keysAndValues ...interface{}) {
	z.Log(types.DPanicLevel, msg, keysAndValues...)
}

// Panic logs a message and panics.
func (z *ZapLoggerAdapter) Panic(msg string, keysAndValues ...interface{}) {
	z.Log(types.PanicLevel, msg, keysAndValues...)
}

// Fatal logs a fatal message.
func (z *ZapLoggerAdapter) Fatal(msg string, keysAndValues ...interface{}) {
	z.Log(types.FatalLevel, msg, keysAndValues...)
}

// GetLevel returns the configured log level.
func (z *ZapLoggerAdapter) GetLevel() types.LogLevel {
	return convertZapLevel(z.atomicLevel.Level())
}

// SetLevel updates the logger's minimum level.
func (z *ZapLoggerAdapter) SetLevel(level types.LogLevel) {
	zapLevel := ConvertLevel(level)
	z.atomicLevel.SetLevel(zapLevel)
}

// Flush syncs the logger's outputs.
func (z *ZapLoggerAdapter) Flush() error {
	z.mu.Lock()
	logger := z.logger
	z.mu.Unlock()

	if logger == nil {
		return nil
	}

	if err := logger.Sync(); err != nil {
		if strings.Contains(err.Error(), "inappropriate ioctl for device") ||
			strings.Contains(err.Error(), "bad file descriptor") ||
			strings.Contains(err.Error(), "invalid argument") {
			return nil
		}
		return err
	}
	return nil
}

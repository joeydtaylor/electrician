package internallogger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (z *ZapLoggerAdapter) AddSink(identifier string, config types.SinkConfig) error {
	z.mu.Lock()
	defer z.mu.Unlock()

	var ws zapcore.WriteSyncer

	switch config.Type {
	case "file":
		if path, ok := config.Config["path"].(string); ok {
			// Ensure the directory for the file exists
			dir := filepath.Dir(path)
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				// Create the directory with appropriate permissions
				if err := os.MkdirAll(dir, 0755); err != nil {
					return fmt.Errorf("failed to create directory %s: %v", dir, err)
				}
			}

			// Open or create the file
			file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %v", path, err)
			}
			ws = zapcore.AddSync(file)
		} else {
			return fmt.Errorf("file path configuration is missing or invalid")
		}
	case "stdout":
		ws = zapcore.Lock(os.Stdout)
	case "network":
		// Handle network-based log streaming
	default:
		return fmt.Errorf("unsupported sink type: %s", config.Type)
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // Ensuring time is logged in a readable format
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), ws, z.level)
	z.sinks[identifier] = core

	// Update logger with new combined cores
	if z.combinedCore == nil {
		z.combinedCore = zapcore.NewTee(core)
	} else {
		z.combinedCore = zapcore.NewTee(append([]zapcore.Core{z.combinedCore}, core)...)
	}
	z.logger = zap.New(z.combinedCore, zap.AddCaller(), zap.AddCallerSkip(z.callerDepth)) // Reapply caller settings

	return nil
}

// RemoveSink removes a sink based on its identifier.
func (z *ZapLoggerAdapter) RemoveSink(identifier string) error {
	z.mu.Lock()
	defer z.mu.Unlock()

	if _, ok := z.sinks[identifier]; ok {
		delete(z.sinks, identifier)
		return nil
	}

	return fmt.Errorf("sink not found: %s", identifier)
}

// ListSinks lists all configured sinks.
func (z *ZapLoggerAdapter) ListSinks() ([]string, error) {
	z.mu.Lock()
	defer z.mu.Unlock()

	var identifiers []string
	for id := range z.sinks {
		identifiers = append(identifiers, id)
	}
	return identifiers, nil
}

func (z *ZapLoggerAdapter) Flush() error {
	// Attempt to flush the logger, catching and handling the specific error
	if err := z.logger.Sync(); err != nil {
		// Check if the error is due to an inappropriate ioctl for device, which is common
		// when trying to sync stdout or stderr, and handle it gracefully
		if strings.Contains(err.Error(), "inappropriate ioctl for device") {
			return nil // Ignore this error or handle it as needed
		}
		return err // Return other errors that might be significant
	}
	return nil
}

func (z *ZapLoggerAdapter) Log(level types.LogLevel, msg string, keysAndValues ...interface{}) {
	if z.logger == nil || z.logger.Core() == nil {
		fmt.Println("Logger or logger core is not initialized.")
		return // Or handle the error as appropriate
	}
	zapLevel := ConvertLevel(level)
	if !z.logger.Core().Enabled(zapLevel) {
		return // Skip logging if the level is not enabled.
	}
	fields := make([]zap.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			continue // Skip non-string keys
		}
		value := keysAndValues[i+1]
		// Convert complex types to string or handle them appropriately
		switch v := value.(type) {
		case string, int, int32, int64, float64, bool:
			fields = append(fields, zap.Any(key, v))
		default:
			// Default case to handle types that might cause serialization issues
			fields = append(fields, zap.String(key, fmt.Sprintf("%v", v)))
		}
	}
	z.logger.Check(zapLevel, msg).Write(fields...)
}

// Implement the Logger interface methods for each log level
func (z *ZapLoggerAdapter) Debug(msg string, keysAndValues ...interface{}) {
	z.Log(types.DebugLevel, msg, keysAndValues...)
}

func (z *ZapLoggerAdapter) Info(msg string, keysAndValues ...interface{}) {
	z.Log(types.InfoLevel, msg, keysAndValues...)
}

func (z *ZapLoggerAdapter) Warn(msg string, keysAndValues ...interface{}) {
	z.Log(types.WarnLevel, msg, keysAndValues...)
}

func (z *ZapLoggerAdapter) Error(msg string, keysAndValues ...interface{}) {
	z.Log(types.ErrorLevel, msg, keysAndValues...)
}

func (z *ZapLoggerAdapter) DPanic(msg string, keysAndValues ...interface{}) {
	z.Log(types.DPanicLevel, msg, keysAndValues...)
}

func (z *ZapLoggerAdapter) Panic(msg string, keysAndValues ...interface{}) {
	z.Log(types.PanicLevel, msg, keysAndValues...)
}

func (z *ZapLoggerAdapter) Fatal(msg string, keysAndValues ...interface{}) {
	z.Log(types.FatalLevel, msg, keysAndValues...)
}

func (z *ZapLoggerAdapter) GetLevel() types.LogLevel {
	return types.LogLevel(z.level)
}

func (z *ZapLoggerAdapter) SetLevel(level types.LogLevel) {
	z.level = ConvertLevel(level)
}

// ConvertLevel converts types.LogLevel to zapcore.Level
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

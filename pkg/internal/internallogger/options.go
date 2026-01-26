package internallogger

import (
	"github.com/joeydtaylor/electrician/pkg/logschema"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerWithLevel configures the logger to use the specified log level.
// It converts the level string to zapcore.Level and sets it in the zap config.
func LoggerWithLevel(levelStr string) LoggerOption {
	return func(cfg *zap.Config, lvl *zapcore.Level, callerDepth *int) {
		level := parseLogLevel(levelStr)                 // Convert string level to types.LogLevel
		convertedLevel := ConvertLevel(level)            // Convert types.LogLevel to zapcore.Level
		cfg.Level = zap.NewAtomicLevelAt(convertedLevel) // Set the atomic level
		*lvl = convertedLevel                            // Update the level reference
	}
}

// LoggerWithDevelopment enables or disables development mode in the logger configuration.
func LoggerWithDevelopment(dev bool) LoggerOption {
	return func(cfg *zap.Config, lvl *zapcore.Level, callerDepth *int) {
		cfg.Development = dev // Set the development mode based on the dev parameter
	}
}

// LoggerWithFields attaches fields to every log line.
func LoggerWithFields(fields map[string]interface{}) LoggerOption {
	return func(cfg *zap.Config, lvl *zapcore.Level, callerDepth *int) {
		if cfg.InitialFields == nil {
			cfg.InitialFields = map[string]interface{}{}
		}
		for key, value := range fields {
			if key == "" {
				continue
			}
			cfg.InitialFields[key] = value
		}
	}
}

// LoggerWithSchema overrides the log schema identifier field.
func LoggerWithSchema(schema string) LoggerOption {
	return func(cfg *zap.Config, lvl *zapcore.Level, callerDepth *int) {
		if cfg.InitialFields == nil {
			cfg.InitialFields = map[string]interface{}{}
		}
		cfg.InitialFields[logschema.FieldSchema] = schema
	}
}

// ZapAdapterWithCallerSkip sets the number of caller frames to skip.
func ZapAdapterWithCallerSkip(skip int) LoggerOption {
	return func(cfg *zap.Config, lvl *zapcore.Level, callerDepth *int) {
		*callerDepth += skip // Adjust the existing skip by adding the specified skip
	}
}

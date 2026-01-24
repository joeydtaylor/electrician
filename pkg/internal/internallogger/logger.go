package internallogger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerOption configures the zap logger adapter.
type LoggerOption func(*zap.Config, *zapcore.Level, *int)

// ZapLoggerAdapter implements the internal logger contract with zap.
type ZapLoggerAdapter struct {
	logger      *zap.Logger
	atomicLevel zap.AtomicLevel
	callerDepth int
	mu          sync.Mutex
	sinks       map[string]zapcore.Core
	baseCore    zapcore.Core
}

// NewLogger initializes a ZapLoggerAdapter with optional configuration.
func NewLogger(options ...LoggerOption) *ZapLoggerAdapter {
	config := zap.NewProductionConfig()
	level := config.Level.Level()
	callerDepth := 3

	for _, option := range options {
		if option == nil {
			continue
		}
		option(&config, &level, &callerDepth)
	}

	config.Level = zap.NewAtomicLevelAt(level)
	baseCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(config.EncoderConfig),
		zapcore.AddSync(os.Stdout),
		config.Level,
	)

	logger := zap.New(zapcore.NewTee(baseCore), zap.AddCaller(), zap.AddCallerSkip(callerDepth))

	return &ZapLoggerAdapter{
		logger:      logger,
		atomicLevel: config.Level,
		callerDepth: callerDepth,
		sinks:       make(map[string]zapcore.Core),
		baseCore:    baseCore,
	}
}

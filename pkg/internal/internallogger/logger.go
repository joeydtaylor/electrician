package internallogger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerOption func(*zap.Config, *zapcore.Level, *int) // Updated to include caller skip management

type ZapLoggerAdapter struct {
	logger       *zap.Logger
	level        zapcore.Level
	callerDepth  int
	mu           sync.Mutex
	sinks        map[string]zapcore.Core
	combinedCore zapcore.Core
}

// NewLogger initializes a new ZapLoggerAdapter with configurable options.
func NewLogger(options ...LoggerOption) *ZapLoggerAdapter {
	config := zap.NewProductionConfig()
	var level zapcore.Level
	var callerDepth int = 3 // Default caller depth

	// Apply each provided option to the configuration
	for _, option := range options {
		option(&config, &level, &callerDepth)
	}

	// Ensure at least one core is created to prevent nil logger
	defaultCore := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(os.Stdout), zap.NewAtomicLevelAt(zapcore.InfoLevel))
	cores := []zapcore.Core{defaultCore} // Start with a default core

	logger := zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddCallerSkip(callerDepth))

	return &ZapLoggerAdapter{
		logger:       logger,
		level:        level,
		callerDepth:  callerDepth,
		sinks:        make(map[string]zapcore.Core), // Initialize sinks map
		combinedCore: zapcore.NewTee(cores...),      // Combine cores
	}
}

package internallogger

import (
	"os"
	"sync"

	"github.com/joeydtaylor/electrician/pkg/logschema"
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
	callerOn    bool
	mu          sync.Mutex
	sinks       map[string]sinkEntry
	baseCore    zapcore.Core
	baseFields  []zap.Field
	encConfig   zapcore.EncoderConfig
}

// NewLogger initializes a ZapLoggerAdapter with optional configuration.
func NewLogger(options ...LoggerOption) *ZapLoggerAdapter {
	config := zap.NewProductionConfig()
	config.EncoderConfig = standardEncoderConfig()
	level := config.Level.Level()
	callerDepth := 3

	for _, option := range options {
		if option == nil {
			continue
		}
		option(&config, &level, &callerDepth)
	}

	if config.InitialFields == nil {
		config.InitialFields = map[string]interface{}{}
	}
	if _, ok := config.InitialFields[logschema.FieldSchema]; !ok {
		config.InitialFields[logschema.FieldSchema] = logschema.SchemaID
	}

	config.Level = zap.NewAtomicLevelAt(level)

	baseCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(config.EncoderConfig),
		zapcore.AddSync(os.Stdout),
		config.Level,
	)

	baseFields := fieldsFromMap(config.InitialFields)
	opts := []zap.Option{zap.AddCallerSkip(callerDepth)}
	callerOn := !config.DisableCaller
	if callerOn {
		opts = append(opts, zap.AddCaller())
	}
	logger := zap.New(zapcore.NewTee(baseCore), opts...)
	if len(baseFields) > 0 {
		logger = logger.With(baseFields...)
	}

	return &ZapLoggerAdapter{
		logger:      logger,
		atomicLevel: config.Level,
		callerDepth: callerDepth,
		callerOn:    callerOn,
		sinks:       make(map[string]sinkEntry),
		baseCore:    baseCore,
		baseFields:  baseFields,
		encConfig:   config.EncoderConfig,
	}
}

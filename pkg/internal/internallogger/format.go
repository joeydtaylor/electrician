package internallogger

import (
	"time"

	"github.com/joeydtaylor/electrician/pkg/logschema"
	"go.uber.org/zap/zapcore"
)

func standardEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        logschema.FieldTimestamp,
		LevelKey:       logschema.FieldLevel,
		NameKey:        logschema.FieldLogger,
		CallerKey:      logschema.FieldCaller,
		MessageKey:     logschema.FieldMessage,
		StacktraceKey:  logschema.FieldStack,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     encodeRFC3339NanoUTC,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func encodeRFC3339NanoUTC(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.UTC().Format(time.RFC3339Nano))
}

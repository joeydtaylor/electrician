package internallogger

import (
	"testing"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestLog_WritesFields(t *testing.T) {
	logger := NewLogger()
	core, obs := observer.New(zapcore.DebugLevel)

	logger.mu.Lock()
	logger.logger = zap.New(core)
	logger.mu.Unlock()

	logger.Log(types.InfoLevel, "msg", "a", "b", "c", 3, "orphan")

	entries := obs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	fields := entries[0].Context
	if len(fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(fields))
	}

	if fields[0].Key != "a" || fields[1].Key != "c" {
		t.Fatalf("unexpected field keys: %v, %v", fields[0].Key, fields[1].Key)
	}
}

func TestLog_IgnoresNonStringKeys(t *testing.T) {
	logger := NewLogger()
	core, obs := observer.New(zapcore.DebugLevel)

	logger.mu.Lock()
	logger.logger = zap.New(core)
	logger.mu.Unlock()

	logger.Log(types.InfoLevel, "msg", 123, "skip", "k", "v")

	entries := obs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	fields := entries[0].Context
	if len(fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(fields))
	}
	if fields[0].Key != "k" {
		t.Fatalf("expected field key 'k', got %q", fields[0].Key)
	}
}

func TestLog_RespectsCoreLevel(t *testing.T) {
	logger := NewLogger()
	core, obs := observer.New(zapcore.WarnLevel)

	logger.mu.Lock()
	logger.logger = zap.New(core)
	logger.mu.Unlock()

	logger.Log(types.InfoLevel, "info")
	logger.Log(types.WarnLevel, "warn")

	entries := obs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Entry.Level != zapcore.WarnLevel {
		t.Fatalf("expected warn entry, got %v", entries[0].Entry.Level)
	}
}

func TestLog_NilLoggerNoPanic(t *testing.T) {
	logger := NewLogger()
	logger.mu.Lock()
	logger.logger = nil
	logger.mu.Unlock()

	logger.Log(types.InfoLevel, "msg")
}

func TestFlush_NilLogger(t *testing.T) {
	logger := NewLogger()
	logger.mu.Lock()
	logger.logger = nil
	logger.mu.Unlock()

	if err := logger.Flush(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestConvertLevel_Defaults(t *testing.T) {
	if got := ConvertLevel(types.LogLevel(99)); got != zapcore.InfoLevel {
		t.Fatalf("expected default zapcore.InfoLevel, got %v", got)
	}
	if got := convertZapLevel(zapcore.Level(99)); got != types.InfoLevel {
		t.Fatalf("expected default types.InfoLevel, got %v", got)
	}
}

func TestParseLogLevel(t *testing.T) {
	cases := map[string]types.LogLevel{
		"debug":  types.DebugLevel,
		"info":   types.InfoLevel,
		"warn":   types.WarnLevel,
		"error":  types.ErrorLevel,
		"dpanic": types.DPanicLevel,
		"panic":  types.PanicLevel,
		"fatal":  types.FatalLevel,
		"bogus":  types.InfoLevel,
	}

	for input, expect := range cases {
		if got := parseLogLevel(input); got != expect {
			t.Fatalf("parseLogLevel(%q) = %v, expected %v", input, got, expect)
		}
	}
}

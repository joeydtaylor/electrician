package internallogger_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joeydtaylor/electrician/pkg/internal/internallogger"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func TestNewLogger_DefaultLevel(t *testing.T) {
	logger := internallogger.NewLogger()
	if got := logger.GetLevel(); got != types.InfoLevel {
		t.Fatalf("expected InfoLevel, got %v", got)
	}
}

func TestNewLogger_WithLevel(t *testing.T) {
	logger := internallogger.NewLogger(internallogger.LoggerWithLevel("debug"))
	if got := logger.GetLevel(); got != types.DebugLevel {
		t.Fatalf("expected DebugLevel, got %v", got)
	}

	logger = internallogger.NewLogger(internallogger.LoggerWithLevel("unknown"))
	if got := logger.GetLevel(); got != types.InfoLevel {
		t.Fatalf("expected InfoLevel on unknown level, got %v", got)
	}
}

func TestLogger_SetLevel(t *testing.T) {
	logger := internallogger.NewLogger()
	logger.SetLevel(types.ErrorLevel)
	if got := logger.GetLevel(); got != types.ErrorLevel {
		t.Fatalf("expected ErrorLevel, got %v", got)
	}
}

func TestLogger_AddRemoveListSinks(t *testing.T) {
	logger := internallogger.NewLogger(internallogger.LoggerWithLevel("debug"))

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "app.log")

	if err := logger.AddSink("file", types.SinkConfig{Type: "file", Config: map[string]interface{}{"path": path}}); err != nil {
		t.Fatalf("AddSink(file) error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected log file to exist: %v", err)
	}

	if err := logger.AddSink("stdout", types.SinkConfig{Type: "stdout"}); err != nil {
		t.Fatalf("AddSink(stdout) error: %v", err)
	}

	sinks, err := logger.ListSinks()
	if err != nil {
		t.Fatalf("ListSinks error: %v", err)
	}
	if len(sinks) != 2 {
		t.Fatalf("expected 2 sinks, got %d", len(sinks))
	}

	if err := logger.RemoveSink("stdout"); err != nil {
		t.Fatalf("RemoveSink error: %v", err)
	}
	if err := logger.RemoveSink("missing"); err == nil {
		t.Fatalf("expected error removing missing sink")
	}
}

func TestLogger_AddSinkInvalidConfig(t *testing.T) {
	logger := internallogger.NewLogger()

	if err := logger.AddSink("file", types.SinkConfig{Type: "file", Config: map[string]interface{}{}}); err == nil {
		t.Fatalf("expected error for missing file path")
	}
	if err := logger.AddSink("network", types.SinkConfig{Type: "network"}); err == nil {
		t.Fatalf("expected error for unsupported sink type")
	}
}

func TestLogger_LogHandlesOddKeys(t *testing.T) {
	logger := internallogger.NewLogger(internallogger.LoggerWithLevel("debug"))

	logger.Log(types.InfoLevel, "odd keys", "key", "value", "orphan")
	logger.Log(types.InfoLevel, "non-string key", 123, "value")
}

func TestLogger_Flush(t *testing.T) {
	logger := internallogger.NewLogger()
	if err := logger.Flush(); err != nil {
		t.Fatalf("Flush error: %v", err)
	}
}

func TestLogger_OptionsCoverage(t *testing.T) {
	logger := internallogger.NewLogger(
		internallogger.LoggerWithDevelopment(true),
		internallogger.ZapAdapterWithCallerSkip(1),
	)
	logger.Info("options")
}

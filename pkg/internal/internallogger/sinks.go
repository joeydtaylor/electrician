package internallogger

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type sinkEntry struct {
	core zapcore.Core
	stop func()
}

// AddSink adds a sink based on its identifier and config.
func (z *ZapLoggerAdapter) AddSink(identifier string, config types.SinkConfig) error {
	z.mu.Lock()
	defer z.mu.Unlock()

	var (
		ws   zapcore.WriteSyncer
		stop func()
	)
	cfg := config.Config
	if cfg == nil {
		cfg = map[string]interface{}{}
	}

	switch config.Type {
	case "file":
		path, ok := cfg["path"].(string)
		if !ok || path == "" {
			return fmt.Errorf("file path configuration is missing or invalid")
		}
		dir := filepath.Dir(path)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", dir, err)
			}
		}
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %v", path, err)
		}
		stop = func() {
			_ = file.Close()
		}
		ws = zapcore.AddSync(file)
	case "stdout":
		ws = zapcore.Lock(os.Stdout)
	case "network":
		relaySink, err := newRelayWriteSyncer(config)
		if err != nil {
			return err
		}
		ws = relaySink
		stop = relaySink.Close
	case "relay":
		relaySink, err := newRelayWriteSyncer(config)
		if err != nil {
			return err
		}
		ws = relaySink
		stop = relaySink.Close
	default:
		return fmt.Errorf("unsupported sink type: %s", config.Type)
	}

	core := zapcore.NewCore(zapcore.NewJSONEncoder(z.encConfig), ws, z.atomicLevel)
	z.sinks[identifier] = sinkEntry{core: core, stop: stop}

	z.rebuildLoggerLocked()
	return nil
}

// RemoveSink removes a sink based on its identifier.
func (z *ZapLoggerAdapter) RemoveSink(identifier string) error {
	z.mu.Lock()
	defer z.mu.Unlock()

	entry, ok := z.sinks[identifier]
	if !ok {
		return fmt.Errorf("sink not found: %s", identifier)
	}
	delete(z.sinks, identifier)
	if entry.stop != nil {
		entry.stop()
	}

	z.rebuildLoggerLocked()
	return nil
}

// ListSinks lists all configured sinks.
func (z *ZapLoggerAdapter) ListSinks() ([]string, error) {
	z.mu.Lock()
	defer z.mu.Unlock()

	identifiers := make([]string, 0, len(z.sinks))
	for id := range z.sinks {
		identifiers = append(identifiers, id)
	}
	return identifiers, nil
}

func (z *ZapLoggerAdapter) rebuildLoggerLocked() {
	cores := make([]zapcore.Core, 0, 1+len(z.sinks))
	cores = append(cores, z.baseCore)
	for _, entry := range z.sinks {
		cores = append(cores, entry.core)
	}
	combined := zapcore.NewTee(cores...)
	opts := []zap.Option{zap.AddCallerSkip(z.callerDepth)}
	if z.callerOn {
		opts = append(opts, zap.AddCaller())
	}
	logger := zap.New(combined, opts...)
	if len(z.baseFields) > 0 {
		logger = logger.With(z.baseFields...)
	}
	z.logger = logger
}

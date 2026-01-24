package internallogger

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// AddSink adds a sink based on its identifier and config.
func (z *ZapLoggerAdapter) AddSink(identifier string, config types.SinkConfig) error {
	z.mu.Lock()
	defer z.mu.Unlock()

	var ws zapcore.WriteSyncer

	switch config.Type {
	case "file":
		path, ok := config.Config["path"].(string)
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
		ws = zapcore.AddSync(file)
	case "stdout":
		ws = zapcore.Lock(os.Stdout)
	case "network":
		return fmt.Errorf("unsupported sink type: %s", config.Type)
	default:
		return fmt.Errorf("unsupported sink type: %s", config.Type)
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), ws, z.atomicLevel)
	z.sinks[identifier] = core

	z.rebuildLoggerLocked()
	return nil
}

// RemoveSink removes a sink based on its identifier.
func (z *ZapLoggerAdapter) RemoveSink(identifier string) error {
	z.mu.Lock()
	defer z.mu.Unlock()

	if _, ok := z.sinks[identifier]; !ok {
		return fmt.Errorf("sink not found: %s", identifier)
	}
	delete(z.sinks, identifier)

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
	for _, core := range z.sinks {
		cores = append(cores, core)
	}
	combined := zapcore.NewTee(cores...)
	z.logger = zap.New(combined, zap.AddCaller(), zap.AddCallerSkip(z.callerDepth))
}

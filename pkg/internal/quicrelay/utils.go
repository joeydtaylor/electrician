package quicrelay

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func normalizeTargets(parts []string) []string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func (fr *ForwardRelay[T]) requireNotFrozen(action string) {
	if atomic.LoadInt32(&fr.configFrozen) == 1 {
		panic(fmt.Sprintf("%s: config frozen", action))
	}
}

func (rr *ReceivingRelay[T]) requireNotFrozen(action string) {
	if atomic.LoadInt32(&rr.configFrozen) == 1 {
		panic(fmt.Sprintf("%s: config frozen", action))
	}
}

func (fr *ForwardRelay[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	loggers := fr.snapshotLoggers()
	if len(loggers) == 0 {
		return
	}
	msg := fmt.Sprintf(format, args...)
	for _, logger := range loggers {
		if logger == nil || logger.GetLevel() > level {
			continue
		}
		switch level {
		case types.DebugLevel:
			logger.Debug(msg)
		case types.InfoLevel:
			logger.Info(msg)
		case types.WarnLevel:
			logger.Warn(msg)
		case types.ErrorLevel:
			logger.Error(msg)
		case types.DPanicLevel:
			logger.DPanic(msg)
		case types.PanicLevel:
			logger.Panic(msg)
		case types.FatalLevel:
			logger.Fatal(msg)
		}
	}
}

func (rr *ReceivingRelay[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	loggers := rr.snapshotLoggers()
	if len(loggers) == 0 {
		return
	}
	msg := fmt.Sprintf(format, args...)
	for _, logger := range loggers {
		if logger == nil || logger.GetLevel() > level {
			continue
		}
		switch level {
		case types.DebugLevel:
			logger.Debug(msg)
		case types.InfoLevel:
			logger.Info(msg)
		case types.WarnLevel:
			logger.Warn(msg)
		case types.ErrorLevel:
			logger.Error(msg)
		case types.DPanicLevel:
			logger.DPanic(msg)
		case types.PanicLevel:
			logger.Panic(msg)
		case types.FatalLevel:
			logger.Fatal(msg)
		}
	}
}

func (fr *ForwardRelay[T]) snapshotLoggers() []types.Logger {
	fr.loggersLock.Lock()
	defer fr.loggersLock.Unlock()
	if len(fr.Loggers) == 0 {
		return nil
	}
	loggers := make([]types.Logger, len(fr.Loggers))
	copy(loggers, fr.Loggers)
	return loggers
}

func (rr *ReceivingRelay[T]) snapshotLoggers() []types.Logger {
	rr.loggersLock.Lock()
	defer rr.loggersLock.Unlock()
	if len(rr.Loggers) == 0 {
		return nil
	}
	loggers := make([]types.Logger, len(rr.Loggers))
	copy(loggers, rr.Loggers)
	return loggers
}

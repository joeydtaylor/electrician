package httpclient

import (
	"fmt"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// NotifyLoggers sends a formatted message to all attached loggers.
func (hp *HTTPClientAdapter[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	loggers := hp.snapshotLoggers()
	if len(loggers) == 0 {
		return
	}

	msg := fmt.Sprintf(format, args...)
	for _, logger := range loggers {
		if logger == nil {
			continue
		}
		if logger.GetLevel() > level {
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

func (hp *HTTPClientAdapter[T]) notifyHTTPClientRequestStart() {
	metadata := hp.GetComponentMetadata()
	for _, sensor := range hp.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnHTTPClientRequestStart(metadata)
	}
}

func (hp *HTTPClientAdapter[T]) notifyHTTPClientError(err error) {
	metadata := hp.GetComponentMetadata()
	for _, sensor := range hp.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnHTTPClientError(metadata, err)
	}
}

func (hp *HTTPClientAdapter[T]) notifyHTTPClientResponseReceived() {
	metadata := hp.GetComponentMetadata()
	for _, sensor := range hp.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnHTTPClientResponseReceived(metadata)
	}
}

func (hp *HTTPClientAdapter[T]) notifyHTTPClientRequestComplete() {
	metadata := hp.GetComponentMetadata()
	for _, sensor := range hp.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnHTTPClientRequestComplete(metadata)
	}
}

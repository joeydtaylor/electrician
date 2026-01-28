package httpserver

import "github.com/joeydtaylor/electrician/pkg/internal/types"

func (h *httpServerAdapter[T]) notifyHTTPServerError(err error) {
	for _, sensor := range h.snapshotSensors() {
		sensor.InvokeOnHTTPClientError(h.componentMetadata, err)
	}

	h.NotifyLoggers(
		types.ErrorLevel,
		"HTTP server error",
		"component", h.componentMetadata,
		"event", "HTTPServerError",
		"error", err,
	)
}

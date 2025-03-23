// File: internal.go
package httpserver

import (
	"net/http"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// -----------------------------------------------------------------------------
// Helper Methods
// -----------------------------------------------------------------------------

// parseRequest reads and decodes the request as JSON (simple placeholder).
func (h *httpServerAdapter[T]) parseRequest(r *http.Request) (*types.WrappedResponse[T], error) {
	var decoded T
	err := utils.DecodeJSON(r.Body, &decoded)
	if err != nil {
		return nil, err
	}
	return types.NewWrappedResponse(decoded, nil, nil, r.Header), nil
}

// notifyHTTPServerError notifies sensors & logs an error event.
func (h *httpServerAdapter[T]) notifyHTTPServerError(err error) {
	h.sensorsLock.Lock()
	defer h.sensorsLock.Unlock()

	for _, sensor := range h.sensors {
		// Optionally define a dedicated OnHTTPServerError method in sensors
		// but we can reuse OnHTTPClientError for the example
		sensor.InvokeOnHTTPClientError(h.componentMetadata, err)
	}

	h.NotifyLoggers(types.ErrorLevel,
		"%s => level: ERROR, event: notifyHTTPServerError, error: %v => Notified sensors",
		h.componentMetadata, err)
}

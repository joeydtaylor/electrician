// api.go file
package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// Serve starts listening for incoming HTTP requests on the configured address/endpoint.
// For each request matching the specified method, it decodes into T, then calls submitFunc.
func (h *httpServerAdapter[T]) Serve(ctx context.Context, submitFunc func(ctx context.Context, req T) error) error {
	h.serverMu.Lock()
	defer h.serverMu.Unlock()

	mux := http.NewServeMux()

	// Simple approach: one method & endpoint
	mux.HandleFunc(h.endpoint, func(w http.ResponseWriter, r *http.Request) {
		// Check method
		if r.Method != h.method {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Decode request body into T (placeholder logic)
		var parsedData T
		wrappedResp, err := h.parseRequest(r)
		if err != nil {
			h.notifyHTTPServerError(err)
			http.Error(w, fmt.Sprintf("Error parsing request: %v", err), http.StatusBadRequest)
			return
		}
		parsedData = wrappedResp.Data

		// Feed the data into the pipeline
		if err := submitFunc(ctx, parsedData); err != nil {
			h.notifyHTTPServerError(err)
			http.Error(w, fmt.Sprintf("Pipeline error: %v", err), http.StatusInternalServerError)
			return
		}

		// Optional default response headers
		for key, val := range h.headers {
			w.Header().Set(key, val)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Success")
	})

	// Build our server
	h.server = &http.Server{
		Addr:         h.address,
		Handler:      mux,
		ReadTimeout:  h.timeout,
		WriteTimeout: h.timeout,
		TLSConfig:    h.tlsConfig,
	}

	// Start the server in a separate goroutine so Serve() can monitor ctx.Done()
	errChan := make(chan error, 1)
	go func() {
		var err error
		if h.tlsConfig != nil {
			h.NotifyLoggers(types.InfoLevel,
				"%s => level: INFO, event: Serve, message: Starting HTTPS server on %s %s",
				h.componentMetadata, h.address, h.endpoint)
			// We do not pass certPath, keyPath here; they're in h.tlsConfig.Certificates
			err = h.server.ListenAndServeTLS("", "")
		} else {
			h.NotifyLoggers(types.InfoLevel,
				"%s => level: INFO, event: Serve, message: Starting HTTP server on %s %s",
				h.componentMetadata, h.address, h.endpoint)
			err = h.server.ListenAndServe()
		}
		errChan <- err
	}()

	// Wait until context is canceled or server errors out
	select {
	case <-ctx.Done():
		h.NotifyLoggers(types.WarnLevel,
			"%s => level: WARN, event: Serve, message: Context canceled; shutting down server.",
			h.componentMetadata)
		_ = h.server.Close()
		return ctx.Err()
	case err := <-errChan:
		if err != nil && err != http.ErrServerClosed {
			h.NotifyLoggers(types.ErrorLevel,
				"%s => level: ERROR, event: Serve, message: Server error => %v",
				h.componentMetadata, err)
			return err
		}
		return nil
	}
}

// ConnectLogger attaches logger(s).
func (h *httpServerAdapter[T]) ConnectLogger(loggers ...types.Logger) {
	h.loggersLock.Lock()
	defer h.loggersLock.Unlock()
	h.loggers = append(h.loggers, loggers...)
}

// ConnectSensor attaches sensor(s).
func (h *httpServerAdapter[T]) ConnectSensor(sensors ...types.Sensor[T]) {
	h.sensorsLock.Lock()
	defer h.sensorsLock.Unlock()
	h.sensors = append(h.sensors, sensors...)
}

// SetAddress configures the listen address.
func (h *httpServerAdapter[T]) SetAddress(address string) {
	h.address = address
}

// SetServerConfig sets the HTTP method and endpoint (e.g. POST, /webhook).
func (h *httpServerAdapter[T]) SetServerConfig(method, endpoint string) {
	h.method = method
	h.endpoint = endpoint
}

// AddHeader adds default response headers.
func (h *httpServerAdapter[T]) AddHeader(key, value string) {
	h.headers[key] = value
}

// GetComponentMetadata returns metadata (ID, Name, Type).
func (h *httpServerAdapter[T]) GetComponentMetadata() types.ComponentMetadata {
	return h.componentMetadata
}

// SetComponentMetadata sets Name and ID.
func (h *httpServerAdapter[T]) SetComponentMetadata(name string, id string) {
	h.componentMetadata.Name = name
	h.componentMetadata.ID = id
}

// NotifyLoggers logs a formatted message to all attached loggers.
func (h *httpServerAdapter[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	h.loggersLock.Lock()
	defer h.loggersLock.Unlock()
	for _, logger := range h.loggers {
		if logger == nil {
			continue
		}
		if logger.GetLevel() <= level {
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
}

// SetTimeout sets read/write timeouts.
func (h *httpServerAdapter[T]) SetTimeout(timeout time.Duration) {
	h.timeout = timeout
}

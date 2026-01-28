package httpserver

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func (h *httpServerAdapter[T]) authorizeRequest(ctx context.Context, r *http.Request, cfg serverConfig) error {
	if len(cfg.staticHeaders) == 0 && cfg.authValidator == nil {
		return nil
	}

	headers := collectHeaders(r)

	if err := checkStaticHeaders(headers, cfg.staticHeaders); err != nil {
		return h.applyAuthPolicy(err, cfg.authRequired)
	}

	if cfg.authValidator != nil {
		if err := cfg.authValidator(ctx, headers); err != nil {
			return h.applyAuthPolicy(err, cfg.authRequired)
		}
	}

	return nil
}

func (h *httpServerAdapter[T]) applyAuthPolicy(err error, required bool) error {
	if err == nil {
		return nil
	}
	if required {
		return err
	}
	h.NotifyLoggers(
		types.WarnLevel,
		"Auth: soft-failing policy error",
		"component", h.componentMetadata,
		"event", "AuthPolicy",
		"error", err,
	)
	return nil
}

func checkStaticHeaders(headers map[string]string, required map[string]string) error {
	for key, expected := range required {
		lk := strings.ToLower(key)
		got, ok := headers[lk]
		if !ok || got != expected {
			return errors.New("missing/invalid header " + key)
		}
	}
	return nil
}

func collectHeaders(r *http.Request) map[string]string {
	out := make(map[string]string)
	for key, vals := range r.Header {
		if len(vals) == 0 {
			continue
		}
		lk := strings.ToLower(key)
		out[lk] = vals[0]
	}
	return out
}

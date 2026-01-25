package websocket

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func (s *serverAdapter[T]) authorizeRequest(ctx context.Context, r *http.Request, cfg serverConfig) error {
	if len(cfg.staticHeaders) == 0 && cfg.authValidator == nil {
		return nil
	}

	headers := collectHeaders(r, cfg.tokenQueryParam)

	if err := checkStaticHeaders(headers, cfg.staticHeaders); err != nil {
		return s.applyAuthPolicy(err, cfg.authRequired)
	}

	if cfg.authValidator != nil {
		if err := cfg.authValidator(ctx, headers); err != nil {
			return s.applyAuthPolicy(err, cfg.authRequired)
		}
	}

	return nil
}

func (s *serverAdapter[T]) applyAuthPolicy(err error, required bool) error {
	if err == nil {
		return nil
	}
	if required {
		return err
	}
	s.NotifyLoggers(types.WarnLevel, "Auth: soft-failing policy error: %v", err)
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

func collectHeaders(r *http.Request, tokenQueryParam string) map[string]string {
	out := make(map[string]string)
	for key, vals := range r.Header {
		if len(vals) == 0 {
			continue
		}
		lk := strings.ToLower(key)
		out[lk] = vals[0]
	}

	if tokenQueryParam != "" {
		token := strings.TrimSpace(r.URL.Query().Get(tokenQueryParam))
		if token != "" {
			if _, ok := out["authorization"]; !ok {
				lt := strings.ToLower(token)
				if strings.HasPrefix(lt, "bearer ") {
					out["authorization"] = token
				} else {
					out["authorization"] = "Bearer " + token
				}
			}
		}
	}

	return out
}

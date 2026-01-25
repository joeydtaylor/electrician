package httpserver

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
)

// SetAuthenticationOptions configures OAuth2 or other authentication options.
func (h *httpServerAdapter[T]) SetAuthenticationOptions(opts *relay.AuthenticationOptions) {
	h.requireNotFrozen("SetAuthenticationOptions")
	h.configLock.Lock()
	h.authOptions = cloneAuthOptions(opts)
	h.configLock.Unlock()
}

// SetStaticHeaders enforces constant header key/value pairs on incoming requests.
func (h *httpServerAdapter[T]) SetStaticHeaders(headers map[string]string) {
	h.requireNotFrozen("SetStaticHeaders")
	h.configLock.Lock()
	h.staticHeaders = cloneHeaderMap(headers)
	h.configLock.Unlock()
}

// SetDynamicAuthValidator registers a per-request validation callback.
func (h *httpServerAdapter[T]) SetDynamicAuthValidator(fn func(ctx context.Context, headers map[string]string) error) {
	h.requireNotFrozen("SetDynamicAuthValidator")
	h.configLock.Lock()
	h.dynamicAuthValidator = fn
	h.configLock.Unlock()
}

// SetAuthRequired toggles strict auth enforcement.
func (h *httpServerAdapter[T]) SetAuthRequired(required bool) {
	h.requireNotFrozen("SetAuthRequired")
	h.configLock.Lock()
	h.authRequired = required
	h.configLock.Unlock()
}

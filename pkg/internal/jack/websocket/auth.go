package websocket

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
)

// SetAuthenticationOptions configures OAuth2 or other authentication options.
func (s *serverAdapter[T]) SetAuthenticationOptions(opts *relay.AuthenticationOptions) {
	s.requireNotFrozen("SetAuthenticationOptions")
	s.configLock.Lock()
	s.authOptions = cloneAuthOptions(opts)
	s.configLock.Unlock()
}

// SetStaticHeaders enforces constant header key/value pairs on incoming requests.
func (s *serverAdapter[T]) SetStaticHeaders(headers map[string]string) {
	s.requireNotFrozen("SetStaticHeaders")
	s.configLock.Lock()
	s.staticHeaders = cloneHeaderMap(headers)
	s.configLock.Unlock()
}

// SetDynamicAuthValidator registers a per-request validation callback.
func (s *serverAdapter[T]) SetDynamicAuthValidator(fn func(ctx context.Context, headers map[string]string) error) {
	s.requireNotFrozen("SetDynamicAuthValidator")
	s.configLock.Lock()
	s.dynamicAuthValidator = fn
	s.configLock.Unlock()
}

// SetAuthRequired toggles strict auth enforcement.
func (s *serverAdapter[T]) SetAuthRequired(required bool) {
	s.requireNotFrozen("SetAuthRequired")
	s.configLock.Lock()
	s.authRequired = required
	s.configLock.Unlock()
}

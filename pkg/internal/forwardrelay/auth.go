package forwardrelay

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// SetAuthenticationOptions attaches auth hints for receivers.
func (fr *ForwardRelay[T]) SetAuthenticationOptions(opts *relay.AuthenticationOptions) {
	fr.requireNotFrozen("SetAuthenticationOptions")
	fr.authOptions = opts
	fr.NotifyLoggers(types.DebugLevel, "SetAuthenticationOptions: %+v", opts)
}

// SetOAuth2 configures a per-RPC token source.
func (fr *ForwardRelay[T]) SetOAuth2(ts types.OAuth2TokenSource) {
	fr.requireNotFrozen("SetOAuth2")
	fr.tokenSource = ts
	if ts == nil {
		fr.NotifyLoggers(types.DebugLevel, "SetOAuth2: disabled")
		return
	}
	fr.NotifyLoggers(types.DebugLevel, "SetOAuth2: enabled")
}

// SetStaticHeaders sets constant gRPC metadata headers.
func (fr *ForwardRelay[T]) SetStaticHeaders(headers map[string]string) {
	fr.requireNotFrozen("SetStaticHeaders")
	fr.staticHeaders = make(map[string]string, len(headers))
	for k, v := range headers {
		fr.staticHeaders[k] = v
	}
	fr.NotifyLoggers(types.DebugLevel, "SetStaticHeaders: keys=%d", len(fr.staticHeaders))
}

// SetDynamicHeaders registers a per-request header callback.
func (fr *ForwardRelay[T]) SetDynamicHeaders(fn func(ctx context.Context) map[string]string) {
	fr.requireNotFrozen("SetDynamicHeaders")
	fr.dynamicHeaders = fn
	fr.NotifyLoggers(types.DebugLevel, "SetDynamicHeaders: installed=%t", fn != nil)
}

// SetAuthRequired enforces that OAuth2 tokens are present before RPCs.
func (fr *ForwardRelay[T]) SetAuthRequired(required bool) {
	fr.requireNotFrozen("SetAuthRequired")
	fr.authRequired = required
	fr.NotifyLoggers(types.DebugLevel, "SetAuthRequired: %t", required)
}

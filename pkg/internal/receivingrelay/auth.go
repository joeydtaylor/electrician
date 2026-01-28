package receivingrelay

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"google.golang.org/grpc"
)

// SetAuthenticationOptions configures expected auth mode and parameters.
func (rr *ReceivingRelay[T]) SetAuthenticationOptions(opts *relay.AuthenticationOptions) {
	rr.requireNotFrozen("SetAuthenticationOptions")
	rr.authOptions = opts
	rr.logKV(
		types.DebugLevel,
		"Authentication options updated",
		"event", "SetAuthenticationOptions",
		"result", "SUCCESS",
		"auth_options", opts,
	)
}

// SetAuthInterceptor installs a unary auth interceptor.
func (rr *ReceivingRelay[T]) SetAuthInterceptor(interceptor grpc.UnaryServerInterceptor) {
	rr.requireNotFrozen("SetAuthInterceptor")
	rr.authUnary = interceptor
	rr.logKV(
		types.DebugLevel,
		"Auth interceptor updated",
		"event", "SetAuthInterceptor",
		"result", "SUCCESS",
		"installed", interceptor != nil,
	)
}

// SetAuthInterceptors installs unary and stream auth interceptors.
func (rr *ReceivingRelay[T]) SetAuthInterceptors(unary grpc.UnaryServerInterceptor, stream grpc.StreamServerInterceptor) {
	rr.requireNotFrozen("SetAuthInterceptors")
	rr.authUnary = unary
	rr.authStream = stream
	rr.logKV(
		types.DebugLevel,
		"Auth interceptors updated",
		"event", "SetAuthInterceptors",
		"result", "SUCCESS",
		"unary_installed", unary != nil,
		"stream_installed", stream != nil,
	)
}

// SetStaticHeaders defines required metadata keys/values for incoming requests.
func (rr *ReceivingRelay[T]) SetStaticHeaders(h map[string]string) {
	rr.requireNotFrozen("SetStaticHeaders")
	rr.staticHeaders = make(map[string]string, len(h))
	for k, v := range h {
		rr.staticHeaders[k] = v
	}
	rr.logKV(
		types.DebugLevel,
		"Static headers updated",
		"event", "SetStaticHeaders",
		"result", "SUCCESS",
		"keys", len(rr.staticHeaders),
	)
}

// SetDynamicAuthValidator registers a per-request validator.
func (rr *ReceivingRelay[T]) SetDynamicAuthValidator(fn func(ctx context.Context, md map[string]string) error) {
	rr.requireNotFrozen("SetDynamicAuthValidator")
	rr.dynamicAuthValidator = fn
	rr.logKV(
		types.DebugLevel,
		"Dynamic auth validator updated",
		"event", "SetDynamicAuthValidator",
		"result", "SUCCESS",
		"installed", fn != nil,
	)
}

// SetAuthRequired toggles strict enforcement of authentication.
func (rr *ReceivingRelay[T]) SetAuthRequired(required bool) {
	rr.requireNotFrozen("SetAuthRequired")
	rr.authRequired = required
	rr.logKV(
		types.DebugLevel,
		"Auth requirement updated",
		"event", "SetAuthRequired",
		"result", "SUCCESS",
		"required", required,
	)
}

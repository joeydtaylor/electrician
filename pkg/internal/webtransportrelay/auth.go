//go:build webtransport

package webtransportrelay

import (
	"context"
	"strings"

	"github.com/joeydtaylor/electrician/pkg/internal/auth"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func (rr *ReceivingRelay[T]) ensureDefaultAuthValidator() {
	if rr.dynamicAuthValidator != nil {
		return
	}
	if rr.authOptions == nil || rr.authOptions.Mode != relay.AuthMode_AUTH_OAUTH2 {
		return
	}
	o := rr.authOptions.GetOauth2()
	if o == nil {
		return
	}

	var jwtValidator *auth.JWTValidator
	if o.GetAcceptJwt() && o.GetJwksUri() != "" {
		jwtValidator = auth.NewJWTValidator(o)
		if jwtValidator != nil {
			rr.logKV(types.InfoLevel, "Auth JWT validator installed",
				"event", "AuthSetup",
				"result", "SUCCESS",
				"issuer", o.GetIssuer(),
				"jwks_url", o.GetJwksUri(),
				"required_audience", o.GetRequiredAudience(),
				"required_scopes", o.GetRequiredScopes(),
			)
		}
	}

	var introspectionValidator *cachingIntrospectionValidator
	if o.GetAcceptIntrospection() && o.GetIntrospectionUrl() != "" {
		introspectionValidator = newCachingIntrospectionValidator(o)
		rr.logKV(types.InfoLevel, "Auth introspection validator installed",
			"event", "AuthSetup",
			"result", "SUCCESS",
			"introspection_url", o.GetIntrospectionUrl(),
			"auth_type", o.GetIntrospectionAuthType(),
			"required_scopes", o.GetRequiredScopes(),
		)
	}

	if jwtValidator == nil && introspectionValidator == nil {
		return
	}

	rr.dynamicAuthValidator = func(ctx context.Context, md map[string]string) error {
		var token string
		if v, ok := md["authorization"]; ok && strings.HasPrefix(strings.ToLower(v), "bearer ") {
			token = strings.TrimSpace(v[len("bearer "):])
		}
		if token == "" {
			return errMissingToken
		}
		if jwtValidator != nil && (auth.LooksLikeJWT(token) || introspectionValidator == nil) {
			return jwtValidator.Validate(ctx, token)
		}
		if introspectionValidator != nil {
			return introspectionValidator.validate(ctx, token)
		}
		return errMissingToken
	}
}

func (rr *ReceivingRelay[T]) checkStaticHeaders(md map[string]string) error {
	for k, v := range rr.staticHeaders {
		lk := strings.ToLower(k)
		got, ok := md[lk]
		if !ok || got != v {
			return errMissingHeader(k)
		}
	}
	return nil
}

func (rr *ReceivingRelay[T]) validateHeaders(ctx context.Context, md map[string]string) error {
	needPolicy := rr.dynamicAuthValidator != nil || len(rr.staticHeaders) > 0
	if !needPolicy {
		return nil
	}
	if len(md) == 0 {
		return errMissingHeaders
	}
	if err := rr.checkStaticHeaders(md); err != nil {
		return err
	}
	if rr.dynamicAuthValidator != nil {
		if err := rr.dynamicAuthValidator(ctx, md); err != nil {
			return err
		}
	}
	return nil
}

func headersFromMetadata(meta *relay.MessageMetadata) map[string]string {
	out := make(map[string]string)
	if meta == nil {
		return out
	}
	for k, v := range meta.GetHeaders() {
		if k == "" {
			continue
		}
		out[strings.ToLower(k)] = v
	}
	return out
}

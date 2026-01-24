package forwardrelay

import "github.com/joeydtaylor/electrician/pkg/internal/relay"

// NewAuthenticationOptionsOAuth2 builds AuthenticationOptions for OAuth2.
func NewAuthenticationOptionsOAuth2(oauth *relay.OAuth2Options) *relay.AuthenticationOptions {
	return &relay.AuthenticationOptions{
		Enabled: true,
		Mode:    relay.AuthMode_AUTH_OAUTH2,
		Oauth2:  oauth,
	}
}

// NewAuthenticationOptionsMTLS builds AuthenticationOptions for mTLS-only expectations.
func NewAuthenticationOptionsMTLS(allowedPrincipals []string, trustDomain string) *relay.AuthenticationOptions {
	return &relay.AuthenticationOptions{
		Enabled: true,
		Mode:    relay.AuthMode_AUTH_MUTUAL_TLS,
		Mtls: &relay.MTLSOptions{
			AllowedPrincipals: cloneStrings(allowedPrincipals),
			TrustDomain:       trustDomain,
		},
	}
}

// NewAuthenticationOptionsNone builds a disabled auth options object.
func NewAuthenticationOptionsNone() *relay.AuthenticationOptions {
	return &relay.AuthenticationOptions{
		Enabled: false,
		Mode:    relay.AuthMode_AUTH_NONE,
	}
}

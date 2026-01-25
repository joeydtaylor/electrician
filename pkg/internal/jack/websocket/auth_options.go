package websocket

import "github.com/joeydtaylor/electrician/pkg/internal/relay"

// NewAuthenticationOptionsOAuth2 composes AuthenticationOptions for OAuth2.
func NewAuthenticationOptionsOAuth2(oauth *relay.OAuth2Options) *relay.AuthenticationOptions {
	return &relay.AuthenticationOptions{
		Enabled: true,
		Mode:    relay.AuthMode_AUTH_OAUTH2,
		Oauth2:  oauth,
	}
}

// NewAuthenticationOptionsNone composes a disabled auth options object.
func NewAuthenticationOptionsNone() *relay.AuthenticationOptions {
	return &relay.AuthenticationOptions{
		Enabled: false,
		Mode:    relay.AuthMode_AUTH_NONE,
	}
}

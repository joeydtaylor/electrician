package receivingrelay

import "github.com/joeydtaylor/electrician/pkg/internal/relay"

// NewOAuth2JWTOptions builds JWT validation settings for resource servers.
func NewOAuth2JWTOptions(
	issuer string,
	jwksURI string,
	audiences []string,
	scopes []string,
	jwksCacheSeconds int32,
) *relay.OAuth2Options {
	return &relay.OAuth2Options{
		AcceptJwt:             true,
		AcceptIntrospection:   false,
		Issuer:                issuer,
		JwksUri:               jwksURI,
		RequiredAudience:      cloneStrings(audiences),
		RequiredScopes:        cloneStrings(scopes),
		JwksCacheSeconds:      jwksCacheSeconds,
		ForwardBearerToken:    false,
		ForwardMetadataKey:    "",
		IntrospectionUrl:      "",
		IntrospectionAuthType: "",
	}
}

// NewOAuth2IntrospectionOptions builds RFC 7662 introspection settings.
func NewOAuth2IntrospectionOptions(
	introspectionURL string,
	authType string,
	clientID string,
	clientSecret string,
	bearerToken string,
	introspectionCacheSeconds int32,
) *relay.OAuth2Options {
	return &relay.OAuth2Options{
		AcceptJwt:                 false,
		AcceptIntrospection:       true,
		IntrospectionUrl:          introspectionURL,
		IntrospectionAuthType:     authType,
		IntrospectionClientId:     clientID,
		IntrospectionClientSecret: clientSecret,
		IntrospectionBearerToken:  bearerToken,
		IntrospectionCacheSeconds: introspectionCacheSeconds,
	}
}

// NewOAuth2Forwarding controls forwarding of inbound bearer tokens to downstream services.
func NewOAuth2Forwarding(forward bool, forwardMetadataKey string) *relay.OAuth2Options {
	return &relay.OAuth2Options{
		ForwardBearerToken: forward,
		ForwardMetadataKey: forwardMetadataKey,
	}
}

// MergeOAuth2Options merges non-zero fields from src into dst.
func MergeOAuth2Options(dst *relay.OAuth2Options, src *relay.OAuth2Options) *relay.OAuth2Options {
	if dst == nil {
		return cloneOAuth2(src)
	}
	if src == nil {
		return dst
	}
	if src.AcceptJwt {
		dst.AcceptJwt = true
	}
	if src.AcceptIntrospection {
		dst.AcceptIntrospection = true
	}
	if src.Issuer != "" {
		dst.Issuer = src.Issuer
	}
	if src.JwksUri != "" {
		dst.JwksUri = src.JwksUri
	}
	if len(src.RequiredAudience) > 0 {
		dst.RequiredAudience = cloneStrings(src.RequiredAudience)
	}
	if len(src.RequiredScopes) > 0 {
		dst.RequiredScopes = cloneStrings(src.RequiredScopes)
	}
	if src.IntrospectionUrl != "" {
		dst.IntrospectionUrl = src.IntrospectionUrl
	}
	if src.IntrospectionAuthType != "" {
		dst.IntrospectionAuthType = src.IntrospectionAuthType
	}
	if src.IntrospectionClientId != "" {
		dst.IntrospectionClientId = src.IntrospectionClientId
	}
	if src.IntrospectionClientSecret != "" {
		dst.IntrospectionClientSecret = src.IntrospectionClientSecret
	}
	if src.IntrospectionBearerToken != "" {
		dst.IntrospectionBearerToken = src.IntrospectionBearerToken
	}
	if src.JwksCacheSeconds != 0 {
		dst.JwksCacheSeconds = src.JwksCacheSeconds
	}
	if src.IntrospectionCacheSeconds != 0 {
		dst.IntrospectionCacheSeconds = src.IntrospectionCacheSeconds
	}
	if src.ForwardMetadataKey != "" || src.ForwardBearerToken {
		dst.ForwardBearerToken = src.ForwardBearerToken
		dst.ForwardMetadataKey = src.ForwardMetadataKey
	}
	return dst
}

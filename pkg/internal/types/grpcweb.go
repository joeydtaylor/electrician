package types

// GRPCWebConfig controls CORS and transport behavior for gRPC-Web servers.
type GRPCWebConfig struct {
	// DisableWebsockets turns off gRPC-Web websocket transport upgrades.
	DisableWebsockets bool
	// AllowAllOrigins controls browser CORS behavior only; it does not bypass auth or TLS.
	AllowAllOrigins bool
	// AllowedOrigins restricts CORS origins when AllowAllOrigins is false.
	AllowedOrigins []string
	// AllowedHeaders augments the allowed request headers for CORS preflight.
	AllowedHeaders []string
}

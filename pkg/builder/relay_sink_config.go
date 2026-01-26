package builder

import (
	"crypto/tls"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// BoolPtr returns a pointer to the provided bool.
func BoolPtr(v bool) *bool {
	return &v
}

// RelaySinkConfig provides a typed way to build relay sink configuration.
type RelaySinkConfig struct {
	Targets        []string
	QueueSize      int
	SubmitTimeout  time.Duration
	FlushTimeout   time.Duration
	DropOnFull     *bool
	StaticHeaders  map[string]string
	TLS            *types.TLSConfig
	BearerToken    string
	BearerTokenEnv string
	AuthRequired   *bool
}

// ToSinkConfig converts the typed config into a SinkConfig map.
func (c RelaySinkConfig) ToSinkConfig() SinkConfig {
	cfg := map[string]interface{}{}

	if len(c.Targets) > 0 {
		cfg["targets"] = append([]string(nil), c.Targets...)
	}
	if c.QueueSize > 0 {
		cfg["queue_size"] = c.QueueSize
	}
	if c.SubmitTimeout > 0 {
		cfg["submit_timeout"] = c.SubmitTimeout
	}
	if c.FlushTimeout > 0 {
		cfg["flush_timeout"] = c.FlushTimeout
	}
	if c.DropOnFull != nil {
		cfg["drop_on_full"] = *c.DropOnFull
	}
	if len(c.StaticHeaders) > 0 {
		headers := make(map[string]string, len(c.StaticHeaders))
		for key, value := range c.StaticHeaders {
			headers[key] = value
		}
		cfg["static_headers"] = headers
	}
	if c.TLS != nil && c.TLS.UseTLS {
		cfg["tls"] = tlsConfigToMap(c.TLS)
	}
	if c.BearerToken != "" {
		cfg["bearer_token"] = c.BearerToken
	}
	if c.BearerTokenEnv != "" {
		cfg["bearer_token_env"] = c.BearerTokenEnv
	}
	if c.AuthRequired != nil {
		cfg["auth_required"] = *c.AuthRequired
	}

	return SinkConfig{
		Type:   string(RelaySink),
		Config: cfg,
	}
}

func tlsConfigToMap(cfg *types.TLSConfig) map[string]interface{} {
	out := map[string]interface{}{
		"cert":        cfg.CertFile,
		"key":         cfg.KeyFile,
		"ca":          cfg.CAFile,
		"server_name": cfg.SubjectAlternativeName,
	}

	if v := tlsVersionString(cfg.MinTLSVersion); v != "" {
		out["min_version"] = v
	}
	if v := tlsVersionString(cfg.MaxTLSVersion); v != "" {
		out["max_version"] = v
	}

	return out
}

func tlsVersionString(v uint16) string {
	switch v {
	case tls.VersionTLS10:
		return "1.0"
	case tls.VersionTLS11:
		return "1.1"
	case tls.VersionTLS12:
		return "1.2"
	case tls.VersionTLS13:
		return "1.3"
	default:
		return ""
	}
}

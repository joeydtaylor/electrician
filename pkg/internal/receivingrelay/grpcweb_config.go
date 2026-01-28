package receivingrelay

import (
	"strings"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

var defaultGRPCWebAllowedHeaders = []string{
	"authorization",
	"content-type",
	"connect-protocol-version",
	"grpc-accept-encoding",
	"grpc-encoding",
	"grpc-timeout",
	"x-tenant",
	"x-grpc-web",
	"x-trace-id",
	"x-user-agent",
}

func cloneGRPCWebConfig(cfg *types.GRPCWebConfig) *types.GRPCWebConfig {
	if cfg == nil {
		return nil
	}
	out := *cfg
	out.AllowedOrigins = cloneStringSlice(cfg.AllowedOrigins)
	out.AllowedHeaders = cloneStringSlice(cfg.AllowedHeaders)
	return &out
}

func (rr *ReceivingRelay[T]) snapshotGRPCWebConfig() types.GRPCWebConfig {
	if rr.grpcWebConfig == nil {
		return types.GRPCWebConfig{
			AllowAllOrigins: true,
			AllowedHeaders:  cloneStringSlice(defaultGRPCWebAllowedHeaders),
		}
	}
	cfg := *rr.grpcWebConfig
	cfg.AllowedOrigins = cloneStringSlice(cfg.AllowedOrigins)
	cfg.AllowedHeaders = mergeHeaders(defaultGRPCWebAllowedHeaders, cfg.AllowedHeaders)
	return cfg
}

func mergeHeaders(base []string, extra []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(base)+len(extra))
	for _, h := range append(cloneStringSlice(base), extra...) {
		h = strings.TrimSpace(strings.ToLower(h))
		if h == "" || seen[h] {
			continue
		}
		seen[h] = true
		out = append(out, h)
	}
	return out
}

func cloneStringSlice(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

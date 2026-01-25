package websocket

import (
	"strings"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"google.golang.org/protobuf/proto"
)

func cloneHeaderMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return make(map[string]string)
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneAuthOptions(in *relay.AuthenticationOptions) *relay.AuthenticationOptions {
	if in == nil {
		return nil
	}
	return proto.Clone(in).(*relay.AuthenticationOptions)
}

func cloneOAuth2(in *relay.OAuth2Options) *relay.OAuth2Options {
	if in == nil {
		return nil
	}
	return proto.Clone(in).(*relay.OAuth2Options)
}

func cloneStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func normalizeHeaderKey(key string) string {
	return strings.ToLower(strings.TrimSpace(key))
}

func trimStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	return out
}

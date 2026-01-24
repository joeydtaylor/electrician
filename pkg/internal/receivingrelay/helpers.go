package receivingrelay

import (
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"google.golang.org/protobuf/proto"
)

func cloneStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func cloneOAuth2(in *relay.OAuth2Options) *relay.OAuth2Options {
	if in == nil {
		return nil
	}
	return proto.Clone(in).(*relay.OAuth2Options)
}

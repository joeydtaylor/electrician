package quicrelay

import (
	"context"
	"fmt"
	"strings"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

func (fr *ForwardRelay[T]) buildStreamDefaults(ctx context.Context) (*relay.MessageMetadata, error) {
	headers := make(map[string]string)
	for k, v := range fr.staticHeaders {
		if k == "" {
			continue
		}
		headers[k] = v
	}
	if fr.dynamicHeaders != nil {
		for k, v := range fr.dynamicHeaders(ctx) {
			if strings.TrimSpace(k) == "" || strings.TrimSpace(v) == "" {
				continue
			}
			headers[k] = v
		}
	}

	if fr.tokenSource != nil {
		tok, err := fr.tokenSource.AccessToken(ctx)
		if err != nil || tok == "" {
			if fr.authRequired {
				if err == nil {
					err = fmt.Errorf("oauth token empty")
				}
				return nil, err
			}
		} else {
			headers["authorization"] = "Bearer " + tok
		}
	}

	if _, ok := headers["trace-id"]; !ok {
		headers["trace-id"] = utils.GenerateUniqueHash()
	}

	return &relay.MessageMetadata{
		Headers:     headers,
		ContentType: "application/octet-stream",
		Version: &relay.VersionInfo{
			Major: 1,
			Minor: 0,
		},
		Performance:    fr.PerformanceOptions,
		Security:       fr.SecurityOptions,
		Authentication: fr.authOptions,
	}, nil
}

package forwardrelay

import (
	"context"
	"fmt"

	"github.com/joeydtaylor/electrician/pkg/internal/utils"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func (fr *ForwardRelay[T]) buildPerRPCContext(ctx context.Context) (context.Context, error) {
	var md metadata.MD
	if existing, ok := metadata.FromOutgoingContext(ctx); ok {
		md = existing.Copy()
	} else {
		md = metadata.New(nil)
	}

	for k, v := range fr.staticHeaders {
		if k == "" {
			continue
		}
		md.Set(k, v)
	}

	if fr.dynamicHeaders != nil {
		for k, v := range fr.dynamicHeaders(ctx) {
			if k == "" || v == "" {
				continue
			}
			md.Set(k, v)
		}
	}

	if fr.tokenSource != nil {
		tok, err := fr.tokenSource.AccessToken(ctx)
		if err != nil || tok == "" {
			if fr.authRequired {
				if err == nil {
					err = fmt.Errorf("oauth token empty")
				}
				return ctx, fmt.Errorf("oauth token source error: %w", err)
			}
		} else {
			md.Set("authorization", "Bearer "+tok)
		}
	}

	if _, present := md["trace-id"]; !present {
		md.Set("trace-id", utils.GenerateUniqueHash())
	}

	return metadata.NewOutgoingContext(ctx, md), nil
}

func (fr *ForwardRelay[T]) makeDialOptions() ([]credentials.TransportCredentials, bool, error) {
	if fr.TlsConfig != nil && fr.TlsConfig.UseTLS {
		creds, err := fr.loadTLSCredentials(fr.TlsConfig)
		if err != nil {
			return nil, false, err
		}
		return []credentials.TransportCredentials{creds}, false, nil
	}

	if fr.tokenSource != nil && fr.authRequired {
		return nil, false, fmt.Errorf("oauth2 enabled and required but TLS is disabled; refuse to dial insecure")
	}

	return nil, true, nil
}

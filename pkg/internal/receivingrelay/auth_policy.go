package receivingrelay

import (
	"context"
	"strings"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (rr *ReceivingRelay[T]) maybePolicyError(err error) error {
	if err == nil {
		return nil
	}
	if rr.authRequired {
		return err
	}
	rr.NotifyLoggers(types.WarnLevel, "Auth: soft-failing policy error: %v", err)
	return nil
}

func (rr *ReceivingRelay[T]) buildUnaryPolicyInterceptor() grpc.UnaryServerInterceptor {
	needPolicy := rr.dynamicAuthValidator != nil || len(rr.staticHeaders) > 0
	if !needPolicy {
		return nil
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		mdMap := rr.collectIncomingMD(ctx)

		if err := rr.checkStaticHeaders(mdMap); err != nil {
			if e := rr.maybePolicyError(err); e != nil {
				return nil, e
			}
		}

		if rr.dynamicAuthValidator != nil {
			if err := rr.dynamicAuthValidator(ctx, mdMap); err != nil {
				if e := rr.maybePolicyError(status.Errorf(codes.Unauthenticated, "auth validation failed: %v", err)); e != nil {
					return nil, e
				}
			}
		}

		return handler(ctx, req)
	}
}

func (rr *ReceivingRelay[T]) buildStreamPolicyInterceptor() grpc.StreamServerInterceptor {
	needPolicy := rr.dynamicAuthValidator != nil || len(rr.staticHeaders) > 0
	if !needPolicy {
		return nil
	}
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		mdMap := rr.collectIncomingMD(ctx)

		if err := rr.checkStaticHeaders(mdMap); err != nil {
			if e := rr.maybePolicyError(err); e != nil {
				return e
			}
		}

		if rr.dynamicAuthValidator != nil {
			if err := rr.dynamicAuthValidator(ctx, mdMap); err != nil {
				if e := rr.maybePolicyError(status.Errorf(codes.Unauthenticated, "auth validation failed: %v", err)); e != nil {
					return e
				}
			}
		}

		return handler(srv, ss)
	}
}

func (rr *ReceivingRelay[T]) appendAuthServerOptions(opts []grpc.ServerOption) []grpc.ServerOption {
	rr.ensureDefaultAuthValidator()

	var unaryChain []grpc.UnaryServerInterceptor
	if p := rr.buildUnaryPolicyInterceptor(); p != nil {
		unaryChain = append(unaryChain, p)
	}
	if rr.authUnary != nil {
		unaryChain = append(unaryChain, rr.authUnary)
	}
	if len(unaryChain) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(unaryChain...))
	}

	var streamChain []grpc.StreamServerInterceptor
	if p := rr.buildStreamPolicyInterceptor(); p != nil {
		streamChain = append(streamChain, p)
	}
	if rr.authStream != nil {
		streamChain = append(streamChain, rr.authStream)
	}
	if len(streamChain) > 0 {
		opts = append(opts, grpc.ChainStreamInterceptor(streamChain...))
	}
	return opts
}

func (rr *ReceivingRelay[T]) collectIncomingMD(ctx context.Context) map[string]string {
	out := make(map[string]string)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for k, vals := range md {
			if len(vals) > 0 {
				out[strings.ToLower(k)] = vals[0]
			}
		}
	}
	return out
}

func (rr *ReceivingRelay[T]) checkStaticHeaders(md map[string]string) error {
	for k, v := range rr.staticHeaders {
		lk := strings.ToLower(k)
		got, ok := md[lk]
		if !ok || got != v {
			return status.Errorf(codes.Unauthenticated, "missing/invalid header %s", k)
		}
	}
	return nil
}

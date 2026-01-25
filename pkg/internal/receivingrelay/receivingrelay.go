package receivingrelay

import (
	"context"
	"net/http"
	"sync"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
	"google.golang.org/grpc"
)

// ReceivingRelay implements the relay gRPC server and forwards decoded items to submitters.
type ReceivingRelay[T any] struct {
	relay.UnimplementedRelayServiceServer

	ctx    context.Context
	cancel context.CancelFunc

	Outputs []types.Submitter[T]

	componentMetadata types.ComponentMetadata
	Address           string

	DataCh chan T

	Loggers     []types.Logger
	loggersLock sync.Mutex

	TlsConfig    *types.TLSConfig
	isRunning    int32
	configFrozen int32

	DecryptionKey string

	authOptions *relay.AuthenticationOptions
	authUnary   grpc.UnaryServerInterceptor
	authStream  grpc.StreamServerInterceptor

	staticHeaders map[string]string

	dynamicAuthValidator func(ctx context.Context, md map[string]string) error
	authRequired         bool

	passthrough bool

	grpcWebConfig   *types.GRPCWebConfig
	grpcWebServer   *http.Server
	grpcWebServerMu sync.Mutex

	outputsOnce sync.Once
}

// NewReceivingRelay initializes a receiving relay with optional configuration.
func NewReceivingRelay[T any](ctx context.Context, options ...types.Option[types.ReceivingRelay[T]]) types.ReceivingRelay[T] {
	ctx, cancel := context.WithCancel(ctx)

	rr := &ReceivingRelay[T]{
		ctx:    ctx,
		cancel: cancel,
		DataCh: make(chan T, 1024),
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "RECEIVING_RELAY",
		},
		Loggers:       make([]types.Logger, 0),
		DecryptionKey: "",
		staticHeaders: make(map[string]string),
		authRequired:  true,
	}

	for _, option := range options {
		option(rr)
	}

	return rr
}

// Codec abstracts payload encoding/decoding.
type Codec[T any] interface {
	Encode(v T) ([]byte, error)
	Decode(b []byte, out *T) error
	ContentType() string
}

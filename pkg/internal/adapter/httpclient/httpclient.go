package httpclient

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// Use a UTC time format, for clear, universal understanding
const TimeFormat = "2006-01-02 15:04:05 UTC"

type RequestConfig struct {
	Method   string
	Endpoint string
	Body     io.Reader
}

type OAuth2Config struct {
	ClientID string
	Secret   string
	TokenURL string
	Audience string
	Scopes   []string
}

type OAuthToken struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time // Calculate the expiry time when token is set
}

type HTTPClientAdapter[T any] struct {
	componentMetadata types.ComponentMetadata
	ctx               context.Context
	cancel            context.CancelFunc
	configLock        sync.Mutex
	plugFunc          []types.AdapterFunc[T]
	baseURL           string
	httpClient        *http.Client
	headers           map[string]string
	interval          time.Duration
	sensors           []types.Sensor[T]
	sensorLock        sync.Mutex
	loggers           []types.Logger
	loggersLock       sync.Mutex
	requestConfig     RequestConfig // Stores the current request configuration
	isServing         int32         // atomic flag to check if Serve is already running
	maxRetries        int
	timeout           time.Duration
	oAuthToken        *OAuthToken
	tokenMutex        sync.Mutex // Ensure thread-safe token access
	oauth2Config      *OAuth2Config
	tlsPinningEnabled bool
	pinnedCert        []byte // This could be the DER-encoded bytes of the certificate
}

// NewHTTPClientAdapter creates a new HTTP plug with a base URL, custom headers, and an interval.
func NewHTTPClientAdapter[T any](ctx context.Context, options ...types.Option[types.HTTPClientAdapter[T]]) types.HTTPClientAdapter[T] {
	ctx, cancel := context.WithCancel(ctx)

	http := &HTTPClientAdapter[T]{
		ctx:    ctx,
		cancel: cancel,
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "HTTP_ADAPTER",
		},
		headers:    make(map[string]string), // Initialize the map to prevent nil map assignment panic
		loggers:    make([]types.Logger, 0),
		sensors:    make([]types.Sensor[T], 0),
		httpClient: &http.Client{Timeout: 30 * time.Second},
		plugFunc:   make([]types.AdapterFunc[T], 0),
		maxRetries: 0,
		timeout:    30 * time.Second,
	}

	for _, opt := range options {
		opt(http)
	}

	return http
}

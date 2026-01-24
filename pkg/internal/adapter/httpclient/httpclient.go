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

// TimeFormat provides a stable UTC format for timestamps in logs.
const TimeFormat = "2006-01-02 15:04:05 UTC"

// RequestConfig describes the HTTP request to execute.
type RequestConfig struct {
	Method   string
	Endpoint string
	Body     io.Reader
}

// OAuth2Config contains OAuth client credentials settings.
type OAuth2Config struct {
	ClientID string
	Secret   string
	TokenURL string
	Audience string
	Scopes   []string
}

// OAuthToken stores access token metadata.
type OAuthToken struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   time.Time
}

// HTTPClientAdapter fetches remote resources and decodes responses into T.
type HTTPClientAdapter[T any] struct {
	componentMetadata types.ComponentMetadata

	ctx context.Context

	httpClient *http.Client

	configLock  sync.Mutex
	request     RequestConfig
	headers     map[string]string
	interval    time.Duration
	maxRetries  int
	timeout     time.Duration
	oauthConfig *OAuth2Config
	pinnedCert  []byte
	pinEnabled  bool

	tokenLock  sync.Mutex
	oAuthToken *OAuthToken

	sensorsLock sync.Mutex
	sensors     []types.Sensor[T]

	loggersLock sync.Mutex
	loggers     []types.Logger

	isServing int32
}

// NewHTTPClientAdapter constructs an adapter with defaults and applies options.
func NewHTTPClientAdapter[T any](ctx context.Context, options ...types.Option[types.HTTPClientAdapter[T]]) types.HTTPClientAdapter[T] {
	if ctx == nil {
		ctx = context.Background()
	}

	http := &HTTPClientAdapter[T]{
		ctx: ctx,
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "HTTP_ADAPTER",
		},
		headers:    make(map[string]string),
		sensors:    make([]types.Sensor[T], 0),
		loggers:    make([]types.Logger, 0),
		interval:   time.Second,
		maxRetries: 0,
		timeout:    30 * time.Second,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	for _, opt := range options {
		if opt == nil {
			continue
		}
		opt(http)
	}

	return http
}

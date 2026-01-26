package internallogger

import (
	"context"
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/forwardrelay"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/logschema"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	logPayloadContentType   = "application/json"
	defaultRelayQueueSize   = 2048
	defaultRelaySubmitDelay = 2 * time.Second
)

type relaySinkConfig struct {
	targets       []string
	tlsConfig     *types.TLSConfig
	staticHeaders map[string]string
	tokenSource   types.OAuth2TokenSource
	authRequired  bool
	queueSize     int
	submitTimeout time.Duration
	flushTimeout  time.Duration
	dropOnFull    bool
}

type relayWriteSyncer struct {
	relay   types.ForwardRelay[*relay.WrappedPayload]
	cfg     relaySinkConfig
	ctx     context.Context
	cancel  context.CancelFunc
	queue   chan *relay.WrappedPayload
	pending int64
	dropped uint64
	wg      sync.WaitGroup
	writeMu sync.Mutex
	closed  bool
}

func newRelayWriteSyncer(config types.SinkConfig) (*relayWriteSyncer, error) {
	cfg, err := parseRelaySinkConfig(config)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	options := []types.Option[types.ForwardRelay[*relay.WrappedPayload]]{
		forwardrelay.WithTarget[*relay.WrappedPayload](cfg.targets...),
		forwardrelay.WithPassthrough[*relay.WrappedPayload](true),
		forwardrelay.WithAuthRequired[*relay.WrappedPayload](cfg.authRequired),
	}
	if cfg.tlsConfig != nil {
		options = append(options, forwardrelay.WithTLSConfig[*relay.WrappedPayload](cfg.tlsConfig))
	}
	if cfg.tokenSource != nil {
		options = append(options, forwardrelay.WithOAuth2[*relay.WrappedPayload](cfg.tokenSource))
	}
	if len(cfg.staticHeaders) > 0 {
		options = append(options, forwardrelay.WithStaticHeaders[*relay.WrappedPayload](cfg.staticHeaders))
	}

	fr := forwardrelay.NewForwardRelay[*relay.WrappedPayload](ctx, options...)

	if cfg.queueSize < 1 {
		cfg.queueSize = 1
	}

	ws := &relayWriteSyncer{
		relay:  fr,
		cfg:    cfg,
		ctx:    ctx,
		cancel: cancel,
		queue:  make(chan *relay.WrappedPayload, cfg.queueSize),
	}
	ws.wg.Add(1)
	go ws.run()
	return ws, nil
}

func (r *relayWriteSyncer) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	payload := r.newPayload(p)

	r.writeMu.Lock()
	defer r.writeMu.Unlock()
	if r.closed {
		return len(p), nil
	}

	if r.cfg.dropOnFull {
		select {
		case r.queue <- payload:
			atomic.AddInt64(&r.pending, 1)
		default:
			atomic.AddUint64(&r.dropped, 1)
		}
		return len(p), nil
	}

	r.queue <- payload
	atomic.AddInt64(&r.pending, 1)
	return len(p), nil
}

func (r *relayWriteSyncer) Sync() error {
	timeout := r.cfg.flushTimeout
	if timeout <= 0 {
		timeout = r.cfg.submitTimeout
	}
	if timeout <= 0 {
		timeout = defaultRelaySubmitDelay
	}

	deadline := time.Now().Add(timeout)
	for atomic.LoadInt64(&r.pending) > 0 {
		if time.Now().After(deadline) {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

func (r *relayWriteSyncer) Close() {
	r.writeMu.Lock()
	if r.closed {
		r.writeMu.Unlock()
		return
	}
	r.closed = true
	close(r.queue)
	r.writeMu.Unlock()

	r.wg.Wait()
	r.relay.Stop()
	r.cancel()
}

func (r *relayWriteSyncer) run() {
	defer r.wg.Done()
	for payload := range r.queue {
		r.submit(payload)
	}
}

func (r *relayWriteSyncer) submit(payload *relay.WrappedPayload) {
	ctx := r.ctx
	var cancel context.CancelFunc
	if r.cfg.submitTimeout > 0 {
		ctx, cancel = context.WithTimeout(r.ctx, r.cfg.submitTimeout)
	}
	_ = r.relay.Submit(ctx, payload)
	if cancel != nil {
		cancel()
	}
	atomic.AddInt64(&r.pending, -1)
}

func (r *relayWriteSyncer) newPayload(p []byte) *relay.WrappedPayload {
	buf := make([]byte, len(p))
	copy(buf, p)

	ts := timestamppb.Now()
	meta := &relay.MessageMetadata{
		Headers: map[string]string{
			"source":              "electrician-logger",
			logschema.FieldSchema: logschema.SchemaID,
		},
		ContentType: logPayloadContentType,
		Version: &relay.VersionInfo{
			Major: 1,
			Minor: 0,
		},
	}

	return &relay.WrappedPayload{
		Id:          ts.AsTime().UTC().Format(time.RFC3339Nano),
		Timestamp:   ts,
		Payload:     buf,
		Metadata:    meta,
		PayloadType: logschema.SchemaID,
	}
}

func parseRelaySinkConfig(config types.SinkConfig) (relaySinkConfig, error) {
	cfg := relaySinkConfig{
		authRequired:  true,
		queueSize:     defaultRelayQueueSize,
		submitTimeout: defaultRelaySubmitDelay,
		flushTimeout:  defaultRelaySubmitDelay,
		dropOnFull:    true,
	}

	if config.Config == nil {
		return cfg, fmt.Errorf("relay sink config missing")
	}

	targets, err := parseTargets(config.Config["targets"])
	if err != nil {
		return cfg, err
	}
	if len(targets) == 0 {
		return cfg, fmt.Errorf("relay sink targets missing")
	}
	cfg.targets = targets

	if raw, ok := config.Config["queue_size"]; ok {
		queueSize, err := parseInt(raw)
		if err != nil || queueSize < 1 {
			return cfg, fmt.Errorf("relay sink queue_size invalid")
		}
		cfg.queueSize = queueSize
	}

	if raw, ok := config.Config["drop_on_full"]; ok {
		dropOnFull, err := parseBool(raw)
		if err != nil {
			return cfg, fmt.Errorf("relay sink drop_on_full invalid")
		}
		cfg.dropOnFull = dropOnFull
	}

	if raw, ok := config.Config["submit_timeout"]; ok {
		timeout, err := parseDuration(raw)
		if err != nil {
			return cfg, fmt.Errorf("relay sink submit_timeout invalid")
		}
		cfg.submitTimeout = timeout
	}

	if raw, ok := config.Config["flush_timeout"]; ok {
		timeout, err := parseDuration(raw)
		if err != nil {
			return cfg, fmt.Errorf("relay sink flush_timeout invalid")
		}
		cfg.flushTimeout = timeout
	}

	if raw, ok := config.Config["auth_required"]; ok {
		required, err := parseBool(raw)
		if err != nil {
			return cfg, fmt.Errorf("relay sink auth_required invalid")
		}
		cfg.authRequired = required
	}

	if raw, ok := config.Config["static_headers"]; ok {
		headers, err := parseStringMap(raw)
		if err != nil {
			return cfg, fmt.Errorf("relay sink static_headers invalid")
		}
		cfg.staticHeaders = headers
	}

	if raw, ok := config.Config["tls"]; ok {
		tlsCfg, err := parseTLSConfig(raw)
		if err != nil {
			return cfg, err
		}
		cfg.tlsConfig = tlsCfg
	}

	token, hasToken := config.Config["bearer_token"]
	tokenEnv, hasTokenEnv := config.Config["bearer_token_env"]
	if hasToken && hasTokenEnv {
		return cfg, fmt.Errorf("relay sink config specifies both bearer_token and bearer_token_env")
	}
	if hasToken {
		val, ok := token.(string)
		if !ok || val == "" {
			return cfg, fmt.Errorf("relay sink bearer_token invalid")
		}
		cfg.tokenSource = forwardrelay.NewStaticBearerTokenSource(val)
	}
	if hasTokenEnv {
		val, ok := tokenEnv.(string)
		if !ok || val == "" {
			return cfg, fmt.Errorf("relay sink bearer_token_env invalid")
		}
		cfg.tokenSource = forwardrelay.NewEnvBearerTokenSource(val)
	}

	return cfg, nil
}

func parseTargets(raw interface{}) ([]string, error) {
	switch v := raw.(type) {
	case string:
		return splitTargets(v), nil
	case []string:
		return normalizeTargets(v), nil
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("relay sink targets must be strings")
			}
			out = append(out, s)
		}
		return normalizeTargets(out), nil
	case nil:
		return nil, fmt.Errorf("relay sink targets missing")
	default:
		return nil, fmt.Errorf("relay sink targets invalid")
	}
}

func splitTargets(raw string) []string {
	parts := strings.Split(raw, ",")
	return normalizeTargets(parts)
}

func normalizeTargets(parts []string) []string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func parseBool(raw interface{}) (bool, error) {
	switch v := raw.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(strings.TrimSpace(v))
	default:
		return false, fmt.Errorf("unsupported bool type %T", raw)
	}
}

func parseInt(raw interface{}) (int, error) {
	switch v := raw.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case float32:
		return int(v), nil
	case string:
		val, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return 0, err
		}
		return val, nil
	default:
		return 0, fmt.Errorf("unsupported int type %T", raw)
	}
}

func parseDuration(raw interface{}) (time.Duration, error) {
	switch v := raw.(type) {
	case time.Duration:
		return v, nil
	case string:
		return time.ParseDuration(strings.TrimSpace(v))
	case int:
		return time.Duration(v) * time.Millisecond, nil
	case int64:
		return time.Duration(v) * time.Millisecond, nil
	case float64:
		return time.Duration(v) * time.Millisecond, nil
	default:
		return 0, fmt.Errorf("unsupported duration type %T", raw)
	}
}

func parseStringMap(raw interface{}) (map[string]string, error) {
	switch v := raw.(type) {
	case map[string]string:
		return v, nil
	case map[string]interface{}:
		out := make(map[string]string, len(v))
		for key, val := range v {
			s, ok := val.(string)
			if !ok {
				return nil, fmt.Errorf("static header %q must be string", key)
			}
			out[key] = s
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported map type %T", raw)
	}
}

func parseTLSConfig(raw interface{}) (*types.TLSConfig, error) {
	cfgMap, ok := raw.(map[string]interface{})
	if !ok {
		if strMap, ok := raw.(map[string]string); ok {
			cfgMap = make(map[string]interface{}, len(strMap))
			for key, val := range strMap {
				cfgMap[key] = val
			}
		} else {
			return nil, fmt.Errorf("relay sink tls must be map")
		}
	}
	if enabledRaw, ok := cfgMap["enabled"]; ok {
		enabled, err := parseBool(enabledRaw)
		if err != nil {
			return nil, fmt.Errorf("relay sink tls enabled invalid")
		}
		if !enabled {
			return nil, nil
		}
	}

	cert, _ := cfgMap["cert"].(string)
	key, _ := cfgMap["key"].(string)
	ca, _ := cfgMap["ca"].(string)
	serverName, _ := cfgMap["server_name"].(string)

	if cert == "" || key == "" || ca == "" {
		return nil, fmt.Errorf("relay sink tls requires cert, key, and ca")
	}

	out := &types.TLSConfig{
		UseTLS:                 true,
		CertFile:               cert,
		KeyFile:                key,
		CAFile:                 ca,
		SubjectAlternativeName: serverName,
	}

	if rawMin, ok := cfgMap["min_version"]; ok {
		minVersion, err := parseTLSVersion(rawMin)
		if err != nil {
			return nil, err
		}
		out.MinTLSVersion = minVersion
	}
	if rawMax, ok := cfgMap["max_version"]; ok {
		maxVersion, err := parseTLSVersion(rawMax)
		if err != nil {
			return nil, err
		}
		out.MaxTLSVersion = maxVersion
	}

	return out, nil
}

func parseTLSVersion(raw interface{}) (uint16, error) {
	switch v := raw.(type) {
	case string:
		return tlsVersionFromString(v)
	case float64:
		return tlsVersionFromFloat(v)
	case int:
		return tlsVersionFromFloat(float64(v))
	default:
		return 0, fmt.Errorf("unsupported tls version type %T", raw)
	}
}

func tlsVersionFromString(raw string) (uint16, error) {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "1.0", "tls1.0":
		return tls.VersionTLS10, nil
	case "1.1", "tls1.1":
		return tls.VersionTLS11, nil
	case "1.2", "tls1.2":
		return tls.VersionTLS12, nil
	case "1.3", "tls1.3":
		return tls.VersionTLS13, nil
	default:
		return 0, fmt.Errorf("unsupported tls version %q", raw)
	}
}

func tlsVersionFromFloat(v float64) (uint16, error) {
	switch v {
	case 1.0:
		return tls.VersionTLS10, nil
	case 1.1:
		return tls.VersionTLS11, nil
	case 1.2:
		return tls.VersionTLS12, nil
	case 1.3:
		return tls.VersionTLS13, nil
	default:
		return 0, fmt.Errorf("unsupported tls version %v", v)
	}
}

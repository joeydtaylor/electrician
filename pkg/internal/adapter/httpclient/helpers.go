package httpclient

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

type adapterConfig struct {
	request     RequestConfig
	headers     map[string]string
	interval    time.Duration
	maxRetries  int
	timeout     time.Duration
	oauthConfig *OAuth2Config
	ctx         context.Context
}

func (hp *HTTPClientAdapter[T]) snapshotConfig() adapterConfig {
	hp.configLock.Lock()
	cfg := adapterConfig{
		request:    hp.request,
		interval:   hp.interval,
		maxRetries: hp.maxRetries,
		timeout:    hp.timeout,
		ctx:        hp.ctx,
	}

	if hp.headers != nil {
		cfg.headers = make(map[string]string, len(hp.headers))
		for k, v := range hp.headers {
			cfg.headers[k] = v
		}
	}

	if hp.oauthConfig != nil {
		copyConfig := *hp.oauthConfig
		cfg.oauthConfig = &copyConfig
	}
	hp.configLock.Unlock()

	if cfg.ctx == nil {
		cfg.ctx = context.Background()
	}

	return cfg
}

func (hp *HTTPClientAdapter[T]) setContext(ctx context.Context) {
	if ctx == nil {
		return
	}
	hp.configLock.Lock()
	hp.ctx = ctx
	hp.configLock.Unlock()
}

func (hp *HTTPClientAdapter[T]) snapshotSensors() []types.Sensor[T] {
	hp.sensorsLock.Lock()
	sensors := append([]types.Sensor[T](nil), hp.sensors...)
	hp.sensorsLock.Unlock()
	return sensors
}

func (hp *HTTPClientAdapter[T]) snapshotLoggers() []types.Logger {
	hp.loggersLock.Lock()
	loggers := append([]types.Logger(nil), hp.loggers...)
	hp.loggersLock.Unlock()
	return loggers
}

func backoffDuration(attempt int) time.Duration {
	if attempt < 0 {
		return 0
	}
	exp := time.Duration(math.Pow(2, float64(attempt))) * time.Second
	if exp <= 0 {
		return 0
	}
	return time.Duration(rand.Int63n(int64(exp)))
}

func sleepWithContext(ctx context.Context, duration time.Duration) error {
	if duration <= 0 {
		return nil
	}
	if ctx == nil {
		time.Sleep(duration)
		return nil
	}

	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

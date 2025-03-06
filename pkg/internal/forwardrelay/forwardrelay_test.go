package forwardrelay_test

import (
	"context"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

func TestForwardRelayInitialization(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tlsConfig := builder.NewTlsClientConfig(true, "../../../cmd/example/relay_example/tls/client.crt", "../../../cmd/example/relay_example/tls/client.key", "../../../cmd/example/relay_example/tls/ca.crt")

	relay := builder.NewForwardRelay[string](
		ctx,
		builder.ForwardRelayWithTarget[string]("localhost:50051"),
		builder.ForwardRelayWithTLSConfig[string](tlsConfig),
	)

	if relay == nil {
		t.Errorf("Failed to initialize ForwardRelay")
	}
}

func TestForwardRelayShutdown(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	inputConduit := builder.NewConduit[string](ctx)

	tlsConfig := builder.NewTlsClientConfig(true, "../../../cmd/example/relay_example/tls/client.crt", "../../../cmd/example/relay_example/tls/client.key", "../../../cmd/example/relay_example/tls/ca.crt")
	relay := builder.NewForwardRelay[string](
		ctx,
		builder.ForwardRelayWithTarget[string]("localhost:50051"),
		builder.ForwardRelayWithTLSConfig[string](tlsConfig),
		builder.ForwardRelayWithInput[string](inputConduit),
	)

	relay.Start(ctx) // This should block until ctx is done.

	<-ctx.Done()

	relay.Stop()

	if relay.IsRunning() { // Assuming IsRunning() checks if the relay is still active
		t.Errorf("Relay did not shut down gracefully")
	}
}

/* func TestForwardRelayThroughput(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	concurrencyConfig := builder.NewConcurrencyConfig[string](1000, 10000)

	logger := builder.NewLogger(builder.InfoLevel)

	inputConduit := builder.NewConduit[string](ctx, concurrencyConfig, nil, nil, nil)
	tlsConfig := builder.NewTlsClientConfig(true, "../../../cmd/example/relay_example/tls/client.crt", "../../../cmd/example/relay_example/tls/client.key", "../../../cmd/example/relay_example/tls/ca.crt")
	relay := builder.NewForwardRelay[string](ctx, "localhost:50051", inputConduit, logger, tlsConfig, nil)

	relay.Start(ctx)
	defer relay.Stop()

	// Bombard the relay with a high volume of messages
	go func() {
		for i := 0; i < 10000; i++ {
			if ctx.Err() != nil {
				return
			}
			inputConduit.Submit(ctx, fmt.Sprintf("Message %d", i))
		}
	}()

	<-ctx.Done()
	if ctx.Err() == context.DeadlineExceeded {
		t.Error("Timeout reached, potential performance issues with handling high throughput")
	} else {
		t.Log("High throughput test completed successfully")
	}
}
*/

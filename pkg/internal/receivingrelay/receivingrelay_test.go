package receivingrelay_test

import (
	"context"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

func TestReceivingRelayInitialization(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tlsConfig := builder.NewTlsServerConfig(true, "../../../cmd/example/relay_example/tls/server.crt", "../../../cmd/example/relay_example/tls/server.key", "../../../cmd/example/relay_example/tls/ca.crt", "localhost")

	relay := builder.NewReceivingRelay[string](
		ctx,
		builder.ReceivingRelayWithAddress[string]("localhost:50051"),
		builder.ReceivingRelayWithTLSConfig[string](tlsConfig),
	)

	<-ctx.Done()

	if relay == nil {
		t.Errorf("Failed to initialize ReceivingRelay")
	}
}

func TestReceivingRelayShutdown(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	output := builder.NewConduit[string](ctx)

	tlsConfig := builder.NewTlsServerConfig(true, "../../../cmd/example/relay_example/tls/server.crt", "../../../cmd/example/relay_example/tls/server.key", "../../../cmd/example/relay_example/tls/ca.crt", "localhost")
	relay := builder.NewReceivingRelay[string](
		ctx,
		builder.ReceivingRelayWithAddress[string]("localhost:50051"),
		builder.ReceivingRelayWithTLSConfig[string](tlsConfig),
		builder.ReceivingRelayWithOutput[string](output),
	)

	relay.Start(ctx) // This should block until ctx is done.

	<-ctx.Done()

	relay.Stop()

	if relay.IsRunning() { // Assuming IsRunning() checks if the relay is still active
		t.Errorf("Relay did not shut down gracefully")
	}
}

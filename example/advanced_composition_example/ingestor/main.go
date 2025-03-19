package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

type EDRData struct {
	EndpointID string
	EventType  string
	Content    string
	IsThreat   bool
}

// Function to simulate generating EDR data from multiple endpoints.
func generateEDRData(ctx context.Context, endpointID string, submitFunc func(ctx context.Context, data EDRData) error) {
	ticker := time.NewTicker(time.Duration(rand.Intn(200)+300) * time.Millisecond)
	defer ticker.Stop()

	eventTypes := []string{"Login", "FileAccess", "NetworkActivity"}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			data := EDRData{
				EndpointID: endpointID,
				EventType:  eventTypes[rand.Intn(len(eventTypes))],
				Content:    fmt.Sprintf("Random event data from %s", endpointID),
			}
			// Increase the chance of marking something as a threat
			if rand.Float32() < 0.2 {
				data.IsThreat = true
				data.Content += " - Threat Detected"
			}
			if err := submitFunc(ctx, data); err != nil {
				fmt.Printf("Error submitting EDR data from %s: %v\n", endpointID, err)
			}
		}
	}
}

func standardizeData(data EDRData) (EDRData, error) {
	data.Content = strings.ToUpper(data.Content)
	return data, nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := builder.NewLogger(builder.LoggerWithLevel("debug"), builder.LoggerWithDevelopment(true))

	tipOffNotifierWire := builder.NewWire[EDRData](ctx)

	// Sensor for monitoring
	sensor := builder.NewSensor(
		builder.SensorWithOnElementProcessedFunc(
			func(c builder.ComponentMetadata, elem EDRData) {
				if elem.IsThreat {
					tipOffNotifierWire.Submit(ctx, elem)
				}
			},
		),
	)

	// Main processing conduit
	mainConduit := builder.NewConduit[EDRData](
		ctx,
	)

	// Set up multiple generators for each endpoint
	endpoints := []string{"Endpoint1", "Endpoint2", "Endpoint3"}
	for _, endpoint := range endpoints {
		plug := builder.NewPlug(
			ctx,
			builder.PlugWithAdapterFunc(func(ctx context.Context, submit func(ctx context.Context, data EDRData) error) {
				generateEDRData(ctx, endpoint, submit)
			}),
			/* 			builder.PlugWithLogger(logger), */
		)

		generator := builder.NewGenerator(
			ctx,
			builder.GeneratorWithPlug(plug),
			/* 			builder.GeneratorWithLogger[EDRData](logger), */
		)

		// Create wires for each generator and add them to the conduit
		wire := builder.NewWire(
			ctx,
			builder.WireWithTransformer(standardizeData),
			builder.WireWithGenerator(generator),
			builder.WireWithSensor(sensor),
			/* 			builder.WireWithLogger[EDRData](logger), */
		)

		mainConduit.ConnectWire(wire)
	}

	tlsConfig := builder.NewTlsClientConfig(
		true,                // UseTLS should be true to use TLS
		"../tls/client.crt", // Path to the client's certificate
		"../tls/client.key", // Path to the client's private key
		"../tls/ca.crt",     // Path to the CA certificate
		tls.VersionTLS13,    // MinVersion: Only allow TLS 1.3
		tls.VersionTLS13,    // MaxVersion: Only allow TLS 1.3
	)

	tipOffNotifier := builder.NewForwardRelay(
		ctx,
		builder.ForwardRelayWithLogger[EDRData](logger),
		builder.ForwardRelayWithInput(tipOffNotifierWire),
		builder.ForwardRelayWithTarget[EDRData]("localhost:50051", "localhost:50052"),
		builder.ForwardRelayWithTLSConfig[EDRData](tlsConfig),
	)

	// Start the entire conduit
	tipOffNotifierWire.Start(ctx)
	mainConduit.Start(ctx)
	tipOffNotifier.Start(ctx)

	// Wait for the context to be canceled (signal or internal cancellation)
	<-ctx.Done()

	// Stop and clean up resources
	tipOffNotifierWire.Stop()
	mainConduit.Stop()
	tipOffNotifier.Stop()
	fmt.Println("Service shutdown gracefully")
}

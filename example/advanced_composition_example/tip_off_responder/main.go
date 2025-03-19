package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

type EDRData struct {
	EndpointID string
	EventType  string
	Content    string
	IsThreat   bool
}

// Enhanced real-time analysis function with random threat detection.
func threatAnalysis(data EDRData) (EDRData, error) {
	// Randomly decide to flag an event as a threat to simulate occasional detections
	if strings.Contains(strings.ToLower(data.Content), "threat") || rand.Float32() < 0.05 {
		data.IsThreat = true
	}
	return data, nil
}

func main() {
	// Setup cancellation context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize the logger
	logger := builder.NewLogger(builder.LoggerWithDevelopment(true), builder.LoggerWithLevel("debug"))

	// Create the TLS configuration for secure connections
	tlsConfig := builder.NewTlsServerConfig(
		true,
		"../tls/server.crt",
		"../tls/server.key",
		"../tls/ca.crt",
		"localhost",
	)

	// Sensor to monitor and react to threat detections
	sensor := builder.NewSensor(
		builder.SensorWithOnElementProcessedFunc(func(c builder.ComponentMetadata, elem EDRData) {
			if elem.IsThreat {
				fmt.Printf("Security team notified about threat from %s.\n", elem.EndpointID)
			}
		}),
	)

	// Wire for real-time threat analysis
	threatWire := builder.NewWire(
		ctx,
		builder.WireWithTransformer(threatAnalysis),
		builder.WireWithLogger[EDRData](logger),
		builder.WireWithSensor(sensor),
	)

	// Wire for storing data in a data lake
	dataLakeWire := builder.NewWire(
		ctx,
		builder.WireWithLogger[EDRData](logger),
	)

	// Receiving relays for different purposes
	responderService := builder.NewReceivingRelay(
		ctx,
		builder.ReceivingRelayWithAddress[EDRData]("localhost:50051"),
		builder.ReceivingRelayWithLogger[EDRData](logger),
		builder.ReceivingRelayWithOutput(threatWire),
		builder.ReceivingRelayWithTLSConfig[EDRData](tlsConfig),
	)

	dataLakeService := builder.NewReceivingRelay(
		ctx,
		builder.ReceivingRelayWithAddress[EDRData]("localhost:50052"),
		builder.ReceivingRelayWithLogger[EDRData](logger),
		builder.ReceivingRelayWithOutput(dataLakeWire),
		builder.ReceivingRelayWithTLSConfig[EDRData](tlsConfig),
	)

	// Start services
	responderService.Start(ctx)
	dataLakeService.Start(ctx)

	// Wait for the context to be canceled (signal or internal cancellation)
	<-ctx.Done()

	// Aggregate and display results after services are stopped
	threatOutput, err := threatWire.LoadAsJSONArray()
	if err != nil {
		fmt.Printf("Error converting threat analysis output to JSON: %v\n", err)
	} else {
		fmt.Println("Threat Analysis Summary:")
		fmt.Println(string(threatOutput))
	}

	dataLakeOutput, err := dataLakeWire.LoadAsJSONArray()
	if err != nil {
		fmt.Printf("Error converting data lake output to JSON: %v\n", err)
	} else {
		fmt.Println("Data Lake Storage Summary:")
		fmt.Println(string(dataLakeOutput))
	}
}

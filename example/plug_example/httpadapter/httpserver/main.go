package main

import (
	"context"
	"crypto/tls"
	"encoding/json" // Needed for JSON marshalling
	"fmt"
	"time"

	// Import your builder package, which wraps the httpserver functional options.
	"github.com/joeydtaylor/electrician/pkg/builder"
)

// MyRequest represents the request body for incoming JSON data.
type MyRequest struct {
	Name string `json:"name"`
}

// MyResponse is the JSON body we want to send back to the client.
type MyResponse struct {
	Greeting string `json:"greeting"`
}

func main() {
	// Create a background context to run the server.
	ctx := context.Background()

	// Build a simple logger using the builder’s logger factory.
	myLogger := builder.NewLogger(
		builder.LoggerWithLevel("info"),      // e.g. "debug", "info", "warn", "error"
		builder.LoggerWithDevelopment(false), // toggles dev mode in the internal logger
	)

	// Create a TLSConfig object. Adjust paths to your cert files in the tls/ directory.
	// The example below sets:
	//   - UseTLS = true => Enables HTTPS
	//   - CertFile = "../tls/server.crt"
	//   - KeyFile  = "../tls/server.key"
	//   - CAFile   = "../tls/ca.crt"
	//   - MinTLSVersion = tls.VersionTLS12
	//   - MaxTLSVersion = tls.VersionTLS13
	tlsConfig := builder.NewTlsServerConfig(
		true,                // Enable TLS
		"../tls/server.crt", // Path to server cert
		"../tls/server.key", // Path to server key
		"../tls/ca.crt",     // Path to CA cert, if needed
		"localhost",         // SubjectAlternativeName (optional)
		tls.VersionTLS12,    // Minimum TLS version
		tls.VersionTLS13,    // Maximum TLS version
	)

	// Create the server using the builder’s NewHTTPServerAdapter function with TLS.
	server := builder.NewHTTPServerAdapter[MyRequest](ctx,
		builder.HTTPServerAdapterWithAddress[MyRequest](":8443"),               // Listen on port 8443 (HTTPS)
		builder.HTTPServerAdapterWithServerConfig[MyRequest]("POST", "/hello"), // Handle POST /hello
		builder.HTTPServerAdapterWithLogger[MyRequest](myLogger),               // Attach our logger
		builder.HTTPServerAdapterWithHeader[MyRequest]("Server", "ExampleServer/1.0"),
		builder.HTTPServerAdapterWithTimeout[MyRequest](5*time.Second), // 5-second read/write timeout
		builder.HTTPServerAdapterWithTLS[MyRequest](*tlsConfig),        // Pass our TLS config
	)

	// Define the callback function that processes incoming requests of type MyRequest.
	// Because builder.HTTPServerResponse.Body is []byte, we must JSON-marshal any struct ourselves.
	handleRequest := func(ctx context.Context, req MyRequest) (builder.HTTPServerResponse, error) {
		fmt.Printf("Received request with Name=%q\n", req.Name)

		// 1) Create your response struct.
		responseData := MyResponse{
			Greeting: "Hello, " + req.Name,
		}

		// 2) Marshal the struct into a JSON byte slice for the Body.
		bodyBytes, err := json.Marshal(responseData)
		if err != nil {
			return builder.HTTPServerResponse{}, fmt.Errorf("failed to marshal response: %w", err)
		}

		// 3) Construct a custom HTTPServerResponse with your desired status code, headers, and JSON bytes.
		response := builder.HTTPServerResponse{
			StatusCode: 201, // e.g., “Created”
			Headers: map[string]string{
				"X-Custom-Header": "MyExample",
				"Content-Type":    "application/json",
			},
			Body: bodyBytes, // Must be []byte
		}

		return response, nil
	}

	fmt.Println("[Main] Starting HTTPS server. Send POST requests to https://localhost:8443/hello")

	// Start the server. This call blocks until ctx is canceled or there's a fatal error.
	if err := server.Serve(ctx, handleRequest); err != nil {
		fmt.Printf("[Main] Server stopped with error: %v\n", err)
	} else {
		fmt.Println("[Main] Server stopped gracefully.")
	}
}

package main

import (
	"context"
	"fmt"
	"time"

	// Import your builder package, which wraps the httpserver functional options.
	"github.com/joeydtaylor/electrician/pkg/builder"
)

// MyRequest represents the request body for incoming data.
type MyRequest struct {
	Name string `json:"name"`
}

// main sets up and runs the HTTP server adapter.
func main() {
	// Create a background context that we can use to run the server.
	ctx := context.Background()

	// Build a simple logger using the builder’s logger factory.
	myLogger := builder.NewLogger(
		builder.LoggerWithLevel("info"),      // e.g. "debug", "info", "warn", "error"
		builder.LoggerWithDevelopment(false), // toggles dev mode in the internal logger
	)

	// Create the server using the builder’s NewHTTPServerAdapter function.
	// We supply a few functional options to demonstrate usage.
	server := builder.NewHTTPServerAdapter(ctx,
		builder.HTTPServerAdapterWithAddress[MyRequest](":8080"),               // Listen on port 8080
		builder.HTTPServerAdapterWithServerConfig[MyRequest]("POST", "/hello"), // Handle POST /hello
		builder.HTTPServerAdapterWithLogger[MyRequest](myLogger),               // Attach our logger
		builder.HTTPServerAdapterWithHeader[MyRequest]("Server", "ExampleServer/1.0"),
		builder.HTTPServerAdapterWithTimeout[MyRequest](5*time.Second), // 5-second read/write timeout
	)

	// Define the callback function that processes incoming requests of type MyRequest.
	// This is our “pipeline” function that the server calls after JSON decoding.
	handleRequest := func(ctx context.Context, req MyRequest) error {
		// For simplicity, we’ll just log the request data.
		fmt.Printf("Received request with Name=%q\n", req.Name)
		// ... do any processing here ...

		// Return nil to indicate success (200 OK).
		return nil
	}

	// Start the server. This call will block until ctx is canceled
	// or the server encounters a fatal error.
	fmt.Println("[Main] Starting HTTP server. Send POST requests to http://localhost:8080/hello")
	err := server.Serve(ctx, handleRequest)
	if err != nil {
		fmt.Printf("[Main] Server stopped with error: %v\n", err)
	} else {
		fmt.Println("[Main] Server stopped gracefully.")
	}
}

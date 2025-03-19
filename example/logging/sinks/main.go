package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

func plugFunc(ctx context.Context, submitFunc func(ctx context.Context, message string) error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return // Stop the generator if the context is cancelled
		case <-ticker.C:
			submitFunc(ctx, "hello")
			submitFunc(ctx, "world")
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger := builder.NewLogger(builder.LoggerWithDevelopment(true), builder.LoggerWithLevel("debug"))
	defer logger.Flush()

	// Add a file sink
	fileSinkConfig := builder.SinkConfig{
		Type: string(builder.FileSink),
		Config: map[string]interface{}{
			"path": "logs/output.log",
		},
	}
	if err := logger.AddSink("fileSink", fileSinkConfig); err != nil {
		fmt.Printf("Failed to add file sink: %v\n", err)
		return
	}

	// Add a console sink (stdout)
	consoleSinkConfig := builder.SinkConfig{Type: "stdout"}
	if err := logger.AddSink("consoleSink", consoleSinkConfig); err != nil {
		fmt.Printf("Failed to add console sink: %v\n", err)
		return
	}

	plug := builder.NewPlug(
		ctx,
		builder.PlugWithAdapterFunc(plugFunc),
	)

	// Initialize the generator with one plug.
	generator := builder.NewGenerator(
		ctx,
		builder.GeneratorWithPlug(plug), // First plug
		builder.GeneratorWithLogger[string](logger),
	)

	transform := func(input string) (string, error) {
		return strings.ToUpper(input), nil
	}

	wire := builder.NewWire(
		ctx,
		builder.WireWithLogger[string](logger),
		builder.WireWithTransformer(transform),
		builder.WireWithGenerator(generator),
	)

	wire.Start(ctx)

	<-ctx.Done()

	wire.Stop()

	output, err := wire.LoadAsJSONArray()
	if err != nil {
		fmt.Printf("Error converting output to JSON: %v\n", err)
		return
	}

	fmt.Println("Output Summary:")
	fmt.Println(string(output))
}

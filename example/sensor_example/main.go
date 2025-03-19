package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

func plugFunc(ctx context.Context, submitFunc func(ctx context.Context, message string) error) {
	// Simulate high frequency message generation
	for i := 0; i < 10000; i++ {
		if err := submitFunc(ctx, fmt.Sprintf("message %d", i)); err != nil {
			fmt.Printf("Error submitting message %d: %v\n", i, err)
			return
		}
		if i%10 == 0 { // occasionally simulate an error
			if err := submitFunc(ctx, "error"); err != nil {
				fmt.Printf("Error submitting 'error': %v\n", err)
				return
			}
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	plug := builder.NewPlug(
		ctx,
		builder.PlugWithAdapterFunc(plugFunc),
	)

	// Set up the sensor to count elements and report on completion.
	sensor := builder.NewSensor(
		builder.SensorWithOnStartFunc[string](func(c builder.ComponentMetadata) { fmt.Printf("Wire started: %v", c) }),
		builder.SensorWithOnElementProcessedFunc(func(c builder.ComponentMetadata, elem string) { fmt.Printf("%v -> Processed element: %+v\n", c, elem) }),
		builder.SensorWithOnCancelFunc(func(c builder.ComponentMetadata, elem string) {
			fmt.Printf("%v -> Context cancelled processing element: %+v\n", c, elem)
		}),
		builder.SensorWithOnErrorFunc(func(c builder.ComponentMetadata, err error, elem string) {
			fmt.Printf("%v -> Error processing element: %+v, Error: %+v\n", c, elem, err)
		}),
		builder.SensorWithOnCompleteFunc[string](func(c builder.ComponentMetadata) { fmt.Printf("%v -> Processing complete", c) }),
	)

	// Create a transformation function: convert string to uppercase.
	transform := func(input string) (string, error) {
		time.Sleep(1 * time.Second)
		if input == "error" {
			return "", fmt.Errorf("simulated processing error")
		}
		return strings.ToUpper(input), nil
	}

	generator := builder.NewGenerator(
		ctx,
		builder.GeneratorWithPlug(plug),
	)

	// Create a wire with transformation, sensor, and circuit breaker.
	wire := builder.NewWire(
		ctx,
		builder.WireWithTransformer(transform),
		builder.WireWithSensor(sensor),
		builder.WireWithGenerator(generator),
	)

	wire.Start(ctx)

	// Let the wire run for a short time and then stop it.
	time.Sleep(2 * time.Second) // Short sleep to show rapid cancellation
	cancel()                    // Trigger the cancellation
	wire.Stop()
	fmt.Println("Wire terminated.")

	// Load the output if needed
	output, err := wire.LoadAsJSONArray()
	if err != nil {
		fmt.Printf("Error converting output to JSON: %v\n", err)
		return
	}
	fmt.Println(string(output))
}

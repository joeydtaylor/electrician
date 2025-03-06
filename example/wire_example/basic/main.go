package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	fmt.Println("Initializing transformer to convert strings to uppercase...")

	transform := func(input string) (string, error) {
		return strings.ToUpper(input), nil
	}

	fmt.Println("Creating a new Wire with the transformer...")

	wire := builder.NewWire(
		ctx,
		builder.WireWithTransformer(transform),
	)

	fmt.Println("Starting the Wire...")
	wire.Start(ctx)

	fmt.Println("Submitting 'hello' to the Wire...")
	wire.Submit(ctx, "hello")

	fmt.Println("Submitting 'world' to the Wire...")
	wire.Submit(ctx, "world")

	// Wait for the context to be done
	<-ctx.Done()

	fmt.Println("Stopping the Wire...")
	wire.Stop()

	fmt.Println("Loading output as JSON array...")
	output, err := wire.LoadAsJSONArray()
	if err != nil {
		fmt.Printf("Error converting output to JSON: %v\n", err)
		return
	}

	fmt.Println("Output Summary:")
	fmt.Println(string(output))
}

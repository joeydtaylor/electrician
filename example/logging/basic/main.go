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

	logger := builder.NewLogger()

	transform := func(input string) (string, error) {
		return strings.ToUpper(input), nil
	}

	wire := builder.NewWire(
		ctx,
		builder.WireWithLogger[string](logger),
		builder.WireWithTransformer(transform),
	)

	wire.Start(ctx)

	wire.Submit(ctx, "hello")
	wire.Submit(ctx, "world")

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

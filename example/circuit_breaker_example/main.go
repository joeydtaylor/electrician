package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

type Item struct {
	ID      int
	Content string
}

// Simulate processing that can randomly fail based on item ID.
func processItem(item Item) (Item, error) {
	if item.ID%10 == 0 { // Simulate an error on every 10th item.
		return Item{}, errors.New("simulated processing error")
	}
	return item, nil
}

// Plug simulates external input by generating items with increasing IDs.
func plugFunc(ctx context.Context, submitFunc func(ctx context.Context, item Item) error) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	id := 1
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			item := Item{
				ID:      id,
				Content: fmt.Sprintf("This is item number %d", id),
			}
			if err := submitFunc(ctx, item); err != nil {
				continue
			}
			id++
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	/* 	logger := builder.NewLogger() */

	meter := builder.NewMeter[Item](ctx)

	sensor := builder.NewSensor[Item](builder.SensorWithMeter[Item](meter))

	plug := builder.NewPlug[Item](
		ctx,
		builder.PlugWithAdapterFunc[Item](plugFunc),
		builder.PlugWithSensor[Item](sensor),
	)

	generator := builder.NewGenerator[Item](
		ctx,
		builder.GeneratorWithPlug[Item](plug),
		builder.GeneratorWithSensor[Item](sensor),
	)

	/* 	groundWire := builder.NewWire[Item](ctx, builder.WireWithLogger[Item](logger)) */
	circuitBreaker := builder.NewCircuitBreaker[Item](
		ctx,
		1,             // Trip after 3 errors.
		5*time.Second, // Reset after 5 seconds.
		/* 		builder.CircuitBreakerWithLogger[Item](logger), */
		/* 		builder.CircuitBreakerWithGroundWire[Item](groundWire), */
		builder.CircuitBreakerWithComponentMetadata[Item]("MyCircuitBreakerNameMetadata", "123456789"),
		builder.CircuitBreakerWithSensor[Item](sensor),
	)

	processingWire := builder.NewWire[Item](
		ctx,
		/* 		builder.WireWithLogger[Item](logger), */
		builder.WireWithTransformer[Item](processItem),
		builder.WireWithCircuitBreaker[Item](circuitBreaker),
		builder.WireWithGenerator[Item](generator),
		builder.WireWithSensor[Item](sensor),
	)

	/* 	groundWire.Start(ctx) */
	processingWire.Start(ctx)

	/* 	// Wait for processing to complete or timeout.
	   	<-ctx.Done() */

	meter.Monitor()

	processingWire.Stop()

	output, err := processingWire.LoadAsJSONArray()
	if err != nil {
		fmt.Printf("Error converting output to JSON: %v\n", err)
		return
	}

	fmt.Println("GroundWire Summary:")
	fmt.Println(string(output))

}

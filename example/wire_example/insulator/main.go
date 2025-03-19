package main

import (
	"context"
	"fmt"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

type Item struct {
	ID      int
	Content string
}

func processItem(item Item) (Item, error) {
	switch item.ID {
	case 3:
		return item, fmt.Errorf("simulated processing error for item 3")
	case 20:
		return item, fmt.Errorf("simulated processing error for item 20")
	default:
		return item, nil
	}
}

// retryItem retries based on specific error content using switch
func retryItem(ctx context.Context, item Item, elementError error) (Item, error) {
	switch {
	case item.ID == 3 && elementError != nil && elementError.Error() == "simulated processing error for item 3":
		// Handle recovery attempt for item 3
		item.Content = "Recovered content for item 3"
		return item, nil // Return the item with recovered content
	case item.ID == 20 && elementError != nil && elementError.Error() == "simulated processing error for item 20":
		// Fail the recovery for item 20 after showing intent to handle it
		return item, elementError // Explicitly returning the error
	default:
		// Return the item and any error that is not handled specifically
		return item, elementError
	}
}

func plugFunc(ctx context.Context, submitFunc func(ctx context.Context, item Item) error) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	id := 1
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			item := Item{ID: id, Content: fmt.Sprintf("This is item number %d", id)}
			if err := submitFunc(ctx, item); err != nil {
				fmt.Printf("Error submitting item %d: %v\n", id, err)
				continue
			}
			id++
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	logger := builder.NewLogger()

	plug := builder.NewPlug(
		ctx,
		builder.PlugWithAdapterFunc(plugFunc),
	)

	generator := builder.NewGenerator(
		ctx,
		builder.GeneratorWithPlug(plug),
	)

	circuitBreaker := builder.NewCircuitBreaker(
		ctx,
		1, // Trips after 1 error.
		5*time.Second,
		builder.CircuitBreakerWithLogger[Item](logger),
		builder.CircuitBreakerWithComponentMetadata[Item]("MyCircuitBreakerNameMetadata", "123456789"),
	)

	processingWire := builder.NewWire(
		ctx,
		builder.WireWithLogger[Item](logger),
		builder.WireWithEncoder(builder.NewJSONEncoder[Item]()),
		builder.WireWithTransformer(processItem),
		builder.WireWithCircuitBreaker(circuitBreaker),
		builder.WireWithGenerator(generator),
		builder.WireWithInsulator(retryItem, 3, 1*time.Second),
	)

	processingWire.Start(ctx)
	<-ctx.Done()
	processingWire.Stop()

	output, err := processingWire.LoadAsJSONArray()
	if err != nil {
		fmt.Printf("Error converting output to JSON: %v\n", err)
		return
	}

	fmt.Println("ProcessingWire Summary:")
	fmt.Println(string(output))

	if ctx.Err() == context.DeadlineExceeded {
		fmt.Println("Processing timed out.")
	} else {
		fmt.Println("Processing finished.")
	}
}

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

func adapterFunc(ctx context.Context, submitFunc func(ctx context.Context, item Item) error) {
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

	plug := builder.NewPlug[Item](
		ctx,
		builder.PlugWithAdapterFunc[Item](adapterFunc),
	)

	generator := builder.NewGenerator[Item](
		ctx,
		builder.GeneratorWithPlug[Item](plug),
	)

	processingWire := builder.NewWire[Item](
		ctx,
		builder.WireWithLogger[Item](logger),
		builder.WireWithTransformer[Item](processItem),
		builder.WireWithGenerator[Item](generator),
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

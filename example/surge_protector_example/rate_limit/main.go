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

// generator produces a fixed number of items at random intervals
func plugFunc(ctx context.Context, submitFunc func(ctx context.Context, item Item) error) {

	id := 1
	totalItems := 10 // Only generate 10 items
	for id <= totalItems {
		interval := time.Duration(100) * time.Millisecond // Random interval between 0ms to 1000ms
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
			item := Item{ID: id, Content: fmt.Sprintf("This is item number %d", id)}
			if err := submitFunc(ctx, item); err != nil {
				continue
			}
			if err := submitFunc(ctx, item); err != nil {
				continue
			}
			id++
		}
	}
}

func main() {
	// Context timeout slightly longer than the time required to produce all items
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	meter := builder.NewMeter[Item](
		ctx,
		builder.MeterWithIdleTimeout[Item](10*time.Second),
	)

	sensor := builder.NewSensor(
		builder.SensorWithMeter[Item](meter),
	)

	plug := builder.NewPlug(
		ctx,
		builder.PlugWithAdapterFunc(plugFunc),
		builder.PlugWithSensor(sensor),
	)

	generator := builder.NewGenerator(
		ctx,
		builder.GeneratorWithPlug(plug),
		builder.GeneratorWithSensor(sensor),
	)

	resister := builder.NewResister(
		ctx,
		builder.ResisterWithSensor(sensor),
	)

	surgeProtector := builder.NewSurgeProtector(
		ctx,
		builder.SurgeProtectorWithSensor(sensor),
		builder.SurgeProtectorWithResister(resister),
		builder.SurgeProtectorWithRateLimit[Item](
			1,
			5000*time.Millisecond,
			13,
		),
	)

	processingWire := builder.NewWire(
		ctx,
		builder.WireWithSensor(sensor),
		builder.WireWithGenerator(generator),
		builder.WireWithSurgeProtector(surgeProtector),
	)

	processingWire.Start(ctx)

	meter.Monitor()

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

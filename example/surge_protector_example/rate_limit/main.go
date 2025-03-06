package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

type Item struct {
	ID      int
	Content string
}

// generator produces a fixed number of items at random intervals
func plugFunc(ctx context.Context, submitFunc func(ctx context.Context, item Item) error) {
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator

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

	sensor := builder.NewSensor[Item](
		builder.SensorWithMeter[Item](meter),
	)

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

	resister := builder.NewResister[Item](
		ctx,
		builder.ResisterWithSensor[Item](sensor),
	)

	surgeProtector := builder.NewSurgeProtector[Item](
		ctx,
		builder.SurgeProtectorWithSensor[Item](sensor),
		builder.SurgeProtectorWithResister[Item](resister),
		builder.SurgeProtectorWithRateLimit[Item](
			1,
			5000*time.Millisecond,
			13,
		),
	)

	processingWire := builder.NewWire[Item](
		ctx,
		builder.WireWithSensor[Item](sensor),
		builder.WireWithGenerator[Item](generator),
		builder.WireWithSurgeProtector[Item](surgeProtector),
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

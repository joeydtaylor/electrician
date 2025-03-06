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

// plug produces items continuously at random intervals
func plugFunc(ctx context.Context, submitFunc func(ctx context.Context, item Item) error) {
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator

	id := 1
	for {
		interval := time.Duration(rand.Intn(1000)) * time.Millisecond // Random interval between 0ms to 1000ms
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
			item := Item{ID: id, Content: fmt.Sprintf("This is item number %d", id)}
			if err := submitFunc(ctx, item); err != nil {
				fmt.Printf("Error submitting item %d: %v\n", id, err)
				continue // Optionally handle errors here, perhaps with a retry mechanism or logging
			}
			id++
		}
	}
}

func main() {
	// Context timeout slightly longer than the time required to produce all items
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	meter := builder.NewMeter[Item](ctx)
	sensor := builder.NewSensor[Item](builder.SensorWithMeter[Item](meter))

	plug := builder.NewPlug[Item](
		ctx,
		builder.PlugWithAdapterFunc[Item](plugFunc),
		builder.PlugWithSensor[Item](sensor),
	)

	backupWire := builder.NewWire[Item](
		ctx,
		builder.WireWithSensor[Item](sensor),
	)

	generator := builder.NewGenerator[Item](
		ctx,
		builder.GeneratorWithPlug[Item](plug),
		builder.GeneratorWithSensor[Item](sensor),
	)

	surgeProtector := builder.NewSurgeProtector[Item](
		ctx,
		builder.SurgeProtectorWithBackupSystem[Item](backupWire),
		builder.SurgeProtectorWithSensor[Item](sensor),
	)

	processingWire := builder.NewWire[Item](
		ctx,
		builder.WireWithSensor[Item](sensor),
		builder.WireWithGenerator[Item](generator),
		builder.WireWithSurgeProtector[Item](surgeProtector),
	)

	backupWire.Start(ctx)
	processingWire.Start(ctx)

	go func() {
		ticker := time.NewTicker(10 * time.Second) // Sets a ticker that ticks every 10 seconds
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done(): // Always good to handle the context cancellation
				return
			case <-ticker.C:
				surgeProtector.Trip()
				time.Sleep(5 * time.Second) // Wait for 5 seconds after tripping before resetting
				surgeProtector.Reset()
			}
		}
	}()

	meter.Monitor()

	output, err := backupWire.LoadAsJSONArray()
	if err != nil {
		fmt.Printf("Error converting output to JSON: %v\n", err)
		return
	}

	fmt.Println("BackupWire Summary:")
	fmt.Println(string(output))

	processingOutput, err := processingWire.LoadAsJSONArray()
	if err != nil {
		fmt.Printf("Error converting output to JSON: %v\n", err)
		return
	}

	fmt.Println("processingOutput Summary after reset:")
	fmt.Println(string(processingOutput))

	if ctx.Err() == context.DeadlineExceeded {
		fmt.Println("Processing timed out.")
	} else {
		fmt.Println("Processing finished.")
	}
}

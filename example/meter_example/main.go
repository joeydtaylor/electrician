package main

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

type Feedback struct {
	CustomerID string
	Content    string
	Category   string
	IsNegative bool
	Tags       []string
}

var counter uint64

func errorSimulator(feedback Feedback) (Feedback, error) {
	count := atomic.AddUint64(&counter, 1)
	if count%5 == 0 {
		return Feedback{}, errors.New("simulated processing error")
	}
	return feedback, nil
}

func processor(feedback Feedback) (Feedback, error) {
	feedback.Content += "-PROCESSED"
	time.Sleep(4000 * time.Millisecond)
	return feedback, nil
}

func plugFunc(ctx context.Context, submitFunc func(ctx context.Context, feedback Feedback) error) {
	totalItems := 1000000
	for i := 0; i < totalItems; i++ {
		feedback := Feedback{
			CustomerID: fmt.Sprintf("Feedback%d", i),
			Content:    fmt.Sprintf("This is feedback item number %d", i),
			IsNegative: i%10 == 0,
		}
		if err := submitFunc(ctx, feedback); err != nil {
			fmt.Printf("\nError submitting feedback: %v\n", err)
			return
		}
		if ctx.Err() != nil {
			fmt.Printf("Plug stopped: %v\n", ctx.Err())
			return
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a new Meter
	meter := builder.NewMeter[Feedback](ctx,
		builder.MeterWithTotalItems[Feedback](1000000),
		builder.MeterWithIdleTimeout[Feedback](10*time.Second),
	)

	sensor := builder.NewSensor(
		builder.SensorWithMeter[Feedback](meter),
	)

	plug := builder.NewPlug(
		ctx,
		builder.PlugWithAdapterFunc(plugFunc),
	)

	generator := builder.NewGenerator(
		ctx,
		builder.GeneratorWithPlug(plug),
	)

	wire := builder.NewWire(
		ctx,
		builder.WireWithTransformer(processor, errorSimulator),
		builder.WireWithGenerator(generator),
		builder.WireWithSensor(sensor),
	)

	conduit := builder.NewConduit(
		ctx,
		builder.ConduitWithWire(wire),
		builder.ConduitWithConcurrencyControl[Feedback](1000000, 1000000),
	)

	conduit.Start(ctx)

	meter.Monitor()

	conduit.Stop()

	output, err := conduit.LoadAsJSONArray()
	if err != nil {
		fmt.Printf("Error converting output to JSON: %v\n", err)
		return
	}

	fmt.Println("Feedback Analysis Summary:")
	fmt.Println(string(output))
}

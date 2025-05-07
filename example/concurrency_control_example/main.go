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
	totalItems := 10000000
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

	totalItems := 10000000

	// Create a new Meter
	meter := builder.NewMeter[Feedback](ctx,
		builder.MeterWithTotalItems[Feedback](uint64(totalItems)),
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
		builder.WireWithConcurrencyControl[Feedback](1000000, 10000000),
		builder.WireWithSensor(sensor),
	)

	wire.Start(ctx)

	meter.Monitor()

	wire.Stop()

	output, _ := wire.LoadAsJSONArray()

	fmt.Printf("Output Results, RIP stdout")
	fmt.Print(string(output))

}

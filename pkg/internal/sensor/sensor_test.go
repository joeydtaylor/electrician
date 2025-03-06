package sensor_test

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

func plugFunc(ctx context.Context, submitFunc func(ctx context.Context, message string) error) {
	for i := 0; i < 10000; i++ { // Increased count to simulate high volume
		if err := submitFunc(ctx, fmt.Sprintf("message %d", i)); err != nil {
			return
		}
		if i%2500 == 0 { // Simulate an error occasionally
			submitFunc(ctx, "error")
		}
	}
}

func TestSensorCallbacks(t *testing.T) {
	var startCount, processCount, cancelCount, errorCount, terminateCount int64

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	plug := builder.NewPlug[string](
		ctx,
		builder.PlugWithAdapterFunc[string](plugFunc),
	)

	// Set up the sensor with callbacks that increment counters.
	sensor := builder.NewSensor[string](
		builder.SensorWithOnStartFunc[string](func(c builder.ComponentMetadata) { atomic.AddInt64(&startCount, 1) }),
		builder.SensorWithOnElementProcessedFunc[string](func(c builder.ComponentMetadata, elem string) { atomic.AddInt64(&processCount, 1) }),
		builder.SensorWithOnCancelFunc[string](func(c builder.ComponentMetadata, elem string) { atomic.AddInt64(&cancelCount, 1) }),
		builder.SensorWithOnErrorFunc[string](func(c builder.ComponentMetadata, err error, elem string) { atomic.AddInt64(&errorCount, 1) }),
		builder.SensorWithOnStopFunc[string](func(c builder.ComponentMetadata) { atomic.AddInt64(&terminateCount, 1) }),
	)

	transform := func(input string) (string, error) {
		if input == "error" {
			return "", fmt.Errorf("simulated processing error")
		}
		return strings.ToUpper(input), nil
	}

	generator := builder.NewGenerator[string](
		ctx,
		builder.GeneratorWithPlug[string](plug),
	)

	// Create a wire with transformation, sensor, and generator that sends many messages quickly.
	wire := builder.NewWire[string](
		ctx,
		builder.WireWithTransformer[string](transform),
		builder.WireWithSensor[string](sensor),
		builder.WireWithGenerator[string](generator),
	)

	wire.Start(ctx)
	time.Sleep(5 * time.Second)
	wire.Stop()

	// Test assertions using atomic reads
	if atomic.LoadInt64(&startCount) != 1 {
		t.Errorf("Expected start to be called once, got %d", startCount)
	}
	if atomic.LoadInt64(&processCount) == 0 {
		t.Errorf("Expected processed count to be greater than 0, got %d", processCount)
	}
	if atomic.LoadInt64(&errorCount) == 0 {
		t.Errorf("Expected error count to be greater than 0, got %d", errorCount)
	}
	if atomic.LoadInt64(&cancelCount) == 0 {
		t.Errorf("Expected cancel count to be greater than 0, got %d", cancelCount)
	}
	if atomic.LoadInt64(&terminateCount) != 1 {
		t.Errorf("Expected terminate to be called once, got %d", terminateCount)
	}
}

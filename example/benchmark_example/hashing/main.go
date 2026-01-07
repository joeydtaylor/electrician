package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

/*
   tiny, real-work benchmark for electrician v1.24.3
   -------------------------------------------------
   • hashes 10 M feedback records (sha256 × 2 000)
   • injects an error on every 5th record
*/

type Feedback struct {
	CustomerID string
	Content    string
	IsNegative bool
	ResultHash string
}

const (
	itemsToProcess = 10000000
	hashRounds     = 2000
	errorEveryNth  = 5
)

var errCounter uint64

// ----- helpers --------------------------------------------------------------

func injectError(f Feedback) (Feedback, error) {
	if atomic.AddUint64(&errCounter, 1)%errorEveryNth == 0 {
		return Feedback{}, errors.New("simulated error")
	}
	return f, nil
}

func shaProcessor(f Feedback) (Feedback, error) {
	sum := sha256.Sum256([]byte(f.Content))
	for i := 1; i < hashRounds; i++ {
		sum = sha256.Sum256(sum[:])
	}
	f.ResultHash = hex.EncodeToString(sum[:])
	return f, nil
}

func plug(ctx context.Context, submit func(context.Context, Feedback) error) {
	for i := 0; i < itemsToProcess; i++ {
		f := Feedback{
			CustomerID: fmt.Sprintf("C%07d", i),
			Content:    fmt.Sprintf("feedback-%d", i),
			IsNegative: i%10 == 0,
		}
		if err := submit(ctx, f); err != nil || ctx.Err() != nil {
			return
		}
	}
}

// ----- main ------------------------------------------------------------------

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	meter := builder.NewMeter[Feedback](ctx,
		builder.MeterWithTotalItems[Feedback](uint64(itemsToProcess)),
	)

	wire := builder.NewWire(
		ctx,
		builder.WireWithTransformer(shaProcessor, injectError),
		builder.WireWithGenerator(
			builder.NewGenerator(ctx,
				builder.GeneratorWithPlug(
					builder.NewPlug(ctx, builder.PlugWithAdapterFunc(plug)),
				),
			),
		),
		builder.WireWithSensor(builder.NewSensor(builder.SensorWithMeter[Feedback](meter))),
		builder.WireWithConcurrencyControl[Feedback](10000000, 10000),
	)

	start := time.Now()
	wire.Start(ctx)
	meter.Monitor() // blocks until done
	wire.Stop()

	elapsed := time.Since(start).Truncate(time.Millisecond)
	fmt.Printf("\nprocessed %d items in %s  →  %.0f items/s\n",
		itemsToProcess, elapsed, float64(itemsToProcess)/elapsed.Seconds())
}

package surgeprotector_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
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
	totalItems := 5 // Only generate 11 items
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
			id++
		}
	}
}

func TestProcessingWire(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resister := builder.NewResister[Item](ctx)

	plug := builder.NewPlug[Item](
		ctx,
		builder.PlugWithAdapterFunc[Item](plugFunc),
	)

	generator := builder.NewGenerator[Item](
		ctx,
		builder.GeneratorWithPlug[Item](plug),
	)

	surgeProtector := builder.NewSurgeProtector[Item](
		ctx,
		builder.SurgeProtectorWithRateLimit[Item](
			3,
			500*time.Millisecond,
			10,
		),
		builder.SurgeProtectorWithResister[Item](resister),
	)

	processingWire := builder.NewWire[Item](
		ctx,
		builder.WireWithGenerator[Item](generator),
		builder.WireWithSurgeProtector[Item](surgeProtector),
	)

	processingWire.Start(ctx)

	// Wait for the processing to complete
	<-ctx.Done()

	output, err := processingWire.LoadAsJSONArray()
	if err != nil {
		t.Fatalf("Error converting output to JSON: %v", err)
	}

	var items []Item
	err = json.Unmarshal([]byte(output), &items)
	if err != nil {
		t.Fatalf("JSON unmarshalling should succeed, got error: %v", err)
	}

	expectedIDs := make(map[int]bool, 5)
	for i := 1; i <= 5; i++ {
		expectedIDs[i] = true
	}

	for _, item := range items {
		if !expectedIDs[item.ID] {
			t.Errorf("Unexpected item ID processed: %d", item.ID)
		}
		delete(expectedIDs, item.ID)
	}

	if len(expectedIDs) != 0 {
		t.Errorf("Not all items were processed, missing: %v", expectedIDs)
	}
}

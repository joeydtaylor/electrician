package conduit_test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
	"github.com/joeydtaylor/electrician/pkg/internal/codec"
	"github.com/joeydtaylor/electrician/pkg/internal/conduit"
	"github.com/joeydtaylor/electrician/pkg/internal/wire"
)

type DummyType struct {
	Value string `json:"value"`
}

func dummyTransformer(input DummyType) (DummyType, error) {
	input.Value = "Transformed: " + input.Value
	return input, nil
}

type Feedback struct {
	CustomerID string   `json:"customerId"`
	Content    string   `json:"content"`
	Category   string   `json:"category,omitempty"`
	IsNegative bool     `json:"isNegative"`
	Tags       []string `json:"tags,omitempty"`
}

func TestTransformerErrorHandling(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errorTransformer := func(input DummyType) (DummyType, error) {
		if input.Value == "error" {
			return DummyType{}, fmt.Errorf("simulated error")
		}
		return dummyTransformer(input)
	}

	wire := wire.NewWire[DummyType](ctx, wire.WithEncoder[DummyType](&codec.LineEncoder[DummyType]{}))
	wire.ConnectTransformer(errorTransformer)

	conduit := conduit.NewConduit[DummyType](ctx)
	conduit.ConnectWire(wire)

	conduit.Start(ctx)

	// Submit an element that causes the transformer to error.
	conduit.Submit(ctx, DummyType{Value: "error"})

	// Submit a valid element to ensure the system continues processing.
	conduit.Submit(ctx, DummyType{Value: "Test"})

	time.Sleep(1 * time.Second) // Allow time for processing.

	conduit.Stop()

	// Load the output directly.
	outputBuffer := conduit.Load()
	outputString := outputBuffer.String()

	// Verify that the valid element was processed and the system handled the error gracefully.
	expectedOutput := "{Transformed: Test}"
	if !strings.Contains(outputString, expectedOutput) {
		t.Errorf("Expected output to contain '%s', got '%s'", expectedOutput, outputString)
	}
}

func errorSimulator(feedback Feedback) (Feedback, error) {
	if strings.Contains(strings.ToLower(feedback.Content), "error") {
		return Feedback{}, fmt.Errorf("simulated processing error")
	}
	return feedback, nil
}

func processor(feedback Feedback) (Feedback, error) {
	feedback.Content += "-PROCESSED"
	time.Sleep(1 * time.Second) // Simulate processing time
	return feedback, nil
}

func TestFeedbackProcessing(t *testing.T) {
	totalFeedbackCount := uint64(5000)
	var processedCount uint64
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	plug := builder.NewPlug[Feedback](
		ctx,
		builder.PlugWithAdapterFunc[Feedback](
			func(ctx context.Context, submitFunc func(ctx context.Context, feedback Feedback) error) {
				for i := 0; i < 5000; i++ {
					feedback := Feedback{
						CustomerID: fmt.Sprintf("Feedback%d", i),
						Content:    fmt.Sprintf("This is feedback item number %d", i),
					}
					if err := submitFunc(ctx, feedback); err != nil {
						return
					}
					if ctx.Err() != nil {
						return
					}
				}
			},
		),
	)

	onProcessed := func(c builder.ComponentMetadata, feedback Feedback) {
		atomic.AddUint64(&processedCount, 1)
		wg.Done() // Mark this feedback as processed
	}

	sensors := builder.NewSensor[Feedback](
		builder.SensorWithOnElementProcessedFunc[Feedback](onProcessed),
	)

	wg.Add(int(totalFeedbackCount)) // Expect to process this many feedback items

	wire := builder.NewWire[Feedback](
		ctx,
		builder.WireWithTransformer[Feedback](processor, errorSimulator),
		builder.WireWithGenerator[Feedback](builder.NewGenerator[Feedback](ctx, builder.GeneratorWithPlug[Feedback](plug))),
		builder.WireWithSensor[Feedback](sensors),
	)

	conduit := builder.NewConduit[Feedback](
		ctx,
		builder.ConduitWithWire[Feedback](wire),
		builder.ConduitWithConcurrencyControl[Feedback](100000, 100000),
	)

	conduit.Start(ctx)

	wg.Wait() // Wait here until all feedback is processed
	conduit.Stop()

	if processedCount != totalFeedbackCount {
		t.Errorf("Expected to process %d feedback items, but processed %d", totalFeedbackCount, processedCount)
	} else {
		t.Log("All feedback processed successfully.")
	}

	if ctx.Err() == context.DeadlineExceeded {
		t.Errorf("Processing did not complete within the expected timeframe")
	}
}

func TestConduitConnectPlug(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	plug := builder.NewPlug[DummyType](
		ctx,
		builder.PlugWithAdapterFunc[DummyType](func(ctx context.Context, submitFunc func(context.Context, DummyType) error) {
			_ = submitFunc(ctx, DummyType{Value: "Test"})
		}),
	)

	w := wire.NewWire[DummyType](
		ctx,
		wire.WithTransformer[DummyType](dummyTransformer),
		wire.WithGenerator[DummyType](
			builder.NewGenerator[DummyType](ctx, builder.GeneratorWithPlug[DummyType](plug)),
		),
	)

	con := conduit.NewConduit(ctx, conduit.WithWire[DummyType](w))
	con.Start(ctx)

	// Wait for exactly one output OR timeout.
	select {
	case <-ctx.Done():
		t.Fatalf("timeout waiting for output: %v", ctx.Err())
	case <-time.After(50 * time.Millisecond):
		// give generator a tick; optional
	}

	// Critical: stop so output channel closes, then load.
	con.Stop()

	output, err := con.LoadAsJSONArray()
	if err != nil {
		t.Fatalf("LoadAsJSONArray failed: %v", err)
	}

	var transformed []DummyType
	if err := json.Unmarshal(output, &transformed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(transformed) != 1 || transformed[0].Value != "Transformed: Test" {
		t.Fatalf("unexpected output: %s", string(output))
	}
}

// Define the feedback processing functions.
func negativeFilter(feedback Feedback) (Feedback, error) {
	negativeWords := []string{"bad", "terrible", "horrible", "worst"}
	for _, word := range negativeWords {
		if strings.Contains(strings.ToLower(feedback.Content), word) {
			feedback.IsNegative = true
			return feedback, nil
		}
	}
	return feedback, nil
}

func classifier(feedback Feedback) (Feedback, error) {
	if feedback.IsNegative {
		return feedback, nil
	}

	s := strings.ToLower(feedback.Content)

	// Deterministic precedence: Delivery > Customer Support > Product Quality > General
	switch {
	case strings.Contains(s, "delivery"):
		feedback.Category = "Delivery"
	case strings.Contains(s, "support"):
		feedback.Category = "Customer Support"
	case strings.Contains(s, "product"):
		feedback.Category = "Product Quality"
	default:
		feedback.Category = "General"
	}

	return feedback, nil
}

func sentimentAnalyzer(feedback Feedback) (Feedback, error) {
	positiveWords := []string{"love", "great", "happy"}
	for _, word := range positiveWords {
		if strings.Contains(strings.ToLower(feedback.Content), word) {
			feedback.Tags = append(feedback.Tags, "Positive Sentiment")
			return feedback, nil
		}
	}
	feedback.Tags = append(feedback.Tags, "Needs Attention")
	return feedback, nil
}

func plugFunc(ctx context.Context, submitFunc func(ctx context.Context, feedback Feedback) error) {
	feedbacks := []Feedback{
		{CustomerID: "C001", Content: "The delivery was fast and the product is amazing!", IsNegative: false},
		{CustomerID: "C002", Content: "I had a terrible experience with customer support.", IsNegative: true},
		{CustomerID: "C003", Content: "The product quality is bad, it broke on the first use.", IsNegative: true},
		{CustomerID: "C004", Content: "I'm really happy with the purchase. Great value for the price!", IsNegative: false},
	}
	for _, f := range feedbacks {
		if err := submitFunc(ctx, f); err != nil {
			fmt.Printf("Error submitting feedback: %v\n", err)
			return
		}
	}
}

// A helper function to sort an array of maps by a key.
func sortFeedbacks(feedbacks []map[string]interface{}) {
	sort.Slice(feedbacks, func(i, j int) bool {
		return feedbacks[i]["customerId"].(string) < feedbacks[j]["customerId"].(string)
	})
}

func TestChainedConduit(t *testing.T) {
	// Define the expected output
	expectedJSON := `[{"customerId":"C004","content":"I'm really happy with the purchase. Great value for the price!","category":"General","isNegative":false,"tags":["Positive Sentiment"]},{"customerId":"C001","content":"The delivery was fast and the product is amazing!","category":"Delivery","isNegative":false,"tags":["Needs Attention"]},{"customerId":"C002","content":"I had a terrible experience with customer support.","isNegative":true,"tags":["Needs Attention"]},{"customerId":"C003","content":"The product quality is bad, it broke on the first use.","isNegative":true,"tags":["Needs Attention"]}]`

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	plug := builder.NewPlug[Feedback](
		ctx,
		builder.PlugWithAdapterFunc[Feedback](plugFunc),
	)

	// Set up the generator, wire, and conduit as per the main logic
	generator := builder.NewGenerator[Feedback](
		ctx,
		builder.GeneratorWithPlug[Feedback](plug),
	)

	generatorWire := builder.NewWire[Feedback](
		ctx,
		builder.WireWithTransformer[Feedback](negativeFilter, classifier),
		builder.WireWithGenerator[Feedback](generator),
	)

	firstConduit := builder.NewConduit[Feedback](
		ctx,
		builder.ConduitWithWire[Feedback](generatorWire),
	)

	sentimentWire := builder.NewWire[Feedback](
		ctx,
		builder.WireWithTransformer[Feedback](sentimentAnalyzer),
	)

	secondConduit := builder.NewConduit[Feedback](
		ctx,
		builder.ConduitWithWire[Feedback](sentimentWire),
	)

	firstConduit.ConnectConduit(secondConduit)

	// Start conduits
	firstConduit.Start(ctx)
	secondConduit.Start(ctx)

	<-ctx.Done()

	// Fetch the output and compare with expected JSON
	output, err := secondConduit.LoadAsJSONArray()
	if err != nil {
		t.Fatalf("Error converting output to JSON: %v", err)
	}

	var got, want []map[string]interface{}
	if err := json.Unmarshal([]byte(output), &got); err != nil {
		t.Fatalf("Failed to unmarshal output JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(expectedJSON), &want); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}

	// Sort both slices before comparing
	sortFeedbacks(got)
	sortFeedbacks(want)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Unexpected results: got %v, want %v", got, want)
	}
}

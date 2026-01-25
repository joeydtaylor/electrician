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
	"github.com/joeydtaylor/electrician/pkg/internal/types"
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

func TestConduit_SubmitWithoutWires(t *testing.T) {
	ctx := context.Background()
	con := conduit.NewConduit[int](ctx)
	if err := con.Submit(ctx, 1); err == nil {
		t.Fatalf("expected error submitting with no wires")
	}
}

func TestConduit_InputChannelForwarding(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	w := wire.NewWire[int](ctx,
		wire.WithTransformer[int](func(v int) (int, error) { return v + 1, nil }),
		wire.WithConcurrencyControl[int](8, 1),
	)

	con := conduit.NewConduit[int](ctx, conduit.WithWire[int](w))
	if err := con.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer con.Stop()

	select {
	case con.GetInputChannel() <- 1:
	case <-ctx.Done():
		t.Fatalf("timeout submitting to input channel")
	}

	select {
	case got := <-con.GetOutputChannel():
		if got != 2 {
			t.Fatalf("expected 2, got %d", got)
		}
	case <-ctx.Done():
		t.Fatalf("timeout waiting for output")
	}
}

func TestConduit_NextConduitForwarding(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	firstWire := wire.NewWire[int](ctx,
		wire.WithTransformer[int](func(v int) (int, error) { return v + 1, nil }),
		wire.WithConcurrencyControl[int](8, 1),
	)
	secondWire := wire.NewWire[int](ctx,
		wire.WithTransformer[int](func(v int) (int, error) { return v * 2, nil }),
		wire.WithConcurrencyControl[int](8, 1),
	)

	first := conduit.NewConduit[int](ctx, conduit.WithWire[int](firstWire))
	second := conduit.NewConduit[int](ctx, conduit.WithWire[int](secondWire))
	first.ConnectConduit(second)

	if err := first.Start(ctx); err != nil {
		t.Fatalf("first Start() error: %v", err)
	}
	if err := second.Start(ctx); err != nil {
		t.Fatalf("second Start() error: %v", err)
	}
	defer first.Stop()
	defer second.Stop()

	if err := first.Submit(ctx, 2); err != nil {
		t.Fatalf("Submit() error: %v", err)
	}

	select {
	case got := <-second.GetOutputChannel():
		if got != 6 {
			t.Fatalf("expected 6, got %d", got)
		}
	case <-ctx.Done():
		t.Fatalf("timeout waiting for chained output")
	}
}

func TestConduit_CircuitBreakerNeutralWire(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mainWire := wire.NewWire[int](ctx, wire.WithConcurrencyControl[int](8, 1))
	neutralWire := wire.NewWire[int](ctx, wire.WithConcurrencyControl[int](1, 1))
	cb := &stubCircuitBreaker[int]{
		allow:   false,
		neutral: []types.Wire[int]{neutralWire},
		meta:    types.ComponentMetadata{Name: "cb", ID: "cb-1", Type: "CIRCUIT_BREAKER"},
	}

	con := conduit.NewConduit[int](ctx, conduit.WithWire[int](mainWire))
	con.ConnectCircuitBreaker(cb)

	if err := con.Submit(ctx, 9); err != nil {
		t.Fatalf("Submit() error: %v", err)
	}

	select {
	case got := <-neutralWire.GetInputChannel():
		if got != 9 {
			t.Fatalf("expected 9, got %d", got)
		}
	case <-ctx.Done():
		t.Fatalf("timeout waiting for neutral wire submission")
	}
}

func TestConduit_ConnectWireSetsOutput(t *testing.T) {
	ctx := context.Background()
	con := conduit.NewConduit[int](ctx)
	w := wire.NewWire[int](ctx)
	con.ConnectWire(w)

	if con.GetOutputChannel() != w.GetOutputChannel() {
		t.Fatalf("expected conduit output channel to match last wire output channel")
	}
}

func TestConduit_GetGenerators(t *testing.T) {
	ctx := context.Background()
	con := conduit.NewConduit[int](ctx)
	gen := &stubGenerator[int]{meta: types.ComponentMetadata{Type: "GENERATOR"}}
	con.ConnectGenerator(gen)

	generators := con.GetGenerators()
	if len(generators) != 1 || generators[0] != gen {
		t.Fatalf("expected generator to be registered")
	}
}

func TestConduit_ComponentMetadata(t *testing.T) {
	ctx := context.Background()
	con := conduit.NewConduit[int](ctx)
	con.SetComponentMetadata("name", "id")

	meta := con.GetComponentMetadata()
	if meta.Name != "name" || meta.ID != "id" || meta.Type != "CONDUIT" {
		t.Fatalf("unexpected metadata: %+v", meta)
	}
}

func TestConduit_ConfigurationPanicsAfterStart(t *testing.T) {
	ctx := context.Background()
	con := conduit.NewConduit[int](ctx).(*conduit.Conduit[int])

	if err := con.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer con.Stop()

	assertPanics(t, "ConnectCircuitBreaker", func() {
		con.ConnectCircuitBreaker(types.CircuitBreaker[int](nil))
	})
	assertPanics(t, "ConnectConduit", func() {
		con.ConnectConduit(types.Conduit[int](nil))
	})
	assertPanics(t, "ConnectGenerator", func() {
		con.ConnectGenerator(types.Generator[int](nil))
	})
	assertPanics(t, "ConnectLogger", func() {
		con.ConnectLogger(types.Logger(nil))
	})
	assertPanics(t, "ConnectSensor", func() {
		con.ConnectSensor(types.Sensor[int](nil))
	})
	assertPanics(t, "ConnectSurgeProtector", func() {
		con.ConnectSurgeProtector(types.SurgeProtector[int](nil))
	})
	assertPanics(t, "ConnectWire", func() {
		con.ConnectWire(types.Wire[int](nil))
	})
	assertPanics(t, "SetComponentMetadata", func() {
		con.SetComponentMetadata("name", "id")
	})
	assertPanics(t, "SetConcurrencyControl", func() {
		con.SetConcurrencyControl(1, 1)
	})
	assertPanics(t, "SetInputChannel", func() {
		con.SetInputChannel(make(chan int))
	})
}

func assertPanics(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic: %s", name)
		}
	}()
	fn()
}

type stubCircuitBreaker[T any] struct {
	allow   bool
	neutral []types.Wire[T]
	meta    types.ComponentMetadata
}

func (s *stubCircuitBreaker[T]) ConnectSensor(...types.Sensor[T]) {}
func (s *stubCircuitBreaker[T]) Allow() bool                      { return s.allow }
func (s *stubCircuitBreaker[T]) SetDebouncePeriod(int)            {}
func (s *stubCircuitBreaker[T]) ConnectNeutralWire(wires ...types.Wire[T]) {
	s.neutral = append(s.neutral, wires...)
}
func (s *stubCircuitBreaker[T]) ConnectLogger(types.Logger) {}
func (s *stubCircuitBreaker[T]) GetComponentMetadata() types.ComponentMetadata {
	return s.meta
}
func (s *stubCircuitBreaker[T]) GetNeutralWires() []types.Wire[T] { return s.neutral }
func (s *stubCircuitBreaker[T]) NotifyLoggers(types.LogLevel, string, ...interface{}) {
}
func (s *stubCircuitBreaker[T]) NotifyOnReset() <-chan struct{} { return make(chan struct{}) }
func (s *stubCircuitBreaker[T]) RecordError()                   {}
func (s *stubCircuitBreaker[T]) Reset()                         {}
func (s *stubCircuitBreaker[T]) SetComponentMetadata(name string, id string) {
	s.meta.Name = name
	s.meta.ID = id
}
func (s *stubCircuitBreaker[T]) Trip() { s.allow = false }

type stubGenerator[T any] struct {
	meta types.ComponentMetadata
}

func (s *stubGenerator[T]) ConnectCircuitBreaker(types.CircuitBreaker[T]) {}
func (s *stubGenerator[T]) Start(context.Context) error                   { return nil }
func (s *stubGenerator[T]) Stop() error                                   { return nil }
func (s *stubGenerator[T]) IsStarted() bool                               { return false }
func (s *stubGenerator[T]) Restart(context.Context) error                 { return nil }
func (s *stubGenerator[T]) ConnectPlug(...types.Plug[T])                  {}
func (s *stubGenerator[T]) NotifyLoggers(types.LogLevel, string, ...interface{}) {
}
func (s *stubGenerator[T]) ConnectLogger(...types.Logger) {}
func (s *stubGenerator[T]) ConnectToComponent(...types.Submitter[T]) {
}
func (s *stubGenerator[T]) GetComponentMetadata() types.ComponentMetadata { return s.meta }
func (s *stubGenerator[T]) SetComponentMetadata(name string, id string) {
	s.meta.Name = name
	s.meta.ID = id
}
func (s *stubGenerator[T]) ConnectSensor(...types.Sensor[T]) {}

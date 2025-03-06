package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

type Feedback struct {
	CustomerID string   `json:"customerId"`
	Content    string   `json:"content"`
	Category   string   `json:"category,omitempty"`
	IsNegative bool     `json:"isNegative"`
	Tags       []string `json:"tags,omitempty"`
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
	keywords := map[string]string{"delivery": "Delivery", "product": "Product Quality", "support": "Customer Support"}
	for keyword, category := range keywords {
		if strings.Contains(strings.ToLower(feedback.Content), keyword) {
			feedback.Category = category
			break
		}
	}
	if feedback.Category == "" {
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
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator

	for {
		for _, f := range feedbacks {
			select {
			case <-ctx.Done():
				return // Exit the function if context is cancelled
			case <-time.After(time.Duration(rand.Intn(1000)+500) * time.Millisecond): // Wait between 0.5s to 1.5s before each feedback
				if err := submitFunc(ctx, f); err != nil {
					// Consider whether to stop on error or continue to try to submit the next feedback
					continue // Continue submitting the next feedback
				}
			}
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	meter := builder.NewMeter[Feedback](ctx)
	sensor := builder.NewSensor[Feedback](builder.SensorWithMeter[Feedback](meter))

	surgeProtector := builder.NewSurgeProtector[Feedback](
		ctx,
		builder.SurgeProtectorWithSensor[Feedback](sensor),
	)

	plug := builder.NewPlug[Feedback](
		ctx,
		builder.PlugWithAdapterFunc[Feedback](plugFunc),
		builder.PlugWithSensor[Feedback](sensor),
	)

	generator := builder.NewGenerator[Feedback](
		ctx,
		builder.GeneratorWithPlug[Feedback](plug),
		builder.GeneratorWithSensor[Feedback](sensor),
	)

	generatorWire := builder.NewWire[Feedback](
		ctx,
		builder.WireWithTransformer[Feedback](negativeFilter, classifier),
		builder.WireWithGenerator[Feedback](generator),
	)

	firstConduit := builder.NewConduit[Feedback](
		ctx,
		builder.ConduitWithWire[Feedback](generatorWire),
		builder.ConduitWithSensor[Feedback](sensor),
		builder.ConduitWithSurgeProtector[Feedback](surgeProtector),
	)

	sentimentWire := builder.NewWire[Feedback](
		ctx,
		builder.WireWithTransformer[Feedback](sentimentAnalyzer),
	)

	secondConduit := builder.NewConduit[Feedback](
		ctx,
		builder.ConduitWithWire[Feedback](sentimentWire),
		builder.ConduitWithSensor[Feedback](sensor),
		builder.ConduitWithSurgeProtector[Feedback](surgeProtector),
	)

	firstConduit.ConnectConduit(secondConduit)
	secondConduit.Start(ctx)
	firstConduit.Start(ctx)

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

	firstConduit.Stop()
	secondConduit.Stop()

	output, err := secondConduit.LoadAsJSONArray()
	if err != nil {
		fmt.Printf("Error converting output to JSON: %v\n", err)
		return
	}

	fmt.Println("Feedback Analysis Summary:")
	fmt.Println(string(output))
}

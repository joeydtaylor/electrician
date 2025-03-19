package main

import (
	"context"
	"fmt"
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
	for _, f := range feedbacks {
		if err := submitFunc(ctx, f); err != nil {
			fmt.Printf("Error submitting feedback: %v\n", err)
			return
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	meter := builder.NewMeter[Feedback](ctx)
	sensor := builder.NewSensor(builder.SensorWithMeter[Feedback](meter))

	plug := builder.NewPlug(
		ctx,
		builder.PlugWithAdapterFunc(plugFunc),
		builder.PlugWithSensor(sensor),
	)

	generator := builder.NewGenerator(
		ctx,
		builder.GeneratorWithPlug(plug),
		builder.GeneratorWithSensor(sensor),
	)

	generatorWire := builder.NewWire(
		ctx,
		builder.WireWithSensor(sensor),
		builder.WireWithTransformer(negativeFilter, classifier),
		builder.WireWithGenerator(generator),
	)

	firstConduit := builder.NewConduit(
		ctx,
		builder.ConduitWithWire(generatorWire),
	)

	sentimentWire := builder.NewWire(
		ctx,
		builder.WireWithTransformer(sentimentAnalyzer),
		builder.WireWithSensor(sensor),
	)

	secondConduit := builder.NewConduit(
		ctx,
		builder.ConduitWithWire(sentimentWire),
		/* 		builder.ConduitWithSensor(sensor), */
	)

	firstConduit.ConnectConduit(secondConduit)
	firstConduit.Start(ctx)
	secondConduit.Start(ctx)

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

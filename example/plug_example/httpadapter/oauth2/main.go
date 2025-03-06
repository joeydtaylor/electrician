package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

// Define types corresponding to the JSON structure from JSONPlaceholder.
type Geo struct {
	Lat string `json:"lat"`
	Lng string `json:"lng"`
}

type Address struct {
	Street  string `json:"street"`
	Suite   string `json:"suite"`
	City    string `json:"city"`
	Zipcode string `json:"zipcode"`
	Geo     Geo    `json:"geo"`
}

type Company struct {
	Name        string `json:"name"`
	CatchPhrase string `json:"catchPhrase"`
	BS          string `json:"bs"`
}

type User struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Address  Address `json:"address"`
	Phone    string  `json:"phone"`
	Website  string  `json:"website"`
	Company  Company `json:"company"`
}

type Users []User

func main() {
	// Create a context with a 30-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Set up a Sensor to log key pipeline events.
	sensor := builder.NewSensor[Users](
		builder.SensorWithOnStartFunc[Users](func(c builder.ComponentMetadata) { fmt.Println("Wire started") }),
		builder.SensorWithOnElementProcessedFunc[Users](func(c builder.ComponentMetadata, elem Users) {
			fmt.Printf("Processed API response: %+v\n", elem)
		}),
		builder.SensorWithOnCancelFunc[Users](func(c builder.ComponentMetadata, elem Users) {
			fmt.Printf("Context cancelled processing API response: %+v\n", elem)
		}),
		builder.SensorWithOnErrorFunc[Users](func(c builder.ComponentMetadata, err error, elem Users) {
			fmt.Printf("Error processing API response: %+v, Error: %+v\n", elem, err)
		}),
		builder.SensorWithOnCompleteFunc[Users](func(c builder.ComponentMetadata) { fmt.Println("Processing complete") }),
		builder.SensorWithOnHTTPClientRequestStartFunc[Users](func(c builder.ComponentMetadata) { fmt.Println("HTTP Request started") }),
		builder.SensorWithOnHTTPClientResponseReceivedFunc[Users](func(c builder.ComponentMetadata) { fmt.Println("HTTP Response received") }),
		builder.SensorWithOnHTTPClientErrorFunc[Users](func(c builder.ComponentMetadata, err error) {
			fmt.Printf("HTTP Client error: %v\n", err)
		}),
		builder.SensorWithOnHTTPClientRequestCompleteFunc[Users](func(c builder.ComponentMetadata) { fmt.Println("HTTP Request completed") }),
	)

	// Create a Logger in development mode at debug level.
	logger := builder.NewLogger(builder.LoggerWithDevelopment(true), builder.LoggerWithLevel("debug"))

	// OAuth client credentials for demonstration purposes.
	clientID := "b9gmCiSH15KsmUD0Bc5C767e8xap14q2"
	clientSecret := "EXEsOit6gfw6-Ey4nSELZPoyk3klVWwVssQ5Ik1IWJx-bm8xDYLx2Az6uGWkPYog"
	tokenURL := "https://dev-8qbukrpqlhxkb84a.us.auth0.com/oauth/token"

	// Create an HTTP Adapter that fetches data from JSONPlaceholder.
	// It uses TLS pinning via a certificate chain file ("./chain.crt") and includes OAuth2 client credentials.
	httpAdapter := builder.NewHTTPClientAdapter[Users](
		ctx,
		builder.HTTPClientAdapterWithLogger[Users](logger),
		builder.HTTPClientAdapterWithRequestConfig[Users]("GET", "https://jsonplaceholder.typicode.com/users", nil),
		builder.HTTPClientAdapterWithInterval[Users](4*time.Second),
		builder.HTTPClientAdapterWithTimeout[Users](10*time.Second),
		builder.HTTPClientAdapterWithOAuth2ClientCredentials[Users](clientID, clientSecret, tokenURL, "https://dev-8qbukrpqlhxkb84a.us.auth0.com/api/v2/"),
		builder.HTTPClientAdapterWithSensor[Users](sensor),
		builder.HTTPClientAdapterWithTLSPinning[Users]("./chain.crt"),
	)

	// Set up a Circuit Breaker that trips after 1 error and resets after 4 seconds.
	circuitBreaker := builder.NewCircuitBreaker[Users](
		context.Background(),
		1,
		4*time.Second,
		builder.CircuitBreakerWithLogger[Users](logger),
	)

	// Define a transformer that converts each user's name to uppercase.
	transformer := func(users Users) (Users, error) {
		for i := range users {
			users[i].Name = strings.ToUpper(users[i].Name)
		}
		return users, nil
	}

	// Create a Plug using the HTTP Adapter.
	plug := builder.NewPlug[Users](
		ctx,
		builder.PlugWithAdapter[Users](httpAdapter),
		builder.PlugWithSensor[Users](sensor),
	)

	// Initialize a Generator that uses the Plug, Logger, and Circuit Breaker.
	generator := builder.NewGenerator[Users](
		ctx,
		builder.GeneratorWithPlug[Users](plug),
		builder.GeneratorWithLogger[Users](logger),
		builder.GeneratorWithCircuitBreaker[Users](circuitBreaker),
	)

	// Create a Wire that uses the Sensor, Transformer, and Generator.
	wire := builder.NewWire[Users](
		ctx,
		builder.WireWithLogger[Users](logger),
		builder.WireWithSensor[Users](sensor),
		builder.WireWithTransformer[Users](transformer),
		builder.WireWithGenerator[Users](generator),
	)

	// Start the pipeline.
	wire.Start(ctx)

	// Wait for the context to finish.
	<-ctx.Done()

	// Stop the pipeline.
	wire.Stop()

	// Load the processed output as a JSON array.
	output, err := wire.LoadAsJSONArray()
	if err != nil {
		fmt.Printf("Error converting output to JSON: %v\n", err)
		return
	}

	fmt.Println("Output Summary:")
	fmt.Println(string(output))
}

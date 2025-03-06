package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

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

	// Create a Meter to track performance (expecting 10 items, for example).
	meter := builder.NewMeter[Users](ctx, builder.MeterWithTotalItems[Users](10))

	// Create a Sensor and attach the Meter.
	sensor := builder.NewSensor[Users](builder.SensorWithMeter[Users](meter))

	// Setup a Circuit Breaker that trips after 1 error and resets after 4 seconds.
	circuitBreaker := builder.NewCircuitBreaker[Users](
		ctx,
		1,
		4*time.Second,
		builder.CircuitBreakerWithSensor[Users](sensor),
	)

	// Create an HTTP Adapter to fetch data from JSONPlaceholder,
	// using certificate pinning (with a local certificate chain file "chain.crt").
	httpAdapter := builder.NewHTTPClientAdapter[Users](
		ctx,
		builder.HTTPClientAdapterWithSensor[Users](sensor),
		builder.HTTPClientAdapterWithTLSPinning[Users]("./chain.crt"),
		builder.HTTPClientAdapterWithRequestConfig[Users]("GET", "https://jsonplaceholder.typicode.com/users", nil),
		builder.HTTPClientAdapterWithInterval[Users](4*time.Second),
		builder.HTTPClientAdapterWithTimeout[Users](10*time.Second),
	)

	// Define a transformer that converts each user's name to uppercase.
	transformer := func(users Users) (Users, error) {
		for i := range users {
			users[i].Name = strings.ToUpper(users[i].Name)
		}
		return users, nil
	}

	// Create a Plug that uses the HTTP Adapter and Sensor.
	plug := builder.NewPlug[Users](
		ctx,
		builder.PlugWithAdapter[Users](httpAdapter),
		builder.PlugWithSensor[Users](sensor),
	)

	// Initialize a Generator that pulls data from the Plug and attaches the Sensor and Circuit Breaker.
	generator := builder.NewGenerator[Users](
		ctx,
		builder.GeneratorWithPlug[Users](plug),
		builder.GeneratorWithSensor[Users](sensor),
		builder.GeneratorWithCircuitBreaker[Users](circuitBreaker),
	)

	// Create a Wire that uses the Sensor, Transformer, and Generator.
	wire := builder.NewWire[Users](
		ctx,
		builder.WireWithSensor[Users](sensor),
		builder.WireWithTransformer[Users](transformer),
		builder.WireWithGenerator[Users](generator),
	)

	// Start the pipeline.
	wire.Start(ctx)

	// Monitor performance metrics.
	meter.Monitor()

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

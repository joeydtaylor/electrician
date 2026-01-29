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

const (
	pgConnString = "postgres://REPLACE_WITH_USER:REPLACE_WITH_PASS@localhost:5432/REPLACE_WITH_DB?sslmode=verify-full"
	pgDriver     = "pgx"
	tableName    = "electrician_events"

	clientSideKeyHex = "REPLACE_WITH_32_BYTE_HEX_KEY"
)

func mustSet(label, v string) string {
	if strings.Contains(v, "REPLACE_WITH") {
		panic(fmt.Sprintf("%s must be set before running", label))
	}
	return v
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log := builder.NewLogger(builder.LoggerWithDevelopment(true))

	adapter := builder.NewPostgresClientAdapter[Feedback](
		ctx,
		builder.PostgresAdapterWithConnString[Feedback](mustSet("pgConnString", pgConnString), pgDriver),
		builder.PostgresAdapterWithTable[Feedback](tableName),
		builder.PostgresAdapterWithAutoCreateTable[Feedback](true),
		builder.PostgresAdapterWithSecureDefaults[Feedback](mustSet("clientSideKeyHex", clientSideKeyHex)),
		builder.PostgresAdapterWithLogger[Feedback](log),
	)

	in := make(chan Feedback, 2)
	in <- Feedback{CustomerID: "cust-001", Content: "secure postgres payload", Tags: []string{"postgres", "secure"}}
	in <- Feedback{CustomerID: "cust-002", Content: "second record", Category: "feedback"}
	close(in)

	if err := adapter.ServeWriter(ctx, in); err != nil {
		panic(err)
	}
	fmt.Println("Postgres write complete.")
}

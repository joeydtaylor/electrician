//go:build webtransport

package main

import (
	"context"
	"crypto/tls"
	"fmt"
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

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger := builder.NewLogger(builder.LoggerWithDevelopment(true))

	tlsCfg := builder.NewTlsClientConfig(
		true,
		"../tls/client.crt",
		"../tls/client.key",
		"../tls/ca.crt",
		tls.VersionTLS13,
		tls.VersionTLS13,
	)

	fr := builder.NewWebTransportForwardRelay[Feedback](
		ctx,
		builder.WebTransportForwardRelayWithTarget[Feedback]("https://localhost:8443/relay"),
		builder.WebTransportForwardRelayWithLogger[Feedback](logger),
		builder.WebTransportForwardRelayWithPayloadFormat[Feedback]("json"),
		builder.WebTransportForwardRelayWithTLSConfig[Feedback](tlsCfg),
	)

	for i := 0; i < 5; i++ {
		fb := Feedback{
			CustomerID: fmt.Sprintf("cust-%02d", i),
			Content:    "WebTransport payload",
			Category:   "feedback",
			IsNegative: i%2 == 0,
			Tags:       []string{"webtransport"},
		}
		if err := fr.Submit(ctx, fb); err != nil {
			logger.Error("Submit failed", "error", err, "seq", i)
		}
		time.Sleep(100 * time.Millisecond)
	}
	logger.Info("Send complete")
}

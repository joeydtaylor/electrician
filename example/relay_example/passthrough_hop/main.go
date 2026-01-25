package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
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

const AES256KeyHex = "ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec"

func mustAES() string {
	raw, err := hex.DecodeString(AES256KeyHex)
	if err != nil || len(raw) != 32 {
		log.Fatalf("bad AES key: %v", err)
	}
	return string(raw)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigs)
	go func() {
		<-sigs
		fmt.Println("Shutting down...")
		cancel()
	}()

	logger := builder.NewLogger(builder.LoggerWithDevelopment(true))
	key := mustAES()
	sec := builder.NewSecurityOptions(true, builder.ENCRYPTION_AES_GCM)
	perf := builder.NewPerformanceOptions(true, builder.COMPRESS_SNAPPY)

	// Core wire and receiver (final hop, does decrypt + unwrap).
	coreWire := builder.NewWire[Feedback](
		ctx,
		builder.WireWithLogger[Feedback](logger),
	)
	_ = coreWire.Start(ctx)

	coreRecv := builder.NewReceivingRelay[Feedback](
		ctx,
		builder.ReceivingRelayWithAddress[Feedback]("localhost:50052"),
		builder.ReceivingRelayWithBufferSize[Feedback](4096),
		builder.ReceivingRelayWithLogger[Feedback](logger),
		builder.ReceivingRelayWithOutput[Feedback](coreWire),
		builder.ReceivingRelayWithDecryptionKey[Feedback](key),
	)

	// Mid-hop passthrough wire + relay (forwards raw WrappedPayload).
	rawWire := builder.NewWire[builder.WrappedPayload](
		ctx,
		builder.WireWithLogger[builder.WrappedPayload](logger),
		builder.WireWithTransformer(func(p builder.WrappedPayload) (builder.WrappedPayload, error) {
			return p, nil
		}),
	)
	_ = rawWire.Start(ctx)

	rawForward := builder.NewForwardRelay[builder.WrappedPayload](
		ctx,
		builder.ForwardRelayWithLogger[builder.WrappedPayload](logger),
		builder.ForwardRelayWithTarget[builder.WrappedPayload]("localhost:50052"),
		builder.ForwardRelayWithInput[builder.WrappedPayload](rawWire),
		builder.ForwardRelayWithPassthrough[builder.WrappedPayload](true),
	)

	rawRecv := builder.NewReceivingRelay[builder.WrappedPayload](
		ctx,
		builder.ReceivingRelayWithAddress[builder.WrappedPayload]("localhost:50051"),
		builder.ReceivingRelayWithBufferSize[builder.WrappedPayload](4096),
		builder.ReceivingRelayWithLogger[builder.WrappedPayload](logger),
		builder.ReceivingRelayWithOutput[builder.WrappedPayload](rawWire),
		builder.ReceivingRelayWithPassthrough[builder.WrappedPayload](true),
	)

	// Edge wire + forwarder (wraps + encrypts once at the edge).
	edgeWire := builder.NewWire[Feedback](ctx)
	_ = edgeWire.Start(ctx)

	edgeForward := builder.NewForwardRelay[Feedback](
		ctx,
		builder.ForwardRelayWithLogger[Feedback](logger),
		builder.ForwardRelayWithTarget[Feedback]("localhost:50051"),
		builder.ForwardRelayWithInput[Feedback](edgeWire),
		builder.ForwardRelayWithPerformanceOptions[Feedback](perf),
		builder.ForwardRelayWithSecurityOptions[Feedback](sec, key),
	)

	if err := coreRecv.Start(ctx); err != nil {
		log.Fatal(err)
	}
	if err := rawRecv.Start(ctx); err != nil {
		log.Fatal(err)
	}
	if err := rawForward.Start(ctx); err != nil {
		log.Fatal(err)
	}
	if err := edgeForward.Start(ctx); err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		item := Feedback{
			CustomerID: fmt.Sprintf("C-%d", i+1),
			Content:    fmt.Sprintf("passthrough hop %d", i+1),
			Category:   "demo",
		}
		if err := edgeWire.Submit(ctx, item); err != nil {
			log.Printf("submit failed: %v", err)
		}
	}

	time.Sleep(2 * time.Second)
	cancel()

	rawRecv.Stop()
	rawForward.Stop()
	coreRecv.Stop()
	edgeForward.Stop()

	if out, err := coreWire.LoadAsJSONArray(); err == nil {
		fmt.Println("Core output:")
		fmt.Println(string(out))
	} else {
		fmt.Println("core wire err:", err)
	}
}

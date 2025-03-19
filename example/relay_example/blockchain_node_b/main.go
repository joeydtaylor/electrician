package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

// Block represents a basic block in our blockchain.
type Block struct {
	Index     int       `json:"index"`
	Timestamp time.Time `json:"timestamp"`
	Data      string    `json:"data"`
	PrevHash  string    `json:"prevHash"`
	Hash      string    `json:"hash"`
}

// processBlock is a transformer that processes an incoming block.
// It validates the block and adds it to our local chain.
func processBlock(block Block) (Block, error) {
	fmt.Printf("\nReceived Block #%d with hash: %s\n", block.Index, block.Hash)
	fmt.Printf("  Data: %s\n", block.Data)
	fmt.Printf("  Previous Hash: %s\n", block.PrevHash)
	fmt.Printf("  Timestamp: %s\n", block.Timestamp.Format(time.RFC3339))

	return block, nil
}

func main() {
	fmt.Println("Starting Blockchain Node...")
	fmt.Println("Listening for blocks on localhost:50051")

	// Set up cancellation context with signal handling for graceful shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle termination signals.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		fmt.Println("\nNode: Received termination signal, shutting down...")
		cancel()
	}()

	// Create a logger with development mode enabled.
	logger := builder.NewLogger(builder.LoggerWithDevelopment(true))

	// TLS server configuration for secure communication.
	tlsConfig := builder.NewTlsServerConfig(
		true,                // UseTLS
		"../tls/server.crt", // Server certificate
		"../tls/server.key", // Server private key
		"../tls/ca.crt",     // CA certificate
		"localhost",
		tls.VersionTLS13, // MinVersion: Only allow TLS 1.3
		tls.VersionTLS13, // MaxVersion: Only allow TLS 1.3
	)

	// Create a wire for processing incoming blocks with our transformer.
	blockWire := builder.NewWire(
		ctx,
		builder.WireWithTransformer(processBlock),
		builder.WireWithLogger[Block](logger),
	)

	// Create an output conduit from the block wire.
	outputConduit := builder.NewConduit(
		ctx,
		builder.ConduitWithWire(blockWire),
	)

	// Receiving relay to listen for blocks from the hub.
	receivingRelay := builder.NewReceivingRelay(
		ctx,
		builder.ReceivingRelayWithAddress[Block]("localhost:50052"), // Listen on port 50051.
		builder.ReceivingRelayWithBufferSize[Block](1000),
		builder.ReceivingRelayWithLogger[Block](logger),
		builder.ReceivingRelayWithOutput(outputConduit),
		builder.ReceivingRelayWithTLSConfig[Block](tlsConfig),
	)

	// Start the receiving relay.
	fmt.Println("Starting receiving relay...")
	receivingRelay.Start(ctx)
	fmt.Println("Receiving relay started. Waiting for blocks...")

	// Setup a goroutine to periodically display our blockchain
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Load all received blocks from the conduit
				blocksJSON, err := outputConduit.LoadAsJSONArray()
				if err != nil {
					fmt.Printf("Error loading blocks: %v\n", err)
					continue
				}

				var blocks []Block
				if err := json.Unmarshal([]byte(blocksJSON), &blocks); err != nil {
					fmt.Printf("Error unmarshalling blocks: %v\n", err)
					continue
				}

				fmt.Printf("\n=== Current Blockchain State ===\n")
				fmt.Printf("Total blocks: %d\n", len(blocks))

				// Show the last 3 blocks
				displayCount := 3
				if len(blocks) < displayCount {
					displayCount = len(blocks)
				}

				if displayCount > 0 {
					fmt.Printf("Last %d blocks:\n", displayCount)
					for i := len(blocks) - displayCount; i < len(blocks); i++ {
						fmt.Printf("  Block #%d: %s\n", blocks[i].Index, blocks[i].Hash)
					}
				}
				fmt.Println("===============================")
			}
		}
	}()

	// Block until the context is cancelled.
	<-ctx.Done()

	// Aggregate the received blocks before shutting down.
	blocksJSON, err := outputConduit.LoadAsJSONArray()
	if err != nil {
		fmt.Printf("Error loading final blockchain: %v\n", err)
		return
	}

	var blocks []Block
	if err := json.Unmarshal([]byte(blocksJSON), &blocks); err != nil {
		fmt.Printf("Error unmarshalling final blockchain: %v\n", err)
		return
	}

	fmt.Println("\nNode Shutdown: Final Blockchain State")
	fmt.Printf("Total blocks received: %d\n", len(blocks))

	// Print detailed information about each block in the final chain
	if len(blocks) > 0 {
		fmt.Println("\nComplete Blockchain:")
		for i, block := range blocks {
			fmt.Printf("Block #%d:\n", i)
			fmt.Printf("  Index: %d\n", block.Index)
			fmt.Printf("  Hash: %s\n", block.Hash)
			fmt.Printf("  PrevHash: %s\n", block.PrevHash)
			fmt.Printf("  Data: %s\n", block.Data)
			fmt.Printf("  Timestamp: %s\n", block.Timestamp.Format(time.RFC3339))
			fmt.Println()
		}
	}
}

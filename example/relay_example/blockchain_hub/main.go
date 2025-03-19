package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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

// calculateHash computes the SHA256 hash of a block's content.
func calculateHash(block Block) string {
	record := fmt.Sprintf("%d%s%s%s", block.Index, block.Timestamp.String(), block.Data, block.PrevHash)
	h := sha256.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}

// mineBlock creates a new block based on the previous block and some data.
func mineBlock(prevBlock Block, data string) Block {
	newBlock := Block{
		Index:     prevBlock.Index + 1,
		Timestamp: time.Now(),
		Data:      data,
		PrevHash:  prevBlock.Hash,
	}
	newBlock.Hash = calculateHash(newBlock)
	return newBlock
}

// plugFunc is used by the generator to produce new blocks.
func plugFunc(ctx context.Context, submitFunc func(ctx context.Context, block Block) error) {
	// Create a genesis block.
	genesisBlock := Block{
		Index:     0,
		Timestamp: time.Now(),
		Data:      "Genesis Block",
		PrevHash:  "",
	}
	genesisBlock.Hash = calculateHash(genesisBlock)

	fmt.Println("Created genesis block with hash:", genesisBlock.Hash)

	// Submit the genesis block.
	if err := submitFunc(ctx, genesisBlock); err != nil {
		fmt.Printf("Error submitting genesis block: %v\n", err)
	} else {
		fmt.Println("Successfully submitted genesis block")
	}

	previousBlock := genesisBlock
	blockCount := 1

	ticker := time.NewTicker(5 * time.Second) // Generate a block every 5 seconds.
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Mine a new block with dummy data.
			data := fmt.Sprintf("Block data for block %d", blockCount)
			newBlock := mineBlock(previousBlock, data)
			fmt.Printf("Mined new block: Index %d, Hash %s\n", newBlock.Index, newBlock.Hash)

			if err := submitFunc(ctx, newBlock); err != nil {
				fmt.Printf("Error submitting block %d: %v\n", blockCount, err)
			} else {
				fmt.Printf("Successfully submitted block %d to wire\n", blockCount)
			}
			previousBlock = newBlock
			blockCount++
		}
	}
}

func main() {
	fmt.Println("Starting Blockchain Hub...")

	// Set up a context with a timeout for this prototype.
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Create a logger with development mode enabled.
	logger := builder.NewLogger(builder.LoggerWithDevelopment(true))

	// Create a plug for generating blocks.
	plug := builder.NewPlug(
		ctx,
		builder.PlugWithAdapterFunc(plugFunc),
	)

	// Create a generator that uses the plug.
	generator := builder.NewGenerator(
		ctx,
		builder.GeneratorWithPlug(plug),
	)

	// Create a wire that uses the generator.
	generatorWire := builder.NewWire(
		ctx,
		builder.WireWithLogger[Block](logger),
		builder.WireWithGenerator(generator),
	)

	// TLS configuration for secure communication.
	tlsConfig := builder.NewTlsClientConfig(
		true,                // UseTLS
		"../tls/client.crt", // Path to the client's certificate
		"../tls/client.key", // Path to the client's private key
		"../tls/ca.crt",     // Path to the CA certificate
	)

	// Forward relay that sends blocks to the node.
	// Using generatorWire directly as the input
	forwardRelay := builder.NewForwardRelay(
		ctx,
		builder.ForwardRelayWithLogger[Block](logger),
		builder.ForwardRelayWithTarget[Block]("localhost:50051", "localhost:50052"), // Node address.
		builder.ForwardRelayWithInput(generatorWire),                                // Connect to the generator wire
		builder.ForwardRelayWithTLSConfig[Block](tlsConfig),
	)

	// Start the forward relay.
	fmt.Println("Starting generator and relay...")
	generatorWire.Start(ctx)
	forwardRelay.Start(ctx)
	fmt.Println("Generator and relay started. Sending blocks to node at localhost:50051")

	// Wait for processing to finish (or timeout).
	<-ctx.Done()

	// Clean up.
	fmt.Println("Stopping generator and relay...")
	generatorWire.Stop()
	forwardRelay.Stop()

	if ctx.Err() == context.DeadlineExceeded {
		fmt.Println("Hub processing timeout after 2 minutes.")
	} else {
		fmt.Println("Hub processing finished.")
	}
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	// These imports assume your Electrician library layout:
	"github.com/joeydtaylor/electrician/pkg/builder"
)

// ExampleAES256Key is a 32-byte (256-bit) key for AES-GCM.
const ExampleAES256Key = "0123456789abcdefghijklmnopqrstuv" // 32 chars

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	//---------------------------------------
	// 1) FIRST WIRE: transform + encrypt
	//---------------------------------------
	fmt.Println("1) Creating first wire to UPPERCASE + encrypt...")

	// A simple transform that uppercases the string
	uppercaseTransform := func(input string) (string, error) {
		return strings.ToUpper(input), nil
	}

	// Enable AES-GCM for outbound encryption
	secOptsEnc := builder.NewSecurityOptions(true, builder.ENCRYPTION_AES_GCM)

	// Build the first wire
	firstWire := builder.NewWire(
		ctx,
		builder.WireWithTransformer(uppercaseTransform),
		builder.WireWithEncryptOptions[string](secOptsEnc, ExampleAES256Key),
	)

	// Start the first wire
	if err := firstWire.Start(ctx); err != nil {
		panic(fmt.Sprintf("Failed to start first wire: %v", err))
	}

	// Submit some data
	firstWire.Submit(ctx, "hello")
	firstWire.Submit(ctx, "world")

	// Wait for the context or short pause
	time.Sleep(500 * time.Millisecond)
	fmt.Println("Stopping first wire...")
	firstWire.Stop()

	// Load final output as JSON array of Base64 ciphertext
	ciphertextJSON, err := firstWire.LoadAsJSONArray()
	if err != nil {
		fmt.Printf("Error converting first wire output to JSON: %v\n", err)
		return
	}
	fmt.Println("FIRST WIRE OUTPUT (Base64 ciphertext array):")
	fmt.Println(string(ciphertextJSON))

	//---------------------------------------
	// 2) SECOND WIRE: decrypt to plaintext
	//---------------------------------------
	fmt.Println("\n2) Creating second wire to DECRYPT inbound data...")

	// Enable AES-GCM inbound decryption
	secOptsDec := builder.NewSecurityOptions(true, builder.ENCRYPTION_AES_GCM)

	// Build the second wire
	secondWire := builder.NewWire(
		ctx,
		// We do NOT add any transform here, just pass data through after decryption
		builder.WireWithDecryptOptions[string](secOptsDec, ExampleAES256Key),
	)

	// Start second wire
	if err := secondWire.Start(ctx); err != nil {
		panic(fmt.Sprintf("Failed to start second wire: %v", err))
	}

	// Unmarshal the first wire's JSON output into a slice of strings
	var ciphertexts []string
	if err := json.Unmarshal(ciphertextJSON, &ciphertexts); err != nil {
		fmt.Printf("Error unmarshaling ciphertext JSON: %v\n", err)
		return
	}

	// Submit each Base64 ciphertext string to the second wire
	for _, ciph := range ciphertexts {
		secondWire.Submit(ctx, ciph)
	}

	// Wait briefly so second wire can process
	time.Sleep(500 * time.Millisecond)
	fmt.Println("Stopping second wire...")
	secondWire.Stop()

	// Load second wire's final output as JSON array
	plaintextJSON, err := secondWire.LoadAsJSONArray()
	if err != nil {
		fmt.Printf("Error converting second wire output to JSON: %v\n", err)
		return
	}

	fmt.Println("SECOND WIRE OUTPUT (Decrypted Plaintext Array):")
	fmt.Println(string(plaintextJSON))
}

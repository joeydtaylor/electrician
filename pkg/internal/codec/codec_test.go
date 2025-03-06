package codec_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/joeydtaylor/electrician/pkg/internal/codec"
)

// DummyType is a placeholder type for testing purposes.
type DummyType struct {
	Value string `json:"value"`
}

// TestJSONEncodingDecoding tests JSON encoding and decoding.
func TestJSONEncodingDecoding(t *testing.T) {
	// Create a dummy value
	dummyValue := DummyType{Value: "Test"}

	// Encode the dummy value to JSON
	var encoded bytes.Buffer
	jsonEncoder := codec.JSONEncoder[DummyType]{}
	if err := jsonEncoder.Encode(&encoded, dummyValue); err != nil {
		t.Errorf("JSON encoding error: %v", err)
	}

	// Decode the JSON into a dummy value
	var decoded DummyType
	var jsonReader io.Reader = &encoded // Convert encoded buffer to io.Reader
	jsonDecoder := codec.JSONDecoder[DummyType]{}
	decoded, err := jsonDecoder.Decode(jsonReader)
	if err != nil {
		t.Errorf("JSON decoding error: %v", err)
	}

	// Check if the decoded value matches the original
	if decoded != dummyValue {
		t.Errorf("Decoded value does not match original. Expected: %v, Got: %v", dummyValue, decoded)
	}
}

// TestJSONSliceEncodingDecoding tests JSON slice encoding and decoding.
func TestJSONSliceEncodingDecoding(t *testing.T) {
	// Create a slice of dummy values
	dummySlice := []DummyType{{Value: "Test1"}, {Value: "Test2"}}

	// Create JSON encoder and decoder for DummyType (not a slice of DummyType)
	jsonEncoder := codec.NewJSONEncoder[DummyType]() // Handling DummyType
	jsonDecoder := codec.NewJSONDecoder[DummyType]() // Handling DummyType

	// Encode the dummy slice to JSON
	var encoded bytes.Buffer
	if err := jsonEncoder.EncodeSlice(&encoded, dummySlice); err != nil {
		t.Errorf("JSON slice encoding error: %v", err)
	}

	// Decode the JSON back into a slice of DummyType
	decodedSlice, err := jsonDecoder.DecodeSlice(&encoded)
	if err != nil {
		t.Errorf("JSON slice decoding error: %v", err)
	}

	// Check if the decoded slice matches the original
	if len(decodedSlice) != len(dummySlice) {
		t.Fatalf("Mismatch in slice lengths. Expected %d, got %d", len(dummySlice), len(decodedSlice))
	}
	for i, item := range decodedSlice {
		if item.Value != dummySlice[i].Value {
			t.Errorf("Decoded slice item at index %d does not match original. Expected: %v, Got: %v", i, dummySlice[i], item)
		}
	}
}

// TestLineEncodingDecoding tests line encoding and decoding.
func TestLineEncodingDecoding(t *testing.T) {
	// Create a dummy value
	dummyValue := "Test"

	// Encode the dummy value to a line
	var encoded bytes.Buffer
	lineEncoder := codec.NewLineEncoder[string]() // Create a new LineEncoder instance
	if err := lineEncoder.Encode(&encoded, dummyValue); err != nil {
		t.Errorf("Line encoding error: %v", err)
	}

	// Decode the line into a string
	lineDecoder := codec.NewLineDecoder() // Create a new LineDecoder instance
	decoded, err := lineDecoder.Decode(&encoded)
	if err != nil {
		t.Errorf("Line decoding error: %v", err)
	}

	// Check if the decoded value matches the original
	if decoded != dummyValue {
		t.Errorf("Decoded value does not match original. Expected: %v, Got: %v", dummyValue, decoded)
	}
}

// TestJSONSliceEncodingDecodingEmpty tests JSON slice encoding and decoding for an empty slice.
func TestJSONSliceEncodingDecodingEmpty(t *testing.T) {
	// Create an empty slice of dummy values
	var dummySlice []DummyType
	jsonEncoder := codec.NewJSONEncoder[DummyType]() // Create a new LineEncoder instance
	jsonDecoder := codec.NewJSONDecoder[DummyType]() // Create a new LineEncoder instance

	// Encode the empty slice to JSON
	var encoded bytes.Buffer
	if err := jsonEncoder.EncodeSlice(&encoded, dummySlice); err != nil {
		t.Errorf("JSON slice encoding error for empty slice: %v", err)
	}

	// Decode the JSON back into a slice of DummyType
	decodedSlice, err := jsonDecoder.DecodeSlice(&encoded)
	if err != nil {
		t.Errorf("JSON slice decoding error: %v", err)
	}

	// Check if the decoded slice is empty
	if len(decodedSlice) != 0 {
		t.Errorf("Decoded slice for empty slice should be empty, but got: %v", decodedSlice)
	}
}

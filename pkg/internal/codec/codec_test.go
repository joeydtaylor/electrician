package codec_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/joeydtaylor/electrician/pkg/internal/codec"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
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

func TestJSONDecodeInvalid(t *testing.T) {
	decoder := codec.NewJSONDecoder[DummyType]()
	_, err := decoder.Decode(strings.NewReader("{invalid json"))
	if err == nil {
		t.Fatalf("expected JSON decode error")
	}
}

func TestTextDecoder(t *testing.T) {
	decoder := codec.NewTextDecoder()
	got, err := decoder.Decode(strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if got != "hello" {
		t.Fatalf("expected hello, got %q", got)
	}
}

func TestBinaryDecoder(t *testing.T) {
	decoder := codec.NewBinaryDecoder()
	got, err := decoder.Decode(strings.NewReader("data"))
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if string(got) != "data" {
		t.Fatalf("expected data, got %q", string(got))
	}
}

func TestLineDecoderEmpty(t *testing.T) {
	decoder := codec.NewLineDecoder()
	_, err := decoder.Decode(strings.NewReader(""))
	if err == nil {
		t.Fatalf("expected EOF on empty input")
	}
}

func TestLineDecoderLongLine(t *testing.T) {
	decoder := codec.NewLineDecoder()
	longLine := strings.Repeat("a", 70000) + "\n"
	_, err := decoder.Decode(strings.NewReader(longLine))
	if err == nil {
		t.Fatalf("expected error for long line")
	}
}

func TestXMLCodec(t *testing.T) {
	type xmlItem struct {
		XMLName string `xml:"item"`
		Name    string `xml:"name"`
	}

	encoder := codec.NewXMLEncoder[xmlItem]()
	decoder := codec.NewXMLDecoder[xmlItem]()

	var buf bytes.Buffer
	item := xmlItem{Name: "alpha"}
	if err := encoder.Encode(&buf, item); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	got, err := decoder.Decode(&buf)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if got.Name != item.Name {
		t.Fatalf("expected %q, got %q", item.Name, got.Name)
	}
}

func TestHTMLDecoder(t *testing.T) {
	decoder := codec.NewHTMLDecoder()
	node, err := decoder.Decode(strings.NewReader("<html><body><p>hi</p></body></html>"))
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if node == nil || node.FirstChild == nil {
		t.Fatalf("expected parsed HTML node")
	}
}

func TestWaveCodec(t *testing.T) {
	encoder := codec.WaveEncoder[types.WaveData]{}
	decoder := codec.WaveDecoder[types.WaveData]{}

	wave := types.WaveData{
		ID:               7,
		OriginalWave:     []complex128{complex(1, 2), complex(3, 4)},
		CompressedHex:    "deadbeef",
		DominantFreq:     3.14,
		TotalEnergy:      9.9,
		CompressionRatio: 0.5,
		MSE:              1.2,
		SNR:              2.3,
		FrequencyPeaks: []types.Peak{
			{Freq: 1.0, Value: 2.0},
			{Freq: 3.0, Value: 4.0},
		},
	}

	var buf bytes.Buffer
	if err := encoder.Encode(&buf, wave); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	got, err := decoder.Decode(&buf)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if got.ID != wave.ID || got.CompressedHex != wave.CompressedHex || len(got.OriginalWave) != len(wave.OriginalWave) {
		t.Fatalf("decoded wave mismatch")
	}
}

func TestWaveEncoderWrongType(t *testing.T) {
	encoder := codec.WaveEncoder[any]{}
	if err := encoder.Encode(&bytes.Buffer{}, "not-wave"); err == nil {
		t.Fatalf("expected error for wrong type")
	}
}

func TestWaveDecoderTruncated(t *testing.T) {
	decoder := codec.WaveDecoder[types.WaveData]{}
	_, err := decoder.Decode(bytes.NewReader([]byte{0x01, 0x02}))
	if err == nil {
		t.Fatalf("expected decode error for truncated data")
	}
}

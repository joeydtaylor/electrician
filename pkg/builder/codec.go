package builder

import (
	"github.com/joeydtaylor/electrician/pkg/internal/codec"
)

// NewJSONEncoder creates a new JSONEncoder.
func NewJSONEncoder[T any]() *codec.JSONEncoder[T] {
	return &codec.JSONEncoder[T]{}
}

// NewJSONDecoder creates a new JSONDecoder.
func NewJSONDecoder[T any]() *codec.JSONDecoder[T] {
	return &codec.JSONDecoder[T]{}
}

// NewLineEncoder creates a new LineEncoder.
func NewLineEncoder[T any]() *codec.LineEncoder[T] {
	return &codec.LineEncoder[T]{}
}

// NewLineDecoder creates a new LineDecoder.
func NewLineDecoder() *codec.LineDecoder {
	return &codec.LineDecoder{}
}

func NewWaveEncoder[T any]() *codec.WaveEncoder[T] {
	return &codec.WaveEncoder[T]{}
}

func NewWaveDecoder[T any]() *codec.WaveDecoder[T] {
	return &codec.WaveDecoder[T]{}
}

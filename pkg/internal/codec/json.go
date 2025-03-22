package codec

import (
	"encoding/json"
	"io"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// JSONEncoder encodes a generic type into JSON.
type JSONEncoder[T any] struct{}

// JSONDecoder decodes JSON into a generic type.
type JSONDecoder[T any] struct{}

func NewJSONDecoder[T any]() types.JSONDecoder[T] {
	return &JSONDecoder[T]{}
}

func NewJSONEncoder[T any]() types.JSONEncoder[T] {
	return &JSONEncoder[T]{}
}

// Decode reads from an io.Reader, decodes the JSON data, and stores the result in a value of type T.
func (d *JSONDecoder[T]) Decode(r io.Reader) (T, error) {
	var t T
	err := json.NewDecoder(r).Decode(&t)
	return t, err
}

// Encode writes the JSON encoding of elem to an io.Writer.
func (e *JSONEncoder[T]) Encode(w io.Writer, elem T) error {
	return json.NewEncoder(w).Encode(elem)
}

// DecodeSlice reads from an io.Reader, decodes the JSON data, and stores the result in a slice of type T.
func (d *JSONDecoder[T]) DecodeSlice(r io.Reader) ([]T, error) {
	var slice []T
	err := json.NewDecoder(r).Decode(&slice)
	return slice, err
}

// EncodeSlice writes the JSON encoding of a slice of elems to an io.Writer.
func (e *JSONEncoder[T]) EncodeSlice(w io.Writer, elems []T) error {
	return json.NewEncoder(w).Encode(elems)
}

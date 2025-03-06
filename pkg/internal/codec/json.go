package codec

import "github.com/joeydtaylor/electrician/pkg/internal/types"

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

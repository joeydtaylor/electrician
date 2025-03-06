package codec

import (
	"io"
)

// Decoder interface remains unchanged, suitable for generic types.
type Decoder[T any] interface {
	Decode(io.Reader) (T, error)
}

// Encoder interface also remains unchanged, designed for generic types.
type Encoder[T any] interface {
	Encode(io.Writer, T) error
}

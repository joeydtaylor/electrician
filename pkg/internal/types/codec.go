package types

import "io"

type Decodable interface {
	DecodeJSON(data io.Reader) error
	DecodeXML(data io.Reader) error
	DecodeText(data string) error
	DecodeBinary(data []byte) error
}

// Decoder defines an interface for components capable of decoding data from an io.Reader into an object
// of type T. This is essential for deserialization and input processing in data pipelines.
type Decoder[T any] interface {
	Decode(io.Reader) (T, error) // Decode reads from an io.Reader and decodes the content into type T.
}

// Encoder defines an interface for components that encode objects of type T into a format suitable
// for output, typically writing to an io.Writer. This is crucial for serialization and data export.
type Encoder[T any] interface {
	Encode(io.Writer, T) error // Encode writes the encoded data of type T to an io.Writer.
}

// JSONDecoder extends the Decoder interface specifically for JSON formatted data, providing methods
// for decoding individual objects and slices of objects from JSON.
type JSONDecoder[T any] interface {
	Decode(io.Reader) (T, error)          // Decode decodes a single object of type T from JSON.
	DecodeSlice(r io.Reader) ([]T, error) // DecodeSlice decodes a slice of objects of type T from JSON.
}

// JSONEncoder extends the Encoder interface specifically for JSON formatted data, including methods
// for encoding single objects or slices of objects into JSON.
type JSONEncoder[T any] interface {
	Encode(io.Writer, T) error                // Encode encodes a single object of type T into JSON.
	EncodeSlice(w io.Writer, elems []T) error // EncodeSlice encodes a slice of objects of type T into JSON.
}

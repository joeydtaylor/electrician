package codec

// LineDecoder decodes individual lines of text.
type LineDecoder struct{}

// LineEncoder encodes any type T into lines of text by converting T to a string.
type LineEncoder[T any] struct{}

func NewLineDecoder() *LineDecoder {
	return &LineDecoder{}
}

func NewLineEncoder[T any]() *LineEncoder[T] {
	return &LineEncoder[T]{}
}

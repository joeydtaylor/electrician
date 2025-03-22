package codec

import "io"

// TextDecoder decodes plain text data into a string.
type TextDecoder struct{}

func NewTextDecoder() *TextDecoder {
	return &TextDecoder{}
}

func (d *TextDecoder) Decode(r io.Reader) (string, error) {
	bytes, err := io.ReadAll(r)
	return string(bytes), err
}

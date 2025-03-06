package codec

import (
	"io"
)

// BinaryDecoder decodes binary data into bytes.
type BinaryDecoder struct{}

func NewBinaryDecoder() *BinaryDecoder {
	return &BinaryDecoder{}
}

func (d *BinaryDecoder) Decode(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}

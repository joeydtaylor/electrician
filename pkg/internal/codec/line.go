package codec

import (
	"bufio"
	"fmt"
	"io"
)

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

func (d *LineDecoder) Decode(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", io.EOF
}

// Encode converts the generic element T into a string and writes it as a line of text.
func (e *LineEncoder[T]) Encode(w io.Writer, elem T) error {
	// Convert elem to a string using fmt.Sprintf("%v") and encode it as a line
	_, err := fmt.Fprintf(w, "%v\n", elem)
	return err
}

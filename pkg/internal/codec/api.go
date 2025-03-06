package codec

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

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

package codec

import (
	"encoding/xml"
	"io"
)

// XMLDecoder decodes XML into a generic type.
type XMLDecoder[T any] struct{}

// XMLEncoder encodes a generic type into XML.
type XMLEncoder[T any] struct{}

func NewXMLDecoder[T any]() *XMLDecoder[T] {
	return &XMLDecoder[T]{}
}

func NewXMLEncoder[T any]() *XMLEncoder[T] {
	return &XMLEncoder[T]{}
}

func (d *XMLDecoder[T]) Decode(r io.Reader) (T, error) {
	var t T
	err := xml.NewDecoder(r).Decode(&t)
	return t, err
}

func (e *XMLEncoder[T]) Encode(w io.Writer, elem T) error {
	return xml.NewEncoder(w).Encode(elem)
}

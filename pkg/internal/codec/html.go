package codec

import (
	"io"

	"golang.org/x/net/html"
)

// HTMLDecoder decodes HTML data into a navigable node structure.
type HTMLDecoder struct{}

func NewHTMLDecoder() *HTMLDecoder {
	return &HTMLDecoder{}
}

func (d *HTMLDecoder) Decode(r io.Reader) (*html.Node, error) {
	return html.Parse(r)
}

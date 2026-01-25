package conduit

import (
	"bytes"
	"encoding/json"
)

// Load waits for processing to complete and returns the final output buffer.
func (c *Conduit[T]) Load() *bytes.Buffer {
	_ = c.Stop()

	if last := c.lastWire(); last != nil {
		return last.GetOutputBuffer()
	}
	return new(bytes.Buffer)
}

// LoadAsJSONArray collects output into a JSON array.
func (c *Conduit[T]) LoadAsJSONArray() ([]byte, error) {
	_ = c.Stop()

	if c.OutputChan == nil || c.lastWire() == nil {
		return json.Marshal([]T{})
	}

	var items []T
	for item := range c.OutputChan {
		items = append(items, item)
	}
	return json.Marshal(items)
}

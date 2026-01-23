package wire

import (
	"bytes"
	"encoding/json"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// LoadAsJSONArray stops the wire and drains available output into JSON.
func (w *Wire[T]) LoadAsJSONArray() ([]byte, error) {
	_ = w.Stop()

	out := w.OutputChan
	if out == nil {
		return []byte("[]"), nil
	}

	items := make([]T, 0, len(out))
	for {
		select {
		case v, ok := <-out:
			if !ok {
				return json.Marshal(items)
			}
			items = append(items, v)
		default:
			return json.Marshal(items)
		}
	}
}

// Load stops the wire and returns a copy of the output buffer.
func (w *Wire[T]) Load() *bytes.Buffer {
	_ = w.Stop()

	w.bufferMutex.Lock()
	buf := make([]byte, w.OutputBuffer.Len())
	copy(buf, w.OutputBuffer.Bytes())
	w.bufferMutex.Unlock()

	w.NotifyLoggers(
		types.DebugLevel,
		"Load",
		"component", w.componentMetadata,
		"event", "Load",
		"result", "SUCCESS",
		"bytes", len(buf),
	)

	return bytes.NewBuffer(buf)
}

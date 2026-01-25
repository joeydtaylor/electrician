package conduit

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// SetComponentMetadata updates conduit metadata.
func (c *Conduit[T]) SetComponentMetadata(name string, id string) {
	c.requireNotStarted("SetComponentMetadata")

	c.configLock.Lock()
	c.componentMetadata = types.ComponentMetadata{Name: name, ID: id, Type: c.componentMetadata.Type}
	c.configLock.Unlock()
}

// SetConcurrencyControl sets buffer size and worker count for attached wires.
func (c *Conduit[T]) SetConcurrencyControl(bufferSize int, maxConcurrency int) {
	c.requireNotStarted("SetConcurrencyControl")

	c.configLock.Lock()
	c.MaxBufferSize = bufferSize
	c.MaxConcurrency = maxConcurrency
	c.configLock.Unlock()

	c.ensureDefaults()

	for _, w := range c.snapshotWires() {
		if w == nil {
			continue
		}
		w.SetConcurrencyControl(c.MaxBufferSize, c.MaxConcurrency)
	}
}

// SetInputChannel replaces the conduit input channel.
func (c *Conduit[T]) SetInputChannel(inputChan chan T) {
	c.requireNotStarted("SetInputChannel")

	c.configLock.Lock()
	c.InputChan = inputChan
	c.configLock.Unlock()
}

package conduit

import (
	"context"
	"fmt"
)

// Submit submits an element to the conduit for processing.
func (c *Conduit[T]) Submit(ctx context.Context, elem T) error {
	wires := c.snapshotWires()
	if len(wires) == 0 {
		return fmt.Errorf("no wires connected in conduit")
	}

	if cb := c.snapshotCircuitBreaker(); cb != nil && !cb.Allow() {
		if len(cb.GetNeutralWires()) != 0 {
			for _, gw := range cb.GetNeutralWires() {
				if gw == nil {
					continue
				}
				if err := gw.Submit(ctx, elem); err != nil {
					return err
				}
			}
		}
		return nil
	}

	first := wires[0]
	if first == nil {
		return fmt.Errorf("no wires connected in conduit")
	}
	return first.Submit(ctx, elem)
}

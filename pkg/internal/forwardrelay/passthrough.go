package forwardrelay

import (
	"fmt"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
)

func (fr *ForwardRelay[T]) asPassthroughPayload(item T) (*relay.WrappedPayload, error) {
	var zero T
	switch any(zero).(type) {
	case relay.WrappedPayload:
		val := any(item).(relay.WrappedPayload)
		return &val, nil
	case *relay.WrappedPayload:
		val := any(item).(*relay.WrappedPayload)
		if val == nil {
			return nil, fmt.Errorf("nil payload")
		}
		return val, nil
	default:
		return nil, fmt.Errorf("passthrough requires relay.WrappedPayload or *relay.WrappedPayload, got %T", zero)
	}
}

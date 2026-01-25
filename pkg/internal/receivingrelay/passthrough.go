package receivingrelay

import (
	"fmt"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
)

func (rr *ReceivingRelay[T]) asPassthroughItem(payload *relay.WrappedPayload) (T, error) {
	var zero T
	if payload == nil {
		return zero, fmt.Errorf("nil payload")
	}

	switch any(zero).(type) {
	case relay.WrappedPayload:
		return any(*payload).(T), nil
	case *relay.WrappedPayload:
		return any(payload).(T), nil
	default:
		return zero, fmt.Errorf("passthrough requires relay.WrappedPayload or *relay.WrappedPayload, got %T", zero)
	}
}

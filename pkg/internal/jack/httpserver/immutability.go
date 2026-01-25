package httpserver

import (
	"fmt"
	"sync/atomic"
)

func (h *httpServerAdapter[T]) requireNotFrozen(action string) {
	if atomic.LoadInt32(&h.configFrozen) == 1 {
		panic(fmt.Sprintf("attempted to modify frozen configuration of started component: %s, action=%s", h.componentMetadata, action))
	}
}

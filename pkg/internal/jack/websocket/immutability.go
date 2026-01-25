package websocket

import (
	"fmt"
	"sync/atomic"
)

func (s *serverAdapter[T]) requireNotFrozen(action string) {
	if atomic.LoadInt32(&s.configFrozen) == 1 {
		panic(fmt.Sprintf("attempted to modify frozen configuration of started component: %s, action=%s", s.componentMetadata, action))
	}
}

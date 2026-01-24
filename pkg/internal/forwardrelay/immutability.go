package forwardrelay

import (
	"fmt"
	"sync/atomic"
)

func (fr *ForwardRelay[T]) requireNotFrozen(action string) {
	if atomic.LoadInt32(&fr.configFrozen) == 1 {
		panic(fmt.Sprintf("attempted to modify frozen configuration of started component: %s, action=%s", fr.componentMetadata, action))
	}
}

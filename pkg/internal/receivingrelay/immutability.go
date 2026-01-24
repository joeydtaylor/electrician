package receivingrelay

import (
	"fmt"
	"sync/atomic"
)

func (rr *ReceivingRelay[T]) requireNotFrozen(action string) {
	if atomic.LoadInt32(&rr.configFrozen) == 1 {
		panic(fmt.Sprintf("attempted to modify frozen configuration of started component: %s, action=%s", rr.componentMetadata, action))
	}
}

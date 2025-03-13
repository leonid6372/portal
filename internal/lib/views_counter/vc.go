package vc

import (
	"sync/atomic"
)

type ViewsCounter struct {
	n int64
}

// Count returns the number of connections at the time
// the call.
func (vc *ViewsCounter) Count() int {
	return int(atomic.LoadInt64(&vc.n))
}

// Add adds c to the number of active connections.
func (vc *ViewsCounter) Add(c int64) {
	atomic.AddInt64(&vc.n, c)
}

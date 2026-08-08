package cli

import (
	"fmt"
	"io"
	"sync"

	"github.com/Seeingu/coldmoon/coldmoon"
	"github.com/Seeingu/coldmoon/runtime"
)

// rejectionTracker collects unhandled promise rejections reported through the
// agent's HostPromiseRejectionTracker hook. The CLI surfaces them once the
// scheduler drains, so an async throw cannot exit successfully and silently.
type rejectionTracker struct {
	mu        sync.Mutex
	unhandled []*coldmoon.PromiseObject
}

func newRejectionTracker() *rejectionTracker {
	return &rejectionTracker{}
}

// hook is installed as the agent's HostPromiseRejectionTracker.
func (r *rejectionTracker) hook(promise *coldmoon.PromiseObject, operation coldmoon.PromiseRejectionTrackerOperation) {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch operation {
	case coldmoon.PromiseRejectionTrackerOperationReject:
		r.unhandled = append(r.unhandled, promise)
	case coldmoon.PromiseRejectionTrackerOperationHandle:
		for index, pending := range r.unhandled {
			if pending == promise {
				r.unhandled = append(r.unhandled[:index], r.unhandled[index+1:]...)
				break
			}
		}
	}
}

// take returns the pending rejections and clears the tracker.
func (r *rejectionTracker) take() []*coldmoon.PromiseObject {
	r.mu.Lock()
	defer r.mu.Unlock()
	pending := r.unhandled
	r.unhandled = nil
	return pending
}

// report prints each pending unhandled rejection to writer and reports whether
// any were present.
func (r *rejectionTracker) report(writer io.Writer, sourceName string) bool {
	pending := r.take()
	if len(pending) == 0 {
		return false
	}
	for _, promise := range pending {
		fmt.Fprintf(writer, "%s: Uncaught (in promise) %s\n", sourceName, formatRejectionReason(promise.PromiseResult))
	}
	return true
}

// formatRejectionReason renders an Error object as "Name: message" like the
// sync diagnostic renderer, falling back to the generic value formatting.
func formatRejectionReason(reason coldmoon.Value) string {
	if object, ok := reason.GetObject(); ok {
		if exception, ok := object.(*coldmoon.ErrorObject); ok {
			name := exception.Name
			if name == "" {
				name = "Error"
			}
			if exception.Message == "" {
				return name
			}
			return name + ": " + exception.Message
		}
	}
	return runtime.FormatValue(reason)
}

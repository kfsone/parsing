// concurrency provides functions to assist with concurrent synchronization.
package concurrency

import "sync"

// WithMutexLock wraps a call to 'action' with a lock/unlock of the given mutex.
func WithMutexLock(mx *sync.Mutex, action func()) {
	mx.Lock()
	defer mx.Unlock()
	action()
}

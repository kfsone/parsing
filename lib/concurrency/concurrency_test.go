package concurrency

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestWithMutexLock(t *testing.T) {
	t.Run("blocked", func(t *testing.T) {
		var mutex sync.Mutex
		mutex.Lock()
		assert.Never(t, func() bool { WithMutexLock(&mutex, func() {}); return true }, 25*time.Millisecond, 5*time.Millisecond)
		mutex.Unlock()
	})
	t.Run("unblocked", func(t *testing.T) {
		var mutex sync.Mutex
		called := 0
		assert.Eventually(t, func() bool { WithMutexLock(&mutex, func() { called++ }); return true }, 25*time.Millisecond, 100*time.Microsecond)
		assert.Equal(t, 1, called)
	})
}

// Counters provides an ultra-simple, concurrency-safe, application wide counter set.
package stats

import (
	"fmt"
	"sort"
	"sync"

	"github.com/kfsone/parsing/lib/concurrency"
)

var counters = make(map[string]int)
var counterMutex sync.Mutex

// BumpCounter increments the value of a counter by 'value' and returns the new quantity, in a thread-safe way.
func BumpCounter(name string, value int) (result int) {
	concurrency.WithMutexLock(&counterMutex, func() {
		counters[name] += value
		result = counters[name]
	})
	return
}

// FetchCounter retrieves a counter value in a thread-safe way.
func FetchCounter(name string) (value int) {
	concurrency.WithMutexLock(&counterMutex, func() {
		value = counters[name]
	})
	return
}

// ReportCounters will print the current counters, sorted into alphabetical order.
func ReportCounters() {
	keys := make([]string, 0, len(counters))
	for key := range counters {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		fmt.Printf("%s: %d\n", key, counters[key])
	}
}

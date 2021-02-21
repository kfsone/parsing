// timing provides functions that facilitate tracking of app-wide runtime performance.
package stats

import (
	"log"
	"sort"
	"sync"
	"time"

	"github.com/kfsone/parsing/lib/concurrency"
)

var startTime time.Time
var timingMutex sync.Mutex
var timings = make(map[string]time.Duration)

func init() {
	startTime = time.Now()
}

func Time(label string, verbose bool, action func()) {
	start := time.Now()
	action()
	duration := time.Since(start)
	if duration/time.Millisecond >= 1.0 || !verbose {
		concurrency.WithMutexLock(&timingMutex, func() {
			timings[label] += time.Since(start)
		})
	}
}

func ReportTimings() {
	timings["total"] = time.Since(startTime)

	keys := make([]string, 0, len(timings))
	for key := range timings {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		log.Printf("timing.%s: %s", key, timings[key])
	}
}

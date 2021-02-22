package parsing

import (
	"sync"

	"github.com/kfsone/parsing/lib/stats"
)

// ParseFiles will locate any files matching 'extension' in the current pathlist
// and run the given parsing function on them using a worker pool. Returns when
// all workers have finished.
func ParseFiles(extension string, parseFn func(string), pathlist []string) {
	// Sequence points for background workers.
	var workers sync.WaitGroup

	// Workers who do the initial parse of a file.
	numWorkers := *Concurrency
	if numWorkers < 1 {
		numWorkers = 1
	}
	workQueue := make(chan string, numWorkers+1)

	// Start finding files in the background.
	go stats.Time("parsing.findfiles", true, func () { FindFiles(pathlist, extension, workQueue) })

	// Create a set of workers to consume filenames and parse them.
	stats.Time("parsing.startworkers", true, func () {
		workers.Add(numWorkers)
		for i := 0; i < numWorkers; i++ {
			// Worker implementation
			go func() {
				// reduce worker count by one when this function exits.
				defer workers.Done()

				for filepath := range workQueue {
					if *stats.Verbose > 1 {
						stats.Time(filepath, false, func() { parseFn(filepath) })
					} else {
						parseFn(filepath)
					}
				}
			}()
		}
	})

	// Wait for the work to complete.
	workers.Wait()
}

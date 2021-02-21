// command-line arguments for parsing related operations.
package parsing

import (
	"path/filepath"

	"github.com/kfsone/parsing/lib/stats"

	flag "github.com/spf13/pflag"
)

// ProjectPath is the top-level path under which any non-absolute, non ".", paths specified on the command line
// will be searched.
var ProjectPath = flag.StringP("project", "P", ".", "top-level path to project. non-absolute paths will search under here.")

// Stats enables reporting of statistics in output.
var Stats = flag.Bool("stats", false, "report stats on exit")

// Concurrency determines how many parse workers run simultaneously.
var Concurrency = flag.IntP("concurrency", "j", 8, "number of concurrent workers (jobs). default: 8")

// CommonCommandLine parses common command line options.
func CommonCommandLine() {
	flag.Parse()

	// Get a clean, absolute path to the project directory.
	var err error
	*ProjectPath, err = filepath.Abs(filepath.Clean(*ProjectPath))
	stats.PanicOn(err)
}

// PathList will either return the input list of paths if it is non-empty, or it will return a list
// consisting of just the current directory (".").
func PathList(paths []string) []string {
	if len(paths) > 0 {
		return paths
	}
	return []string{"."}
}

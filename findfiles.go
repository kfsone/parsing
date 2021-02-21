package parsing

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kfsone/parsing/lib/stats"
)

// FindFiles will iterate across the given list of paths or ".", finding files
// that match the given extension, and forwarding their fully-qualified paths
// to the supplied channel.
func FindFiles(listedPaths []string, extension string, pathChan chan<- string) {
	defer close(pathChan)

	uniquePaths := make(map[string]bool)

	for _, toplevelPath := range listedPaths {
		// translate non-absolute paths (including '.') to ProjectPath relative
		toplevelPath = filepath.Clean(toplevelPath)
		if !filepath.IsAbs(toplevelPath) && toplevelPath != "." {
			toplevelPath = filepath.Join(*ProjectPath, toplevelPath)
		}

		// deduplication
		if _, exists := uniquePaths[toplevelPath]; exists == true {
			stats.BumpCounter("duplicate_paths", 1)
			continue
		}
		uniquePaths[toplevelPath] = true

		// annotate
		stats.BumpCounter("paths", 1)
		stats.Info("crawling path: %s", toplevelPath)

		// walk the directory tree
		err := filepath.Walk(toplevelPath,
			func(path string, info os.FileInfo, e error) error {
				if e != nil {
					return e
				}
				// match regular files against the desired extension
				if info.Mode().IsRegular() && strings.HasSuffix(info.Name(), extension) {
					// deduplicate
					if _, exists := uniquePaths[path]; exists == false {
						uniquePaths[path] = true
						// dispatch
						pathChan <- path
						stats.BumpCounter(extension, 1)
					} else {
						stats.BumpCounter("duplicate_paths", 1)
					}
				}
				return nil
			})
		if err != nil {
			fmt.Println(err)
		}
	}

	return
}

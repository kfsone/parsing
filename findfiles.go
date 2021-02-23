package parsing

import (
	"fmt"
	"github.com/kfsone/parsing/lib/stats"
	"os"
	"path/filepath"
	"strings"
)

// FindFiles will iterate across the given list of paths or ".", finding files
// that match the given extension, and forwarding their fully-qualified paths
// to the supplied channel.
func FindFiles(listedPaths []string, extension string, pathChan chan<- string) {
	defer close(pathChan)

	uniquePaths := make(map[string]bool, 1)

	for _, toplevelPath := range listedPaths {
		// translate non-absolute paths (including '.') to ProjectPath relative
		toplevelPath = filepath.Clean(toplevelPath)
		if !filepath.IsAbs(toplevelPath) && toplevelPath != "." {
			toplevelPath = filepath.Join(*ProjectPath, toplevelPath)
		}

		// deduplication
		if _, exists := uniquePaths[toplevelPath]; exists == true {
			stats.BumpCounter("files.duplicate_paths", 1)
			continue
		}
		uniquePaths[toplevelPath] = true

		// annotate
		stats.BumpCounter("files.paths", 1)
		stats.Info("crawling path: %s", toplevelPath)

		// walk the directory tree
		stats.Time("paths.walk", true, func () {
			err := filepath.Walk(toplevelPath,
				func(path string, info os.FileInfo, e error) error {
					if e != nil {
						return e
					}
					// match regular files against the desired extension
					if strings.HasSuffix(info.Name(), extension) && info.Mode().IsRegular() {
						// deduplicate
						if _, exists := uniquePaths[path]; exists == false {
							uniquePaths[path] = true
							// dispatch
							pathChan <- path
							stats.BumpCounter("files.ext" + extension, 1)
						} else {
							stats.BumpCounter("files.duplicate_paths", 1)
						}
					}
					return nil
				})
			if err != nil {
				fmt.Println(err)
			}
		})
	}

	return
}

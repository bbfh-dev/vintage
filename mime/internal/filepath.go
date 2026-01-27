package internal

import (
	"os"
	"path/filepath"

	"github.com/bbfh-dev/mime/cli"
)

func ToAbs(path string) string {
	path, _ = filepath.Abs(path)
	return path
}

func IterateDirsOnly(entries []os.DirEntry) func(func(os.DirEntry) bool) {
	return func(yield func(os.DirEntry) bool) {
		for _, entry := range entries {
			if !entry.IsDir() {
				cli.LogDebug(true, "Skipping file %q", entry.Name())
				continue
			}
			if !yield(entry) {
				return
			}
		}
	}
}

func IterateFilesOnly(entries []os.DirEntry) func(func(os.DirEntry) bool) {
	return func(yield func(os.DirEntry) bool) {
		for _, entry := range entries {
			if entry.IsDir() {
				cli.LogDebug(true, "Skipping folder %q", entry.Name())
				continue
			}
			if !yield(entry) {
				return
			}
		}
	}
}

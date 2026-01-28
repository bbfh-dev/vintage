package internal

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/bbfh-dev/mime/cli"
)

func GetMostRecentIn(dirs ...string) time.Time {
	timestamp := time.UnixMilli(0)
	for _, dir := range dirs {
		new_timestamp := getMostRecent(dir)
		if new_timestamp.Sub(timestamp) > 0 {
			timestamp = new_timestamp
		}
	}
	return timestamp
}

func getMostRecent(dir string) time.Time {
	latest_time := time.UnixMilli(0)
	filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		if timestamp := info.ModTime(); timestamp.Sub(latest_time) > 0 {
			latest_time = timestamp
		}

		return nil
	})
	return latest_time
}

func ToAbs(path string) string {
	path, _ = filepath.Abs(path)
	return path
}

func IterateDirsOnly(entries []os.DirEntry) func(func(os.DirEntry) bool) {
	return func(yield func(os.DirEntry) bool) {
		for _, entry := range entries {
			if !entry.IsDir() {
				cli.LogDebug(2, "Skipping file %q", entry.Name())
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
				cli.LogDebug(2, "Skipping folder %q", entry.Name())
				continue
			}
			if !yield(entry) {
				return
			}
		}
	}
}

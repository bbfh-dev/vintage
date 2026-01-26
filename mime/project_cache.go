package mime

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/bbfh-dev/mime/cli"
)

func (project *Project) checkIfCached(folder, lib, zip_name string, value *bool) func() error {
	if cli.Main.Options.Force {
		return nil
	}
	return func() error {
		data_timestamp := getMostRecentTimestamp(folder)
		data_libs_timestamp := getMostRecentTimestamp(filepath.Join("libs", lib))
		if data_libs_timestamp.Sub(data_timestamp) > 0 {
			data_timestamp = data_libs_timestamp
		}

		info, err := os.Stat(zip_name)
		if err != nil {
			return nil
		}
		data_zip_timestamp := info.ModTime()

		if data_timestamp.Sub(data_zip_timestamp) < 0 {
			*value = true
			cli.LogCached(false, "%q is already up-to-date", filepath.Base(zip_name))
		}

		return nil
	}
}

func getMostRecentTimestamp(dir string) time.Time {
	var latest_time time.Time
	filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil && entry.IsDir() {
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

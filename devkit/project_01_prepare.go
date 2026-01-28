package devkit

import (
	"os"
	"path/filepath"

	"github.com/bbfh-dev/mime/cli"
	"github.com/bbfh-dev/mime/devkit/internal"
)

// These are constants to make it clear no other values are allowed
const (
	FOLDER_DATA   = "data"
	FOLDER_ASSETS = "assets"
)

func (project *Project) LogHeader(header string) internal.Task {
	return func() error {
		cli.LogInfo(0, "%s", header)
		return nil
	}
}

func (project *Project) DetectPackIcon() error {
	_, err := os.Stat("pack.png")
	if os.IsNotExist(err) {
		cli.LogWarn(1, "No pack icon found")
		return nil
	}

	cli.LogInfo(1, "Found 'pack.png'")
	project.extraFilesToCopy = append(project.extraFilesToCopy, "pack.png")
	return nil
}

func (project *Project) CheckIfCached(value *bool, folder string) internal.Task {
	if cli.Main.Options.Force {
		return nil
	}
	var libs_folder, zip_path string
	switch folder {
	case FOLDER_DATA:
		libs_folder, zip_path = "data_packs", project.getZipPath("DP")
	case FOLDER_ASSETS:
		libs_folder, zip_path = "resource_packs", project.getZipPath("RP")
	}
	return func() error {
		timestamp := internal.GetMostRecentIn(
			folder,
			filepath.Join("libs", libs_folder),
			"templates",
		)

		info, err := os.Stat(zip_path)
		if err != nil {
			cli.LogWarn(1, "%q is missing. Caching is impossible", filepath.Base(zip_path))
			return nil
		}
		zip_timestamp := info.ModTime()

		if timestamp.Sub(zip_timestamp) < 0 {
			*value = true
			cli.LogCached(1, "%q is already up-to-date", filepath.Base(zip_path))
		}

		return nil
	}
}

func (project *Project) LoadTemplates() error {
	// TODO: this
	return nil
}

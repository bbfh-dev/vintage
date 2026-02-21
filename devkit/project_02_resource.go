package devkit

import (
	"os"
	"path/filepath"

	liblog "github.com/bbfh-dev/lib-log"
	"github.com/bbfh-dev/vintage/devkit/internal/pipeline"
	"github.com/bbfh-dev/vintage/devkit/minecraft"
)

func (project *Project) GenerateResourcePack() error {
	if project.isAssetsCached {
		return nil
	}

	_, err := os.Stat(FOLDER_ASSETS)
	if os.IsNotExist(err) {
		liblog.Debug(0, "No resource pack found")
		return nil
	}

	liblog.Info(0, "Creating a Resource Pack")
	path := filepath.Join(project.BuildDir, "resource_pack")

	return pipeline.New(
		project.clearDir(path),
		pipeline.Async(
			project.copyPackDirs(FOLDER_ASSETS, path, nil),
		),
		project.copyExtraFiles(path),
		project.createPackMcmeta("resource_pack", "resources", minecraft.ResourcePackFormats),
	)
}

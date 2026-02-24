package devkit

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zip"

	liberrors "github.com/bbfh-dev/lib-errors"
	liblog "github.com/bbfh-dev/lib-log"
)

func (project *Project) ZipPacks() error {
	if project.isDataCached && project.isAssetsCached {
		return nil
	}

	liblog.Info(0, "Creating .zip files")
	if !project.isDataCached {
		if err := project.zip("data_pack"); err != nil {
			return err
		}
	}
	if !project.isAssetsCached {
		if err := project.zip("resource_pack"); err != nil {
			return err
		}
	}
	return nil
}

func (project *Project) zip(folder string) error {
	zip_path := project.getZipPath(getZipLabel(folder))

	folder_path := filepath.Join(project.BuildDir, folder)
	root, err := os.OpenRoot(folder_path)
	if err != nil {
		liblog.Debug(1, "Skipping %q: %s", folder, err.Error())
		return nil
	}

	file, err := os.Create(zip_path)
	if err != nil {
		return liberrors.NewIO(err, zip_path)
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	defer writer.Close()

	err = writer.AddFS(root.FS())
	if err != nil {
		return liberrors.NewIO(err, folder_path)
	}

	liblog.Done(1, "Created %q", zip_path)
	return nil
}

func (project *Project) getZipPath(label string) string {
	// TODO: Would be nice to not add label if pack is only one of a kind (only RP or only DP)
	return filepath.Join(
		project.BuildDir,
		fmt.Sprintf(
			"%s_%s_v%s.zip",
			project.Meta.Name(),
			label,
			project.Meta.VersionFormatted(),
		),
	)
}

func getZipLabel(folder string) string {
	switch folder {
	case "data_pack":
		return "DP"
	case "resource_pack":
		return "RP"
	}
	return "PACK"
}

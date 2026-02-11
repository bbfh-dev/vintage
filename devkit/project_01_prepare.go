package devkit

import (
	"fmt"
	"os"
	"path/filepath"

	liberrors "github.com/bbfh-dev/lib-errors"
	liblog "github.com/bbfh-dev/lib-log"
	"github.com/bbfh-dev/vintage/cli"
	"github.com/bbfh-dev/vintage/devkit/internal/drive"
	"github.com/bbfh-dev/vintage/devkit/internal/pipeline"
	"github.com/bbfh-dev/vintage/devkit/language"
	"github.com/tidwall/gjson"
)

// These are constants to make it clear no other values are allowed
const (
	FOLDER_DATA   = "data"
	FOLDER_ASSETS = "assets"
)

func (project *Project) LogHeader(header string) pipeline.Task {
	return func() error {
		liblog.Info(0, "%s", header)
		return nil
	}
}

func (project *Project) DetectPackIcon() error {
	_, err := os.Stat("pack.png")
	if os.IsNotExist(err) {
		liblog.Warn(1, "No pack icon found")
		return nil
	}

	liblog.Info(1, "Found 'pack.png'")
	project.extraFilesToCopy = append(project.extraFilesToCopy, "pack.png")
	return nil
}

func (project *Project) CheckIfCached(value *bool, folder string) pipeline.Task {
	if cli.Build.Options.Force {
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
		timestamp := drive.GetMostRecentIn(
			folder,
			filepath.Join("libs", libs_folder),
			"templates",
		)

		info, err := os.Stat(zip_path)
		if err != nil {
			liblog.Warn(1, "%q is missing. Caching is impossible", filepath.Base(zip_path))
			return nil
		}
		zip_timestamp := info.ModTime()

		if timestamp.Sub(zip_timestamp) < 0 {
			*value = true
			liblog.Cached(1, "%q is already up-to-date", filepath.Base(zip_path))
		}

		return nil
	}
}

func (project *Project) LoadTemplates() error {
	if project.isDataCached && project.isAssetsCached {
		return nil
	}

	_, err := os.Stat("templates")
	if os.IsNotExist(err) {
		liblog.Debug(1, "No templates found")
		return nil
	}

	liblog.Info(1, "Loading templates")

	entries, err := os.ReadDir("templates")
	if err != nil {
		return liberrors.NewIO(err, drive.ToAbs("templates"))
	}

	for entry := range drive.IterateDirsOnly(entries) {
		path := filepath.Join("templates", entry.Name(), "manifest.json")
		manifest_data, err := os.ReadFile(path)
		if err != nil {
			return liberrors.NewIO(err, drive.ToAbs(path))
		}
		manifest := drive.NewJsonFile(manifest_data)

		if err := manifest.ExpectField("type", gjson.String); err != nil {
			return &liberrors.DetailedError{
				Label:   liberrors.ERR_VALIDATE,
				Context: liberrors.DirContext{Path: path},
				Details: err.Error(),
			}
		}

		dir := filepath.Join("templates", entry.Name())
		template_type := manifest.Get("type").String()
		switch template_type {

		case "inline":
			template, err := language.NewInlineTemplate(dir, manifest)
			if err != nil {
				return err
			}
			project.inlineTemplates[entry.Name()] = template
			liblog.Debug(2, "Loaded inline %q", entry.Name())

		case "generate":
			template, err := language.NewGeneratorTemplate(dir, manifest)
			if err != nil {
				return err
			}
			project.generatorTemplates[entry.Name()] = template
			liblog.Debug(2, "Loaded generator %q", entry.Name())

		default:
			return &liberrors.DetailedError{
				Label:   liberrors.ERR_SYNTAX,
				Context: liberrors.DirContext{Path: path},
				Details: fmt.Sprintf("unknown template type %q", template_type),
			}
		}
	}

	return nil
}

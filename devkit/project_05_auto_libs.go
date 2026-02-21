package devkit

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	liberrors "github.com/bbfh-dev/lib-errors"
	liblog "github.com/bbfh-dev/lib-log"
	"github.com/bbfh-dev/vintage/devkit/internal/autolibs"
	"github.com/bbfh-dev/vintage/devkit/internal/drive"
	"github.com/bbfh-dev/vintage/devkit/internal/mcfunc"
	"github.com/bbfh-dev/vintage/devkit/internal/pipeline"
	"golang.org/x/sync/errgroup"
)

func (project *Project) LoadAutoLibs() error {
	if project.isDataCached && project.isAssetsCached {
		return nil
	}

	_, err := os.Stat("libs")
	if os.IsNotExist(err) {
		liblog.Debug(1, "No libraries found")
		return nil
	}

	liblog.Info(1, "Loading automatic libraries")
	return pipeline.New(
		pipeline.Async(
			pipeline.If[pipeline.AsyncTask](!project.isDataCached).
				Then(project.loadLibsFrom("data_packs")),
			pipeline.If[pipeline.AsyncTask](!project.isAssetsCached).
				Then(project.loadLibsFrom("resource_packs")),
		),
	)
}

func (project *Project) loadLibsFrom(folder string) pipeline.AsyncTask {
	return func(errs *errgroup.Group) error {
		entries, err := os.ReadDir(filepath.Join("libs", folder))
		if err != nil {
			liblog.Debug(2, "Skipping %q: %s", folder, err.Error())
			return nil
		}

		for _, entry := range entries {
			ext := filepath.Ext(entry.Name())
			if ext != ".json" {
				continue
			}
			name := strings.TrimSuffix(entry.Name(), ext)
			name = strings.ReplaceAll(name, ".", "\\.")
			name = strings.ReplaceAll(name, "---", ".*")
			re := regexp.MustCompile(name)

			path := filepath.Join("libs", folder, entry.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				return liberrors.NewIO(err, path)
			}

			lib := &autolibs.Library{
				Namespaces: []string{},
				Path:       path,
				File:       drive.NewJsonFile(data),
			}

			for namespace := range mcfunc.UsedNamespaces {
				if re.MatchString(namespace) {
					lib.Namespaces = append(lib.Namespaces, namespace)
				}
			}

			if err := lib.Validate(); err != nil {
				return &liberrors.DetailedError{
					Label:   liberrors.ERR_VALIDATE,
					Context: liberrors.DirContext{Path: path},
					Details: err.Error(),
				}
			}

			liblog.Debug(2, "Loaded %s / %q", folder, entry.Name())
			project.libraries = append(project.libraries, lib)
		}

		return nil
	}
}

func (project *Project) ManageAutoLibs() error {
	if project.isDataCached && project.isAssetsCached {
		return nil
	}

	if len(project.libraries) == 0 {
		liblog.Debug(0, "No automatic libraries found")
		return nil
	}

	liblog.Info(0, "Managing automatic libraries")

	for _, lib := range project.libraries {
		if err := lib.Manage(); err != nil {
			return err
		}
	}

	return nil
}

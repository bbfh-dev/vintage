package devkit

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	liberrors "github.com/bbfh-dev/lib-errors"
	liblog "github.com/bbfh-dev/lib-log"
	"github.com/bbfh-dev/vintage/devkit/internal/code"
	"github.com/bbfh-dev/vintage/devkit/internal/drive"
	"github.com/bbfh-dev/vintage/devkit/internal/mcfunc"
	cp "github.com/otiai10/copy"
	"golang.org/x/sync/errgroup"
)

func (project *Project) GenerateFromTemplates(errs *errgroup.Group) error {
	// TODO: this code needs refactoring
	if project.isDataCached && project.isAssetsCached {
		return nil
	}

	liblog.Info(0, "Generating from %d template(s)", len(project.generatorTemplates))

	for _, template := range project.generatorTemplates {
		errs.Go(func() error {
			liblog.Debug(
				1,
				"Generating from %q with %d definition(s)",
				template.Root,
				len(template.Definitions),
			)

			files_to_generate := []string{}
			for _, folder := range []string{"data", "assets"} {
				path := filepath.Join(template.Root, folder)
				if _, err := os.Stat(path); os.IsNotExist(err) {
					continue
				}

				err := filepath.WalkDir(path, func(p string, entry fs.DirEntry, err error) error {
					if err != nil || entry.IsDir() {
						return err
					}

					p = strings.TrimPrefix(p, filepath.Dir(path))
					p = strings.TrimPrefix(p, string(filepath.Separator))
					files_to_generate = append(files_to_generate, p)
					return nil
				})

				if err != nil {
					return liberrors.NewIO(err, path)
				}
			}

			liblog.Debug(2, "Loaded %d files to generate per definition", len(files_to_generate))

			for _, definition := range template.Definitions {
				for _, path := range files_to_generate {
					src_path := filepath.Join(template.Root, path)

					var dest_folder string
					if strings.HasPrefix(path, "data") {
						dest_folder = "data_pack"
					} else if strings.HasPrefix(path, "assets") {
						dest_folder = "resource_pack"
					} else {
						return &liberrors.DetailedError{
							Label:   liberrors.ERR_INTERNAL,
							Context: liberrors.DirContext{Path: path},
							Details: fmt.Sprintf("unknown destination for %q", path),
						}
					}

					dest_path, err := code.SubstituteString(path, definition.Env)
					if err != nil {
						return &liberrors.DetailedError{
							Label:   liberrors.ERR_FORMAT,
							Context: liberrors.DirContext{Path: path},
							Details: err.Error(),
						}
					}

					switch filepath.Ext(path) {

					case ".json":
						dest_path = filepath.Join(project.BuildDir, dest_folder, dest_path)
						data, err := os.ReadFile(src_path)
						if err != nil {
							return liberrors.NewIO(err, path)
						}

						file := drive.NewJsonFile(data)

						err = code.SubstituteJsonFile(file, definition.Env)
						if err != nil {
							return &liberrors.DetailedError{
								Label:   liberrors.ERR_FORMAT,
								Context: liberrors.DirContext{Path: path},
								Details: err.Error(),
							}
						}

						liblog.Debug(3, "Writing %q", dest_path)
						err = errors.Join(
							os.MkdirAll(filepath.Dir(dest_path), os.ModePerm),
							os.WriteFile(dest_path, file.Formatted(), os.ModePerm),
						)
						if err != nil {
							return liberrors.NewIO(err, path)
						}

					case ".mcfunction":
						data, err := os.ReadFile(src_path)
						if err != nil {
							return liberrors.NewIO(err, dest_path)
						}

						output, err := code.SubstituteString(string(data), definition.Env)
						if err != nil {
							return &liberrors.DetailedError{
								Label:   liberrors.ERR_FORMAT,
								Context: liberrors.DirContext{Path: dest_path},
								Details: err.Error(),
							}
						}

						scanner := bufio.NewScanner(strings.NewReader(output))
						fn := mcfunc.New(dest_path, scanner, project.inlineTemplates)
						proc := mcfunc.NewProcessor(fn)
						if err := proc.Build(); err != nil {
							return err
						}

					default:
						dest_path = filepath.Join(project.BuildDir, dest_folder, dest_path)
						liblog.Debug(3, "Copying into %q", dest_path)
						err := cp.Copy(src_path, dest_path)
						if err != nil {
							return liberrors.NewIO(err, path)
						}
					}
				}
			}

			return nil
		})
	}

	return nil
}

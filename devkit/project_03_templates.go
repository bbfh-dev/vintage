package devkit

import (
	"bufio"
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

var GeneratedJsonFiles = map[string]*drive.JsonFile{}

func mergeGeneratedJsonFile(path string, file *drive.JsonFile) {
	if original, ok := GeneratedJsonFiles[path]; ok {
		original.MergeWith(file)
	} else {
		GeneratedJsonFiles[path] = file
	}
}

func (project *Project) GenerateFromTemplates(errs *errgroup.Group) error {
	// TODO: this code needs refactoring
	if project.isDataCached && project.isAssetsCached {
		return nil
	}

	liblog.Info(0, "Generating from %d template(s)", len(project.generatorTemplates))

	for _, template := range project.generatorTemplates {
		errs.Go(func() error {
			liblog.Info(
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

						mergeGeneratedJsonFile(dest_path, file)

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

			liblog.Done(
				2,
				"Generated %d file(s)",
				len(template.Definitions)*len(files_to_generate),
			)

			return nil
		})
	}

	return nil
}

func (project *Project) CollectFromTemplates() error {
	return nil
}

func (project *Project) writeGeneratedJsonFiles() error {
	for path, file := range GeneratedJsonFiles {
		err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
		if err != nil {
			return liberrors.NewIO(err, path)
		}

		err = os.WriteFile(path, file.Formatted(), os.ModePerm)
		if err != nil {
			return liberrors.NewIO(err, path)
		}
	}

	return nil
}

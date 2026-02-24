package devkit

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	liberrors "github.com/bbfh-dev/lib-errors"
	liblog "github.com/bbfh-dev/lib-log"
	"github.com/bbfh-dev/vintage/devkit/internal/code"
	"github.com/bbfh-dev/vintage/devkit/internal/drive"
	"github.com/bbfh-dev/vintage/devkit/internal/mcfunc"
	"golang.org/x/sync/errgroup"
)

var GeneratorResults []map[string]*drive.JsonFile

func (project *Project) GenerateFromTemplates(errs *errgroup.Group) error {
	// TODO: this code needs refactoring
	if project.isDataCached && project.isAssetsCached {
		return nil
	}

	liblog.Info(0, "Generating from %d template(s)", len(project.generatorTemplates))

	GeneratorResults = make([]map[string]*drive.JsonFile, 0, len(project.generatorTemplates))

	for _, template := range project.generatorTemplates {
		errs.Go(func() error {
			liblog.Info(
				1,
				"Generating from %q with %d definition(s)",
				template.Root,
				len(template.Definitions),
			)

			// TODO: No need to include image/audio/etc. files
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

			file_cache := make(map[string][]byte)
			for _, path := range files_to_generate {
				src_path := filepath.Join(template.Root, path)
				data, err := os.ReadFile(src_path)
				if err != nil {
					return liberrors.NewIO(err, src_path)
				}
				file_cache[path] = data
			}

			liblog.Debug(2, "Loaded %d files to generate per definition", len(files_to_generate))
			localMap := make(map[string]*drive.JsonFile)

			for _, definition := range template.Definitions {
				for _, path := range files_to_generate {
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
						data := file_cache[path]
						file := drive.NewJsonFile(data)

						err = code.SubstituteJsonFile(file, definition.Env)
						if err != nil {
							return &liberrors.DetailedError{
								Label:   liberrors.ERR_FORMAT,
								Context: liberrors.DirContext{Path: path},
								Details: err.Error(),
							}
						}

						if original, ok := localMap[dest_path]; ok {
							original.MergeWith(file)
						} else {
							localMap[dest_path] = file
						}

					case ".mcfunction":
						data := file_cache[path]

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
						err := os.WriteFile(dest_path, file_cache[path], os.ModePerm)
						if err != nil {
							return liberrors.NewIO(err, path)
						}
					}
				}
			}

			liblog.Done(2, "Generated %d file(s)", len(template.Definitions)*len(files_to_generate))

			GeneratorResults = append(GeneratorResults, localMap)
			return nil
		})
	}

	return nil
}

func (project *Project) CollectFromTemplates() error {
	return nil
}

func (project *Project) RunCustomTemplates() error {
	if project.isDataCached && project.isAssetsCached {
		return nil
	}

	liblog.Info(0, "Running %d custom template(s)", len(project.customTemplates))

	for _, template := range project.customTemplates {
		path := filepath.Join(template.Root, template.Program)
		cmd := exec.Command(path, project.BuildDir)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			path = fmt.Sprintf("%s with [%s]", path, project.BuildDir)
			return &liberrors.DetailedError{
				Label:   liberrors.ERR_EXECUTE,
				Context: liberrors.DirContext{Path: path},
				Details: err.Error(),
			}
		}
	}

	return nil
}

func (project *Project) writeGeneratedJsonFiles(errs *errgroup.Group) error {
	merged := make(map[string]*drive.JsonFile)
	for _, local := range GeneratorResults {
		for path, file := range local {
			if original, ok := merged[path]; ok {
				original.MergeWith(file)
			} else {
				merged[path] = file
			}
		}
	}
	GeneratorResults = nil

	for path, file := range merged {
		errs.Go(func() error {
			err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
			if err != nil {
				return liberrors.NewIO(err, path)
			}

			err = os.WriteFile(path, file.Formatted(), os.ModePerm)
			if err != nil {
				return liberrors.NewIO(err, path)
			}

			return nil
		})
	}

	return nil
}

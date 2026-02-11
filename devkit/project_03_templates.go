package devkit

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"

	liberrors "github.com/bbfh-dev/lib-errors"
	liblog "github.com/bbfh-dev/lib-log"
	"github.com/bbfh-dev/vintage/devkit/internal"
	"github.com/bbfh-dev/vintage/devkit/internal/drive"
	"github.com/bbfh-dev/vintage/devkit/language"
	"golang.org/x/sync/errgroup"
)

func (project *Project) GenerateFromTemplates() error {
	if project.isDataCached && project.isAssetsCached {
		return nil
	}

	if len(project.generatorTemplates) == 0 {
		liblog.Debug(0, "No generator templates defined")
		return nil
	}

	liblog.Info(0, "Generating code from %d template(s)", len(project.generatorTemplates))

	for template_name, template := range project.generatorTemplates {
		liblog.Debug(1, "Generating from %q", template_name)
		var errs errgroup.Group

		for definition_name, definition := range template.Definitions {
			errs.Go(func() error {
				liblog.Debug(2, "Begin processing %q", definition_name)

				root := filepath.Join("templates", template_name)
				tree, err := internal.LoadTree(
					root,
					[2]string{"data", "data_pack"},
					[2]string{"assets", "resource_pack"},
				)
				if err != nil {
					return err
				}

				for path, file := range tree {
					file := file.Clone()
					err = internal.SubstituteGenericFile(file, definition.Env)
					if err != nil {
						path = strings.TrimPrefix(path, "data_pack/")
						path = strings.TrimPrefix(path, "resource_pack/")
						return &liberrors.DetailedError{
							Label:   liberrors.ERR_FORMAT,
							Context: liberrors.DirContext{Path: filepath.Join(root, path)},
							Details: err.Error(),
						}
					}

					new_path, err := internal.SimpleSubstitute(path, definition.Env)
					if err != nil {
						return &liberrors.DetailedError{
							Label:   liberrors.ERR_FORMAT,
							Context: liberrors.DirContext{Path: path},
							Details: err.Error(),
						}
					}

					if file.Ext == ".mcfunction" {
						path = strings.TrimPrefix(path, "data_pack/")
						scanner := bufio.NewScanner(bytes.NewReader(file.Body))
						fn := language.NewMcfunction(filepath.Join(root, new_path), scanner)
						err := fn.BuildTree().ParseAndSave(project.inlineTemplates)
						if err != nil {
							return err
						}
						continue
					}

					if err := project.saveFile(path, new_path, file); err != nil {
						return err
					}
					liblog.Debug(2, "Saved %q", new_path)
				}

				return nil
			})
		}

		if err := errs.Wait(); err != nil {
			return err
		}

		liblog.Done(
			1,
			"Finished generating %q for %d definitions",
			template_name,
			len(template.Definitions),
		)
	}

	return nil
}

func (project *Project) saveFile(path, new_path string, file *drive.GenericFile) error {
	new_path = filepath.Join(project.BuildDir, new_path)
	if err := os.MkdirAll(filepath.Dir(new_path), os.ModePerm); err != nil {
		return liberrors.NewIO(err, path)
	}

	if err := os.WriteFile(new_path, file.Contents(), os.ModePerm); err != nil {
		return liberrors.NewIO(err, path)
	}

	return nil
}

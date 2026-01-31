package devkit

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	liberrors "github.com/bbfh-dev/lib-errors"
	"github.com/bbfh-dev/mime/cli"
	"github.com/bbfh-dev/mime/devkit/internal"
	"github.com/bbfh-dev/mime/devkit/language"
	"github.com/bbfh-dev/mime/devkit/minecraft"
	"golang.org/x/sync/errgroup"
)

func (project *Project) GenerateDataPack() error {
	if project.isDataCached {
		return nil
	}

	_, err := os.Stat(FOLDER_DATA)
	if os.IsNotExist(err) {
		cli.LogDebug(0, "No data pack found")
		return nil
	}

	cli.LogInfo(0, "Creating a Data Pack")
	path := filepath.Join(project.BuildDir, "data_pack")

	var funcFoldersToParse = []string{}

	return internal.Pipeline(
		project.clearDir(path),
		internal.Async(
			project.copyPackDirs(FOLDER_DATA, path, &funcFoldersToParse),
		),
		internal.Async(
			project.parseMcFunctions(&funcFoldersToParse),
		),
		project.writeMcfunctions,
		project.copyExtraFiles(path),
		project.createPackMcmeta("data_pack", minecraft.DataPackFormats),
	)
}

func (project *Project) parseMcFunctions(folders *[]string) internal.AsyncTask {
	return func(errs *errgroup.Group) error {
		for _, path := range *folders {
			filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
				if err != nil || entry.IsDir() {
					return err
				}
				errs.Go(func() error {
					return project.parseFunction(path)
				})
				return nil
			})
		}

		return nil
	}
}

func (project *Project) parseFunction(path string) error {
	file, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return liberrors.NewIO(err, path)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	function := language.NewMcfunction(path, scanner)
	return function.BuildTree().ParseAndSave(project.inlineTemplates)
}

func (project *Project) writeMcfunctions() error {
	for path, lines := range language.Registry {
		path = filepath.Join(project.BuildDir, "data_pack", path)

		err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
		if err != nil {
			return liberrors.NewIO(err, path)
		}

		err = os.WriteFile(path, []byte(strings.Join(lines, "\n")), os.ModePerm)
		if err != nil {
			return liberrors.NewIO(err, path)
		}
	}

	return nil
}

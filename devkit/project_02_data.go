package devkit

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	liberrors "github.com/bbfh-dev/lib-errors"
	liblog "github.com/bbfh-dev/lib-log"
	"github.com/bbfh-dev/vintage/devkit/internal/mcfunc"
	"github.com/bbfh-dev/vintage/devkit/internal/pipeline"
	"github.com/bbfh-dev/vintage/devkit/minecraft"
	"golang.org/x/sync/errgroup"
)

func (project *Project) GenerateDataPack() error {
	if project.isDataCached {
		return nil
	}

	_, err := os.Stat(FOLDER_DATA)
	if os.IsNotExist(err) {
		liblog.Debug(0, "No data pack found")
		return nil
	}

	liblog.Info(0, "Creating a Data Pack")
	path := filepath.Join(project.BuildDir, "data_pack")

	var funcFoldersToParse = []string{}

	return pipeline.New(
		project.clearDir(path),
		pipeline.Async(
			project.copyPackDirs(FOLDER_DATA, path, &funcFoldersToParse),
		),
		pipeline.Async(
			project.parseMcFunctions(&funcFoldersToParse),
		),
		project.copyExtraFiles(path),
		project.createPackMcmeta("data_pack", "data", minecraft.DataPackFormats),
	)
}

func (project *Project) parseMcFunctions(folders *[]string) pipeline.AsyncTask {
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
	fn := mcfunc.New(path, scanner, project.inlineTemplates)
	proc := mcfunc.NewProcessor(fn)

	if err := proc.Build(); err != nil {
		return err
	}

	return nil
}

func (project *Project) writeMcfunctions() error {
	for path, lines := range mcfunc.Registry {
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

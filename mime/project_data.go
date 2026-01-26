package mime

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	liberrors "github.com/bbfh-dev/lib-errors"
	"github.com/bbfh-dev/mime/cli"
	"github.com/bbfh-dev/mime/mime/internal"
	"github.com/bbfh-dev/mime/mime/language"
	"golang.org/x/sync/errgroup"

	cp "github.com/otiai10/copy"
)

func (project *Project) genDataPack() error {
	if project.isDataCached {
		return nil
	}

	_, err := os.Stat("data")
	if os.IsNotExist(err) {
		cli.LogDebug(false, "No data pack found")
		return nil
	}

	cli.LogInfo(false, "Creating a data pack")
	path := filepath.Join(project.BuildDir, "data_pack")

	err = os.RemoveAll(path)
	if err != nil {
		return liberrors.NewIO(err, project.BuildDir)
	}

	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return liberrors.NewIO(err, project.BuildDir)
	}

	data_entries, err := os.ReadDir("data")
	if err != nil {
		return liberrors.NewIO(err, internal.ToAbs("data"))
	}

	funcFoldersToParse := []string{}
	var errs errgroup.Group

	for data_entry := range internal.IterateDirsOnly(data_entries) {
		path := filepath.Join("data", data_entry.Name())
		folder_entries, err := os.ReadDir(path)
		if err != nil {
			return liberrors.NewIO(err, path)
		}

		for folder_entry := range internal.IterateDirsOnly(folder_entries) {
			path := filepath.Join("data", data_entry.Name(), folder_entry.Name())
			switch folder_entry.Name() {
			case "function", "functions":
				funcFoldersToParse = append(funcFoldersToParse, path)
			default:
				cli.LogDebug(true, "Copying directory %q", path)
				errs.Go(func() error {
					return cp.Copy(path, filepath.Join(project.BuildDir, "data_pack", path))
				})
			}
		}
	}

	for _, path := range funcFoldersToParse {
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

	if err := errs.Wait(); err != nil {
		switch err := err.(type) {
		case *liberrors.DetailedError:
			return err
		default:
			return &liberrors.DetailedError{
				Label: "Task Error",
				Context: liberrors.DirContext{
					Path: internal.ToAbs("data"),
				},
				Details: err.Error(),
			}
		}
	}

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

	for _, file := range project.extraFilesToCopy {
		cli.LogDebug(true, "Copying extra %q", file)
		path = filepath.Join(project.BuildDir, "data_pack", file)
		err = cp.Copy(file, path)
		if err != nil {
			return liberrors.NewIO(err, path)
		}
	}

	return nil
}

func (project *Project) parseFunction(path string) error {
	file, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return liberrors.NewIO(err, path)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	function := language.New(path, scanner, 0)
	return function.Parse()
}

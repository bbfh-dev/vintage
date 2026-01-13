package mime

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/bbfh-dev/mime/cli"
	"github.com/bbfh-dev/mime/mime/errors"
	"github.com/bbfh-dev/mime/mime/minecraft"
	cp "github.com/otiai10/copy"
	"golang.org/x/sync/errgroup"
)

type Project struct {
	BuildDir      string
	Meta          *minecraft.PackMcmeta
	has_icon      bool
	has_resources bool
	task_err      error
}

func New(mcmeta *minecraft.PackMcmeta) *Project {
	return &Project{
		BuildDir:      cli.Main.Options.Output,
		Meta:          mcmeta,
		has_icon:      false,
		has_resources: false,
		task_err:      nil,
	}
}

func (project *Project) Build() error {
	start := time.Now()
	cli.LogInfo(
		false,
		"Building v%s for Minecraft %s",
		project.Meta.File.Get("meta.version"),
		project.Meta.File.Get("meta.minecraft"),
	)

	project.do(project.checkBuildDir)
	project.do(project.clearBuildDir)
	project.do(project.detectPackIcon)
	project.do(project.createResourcePack)
	project.do(project.createDataPack)

	if project.task_err != nil {
		return project.task_err
	}

	cli.LogDone(false, "Finished building in %s", time.Since(start))
	return nil
}

func (project *Project) do(task func() error) {
	if project.task_err == nil {
		project.task_err = task()
	}
}

func (project *Project) checkBuildDir() error {
	project.BuildDir, _ = filepath.Abs(project.BuildDir)
	cli.LogInfo(false, "Checking build directory: %s", project.BuildDir)

	stat, err := os.Stat(project.BuildDir)
	if err != nil {
		if os.IsNotExist(err) {
			cli.LogDebug(true, "Directory doesn't exist yet. Skipping checks")
			return nil
		}
		return errors.NewError(errors.ERR_IO, project.BuildDir, err.Error())
	}

	if !stat.IsDir() {
		return errors.NewError(
			errors.ERR_VALID,
			project.BuildDir,
			"build output is a file",
		)
	}

	return nil
}

func (project *Project) clearBuildDir() error {
	cli.LogInfo(false, "Clearing build directory")

	if err := os.RemoveAll(project.BuildDir); err != nil {
		return errors.NewError(errors.ERR_IO, project.BuildDir, err.Error())
	}

	err := os.MkdirAll(project.BuildDir, os.ModePerm)
	if err != nil {
		return errors.NewError(errors.ERR_IO, project.BuildDir, err.Error())
	}

	return nil
}

func (project *Project) detectPackIcon() error {
	_, err := os.Stat("pack.png")
	if os.IsNotExist(err) {
		cli.LogWarn(false, "No pack icon found")
		return nil
	}

	cli.LogInfo(false, "Found 'pack.png'")
	project.has_icon = true
	return nil
}

func (project *Project) createResourcePack() error {
	_, err := os.Stat("assets")
	if os.IsNotExist(err) {
		cli.LogDebug(false, "No resource pack found")
		return nil
	}

	cli.LogInfo(false, "Creating a resource pack")
	project.has_resources = true

	err = os.MkdirAll(filepath.Join(project.BuildDir, "resource_pack"), os.ModePerm)
	if err != nil {
		return errors.NewError(errors.ERR_IO, project.BuildDir, err.Error())
	}

	cli.LogInfo(true, "Copying 'assets/'")
	path := filepath.Join(project.BuildDir, "resource_pack", "assets")
	err = cp.Copy("assets", path)
	if err != nil {
		return errors.NewError(errors.ERR_IO, path, err.Error())
	}

	if project.has_icon {
		path := filepath.Join(project.BuildDir, "resource_pack", "pack.png")
		err = cp.Copy("pack.png", path)
		if err != nil {
			return errors.NewError(errors.ERR_IO, path, err.Error())
		}
	}

	return nil
}

func (project *Project) createDataPack() error {
	_, err := os.Stat("data")
	if os.IsNotExist(err) {
		cli.LogDebug(false, "No data pack found")
		return nil
	}

	cli.LogInfo(false, "Creating a data pack")

	err = os.MkdirAll(filepath.Join(project.BuildDir, "data_pack"), os.ModePerm)
	if err != nil {
		return errors.NewError(errors.ERR_IO, project.BuildDir, err.Error())
	}

	data_entries, err := os.ReadDir("data")
	if err != nil {
		work_dir, _ := os.Getwd()
		return errors.NewError(errors.ERR_IO, filepath.Join(work_dir, "data"), err.Error())
	}

	function_paths := []string{}
	var errs errgroup.Group

	for _, data_entry := range data_entries {
		if !data_entry.IsDir() {
			cli.LogDebug(true, "Skipping file %q", data_entry.Name())
			continue
		}

		folder_entries, err := os.ReadDir(filepath.Join("data", data_entry.Name()))
		if err != nil {
			work_dir, _ := os.Getwd()
			return errors.NewError(
				errors.ERR_IO,
				filepath.Join(work_dir, "data", data_entry.Name()),
				err.Error(),
			)
		}

		for _, folder_entry := range folder_entries {
			path := filepath.Join("data", data_entry.Name(), folder_entry.Name())
			if !folder_entry.IsDir() {
				cli.LogDebug(true, "Skipping file %q", path)
				continue
			}

			switch folder_entry.Name() {
			case "function", "functions":
				function_paths = append(function_paths, path)
			default:
				cli.LogDebug(true, "Copying directory %q", path)
				errs.Go(func() error {
					return cp.Copy(
						path,
						filepath.Join(
							project.BuildDir,
							"data_pack",
							"data",
							data_entry.Name(),
							folder_entry.Name(),
						),
					)
				})
			}
		}
	}

	if err := errs.Wait(); err != nil {
		work_dir, _ := os.Getwd()
		return errors.NewError(
			errors.ERR_INTERNAL,
			filepath.Join(work_dir, "data"),
			err.Error(),
		)
	}

	cli.LogInfo(true, "Parsing mcfunction files")
	for _, path := range function_paths {
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
		work_dir, _ := os.Getwd()
		return errors.NewError(
			errors.ERR_INTERNAL,
			filepath.Join(work_dir, "data"),
			err.Error(),
		)
	}

	cli.LogInfo(true, "Writing mcfunction files to disk")

	// TODO: this

	return nil
}

func (project *Project) parseFunction(path string) error {
	// TODO: this
	return nil
}

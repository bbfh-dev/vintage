package mime

import (
	"os"
	"path/filepath"
	"time"

	"github.com/bbfh-dev/mime/cli"
	"github.com/bbfh-dev/mime/mime/errors"
	"github.com/bbfh-dev/mime/mime/minecraft"
)

type Project struct {
	Meta     *minecraft.PackMcmeta
	has_icon bool
	task_err error
}

func New(mcmeta *minecraft.PackMcmeta) *Project {
	return &Project{
		Meta:     mcmeta,
		has_icon: false,
		task_err: nil,
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
	dir, _ := filepath.Abs(cli.Main.Options.Output)
	cli.LogInfo(false, "Checking build directory: %s", dir)

	stat, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			cli.LogDebug(true, "Directory doesn't exist yet. Skipping checks")
			return nil
		}
		return errors.NewError(errors.ERR_IO, dir, err.Error())
	}

	if !stat.IsDir() {
		return errors.NewError(
			errors.ERR_VALID,
			dir,
			"build output is a file",
		)
	}

	return nil
}

func (project *Project) clearBuildDir() error {
	cli.LogInfo(false, "Clearing build directory")

	if err := os.RemoveAll(cli.Main.Options.Output); err != nil {
		return errors.NewError(errors.ERR_IO, cli.Main.Options.Output, err.Error())
	}

	path := filepath.Join(cli.Main.Options.Output, "data_pack")
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return errors.NewError(errors.ERR_IO, path, err.Error())
	}

	return nil
}

func (project *Project) detectPackIcon() error {
	_, err := os.Stat("pack.png")
	if !os.IsNotExist(err) {
		cli.LogInfo(false, "Found 'pack.png'")
		project.has_icon = true
	}

	return nil
}

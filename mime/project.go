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
	task_err error
}

func New(mcmeta *minecraft.PackMcmeta) *Project {
	return &Project{
		Meta:     mcmeta,
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
			"build output is a file. Must not exist or be a directory",
		)
	}

	return nil
}

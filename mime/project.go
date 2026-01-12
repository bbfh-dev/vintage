package mime

import (
	"os"
	"path/filepath"
	"time"

	"github.com/bbfh-dev/mime/cli"
	"github.com/bbfh-dev/mime/mime/errors"
	"github.com/bbfh-dev/mime/mime/minecraft"
	cp "github.com/otiai10/copy"
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

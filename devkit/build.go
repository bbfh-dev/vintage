package devkit

import (
	"os"
	"time"

	liberrors "github.com/bbfh-dev/lib-errors"
	liblog "github.com/bbfh-dev/lib-log"
	"github.com/bbfh-dev/vintage/cli"
	"github.com/bbfh-dev/vintage/devkit/internal/drive"
	"github.com/bbfh-dev/vintage/devkit/minecraft"
)

func Build(raw_args []string) error {
	if err := cli.ApplyWorkDir(cli.Build.Args.WorkDir); err != nil {
		return err
	}

	// Sync DEBUG
	if cli.Main.Options.Debug || cli.Build.Options.Debug {
		cli.Main.Options.Debug = true
		cli.Build.Options.Debug = true
		liblog.LogLevel = liblog.LEVEL_DEBUG
	}

	mcmeta_body, err := os.ReadFile("pack.mcmeta")
	if err != nil {
		work_dir, _ := os.Getwd()
		return liberrors.NewIO(err, work_dir)
	}

	mcmeta := minecraft.NewPackMcmeta(mcmeta_body)
	if err := mcmeta.Validate(); err != nil {
		return &liberrors.DetailedError{
			Label:   liberrors.ERR_VALIDATE,
			Context: liberrors.DirContext{Path: drive.ToAbs("pack.mcmeta")},
			Details: err.Error(),
		}
	}

	cli.UsesPluralFolderNames = minecraft.UsesPluralFolderNames(mcmeta.Minecraft()[0])

	start := time.Now()
	project := New(mcmeta)
	if err := project.Build(); err != nil {
		return err
	}

	liblog.Done(0, "Finished building in %s", time.Since(start))
	return nil
}

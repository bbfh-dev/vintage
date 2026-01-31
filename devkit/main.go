package devkit

import (
	"os"
	"path/filepath"
	"time"

	liberrors "github.com/bbfh-dev/lib-errors"
	"github.com/bbfh-dev/mime/cli"
	"github.com/bbfh-dev/mime/devkit/minecraft"
)

func Main(raw_args []string) error {
	if cli.Main.Args.WorkDir != nil {
		if err := os.Chdir(*cli.Main.Args.WorkDir); err != nil {
			return liberrors.NewIO(err, *cli.Main.Args.WorkDir)
		}
	}

	mcmeta_body, err := os.ReadFile("pack.mcmeta")
	if err != nil {
		work_dir, _ := os.Getwd()
		return liberrors.NewIO(err, work_dir)
	}

	mcmeta := minecraft.NewPackMcmeta(mcmeta_body)
	if err := mcmeta.Validate(); err != nil {
		path, _ := filepath.Abs("pack.mcmeta")
		return &liberrors.DetailedError{
			Label:   liberrors.ERR_VALIDATE,
			Context: liberrors.DirContext{Path: path},
			Details: err.Error(),
		}
	}

	start := time.Now()
	project := New(mcmeta)
	if err := project.Build(); err != nil {
		return err
	}

	cli.LogDone(0, "Finished building in %s", time.Since(start))
	return nil
}

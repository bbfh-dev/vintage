package cli

import (
	"os"

	liberrors "github.com/bbfh-dev/lib-errors"
)

var Main struct {
	Options struct {
		Debug bool `alt:"d" desc:"Print verbose debug information"`
	}
	Args struct{}
}

var UsesPluralFolderNames bool

func ApplyWorkDir(work_dir *string) error {
	if work_dir != nil {
		return liberrors.NewIO(os.Chdir(*work_dir), *work_dir)
	}
	return nil
}

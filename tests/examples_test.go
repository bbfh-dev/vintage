package vintage_test

import (
	"os"
	"path/filepath"
	"testing"

	liblog "github.com/bbfh-dev/lib-log"
	"github.com/bbfh-dev/vintage/cli"
	"github.com/bbfh-dev/vintage/devkit"
	"gotest.tools/assert"
)

func TestExamples(t *testing.T) {
	work_dir, err := os.Getwd()
	assert.NilError(t, err)

	work_dir = filepath.Join(work_dir, "..")
	assert.NilError(t, os.Chdir(work_dir))

	entries, err := os.ReadDir("examples")
	assert.NilError(t, err)

	cli.Main.Options.Debug = testing.Verbose()
	liblog.LogLevel = liblog.LEVEL_DEBUG
	cli.Build.Options.Force = true
	cli.Build.Options.Output = filepath.Join(os.TempDir(), "vintage-test")
	cli.Build.Options.Zip = true

	for _, entry := range entries {
		path := filepath.Join("examples", entry.Name())

		// Reset
		devkit.Reset()
		assert.NilError(t, os.RemoveAll(cli.Build.Options.Output))
		assert.NilError(t, os.Chdir(work_dir))

		t.Run(entry.Name(), func(t *testing.T) {
			liblog.Output = t.Output()
			cli.Build.Args.WorkDir = &path
			err := devkit.Build([]string{path})
			assert.NilError(t, err)
		})
	}
}

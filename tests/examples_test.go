package mime_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bbfh-dev/mime/cli"
	"github.com/bbfh-dev/mime/devkit"
	"github.com/bbfh-dev/mime/devkit/language"
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
	cli.Main.Options.Force = true
	cli.Main.Options.Output = filepath.Join(os.TempDir(), "mime-test")
	cli.Main.Options.Zip = true

	for _, entry := range entries {
		path := filepath.Join("examples", entry.Name())

		// Reset
		language.Registry = map[string][]string{}
		assert.NilError(t, os.RemoveAll(cli.Main.Options.Output))
		assert.NilError(t, os.Chdir(work_dir))

		t.Run(entry.Name(), func(t *testing.T) {
			cli.Output = t.Output()
			cli.Main.Args.WorkDir = &path
			err := devkit.Main([]string{path})
			assert.NilError(t, err)
		})
	}
}

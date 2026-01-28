package devkit

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	liberrors "github.com/bbfh-dev/lib-errors"
	"github.com/bbfh-dev/mime/cli"
	"github.com/bbfh-dev/mime/devkit/internal"
	cp "github.com/otiai10/copy"
	"golang.org/x/sync/errgroup"
)

func (project *Project) WeldPacks() error {
	if project.isDataCached && project.isAssetsCached {
		return nil
	}

	_, err := os.Stat("libs")
	if os.IsNotExist(err) {
		cli.LogDebug(0, "No libraries found")
		return nil
	}

	cli.LogInfo(0, "Merging with Smithed Weld")
	return internal.Pipeline(
		internal.Async(
			internal.If(
				!project.isDataCached,
				project.weld("data_packs", project.getZipPath("DP")),
			),
			internal.If(
				!project.isAssetsCached,
				project.weld("resource_packs", project.getZipPath("RP")),
			),
		),
	)
}

func (project *Project) weld(dir, zip_name string) internal.AsyncTask {
	return func(errs *errgroup.Group) error {
		start := time.Now()
		output_name := fmt.Sprintf("weld-%s.zip", dir)

		path := filepath.Join("libs", dir)
		entries, err := readLibDir(path)
		if err != nil {
			return err
		}
		entries[len(entries)-1] = zip_name

		if len(entries) < 2 {
			cli.LogDebug(1, "No libraries found for %q. Skipping...", dir)
			return nil
		}

		cmd := exec.Command("weld", append([]string{
			"--dir",
			project.BuildDir,
			"--name",
			output_name,
		}, entries...)...)

		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			os.Stdout.Write(out.Bytes())
			return &liberrors.DetailedError{
				Label:   liberrors.ERR_EXECUTE,
				Context: liberrors.DirContext{Path: path},
				Details: err.Error(),
			}
		}

		path = filepath.Join(project.BuildDir, output_name)
		err = errors.Join(cp.Copy(path, zip_name), os.Remove(path))
		if err != nil {
			return liberrors.NewIO(err, path)
		}

		cli.LogDone(1, "Merged %q in %s", zip_name, time.Since(start))
		return nil
	}
}

func readLibDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, liberrors.NewIO(err, internal.ToAbs("libs"))
	}

	files := make([]string, len(entries)+1)
	for i, entry := range entries {
		files[i] = filepath.Join(dir, entry.Name())
	}

	return files, nil
}

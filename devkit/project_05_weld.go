package devkit

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	liberrors "github.com/bbfh-dev/lib-errors"
	liblog "github.com/bbfh-dev/lib-log"
	"github.com/bbfh-dev/vintage/devkit/internal/drive"
	"github.com/bbfh-dev/vintage/devkit/internal/pipeline"
	cp "github.com/otiai10/copy"
	"golang.org/x/sync/errgroup"
)

func (project *Project) WeldPacks() error {
	if project.isDataCached && project.isAssetsCached {
		return nil
	}

	_, err := os.Stat("libs")
	if os.IsNotExist(err) {
		liblog.Debug(0, "No libraries found")
		return nil
	}

	liblog.Info(0, "Merging with Smithed Weld")
	return pipeline.New(
		pipeline.Async(
			pipeline.If[pipeline.AsyncTask](!project.isDataCached).
				Then(
					project.weld("data_packs", project.getZipPath("DP")),
				),
			pipeline.If[pipeline.AsyncTask](!project.isAssetsCached).
				Then(
					project.weld("resource_packs", project.getZipPath("RP")),
				),
		),
	)
}

func (project *Project) weld(dir, zip_name string) pipeline.AsyncTask {
	return func(errs *errgroup.Group) error {
		start := time.Now()
		output_name := fmt.Sprintf("weld-%s.zip", dir)

		path := filepath.Join("libs", dir)

		if _, err := os.Stat(path); os.IsNotExist(err) {
			liblog.Debug(1, "%q does not exist. Skipping...", dir)
			return nil
		}

		entries, err := readLibDir(path)
		if err != nil {
			return err
		}
		entries = append(entries, zip_name)

		if len(entries) < 2 {
			liblog.Debug(1, "No libraries found for %q. Skipping...", dir)
			return nil
		}

		args := append([]string{"--dir", project.BuildDir, "--name", output_name}, entries...)
		liblog.Debug(1, "$ weld %s", strings.Join(args, " "))
		cmd := exec.Command("weld", args...)

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

		liblog.Done(1, "Merged %q in %s", zip_name, time.Since(start))
		return nil
	}
}

func readLibDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, liberrors.NewIO(err, drive.ToAbs("libs"))
	}

	files := make([]string, 0, len(entries)+1)
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".zip" {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	return files, nil
}

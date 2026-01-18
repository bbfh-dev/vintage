package mime

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bbfh-dev/mime/cli"
	"github.com/bbfh-dev/mime/mime/errors"
	"golang.org/x/sync/errgroup"
)

func (project *Project) runWeld() error {
	_, err := os.Stat("libs")
	if os.IsNotExist(err) {
		cli.LogDebug(false, "No libraries found")
		return nil
	}

	cli.LogInfo(false, "Merging with Smithed Weld")
	var errs errgroup.Group

	if _, err = os.Stat(filepath.Join("libs", "data_packs")); err == nil {
		errs.Go(func() error {
			return project.weldPack("data_packs", project.data_zip_name)
		})
	}

	if _, err = os.Stat(filepath.Join("libs", "resource_packs")); err == nil {
		errs.Go(func() error {
			return project.weldPack("resource_packs", project.resources_zip_name)
		})
	}

	return errs.Wait()
}

func (project *Project) weldPack(dir, zip_name string) error {
	if zip_name == "" {
		cli.LogError(true, "--zip flag must be provided for merging tool to work!")
		return nil
	}

	output_name := fmt.Sprintf("weld-%s.zip", dir)

	work_dir, _ := os.Getwd()
	path := filepath.Join(work_dir, "libs", dir)
	entries, err := readLibDir(path)
	if err != nil {
		return err
	}
	entries[len(entries)-1] = zip_name

	cmd := exec.Command(
		"weld",
		append([]string{"--dir", project.BuildDir, "--name", output_name}, entries...)...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Stdout.Write(out.Bytes())
		return errors.NewError(errors.ERR_EXEC, path, err.Error())
	}

	err = MoveFile(filepath.Join(project.BuildDir, output_name), zip_name)
	if err != nil {
		return errors.NewError(errors.ERR_IO, path, err.Error())
	}
	cli.LogDone(true, "Finished merging %q", zip_name)

	return nil
}

func readLibDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		work_dir, _ := os.Getwd()
		return nil, errors.NewError(errors.ERR_IO, filepath.Join(work_dir, "libs"), err.Error())
	}

	files := make([]string, len(entries)+1)
	for i, entry := range entries {
		files[i] = filepath.Join(dir, entry.Name())
	}

	return files, nil
}

// @source: https://www.tutorialpedia.org/blog/move-a-file-to-a-different-drive-with-go/
func MoveFile(src, dst string) error {
	srcFileInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}
	if srcFileInfo.IsDir() {
		return fmt.Errorf("source is a directory: %s", src)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination: %w", err)
	}

	if err := os.Chmod(dst, srcFileInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	if err := os.Remove(src); err != nil {
		return fmt.Errorf("failed to delete source: %w", err)
	}

	return nil
}

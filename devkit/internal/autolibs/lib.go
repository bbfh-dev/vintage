// TODO: Refactor this entire monstrosity
package autolibs

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	liberrors "github.com/bbfh-dev/lib-errors"
	liblog "github.com/bbfh-dev/lib-log"
	"github.com/bbfh-dev/vintage/cli"
	"github.com/bbfh-dev/vintage/devkit/internal/code"
	"github.com/bbfh-dev/vintage/devkit/internal/drive"
	"github.com/schollz/progressbar/v3"
	"github.com/tidwall/gjson"
)

type Library struct {
	Namespaces []string
	Path       string
	File       *drive.JsonFile
}

func (lib *Library) IsUsed() bool {
	return len(lib.Namespaces) != 0
}

func (lib *Library) Dir() string {
	return filepath.Dir(lib.Path)
}

func (lib *Library) Name() string {
	return filepath.Base(lib.Path)
}

func (lib *Library) DownloadURL(namespace string) (string, error) {
	env := code.NewEnv()
	env.Variables["namespace"] = code.SimpleVariable(namespace)
	return code.SubstituteString(lib.File.Get("download").String(), env)
}

func (lib *Library) Validate() error {
	return lib.File.ExpectField("download", gjson.String)
}

func (lib *Library) GetMissing() []string {
	field := lib.File.Get("installed")
	if !field.Exists() {
		return lib.Namespaces
	}

	missing := make([]string, 0, len(lib.Namespaces))

	for _, namespace := range lib.Namespaces {
		contains := false
		for _, item := range field.Array() {
			if strings.HasPrefix(item.String(), namespace) {
				contains = true
				break
			}
		}
		if !contains {
			missing = append(missing, namespace)
		}
	}

	return missing
}

func (lib *Library) RemoveInstalled(filename string) {
	field := lib.File.Get("installed")
	items := []string{}

	if field.Exists() {
		for _, item := range field.Array() {
			if item.String() != filename {
				items = append(items, item.String())
			}
		}
	}

	lib.File.Set("installed", items)
}

func (lib *Library) AddInstalled(filename string) {
	field := lib.File.Get("installed")
	items := []string{filename}

	if field.Exists() {
		for _, item := range field.Array() {
			items = append(items, item.String())
		}
	}

	lib.File.Set("installed", items)
}

func (lib *Library) SaveToDisk() error {
	err := os.WriteFile(lib.Path, lib.File.Formatted(), os.ModePerm)
	if err != nil {
		return liberrors.NewIO(err, lib.Path)
	}
	return nil
}

func (lib *Library) Install() error {
	missing := lib.GetMissing()
	if len(missing) == 0 {
		liblog.Debug(2, "%q files are already installed", lib.Name())
		return nil
	}

	for _, namespace := range missing {
		url, err := lib.DownloadURL(namespace)
		filename := getDownloadFilename(url)
		path := filepath.Join(lib.Dir(), filename)

		if _, err := os.Stat(path); err == nil {
			goto next_iteration
		}

		if _, err := os.Stat(path + ".disabled"); err == nil {
			err := os.Rename(path+".disabled", path)
			if err != nil {
				return liberrors.NewIO(err, path)
			}
			goto next_iteration
		}

		err = lib.Download(namespace)
		if err != nil {
			return err
		}

	next_iteration:
		lib.AddInstalled(filename)
		if err := lib.SaveToDisk(); err != nil {
			return err
		}

		liblog.Info(2, "Installed %q", namespace)
	}

	return nil
}

func (lib *Library) Manage() error {
	// TODO: Check if the files actually exist
	if lib.IsUsed() {
		if err := lib.Install(); err != nil {
			return err
		}
	}

	field := lib.File.Get("installed")
	if !field.Exists() {
		return nil
	}

	for _, installed := range field.Array() {
		keep := false
		for _, namespace := range lib.Namespaces {
			if strings.HasPrefix(installed.String(), namespace) {
				keep = true
				break
			}
		}
		if keep {
			continue
		}
		path := filepath.Join(lib.Dir(), installed.String())

		if cli.Build.Options.DeleteUnusedLibs {
			err := os.Remove(path)
			if err != nil {
				return liberrors.NewIO(err, path)
			}
			liblog.Info(2, "Deleted %q", path)
		} else {
			err := os.Rename(path, path+".disabled")
			if err != nil {
				return liberrors.NewIO(err, path)
			}
			liblog.Info(2, "Disabled %q", path)
		}

		lib.RemoveInstalled(installed.String())
		if err := lib.SaveToDisk(); err != nil {
			return err
		}
	}

	return nil
}

func (lib *Library) Download(namespace string) error {
	url, err := lib.DownloadURL(namespace)
	if err != nil {
		return &liberrors.DetailedError{
			Label:   liberrors.ERR_FORMAT,
			Context: liberrors.DirContext{Path: lib.Path},
			Details: err.Error(),
		}
	}

	filename := getDownloadFilename(url)
	path := filepath.Join(lib.Dir(), filename)
	liblog.Debug(2, "Downloading %s", url)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return liberrors.NewIO(err, url)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return liberrors.NewIO(err, url)
	}
	defer response.Body.Close()

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return liberrors.NewIO(err, path)
	}
	defer file.Close()

	bar := progressbar.DefaultBytes(response.ContentLength, filename)
	_, err = io.Copy(io.MultiWriter(file, bar), response.Body)
	if err != nil {
		return liberrors.NewIO(err, path)
	}

	return nil
}

func getDownloadFilename(url string) string {
	url_parts := strings.Split(url, "/")
	filename := url_parts[len(url_parts)-1]
	if !strings.HasSuffix(filename, ".zip") {
		filename += ".zip"
	}
	return filename
}

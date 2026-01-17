package addons

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bbfh-dev/mime/mime/errors"
	"github.com/bbfh-dev/mime/mime/minecraft"
	"github.com/tidwall/gjson"
)

type Columns []string

type Rows []Columns

type Addon struct {
	SourceDir string
	BuildDir  string
	Iterators map[string]Rows
	Paths     []string
}

func New(source_dir, build_dir string) *Addon {
	return &Addon{
		SourceDir: source_dir,
		BuildDir:  build_dir,
		Iterators: map[string]Rows{},
		Paths:     []string{},
	}
}

func (addon *Addon) Build() error {
	entries, err := os.ReadDir(addon.SourceDir)
	if err != nil {
		return errors.NewError(errors.ERR_IO, addon.SourceDir, err.Error())
	}

	env := NewEnv()

	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), "_") {
			continue
		}

		path := filepath.Join(addon.SourceDir, entry.Name())
		body, err := os.ReadFile(path)
		if err != nil {
			return errors.NewError(errors.ERR_IO, path, err.Error())
		}
		file := minecraft.NewJsonFile(body)

		iterators := extractIteratorsFrom(entry.Name())
		if len(iterators) == 0 {
			err = addon.define(trimExt(entry.Name()), file, env)
		} else {
			err = addon.BuildWithIterators(entry.Name(), iterators, file)
		}
		if err != nil {
			return errors.NewError(errors.ERR_ITER, path, err.Error())
		}
	}

	return nil
}

func (addon *Addon) BuildWithIterators(
	name string,
	iterators []string,
	file *minecraft.JsonFile,
) error {
	resolved := make([]Rows, len(iterators))
	identifiers := make([]string, len(iterators))

	for i, iterator := range iterators {
		parts := strings.SplitN(iterator, ".", 2)
		identifier, csv_index := parts[0], 0

		if len(parts) == 2 {
			value, err := strconv.Atoi(parts[1])
			if err != nil {
				return err
			}
			csv_index = value
		}

		rows, ok := addon.Iterators[identifier]
		switch {
		case !ok:
			return fmt.Errorf("undefined iterator %q", identifier)
		case len(rows) == 0:
			continue
		case csv_index >= len(rows[0]):
			return fmt.Errorf("csv index out of range")
		}

		resolved[i] = rows
		identifiers[i] = identifier
	}

	indices := make([]int, len(resolved))
	env := NewEnv()
	n := 0

	for {
		for i := range resolved {
			env.Iterators[identifiers[i]] = resolved[i][indices[i]]
		}

		in, err := SubstituteString(name, env)
		if err != nil {
			return errors.NewError(
				errors.ERR_ITER,
				filepath.Join(addon.SourceDir, name),
				err.Error(),
			)
		}

		env.Variables["i"] = gjson.Result{
			Type: gjson.Number,
			Num:  float64(n),
		}
		err = addon.define(in, file, env)
		if err != nil {
			return err
		}
		n++

		pos := len(indices) - 1
		for pos >= 0 {
			indices[pos]++
			if indices[pos] < len(resolved[pos]) {
				break
			}
			indices[pos] = 0
			pos--
		}
		if pos < 0 {
			break
		}
	}

	return nil
}

func (addon *Addon) define(name string, in *minecraft.JsonFile, env Env) error {
	name = trimExt(name)
	fmt.Println("$:", name)

	file := in.Clone()
	err := SubstituteFile(file, env)
	if err != nil {
		return err
	}

	local_env := NewEnv()
	maps.Copy(local_env.Iterators, env.Iterators)
	maps.Copy(local_env.Variables, env.Variables)

	for _, key := range file.Get("@keys").Array() {
		local_env.Variables[key.String()] = file.Get(key.String())
	}
	local_env.Variables["id"] = gjson.Result{
		Type: gjson.String,
		Str:  name,
	}

	for _, path := range addon.Paths {
		new_path, err := SubstituteString(path, local_env)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		fmt.Println(new_path)
	}
	fmt.Println(addon.SourceDir)
	fmt.Println(string(file.Formatted()))
	return nil
}

func trimExt(name string) string {
	return strings.TrimSuffix(name, filepath.Ext(name))
}

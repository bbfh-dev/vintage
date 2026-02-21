package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	liberrors "github.com/bbfh-dev/lib-errors"
	liblog "github.com/bbfh-dev/lib-log"
	"github.com/bbfh-dev/vintage/devkit/internal/code"
	"github.com/bbfh-dev/vintage/devkit/internal/drive"
	"github.com/tidwall/gjson"
	"golang.org/x/sync/errgroup"
)

type Definition struct {
	File *drive.JsonFile
	Env  code.Env
}

type Generator struct {
	Root        string
	Iterators   map[string]code.Rows
	Definitions map[string]Definition
}

func NewGenerator(root string, manifest *drive.JsonFile) (*Generator, error) {
	template := &Generator{
		Root:        root,
		Iterators:   map[string]code.Rows{},
		Definitions: map[string]Definition{},
	}

	if field_iters := manifest.Get("iterators"); field_iters.Exists() {
		if !field_iters.IsObject() {
			return nil, newSyntaxError(
				filepath.Join(root, "manifest.json"),
				"field 'iterators' must be an object",
				field_iters,
			)
		}

		for _, key := range field_iters.Get("@keys").Array() {
			values := field_iters.Get(key.String())
			rows := code.Rows{}
			for i, row := range values.Array() {
				cols := code.Columns{}
				for j, col := range row.Array() {
					if col.Type != gjson.String {
						return nil, newSyntaxError(
							filepath.Join(root, "manifest.json"),
							fmt.Sprintf(
								"field 'iterators.%s[%d][%d]' must be a string",
								key.String(),
								i, j,
							),
							col,
						)
					}
					cols = append(cols, col.String())
				}
				rows = append(rows, cols)
			}
			template.Iterators[key.String()] = rows
		}
	}

	dir := filepath.Join(root, "definitions")
	entries, err := os.ReadDir(dir)
	if err != nil {
		liblog.Warn(2, "%q has no definitions", root)
		return nil, nil
	}

	var errs errgroup.Group
	var mutex sync.Mutex

	for entry := range drive.IterateFilesOnly(entries) {
		errs.Go(func() error {
			path := filepath.Join(dir, entry.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				return liberrors.NewIO(err, path)
			}

			file := drive.NewJsonFile(data)
			mutex.Lock()

			extracted_iters := code.ExtractVariablesFrom(entry.Name())
			if len(extracted_iters) == 0 {
				template.Definitions[entry.Name()] = Definition{
					File: file,
					Env:  code.NewEnv(),
				}
			} else {
				err := template.defineUsingIterators(entry.Name(), extracted_iters, file)
				if err != nil {
					mutex.Unlock()
					return err
				}
			}

			mutex.Unlock()
			return nil
		})
	}

	if err := errs.Wait(); err != nil {
		return nil, err
	}

	for name, definition := range template.Definitions {
		definition.Env.Variables["id"] = code.SimpleVariable(
			strings.TrimSuffix(name, filepath.Ext(name)),
		)
		definition.Env.Variables["filename"] = code.SimpleVariable(name)
		for _, key := range definition.File.Get("@keys").Array() {
			value := definition.File.Get(key.String())
			definition.Env.Variables[key.String()] = value
		}
	}

	return template, nil
}

func (template *Generator) defineUsingIterators(
	name string,
	iterators []string,
	file *drive.JsonFile,
) error {
	resolved := make([]code.Rows, len(iterators))
	identifiers := make([]string, len(iterators))

	for i, iterator := range iterators {
		parts := strings.SplitN(iterator, ".", 2)
		identifier, item_index := parts[0], 0

		if len(parts) == 2 {
			value, err := strconv.Atoi(parts[1])
			if err != nil {
				return &liberrors.DetailedError{
					Label: liberrors.ERR_SYNTAX,
					Context: liberrors.DirContext{
						Path: filepath.Join(template.Root, "templates", name),
					},
					Details: err.Error(),
				}
			}
			item_index = value
		}

		rows, ok := template.Iterators[identifier]
		switch {
		case !ok:
			return &liberrors.DetailedError{
				Label: liberrors.ERR_VALIDATE,
				Context: liberrors.DirContext{
					Path: filepath.Join(template.Root, "templates", name),
				},
				Details: fmt.Sprintf("undefined iterator %q", identifier),
			}
		case len(rows) == 0:
			continue
		case item_index >= len(rows[0]):
			return &liberrors.DetailedError{
				Label: liberrors.ERR_VALIDATE,
				Context: liberrors.DirContext{
					Path: filepath.Join(template.Root, "templates", name),
				},
				Details: fmt.Sprintf("index %d is out of range for %q", item_index, rows[0]),
			}
		}

		resolved[i] = rows
		identifiers[i] = identifier
	}

	indices := make([]int, len(resolved))
	n := 0

	for {
		env := code.NewEnv()

		for i := range resolved {
			env.Iterators[identifiers[i]] = resolved[i][indices[i]]
		}

		in, err := code.SubstituteString(name, env)
		if err != nil {
			return &liberrors.DetailedError{
				Label: liberrors.ERR_FORMAT,
				Context: liberrors.DirContext{
					Path: filepath.Join(template.Root, "templates", name),
				},
				Details: err.Error(),
			}
		}

		env.Variables["i"] = gjson.Result{
			Type: gjson.Number,
			Num:  float64(n),
		}

		file := file.Clone()
		if err := code.SubstituteJsonFile(file, env); err != nil {
			return &liberrors.DetailedError{
				Label: liberrors.ERR_FORMAT,
				Context: liberrors.DirContext{
					Path: filepath.Join(template.Root, "templates", name),
				},
				Details: err.Error(),
			}
		}

		template.Definitions[in] = Definition{
			File: file,
			Env:  env,
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

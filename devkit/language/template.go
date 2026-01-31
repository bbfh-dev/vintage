package language

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	liberrors "github.com/bbfh-dev/lib-errors"
	"github.com/bbfh-dev/mime/cli"
	"github.com/bbfh-dev/mime/devkit/internal"
	"github.com/tidwall/gjson"
	"golang.org/x/sync/errgroup"
)

const BODY_SUBSTITUTION = "%[...]"

// ————————————————————————————————

type Definition struct {
	File *internal.JsonFile
	Env  internal.Env
}

type GeneratorTemplate struct {
	Dir         string
	Iterators   map[string]internal.Rows
	Definitions map[string]Definition
}

func NewGeneratorTemplate(root string, manifest *internal.JsonFile) (*GeneratorTemplate, error) {
	template := &GeneratorTemplate{
		Dir:         root,
		Iterators:   map[string]internal.Rows{},
		Definitions: map[string]Definition{},
	}

	if field_iters := manifest.Get("iterators"); field_iters.Exists() {
		if !field_iters.IsObject() {
			return nil, newSyntaxError(
				filepath.Join(root, "manifest.json"),
				fmt.Sprintf(
					"field 'iterators' must be an object, got (%s) %q",
					field_iters.Type.String(),
					field_iters.String(),
				),
			)
		}

		for _, key := range field_iters.Get("@keys").Array() {
			values := field_iters.Get(key.String())
			rows := internal.Rows{}
			for i, row := range values.Array() {
				cols := internal.Columns{}
				for j, col := range row.Array() {
					if col.Type != gjson.String {
						return nil, newSyntaxError(
							filepath.Join(root, "manifest.json"),
							fmt.Sprintf(
								"field 'iterators.%s[%d][%d]' must be a string, got (%s) %q",
								key.String(),
								i, j,
								col.Type.String(),
								col.String(),
							),
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
		cli.LogWarn(2, "%q has no definitions", root)
		return nil, nil
	}

	var errs errgroup.Group
	var mutex sync.Mutex

	for entry := range internal.IterateFilesOnly(entries) {
		errs.Go(func() error {
			path := filepath.Join(dir, entry.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				return liberrors.NewIO(err, path)
			}

			file := internal.NewJsonFile(data)
			mutex.Lock()

			extracted_iters := internal.ExtractVariablesFrom(entry.Name())
			if len(extracted_iters) == 0 {
				template.Definitions[entry.Name()] = Definition{
					File: file,
					Env:  internal.NewEnv(),
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
		definition.Env.Variables["id"] = gjson.Result{
			Type: gjson.String,
			Str:  strings.TrimSuffix(name, filepath.Ext(name)),
		}
		definition.Env.Variables["filename"] = gjson.Result{
			Type: gjson.String,
			Str:  name,
		}
		for _, key := range definition.File.Get("@keys").Array() {
			value := definition.File.Get(key.String())
			definition.Env.Variables[key.String()] = value
		}
	}

	return template, nil
}

func (template *GeneratorTemplate) defineUsingIterators(
	name string,
	iterators []string,
	file *internal.JsonFile,
) error {
	resolved := make([]internal.Rows, len(iterators))
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
						Path: filepath.Join(template.Dir, "templates", name),
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
					Path: filepath.Join(template.Dir, "templates", name),
				},
				Details: fmt.Sprintf("undefined iterator %q", identifier),
			}
		case len(rows) == 0:
			continue
		case item_index >= len(rows[0]):
			return &liberrors.DetailedError{
				Label: liberrors.ERR_VALIDATE,
				Context: liberrors.DirContext{
					Path: filepath.Join(template.Dir, "templates", name),
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
		env := internal.NewEnv()

		for i := range resolved {
			env.Iterators[identifiers[i]] = resolved[i][indices[i]]
		}

		in, err := internal.SimpleSubstitute(name, env)
		if err != nil {
			return &liberrors.DetailedError{
				Label:   liberrors.ERR_FORMAT,
				Context: liberrors.DirContext{Path: filepath.Join(template.Dir, "templates", name)},
				Details: err.Error(),
			}
		}

		env.Variables["i"] = gjson.Result{
			Type: gjson.Number,
			Num:  float64(n),
		}

		file := file.Clone()
		if err := internal.SubstituteJsonFile(file, env); err != nil {
			return &liberrors.DetailedError{
				Label:   liberrors.ERR_FORMAT,
				Context: liberrors.DirContext{Path: filepath.Join(template.Dir, "templates", name)},
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

// ————————————————————————————————

type InlineTemplate struct {
	RequiredArgs []string
	Call         func(writer io.Writer, reader io.Reader, args []string) error
}

func (template *InlineTemplate) IsArgPassthrough() bool {
	return template.RequiredArgs == nil
}

func NewInlineTemplate(dir string, manifest *internal.JsonFile) (*InlineTemplate, error) {
	template := &InlineTemplate{RequiredArgs: nil}

	field_args := manifest.Get("arguments")
	if field_args.Exists() && field_args.Type != gjson.Null {
		switch {

		case field_args.IsArray():
			template.RequiredArgs = []string{}
			for _, value := range field_args.Array() {
				if value.Type != gjson.String {
					return nil, newSyntaxError(
						internal.ToAbs(dir),
						fmt.Sprintf(
							"field 'arguments' must be an array of strings, but got (%s) %q",
							value.Type.String(),
							value.String(),
						),
					)
				}
				template.RequiredArgs = append(template.RequiredArgs, value.String())
			}

		default:
			return nil, newSyntaxError(
				internal.ToAbs(dir),
				fmt.Sprintf(
					"field 'arguments' must be an object or equal to '*' (string), but got (%s) %q",
					field_args.Type.String(),
					field_args.String(),
				),
			)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, liberrors.NewIO(err, dir)
	}

	for _, entry := range entries {
		switch {

		case strings.HasPrefix(entry.Name(), "call"):
			path := filepath.Join(dir, entry.Name())
			template.Call = func(writer io.Writer, reader io.Reader, args []string) error {
				cmd := exec.Command(path, args...)
				cmd.Stdin = reader
				cmd.Stdout = writer
				cmd.Stderr = os.Stderr

				err := cmd.Run()
				if err != nil {
					return &liberrors.DetailedError{
						Label:   liberrors.ERR_EXECUTE,
						Context: liberrors.DirContext{Path: path},
						Details: err.Error(),
					}
				}
				return nil
			}
			return template, nil

		case strings.HasSuffix(entry.Name(), ".mcfunction"):
			body, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				return nil, liberrors.NewIO(err, internal.ToAbs(dir))
			}
			template.Call = func(writer io.Writer, reader io.Reader, args []string) error {
				env := internal.NewEnv()
				for i, arg := range args {
					env.Variables[template.RequiredArgs[i]] = gjson.Result{
						Type: gjson.String,
						Str:  arg,
					}
				}

				body := string(body)
				lines := strings.Split(body, "\n")

				var before strings.Builder
				var after string
				var ok bool
				for i, line := range lines {
					if strings.Contains(line, BODY_SUBSTITUTION) {
						after = strings.Join(lines[i+1:], "\n")
						ok = true
						break
					}
					before.WriteString(line + "\n")
				}

				if ok {
					str, err := internal.SimpleSubstitute(before.String(), env)
					if err != nil {
						return &liberrors.DetailedError{
							Label:   liberrors.ERR_FORMAT,
							Context: liberrors.DirContext{Path: filepath.Join(dir, entry.Name())},
							Details: err.Error(),
						}
					}
					writer.Write([]byte(str))

					io.Copy(writer, reader)

					str, err = internal.SimpleSubstitute(after, env)
					if err != nil {
						return &liberrors.DetailedError{
							Label:   liberrors.ERR_FORMAT,
							Context: liberrors.DirContext{Path: filepath.Join(dir, entry.Name())},
							Details: err.Error(),
						}
					}
					writer.Write([]byte(str))
				}

				return nil
			}

			return template, nil
		}
	}

	return template, &liberrors.DetailedError{
		Label:   liberrors.ERR_VALIDATE,
		Context: liberrors.DirContext{Path: internal.ToAbs(dir)},
		Details: fmt.Sprintf(
			"template %q contains no logic files. Must contain `*.mcfunction` or `call*`. Refer to documentation",
			filepath.Base(dir),
		),
	}
}

func newSyntaxError(path string, details string) *liberrors.DetailedError {
	return &liberrors.DetailedError{
		Label:   liberrors.ERR_SYNTAX,
		Context: liberrors.DirContext{Path: path},
		Details: details,
	}
}

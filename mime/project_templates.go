package mime

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	liberrors "github.com/bbfh-dev/lib-errors"
	"github.com/bbfh-dev/mime/cli"
	"github.com/bbfh-dev/mime/mime/internal"
	"github.com/bbfh-dev/mime/mime/language/templates"
	"github.com/bbfh-dev/mime/mime/minecraft"
	"github.com/tidwall/gjson"
)

func (project *Project) loadTemplates() error {
	if project.isDataCached && project.isAssetsCached {
		return nil
	}

	_, err := os.Stat("templates")
	if os.IsNotExist(err) {
		cli.LogDebug(false, "No templates found")
		return nil
	}

	cli.LogInfo(false, "Loading templates")
	template_entries, err := os.ReadDir("templates")
	if err != nil {
		return liberrors.NewIO(err, internal.ToAbs("templates"))
	}

	for entry := range internal.IterateDirsOnly(template_entries) {
		if err := project.loadTemplate(filepath.Join("templates", entry.Name())); err != nil {
			return err
		}
	}

	return nil
}

func (project *Project) loadTemplate(dir string) error {
	path := filepath.Join(dir, "manifest.json")
	manifest_data, err := os.ReadFile(path)
	if err != nil {
		return liberrors.NewIO(err, internal.ToAbs(path))
	}

	manifest := minecraft.NewJsonFile(manifest_data)

	if err := manifest.ExpectField("type", gjson.String); err != nil {
		return &liberrors.DetailedError{
			Label:   liberrors.ERR_VALIDATE,
			Context: liberrors.DirContext{Path: path},
			Details: err.Error(),
		}
	}

	manifest_type := manifest.Get("type").String()
	template := templates.New(dir, manifest)

	switch manifest_type {
	case "generate":
		return template.Generate()
	case "inline":
		return project.loadInlineTemplate(dir, manifest)
	}

	return &liberrors.DetailedError{
		Label:   liberrors.ERR_VALIDATE,
		Context: liberrors.DirContext{Path: path},
		Details: fmt.Sprintf(
			"field 'type' must be one of ['generate', 'inline'], but got %q",
			manifest_type,
		),
	}
}

func (project *Project) loadInlineTemplate(dir string, manifest *minecraft.JsonFile) error {
	cli.LogInfo(true, "Loading inline %q", dir)

	var arguments []templates.Argument
	manifest_args := manifest.Get("arguments")
	if manifest_args.Exists() {
		if !manifest_args.IsObject() {
			return &liberrors.DetailedError{
				Label:   liberrors.ERR_SYNTAX,
				Context: liberrors.DirContext{Path: dir},
				Details: "field 'arguments' must be an object with string values",
			}
		}

		for _, key := range manifest_args.Get("@keys").Array() {
			arguments = append(arguments, templates.Argument{
				key.String(),
				manifest_args.Get(key.String()).String(),
			})
		}
	}

	path := filepath.Join(dir, "body.mcfunction")
	if _, err := os.Stat(path); err == nil {
		body, err := os.ReadFile(path)
		if err != nil {
			return liberrors.NewIO(err, path)
		}

		project.inlineTemplates = append(project.inlineTemplates, &templates.InlineTemplate{
			Arguments: arguments,
			Call: func(writer io.Writer, args map[string]string, nested []string) error {
				return templates.InsertBody(writer, string(body), nested, args)
			},
		})
		return nil
	}

	path = filepath.Join(dir, "call")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &liberrors.DetailedError{}
	}

	project.inlineTemplates = append(project.inlineTemplates, &templates.InlineTemplate{
		Arguments: arguments,
		Call: func(writer io.Writer, args map[string]string, nested []string) error {
			in := []string{}
			if arguments != nil {
				panic("TODO")
			}

			cmd := exec.Command(path, in...)
			cmd.Stdin = strings.NewReader(strings.Join(nested, "\n"))
			cmd.Stdout = writer
			cmd.Stderr = os.Stderr
			return cmd.Run()
		},
	})

	return nil
}

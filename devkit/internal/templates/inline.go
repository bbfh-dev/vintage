package templates

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	liberrors "github.com/bbfh-dev/lib-errors"
	liblog "github.com/bbfh-dev/lib-log"
	"github.com/bbfh-dev/vintage/devkit/internal/code"
	"github.com/bbfh-dev/vintage/devkit/internal/drive"
	"github.com/tidwall/gjson"
)

const BODY_SUBSTITUTION = "%[...]"
const SNIPPET_FILENAME = "snippet.mcfunction"
const INLINE_CALL_PREFIX = "#~>"

type Inline struct {
	RequiredArgs []string
	Call         func(out Writer, in Scanner, args []string) error
}

func NewInline(dir string, manifest *drive.JsonFile) (*Inline, error) {
	template := &Inline{RequiredArgs: nil}

	field_args := manifest.Get("arguments")
	if field_args.Exists() {
		switch {

		case field_args.IsArray():
			template.RequiredArgs = []string{}
			for _, value := range field_args.Array() {
				if value.Type != gjson.String {
					return nil, newSyntaxError(
						drive.ToAbs(dir),
						"field 'arguments' must be an array of strings",
						value,
					)
				}
				template.RequiredArgs = append(template.RequiredArgs, value.String())
			}

		default:
			return nil, newSyntaxError(
				drive.ToAbs(dir),
				"field 'arguments' must be an array of strings",
				field_args,
			)
		}
	}

	path := filepath.Join(dir, SNIPPET_FILENAME)
	if _, err := os.Stat(path); err == nil {
		return inlineTemplateUsingSnippet(template, path)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, liberrors.NewIO(err, dir)
	}

	for entry := range drive.IterateFilesOnly(entries) {
		switch {
		case strings.HasPrefix(entry.Name(), "call"):
			path := filepath.Join(dir, entry.Name())
			return inlineTemplateUsingExec(template, path)
		}
	}

	return template, &liberrors.DetailedError{
		Label:   liberrors.ERR_VALIDATE,
		Context: liberrors.DirContext{Path: drive.ToAbs(dir)},
		Details: fmt.Sprintf(
			"template %q contains no logic files. Must contain `*.mcfunction` or `call*`. Refer to documentation",
			filepath.Base(dir),
		),
	}
}

func (template *Inline) IsArgPassthrough() bool {
	return template.RequiredArgs == nil
}

func newSyntaxError(path, details string, field gjson.Result) *liberrors.DetailedError {
	return &liberrors.DetailedError{
		Label:   liberrors.ERR_SYNTAX,
		Context: liberrors.DirContext{Path: path},
		Details: fmt.Sprintf("%s, but got (%s) %q", details, field.Type, field),
	}
}

func inlineTemplateUsingSnippet(template *Inline, path string) (*Inline, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, liberrors.NewIO(err, drive.ToAbs(path))
	}

	template.Call = func(out Writer, in Scanner, args []string) error {
		env := code.NewEnv()
		for i, arg := range args {
			env.Variables[template.RequiredArgs[i]] = code.SimpleVariable(arg)
		}

		body := string(body)
		lines := strings.Split(strings.TrimSuffix(body, "\n"), "\n")

		var before string
		var after string
		var ok bool
		for i, line := range lines {
			if strings.Contains(line, BODY_SUBSTITUTION) {
				if i > 0 {
					before = strings.Join(lines[:i], "\n")
				}
				after = strings.Join(lines[i+1:], "\n")
				ok = true
				break
			}
		}

		if !ok {
			before = strings.Join(lines, "\n")
			return writeSubstituted(out, path, before, env)
		}

		if err := writeSubstituted(out, path, before, env); err != nil {
			return err
		}
		for in.Scan() {
			out.Writeln(in.Text())
		}
		if err := writeSubstituted(out, path, after, env); err != nil {
			return err
		}

		return nil
	}

	return template, nil
}

func inlineTemplateUsingExec(template *Inline, path string) (*Inline, error) {
	template.Call = func(out Writer, in Scanner, args []string) error {

		cmd := exec.Command(path, args...)

		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		cmd.Stdin = in.Reader()
		cmd.Stdout = out

		path := fmt.Sprintf("%s with [%s]", path, strings.Join(args, " "))

		err := cmd.Run()
		if err != nil {
			return &liberrors.DetailedError{
				Label:   liberrors.ERR_EXECUTE,
				Context: liberrors.NewProgramContext(cmd, stderr.String()),
				Details: err.Error(),
			}
		}

		if stderr.Len() != 0 {
			liblog.Error(1, "From: %s", path)
			scanner := bufio.NewScanner(&stderr)
			for scanner.Scan() {
				liblog.Error(2, "%s", scanner.Text())
			}
		}

		return nil
	}

	return template, nil
}

func writeSubstituted(out Writer, path, in string, env code.Env) error {
	str, err := code.SubstituteString(in, env)
	if err != nil {
		return &liberrors.DetailedError{
			Label:   liberrors.ERR_FORMAT,
			Context: liberrors.DirContext{Path: path},
			Details: err.Error(),
		}
	}
	for line := range strings.SplitSeq(str, "\n") {
		out.Writeln(line)
	}
	return nil
}

func IsInlineCall(line string) bool {
	return strings.HasPrefix(line, INLINE_CALL_PREFIX)
}

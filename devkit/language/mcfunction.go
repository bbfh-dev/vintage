package language

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	liberrors "github.com/bbfh-dev/lib-errors"
	"github.com/bbfh-dev/mime/cli"
	"github.com/bbfh-dev/mime/devkit/internal"
)

type Mcfunction struct {
	Path    string
	Scanner *bufio.Scanner

	root    *Line
	current *Line
}

func NewMcfunction(path string, scanner *bufio.Scanner) *Mcfunction {
	root := newLine(nil, "<root>")
	return &Mcfunction{
		Path:    path,
		Scanner: scanner,
		root:    root,
		current: root,
	}
}

func (fn *Mcfunction) BuildTree() *Mcfunction {
	indents := []int{0}

	for fn.Scanner.Scan() {
		parent_indent := indents[len(indents)-1]

		line := fn.Scanner.Text()
		line_indent := internal.GetIndentOf(line) - parent_indent
		line = strings.TrimSpace(line)

		if line == "" {
			goto next_iteration
		}

		if line_indent > 0 {
			last := fn.current.Nested[len(fn.current.Nested)-1]
			if strings.HasSuffix(strings.TrimRight(last.Contents, " "), "\\") {
				line = internal.GetIndentString(line_indent) + line
				goto next_iteration
			}
			indents = append(indents, parent_indent+line_indent)
			fn.current = last
			goto next_iteration
		}

		if line_indent < 0 {
			indents = indents[:len(indents)-1]
			fn.current = fn.current.Parent
			goto next_iteration
		}

	next_iteration:
		fn.current.Append(newLine(fn.current, line))
	}

	return fn
}

func (fn *Mcfunction) Parse(templates map[string]*InlineTemplate) ([]string, error) {
	out := make([]string, 0, len(fn.root.Nested))

	err := fn.parse(templates, &out, fn.root, 0)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (fn *Mcfunction) parse(
	templates map[string]*InlineTemplate,
	out *[]string,
	root *Line,
	indent int,
) error {
	for i, line := range root.Nested {
		if !strings.HasPrefix(line.Contents, "#!/") || line.Contents == "#!/" {
			*out = append(*out, internal.GetIndentString(indent)+line.Contents)
			if err := fn.parse(templates, out, line, indent+4); err != nil {
				return err
			}
			continue
		}

		contents := line.Contents[3:]
		fields := internal.Fields(contents)
		name := fields[0]

		template, ok := templates[name]
		if !ok {
			return &liberrors.DetailedError{
				Label:   liberrors.ERR_SYNTAX,
				Context: fn.makeContext(line, i),
				Details: fmt.Sprintf("undefined inline template %q", name),
			}
		}

		var args []string
		if template.IsArgPassthrough() {
			args = []string{strings.TrimSpace(contents[len(name):])}
		} else {
			args = fields[1:]
			if len(args) != len(template.RequiredArgs) {
				required_args := make([]string, len(template.RequiredArgs))
				for i, required_arg := range template.RequiredArgs {
					required_args[i] = fmt.Sprintf("<%s>", required_arg)
				}
				return &liberrors.DetailedError{
					Label:   liberrors.ERR_VALIDATE,
					Context: fn.makeContext(line, i),
					Details: fmt.Sprintf(
						"template %q requires %d arguments (%s), but got %d (%s)",
						name,
						len(template.RequiredArgs),
						strings.Join(required_args, " "),
						len(args),
						strings.Join(args, " "),
					),
				}
			}
		}

		var stdout strings.Builder
		var stdin bytes.Buffer
		for _, line := range line.Nested {
			line.Write(&stdin, 4)
		}

		cli.LogDebug(1, "$ %s %s", name, strings.Join(args, " "))
		err := template.Call(&stdout, &stdin, args)
		if err != nil {
			return &liberrors.DetailedError{
				Label:   liberrors.ERR_EXECUTE,
				Context: fn.makeContext(line, i),
				Details: err.Error(),
			}
		}
		// TODO: this could probably be done better by making stdout
		// write to out line-by-line automatically.
		for line := range strings.SplitSeq(strings.TrimSuffix(stdout.String(), "\n"), "\n") {
			*out = append(*out, internal.GetIndentString(indent)+line)
		}
	}

	return nil
}

func (fn *Mcfunction) GenerateFiles(lines []string) (map[string][]string, error) {
	// Reset
	fn.Scanner = bufio.NewScanner(strings.NewReader(strings.Join(lines, "\n")))
	fn.current = fn.root
	fn.root.Nested = []*Line{}
	fn.BuildTree()

	tree := map[string][]string{}
	err := fn.generate(tree, internal.PathToResource(fn.Path), fn.root)
	if err != nil {
		return nil, err
	}

	return tree, nil
}

func (fn *Mcfunction) generate(tree map[string][]string, path string, line *Line) error {
	tree[path] = []string{}

	for i, nested_line := range line.Nested {
		nested_line.Format(path)
		tree[path] = append(tree[path], nested_line.Contents)

		if len(nested_line.Nested) != 0 {
			resource := nested_line.ExtractResource()
			if resource == "" {
				cli.LogDebug(1, "Context:\n%s", line)
				return &liberrors.DetailedError{
					Label: liberrors.ERR_SYNTAX,
					// TODO: this context isn't much helpful
					Context: fn.makeContext(nested_line, i),
					Details: "can't nest code without being subsequent to a function call",
				}
			}

			err := fn.generate(tree, resource, nested_line)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (fn *Mcfunction) ParseAndSave(templates map[string]*InlineTemplate) error {
	lines, err := fn.Parse(templates)
	if err != nil {
		return err
	}

	tree, err := fn.GenerateFiles(lines)
	if err != nil {
		return err
	}

	for resource, lines := range tree {
		if err := addFunction("function", resource, lines); err != nil {
			return err
		}
	}

	if cli.UsesPluralFolderNames {
		for resource, lines := range tree {
			if err := addFunction("functions", resource, lines); err != nil {
				return err
			}
		}
	}

	return nil
}

func addFunction(folder, resource string, lines []string) error {
	path := internal.ResourceToPath(folder, resource) + ".mcfunction"
	if err := Add(filepath.Join("data", path), lines); err != nil {
		return &liberrors.DetailedError{
			Label:   liberrors.ERR_INTERNAL,
			Context: liberrors.DirContext{Path: path},
			Details: err.Error(),
		}
	}
	return nil
}

func (fn *Mcfunction) makeContext(line *Line, i int) liberrors.FileContext {
	buffer, j := "", i
	if j > 0 {
		j--
		buffer = fn.root.Nested[j].Contents + "\n...\n"
	}
	return liberrors.FileContext{
		Trace: []liberrors.TraceItem{
			{
				Name: fn.Path,
				Row:  i,
				Col:  3,
			},
		},
		Buffer: liberrors.Buffer{
			FirstLine:   uint(j),
			Buffer:      buffer,
			Highlighted: line.Contents,
		},
	}
}

// ————————————————————————————————

type Line struct {
	Parent   *Line
	Contents string
	Nested   []*Line
}

func newLine(parent *Line, contents string) *Line {
	return &Line{Parent: parent, Contents: contents, Nested: nil}
}

func (line *Line) Append(nested *Line) {
	line.Nested = append(line.Nested, nested)
}

func (line *Line) Write(writer io.Writer, indent int) {
	internal.WriteIndentString(writer, indent)
	writer.Write([]byte(line.Contents + "\n"))
	for _, line := range line.Nested {
		line.Write(writer, indent+4)
	}
}

func (line *Line) String() string {
	var builder strings.Builder
	line.Write(&builder, 0)
	return builder.String()
}

func (line *Line) ExtractResource() string {
	fields := strings.Fields(line.Contents)
	var previous *string
	for i := len(fields) - 1; i >= 0; i-- {
		if fields[i] == "function" {
			if previous == nil {
				return ""
			}
			return *previous
		}
		previous = &fields[i]
	}
	return ""
}

func (line *Line) Format(location string) *Line {
	line.Contents = strings.ReplaceAll(line.Contents, "./", location+"/")
	return line
}

package mcfunc

import (
	"fmt"
	"path/filepath"
	"strings"

	liberrors "github.com/bbfh-dev/lib-errors"
	"github.com/bbfh-dev/vintage/devkit/internal"
	"github.com/bbfh-dev/vintage/devkit/internal/code"
	"github.com/bbfh-dev/vintage/devkit/internal/templates"
)

type Processor struct {
	Function *Function
}

func NewProcessor(fn *Function) *Processor {
	return &Processor{
		Function: fn,
	}
}

func (proc *Processor) Build() error {
	buffer := templates.NewBuffer()

	if err := proc.ExecInlineTemplates(buffer, 0, true); err != nil {
		return err
	}

	if err := proc.Inline(buffer); err != nil {
		return err
	}

	return nil
}

func (proc *Processor) ExecInlineTemplates(
	output *templates.Buffer,
	root_indent int,
	is_root bool,
) error {
	for proc.Function.Scanner.Scan() {
		raw_line := proc.Function.Scanner.Text()
		line_indent := code.GetIndentOf(raw_line) - root_indent
		clean_line := strings.TrimSpace(raw_line)
		aligned_line := code.GetIndentString(line_indent) + clean_line

		if clean_line == "" {
			output.Writeln(aligned_line)
			continue
		}

		if line_indent <= 0 && !is_root {
			proc.Function.Scanner.Unscan()
			return nil
		}

		if !templates.IsInlineCall(clean_line) {
			output.Writeln(aligned_line)
			continue
		}

		contents := clean_line[len(templates.INLINE_CALL_PREFIX):]
		if len(contents) == 0 {
			return proc.errEmptyTemplateCall(clean_line)
		}
		in_args := code.ExtractArgsFrom(contents)
		name := in_args[0]

		template, ok := proc.Function.Templates[name]
		if !ok {
			return proc.errUndefinedTemplate(clean_line, name)
		}

		var call_args []string
		if template.IsArgPassthrough() {
			call_args = []string{strings.TrimSpace(contents[len(name):])}
		} else {
			call_args = in_args[1:]
			if len(call_args) != len(template.RequiredArgs) {
				return proc.errMismatchArgs(clean_line, name, template, call_args)
			}
		}

		buffer_indent := output.Indent
		output.SetIndent(line_indent)

		call_buffer := templates.NewBuffer()
		err := proc.ExecInlineTemplates(call_buffer, line_indent, false)
		if err != nil {
			return err
		}

		err = template.Call(output, call_buffer, call_args)
		if err != nil {
			return err
		}
		output.Indent = buffer_indent
	}

	return nil
}

func (proc *Processor) Inline(input *templates.Buffer) error {
	current_path := proc.Function.Path
	breadcrumbs := []string{current_path}
	current_resource := internal.PathToResource(current_path)
	current_indent := 0

	for i, raw_line := range input.Lines {
	go_again:
		line_indent := code.GetIndentOf(raw_line) - current_indent
		formatted_line := formatLine(raw_line, current_resource)
		clean_line := strings.TrimSpace(formatted_line)

		if line_indent == 0 || clean_line == "" {
			AddLine(breadcrumbs[len(breadcrumbs)-1], clean_line)
			continue
		}

		if line_indent < 0 {
			current_indent -= 4
			breadcrumbs = breadcrumbs[:len(breadcrumbs)-1]
			goto go_again
		}

		if line_indent > 0 {
			if i == 0 {
				return &liberrors.DetailedError{
					Label:   liberrors.ERR_SYNTAX,
					Context: proc.makeBufferErrorContext(input, i),
					Details: "first line cannot be indented",
				}
			}

			previous_line := input.Lines[i-1]
			if strings.HasSuffix(strings.TrimRight(previous_line, " "), "\\") {
				AddLine(breadcrumbs[len(breadcrumbs)-1], formatted_line)
				continue
			}

			resource := code.ExtractResourceFrom(formatLine(previous_line, current_resource))
			if resource == "" {
				return &liberrors.DetailedError{
					Label:   liberrors.ERR_SYNTAX,
					Context: proc.makeBufferErrorContext(input, i-1),
					Details: "indented block must be subsequent to a function call",
				}
			}

			path := internal.ResourceToPath("function", resource)
			if path == "" {
				return &liberrors.DetailedError{
					Label:   liberrors.ERR_SYNTAX,
					Context: proc.makeBufferErrorContext(input, i-1),
					Details: fmt.Sprintf(
						"invalid resource %q in the function call",
						resource,
					),
				}
			}

			breadcrumbs = append(breadcrumbs, filepath.Join("data", path+".mcfunction"))
			current_indent += line_indent
			AddLine(breadcrumbs[len(breadcrumbs)-1], clean_line)
			continue
		}
	}

	return nil
}

func formatLine(line, resource string) string {
	return strings.ReplaceAll(line, "./", resource+"/")
}

func (proc *Processor) makeBufferErrorContext(buffer *templates.Buffer, i int) liberrors.Context {
	return liberrors.FileContext{
		Trace: []liberrors.TraceItem{
			{
				Name: proc.Function.Path + "*",
				Col:  -1,
				Row:  i,
			},
		},
		Buffer: liberrors.Buffer{
			FirstLine:   0,
			Buffer:      strings.Join(buffer.Lines[:i], "\n") + "\n",
			Highlighted: buffer.Lines[i],
		},
	}
}

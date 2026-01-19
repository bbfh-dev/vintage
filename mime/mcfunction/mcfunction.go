package mcfunction

import (
	"bufio"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bbfh-dev/mime/mime/errors"
)

type McFunction struct {
	Path    string
	Scanner *bufio.Scanner
	Locals  map[string][]string
	Indent  int
}

func New(path string, scanner *bufio.Scanner, indent int) *McFunction {
	return &McFunction{
		Path:    path,
		Scanner: scanner,
		Locals:  map[string][]string{},
		Indent:  indent,
	}
}

func (fn *McFunction) Parse() error {
	var previous_line *string
	breadcrumbs := []string{""}
	indents := []int{0}
	fn.Locals[""] = []string{}

	for fn.Scanner.Scan() {
		filename := breadcrumbs[len(breadcrumbs)-1]
		file_indent := indents[len(indents)-1]
		line := fn.Scanner.Text()
		line_indent := getIndentOf(line) - file_indent
		line = strings.TrimSpace(line)

		_, resource := FilepathToResource(fn.Path)
		line = strings.ReplaceAll(line, "function ./", "function "+resource+"/")

		if line_indent == 0 {
			fn.Locals[filename] = append(fn.Locals[filename], line)
			goto next_iteration
		}

		if line_indent < 0 {
			breadcrumbs = breadcrumbs[:len(breadcrumbs)-1]
			indents = indents[:len(indents)-1]
			filename := breadcrumbs[len(breadcrumbs)-1]
			fn.Locals[filename] = append(fn.Locals[filename], line)
			goto next_iteration
		}

		if previous_line == nil {
			return errors.NewError(errors.ERR_MCFUNC, fn.Path, fmt.Sprintf(
				"%q unexpected indentation at the beginning of the block",
				line,
			))
		}

		if strings.HasSuffix(*previous_line, "\\") {
			fn.Locals[filename] = append(fn.Locals[filename], line)
			goto next_iteration
		}

		resource = extractFunctionCall(*previous_line)
		if resource == "" {
			return errors.NewError(errors.ERR_MCFUNC, fn.Path, fmt.Sprintf(
				"%q expected a function call in the previous line %q",
				line,
				*previous_line,
			))
		}
		filename = filepath.Join("data", ResourceToFilepath("function", resource))
		fn.Locals[filename] = append(fn.Locals[filename], line)
		breadcrumbs = append(breadcrumbs, filename)
		indents = append(indents, file_indent+line_indent)

	next_iteration:
		previous_line = &line
	}

	for name, lines := range fn.Locals {
		if name == "" {
			name = fn.Path
		}
		if err := Add(name, lines); err != nil {
			return err
		}
	}

	return nil
}

func getIndentOf(line string) (indent int) {
	for _, char := range line {
		switch char {
		case ' ':
			indent++
		case '\t':
			indent += 4
		default:
			return
		}
	}
	return
}

func FilepathToResource(path string) (folder, resource string) {
	fields := strings.Split(path, "/")

	// Convert into data pack local space
	if index := slices.Index(fields, "data"); index != -1 {
		fields = fields[index+1:]
	}

	switch len(fields) {
	case 0, 1, 2:
		panic(
			fmt.Sprintf("Invalid FilepathToResource(%q). Not enough directories to convert", path),
		)
	default:
		last := len(fields) - 1
		fields[last] = strings.TrimSuffix(
			fields[last],
			filepath.Ext(fields[last]),
		)
		return fields[1], fields[0] + ":" + strings.Join(fields[2:], "/")
	}
}

func ResourceToFilepath(folder_name, resource string) string {
	parts := strings.SplitN(resource, ":", 2)
	if len(parts) == 1 {
		panic(fmt.Sprintf("Invalid ResourceToFilepath(%q, %q)", folder_name, resource))
	}

	return filepath.Join(parts[0], folder_name, parts[1])
}

func extractFunctionCall(line string) string {
	fields := strings.Fields(line)
	for i := len(fields) - 1; i >= 0; i-- {
		if fields[i] == "function" {
			if i+1 >= len(fields) {
				return ""
			}
			return fields[i+1]
		}
	}
	return ""
}

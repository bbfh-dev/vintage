package mcfunc

import (
	"bufio"

	liberrors "github.com/bbfh-dev/lib-errors"
	"github.com/bbfh-dev/vintage/devkit/internal/drive"
	"github.com/bbfh-dev/vintage/devkit/internal/templates"
)

var FunctionPool = drive.NewPool[Function](5000, 5000)

type Function struct {
	Path      string
	Scanner   *templates.BufferedScanner
	Templates map[string]*templates.Inline
}

func New(path string, scanner *bufio.Scanner, registry map[string]*templates.Inline) *Function {
	return FunctionPool.Acquire(func(fn *Function) {
		fn.Path = path
		fn.Scanner = templates.NewBufferedScanner(scanner)
		fn.Templates = registry
	})
}

func (fn *Function) MakeErrorContext(line_index uint, line string) liberrors.Context {
	return liberrors.FileContext{
		Trace: []liberrors.TraceItem{
			{
				Name: fn.Path,
				Col:  -1,
				Row:  int(line_index),
			},
		},
		Buffer: liberrors.Buffer{
			FirstLine:   line_index,
			Buffer:      "",
			Highlighted: line,
		},
	}
}

package templates

import (
	"bufio"

	"github.com/bbfh-dev/vintage/devkit/internal/code"
)

type Writer interface {
	Writeln(line string)
}

type Scanner interface {
	Scan() bool
	Text() string
}

// ————————————————————————————————

type Buffer struct {
	Lines   []string
	pointer int
	Indent  string
}

func NewBuffer() *Buffer {
	return &Buffer{
		Lines:   nil,
		pointer: 0,
	}
}

func (buffer *Buffer) SetIndent(indent int) {
	buffer.Indent = code.GetIndentString(indent)
}

func (buffer *Buffer) Scan() bool {
	if buffer.pointer >= len(buffer.Lines) {
		return false
	}
	buffer.pointer++
	return true
}

func (buffer *Buffer) Text() string {
	return buffer.Lines[buffer.pointer-1]
}

func (buffer *Buffer) Writeln(line string) {
	buffer.Lines = append(buffer.Lines, buffer.Indent+line)
}

// ————————————————————————————————

type BufferedScanner struct {
	Inner      *bufio.Scanner
	IsBuffered bool
	LineNumber uint
}

func NewBufferedScanner(scanner *bufio.Scanner) *BufferedScanner {
	return &BufferedScanner{
		Inner:      scanner,
		IsBuffered: false,
		LineNumber: 0,
	}
}

func (scanner *BufferedScanner) Unscan() {
	scanner.LineNumber--
	scanner.IsBuffered = true
}

func (scanner *BufferedScanner) Scan() bool {
	scanner.LineNumber++
	// Not a single statement because [*bufio.Scanner.Scan] has side-effects
	// and I want to make it clear that we do NOT want to cause them.
	if scanner.IsBuffered {
		scanner.IsBuffered = false
		return true
	}
	return scanner.Inner.Scan()
}

func (scanner *BufferedScanner) Text() string {
	return scanner.Inner.Text()
}

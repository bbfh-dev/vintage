package errors

import (
	"strings"

	libescapes "github.com/bbfh-dev/lib-ansi-escapes"
)

const (
	ERR_IO       = "I/O Error"
	ERR_META     = "Metadata Error"
	ERR_VALID    = "Validation Error"
	ERR_INTERNAL = "Internal Error"
	ERR_ITER     = "Iterator Error"
	ERR_EXEC     = "Execution Error"
	ERR_MCFUNC   = "MCFunction Error"
)

type MimeError struct {
	Name    string
	Context string
	Body    string
}

func NewError(name, context, body string) *MimeError {
	return &MimeError{
		Name:    name,
		Context: context,
		Body:    body,
	}
}

func (err *MimeError) Error() string {
	var builder strings.Builder

	builder.WriteString(libescapes.TextColorWhite)
	builder.WriteString(strings.Repeat("─", 32) + "\n")
	builder.WriteString(" [!] " + libescapes.TextColorBrightRed + err.Name + "\n")
	builder.WriteString(libescapes.TextColorWhite)
	builder.WriteString("     in " + err.Context + "\n")
	builder.WriteString(libescapes.TextColorBrightYellow)
	builder.WriteString("\n >>> ")
	builder.WriteString(libescapes.ColorReset)
	builder.WriteString(err.Body + "\n")

	return builder.String()
}

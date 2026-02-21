package mcfunc

import (
	"fmt"
	"strings"

	liberrors "github.com/bbfh-dev/lib-errors"
	"github.com/bbfh-dev/vintage/devkit/internal/templates"
)

func (proc *Processor) errEmptyTemplateCall(clean_line string) error {
	return &liberrors.DetailedError{
		Label: liberrors.ERR_SYNTAX,
		Context: proc.Function.MakeErrorContext(
			proc.Function.Scanner.LineNumber,
			clean_line,
		),
		Details: fmt.Sprintf(
			"%q expects to run an inline template, but it's not followed by anything.",
			templates.INLINE_CALL_PREFIX,
		),
	}
}

func (proc *Processor) errUndefinedTemplate(clean_line, name string) error {
	return &liberrors.DetailedError{
		Label: liberrors.ERR_SYNTAX,
		Context: proc.Function.MakeErrorContext(
			proc.Function.Scanner.LineNumber,
			clean_line,
		),
		Details: fmt.Sprintf("undefined inline template %q", name),
	}
}

func (proc *Processor) errMismatchArgs(
	clean_line, name string,
	template *templates.Inline,
	call_args []string,
) error {
	return &liberrors.DetailedError{
		Label: liberrors.ERR_VALIDATE,
		Context: proc.Function.MakeErrorContext(
			proc.Function.Scanner.LineNumber,
			clean_line,
		),
		Details: fmt.Sprintf(
			"template %q requires %d arguments (%s), but got %d (%s)",
			name,
			len(template.RequiredArgs),
			strings.Join(template.RequiredArgs, " "),
			len(call_args),
			strings.Join(call_args, " "),
		),
	}
}

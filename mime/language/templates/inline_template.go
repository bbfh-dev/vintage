package templates

import "io"

type Argument [2]string

func (arg Argument) Key() string {
	return arg[0]
}

func (arg Argument) Value() string {
	return arg[1]
}

type InlineTemplate struct {
	// NOTE: It CAN be nil in case arguments should be provided as a raw string.
	// an empty slice indicates that no arguments must be provided.
	Arguments []Argument
	Call      func(writer io.Writer, args map[string]string, nested []string) error
}

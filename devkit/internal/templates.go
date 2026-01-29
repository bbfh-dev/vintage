package internal

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

type Columns []string

type Rows []Columns

type Env struct {
	Iterators map[string]Columns
	Variables map[string]gjson.Result
}

func NewEnv() Env {
	return Env{
		Iterators: map[string]Columns{},
		Variables: map[string]gjson.Result{},
	}
}

func SimpleSubstitute(in string, env Env) (string, error) {
	var builder strings.Builder
	reader := bufio.NewReader(strings.NewReader(in))
	expect_bracket := false

	for {
		char, _, err := reader.ReadRune()
		if err != nil {
			return builder.String(), nil
		}
		if char == '%' {
			expect_bracket = true
			continue
		}
		if !expect_bracket {
			builder.WriteRune(char)
		}
		if expect_bracket {
			expect_bracket = false
			if char == '[' {
				str, err := reader.ReadString(']')
				if err != nil {
					return builder.String(), err
				}
				str = strings.TrimSuffix(str, "]")

				parts := strings.SplitN(str, ".", 2)
				key := parts[0]

				values, ok := env.Iterators[key]
				if !ok {
					value, ok := env.Variables[key]
					if ok {
						if value.Type != gjson.String {
							return "", fmt.Errorf(
								"simple subtitution only supports strings, got (%s) %q",
								value.Type.String(),
								value.String(),
							)
						}
						builder.WriteString(value.String())
						continue
					}
					return "", fmt.Errorf("unknown variable %q", key)
				}
				if len(parts) == 2 {
					num, err := strconv.Atoi(parts[1])
					if err != nil {
						return "", err
					}
					if num >= len(values) {
						return "", fmt.Errorf("index %d out of range of %#v", num, values)
					}
					builder.WriteString(values[num])
					continue
				}
				builder.WriteString(values[0])
			} else {
				builder.WriteRune('%')
				builder.WriteRune(char)
			}
		}
	}
}

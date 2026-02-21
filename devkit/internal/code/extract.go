package code

import (
	"bufio"
	"strings"
)

func ExtractVariablesFrom(in string) []string {
	out := []string{}
	reader := bufio.NewReader(strings.NewReader(in))
	expect_bracket := false

	for {
		char, _, err := reader.ReadRune()
		if err != nil {
			return out
		}

		switch {

		case char == '%':
			expect_bracket = true

		case !expect_bracket:
			// ignore

		case expect_bracket && char != '[':
			expect_bracket = false

		default:
			identifier, err := reader.ReadString(']')
			if err != nil {
				return out
			}
			out = append(out, strings.TrimSuffix(identifier, "]"))
		}
	}
}

// Similar to strings.Fields(), except it recognizes quoted elements.
func ExtractArgsFrom(in string) []string {
	reader := bufio.NewReader(strings.NewReader(in))
	var out []string
	var builder strings.Builder

	for {
		char, _, err := reader.ReadRune()
		if err != nil {
			goto exit
		}

		switch char {

		case ' ':
			out = append(out, builder.String())
			builder.Reset()

		case '"', '\'', '`':
			str, err := reader.ReadString(byte(char))
			if err != nil {
				goto exit
			}
			if char == '`' {
				builder.WriteString(strings.TrimSuffix(str, string(char)))
			} else {
				builder.WriteString(string(char) + str)
			}

		default:
			builder.WriteRune(char)
		}
	}

exit:
	if builder.Len() != 0 {
		out = append(out, builder.String())
	}
	return out
}

func ExtractResourceFrom(line string) string {
	fields := strings.Fields(line)
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

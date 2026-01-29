package internal

import (
	"bufio"
	"io"
	"strings"
)

func GetIndentOf(line string) (indent int) {
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

func WriteIndentString(writer io.Writer, indent int) {
	if indent%4 == 0 {
		char := []byte{'\t'}
		for range indent / 4 {
			writer.Write(char)
		}
		return
	}

	char := []byte{' '}
	for range indent {
		writer.Write(char)
	}
}

func ExtractIteratorsFrom(in string) []string {
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

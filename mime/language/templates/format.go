package templates

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func InsertBody(writer io.Writer, body string, nested []string, env map[string]string) error {
	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "%[...]" {
			line, err := SimpleSubstitute(scanner.Text(), env)
			if err != nil {
				return err
			}
			writer.Write([]byte(line))
			continue
		}

		if len(nested) == 0 {
			continue
		}

		writer.Write([]byte(nested[0]))

		indent := GetIndentOf(scanner.Text())
		for _, nested_line := range nested[1:] {
			writeIndentString(writer, indent)
			writer.Write([]byte(nested_line))
		}
	}

	return nil
}

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

func writeIndentString(writer io.Writer, indent int) {
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

func SimpleSubstitute(in string, env map[string]string) (string, error) {
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

				if value, ok := env[key]; ok {
					builder.WriteString(value)
					continue
				}
				return "", fmt.Errorf("unknown variable %q", key)
			} else {
				builder.WriteRune('%')
				builder.WriteRune(char)
			}
		}
	}
}

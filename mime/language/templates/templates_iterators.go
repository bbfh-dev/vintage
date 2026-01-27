package templates

import (
	"bufio"
	"strings"
)

func extractIteratorsFrom(in string) []string {
	out := []string{}
	reader := bufio.NewReader(strings.NewReader(in))
	expect_bracket := false

	for {
		char, _, err := reader.ReadRune()
		if err != nil {
			return out
		}
		if char == '%' {
			expect_bracket = true
			continue
		}
		if !expect_bracket {
			continue
		}
		if expect_bracket && char != '[' {
			expect_bracket = false
			continue
		}

		identifier, err := reader.ReadString(']')
		if err != nil {
			return out
		}
		out = append(out, strings.TrimSuffix(identifier, "]"))
	}
}

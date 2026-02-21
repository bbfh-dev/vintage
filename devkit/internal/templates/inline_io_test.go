package templates_test

import (
	"bufio"
	"strings"
	"testing"

	"github.com/bbfh-dev/vintage/devkit/internal/templates"
	"gotest.tools/assert"
)

func TestBufferedScanner(t *testing.T) {
	input := "line 1 (repeat)\nline 2\nline 3\nline 4 (repeat)\n"
	expect := "line 1 (repeat)\nline 1 (repeat)\nline 2\nline 3\nline 4 (repeat)\nline 4 (repeat)\n"
	scanner := templates.NewBufferedScanner(
		bufio.NewScanner(
			strings.NewReader(input)))
	i := 0
	repeat_read := false
	var result strings.Builder
	for scanner.Scan() {
		t.Logf("Reading (i=%d) %q", i, scanner.Text())
		result.WriteString(scanner.Text() + "\n")
		if !repeat_read && strings.HasSuffix(scanner.Text(), "(repeat)") {
			scanner.Unscan()
			repeat_read = true
			goto next_iteration
		}

		repeat_read = false
	next_iteration:
		i++
	}

	assert.DeepEqual(t, result.String(), expect)
}

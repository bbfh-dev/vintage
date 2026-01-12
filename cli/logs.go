package cli

import (
	"fmt"

	libescapes "github.com/bbfh-dev/lib-ansi-escapes"
)

func LogDebug(is_nested bool, format string, args ...any) {
	// Uses a different format
	fmt.Println(
		libescapes.TextColorWhite +
			arrow(is_nested) +
			fmt.Sprintf(format, args...) +
			libescapes.ColorReset,
	)
}

func LogInfo(is_nested bool, format string, args ...any) {
	log(
		is_nested,
		"",
		libescapes.TextColorBrightBlue,
		fmt.Sprintf(format, args...),
	)
}

func LogDone(is_nested bool, format string, args ...any) {
	log(
		is_nested,
		"DONE: ",
		libescapes.TextColorBrightGreen,
		fmt.Sprintf(format, args...),
	)
}

func LogWarn(is_nested bool, format string, args ...any) {
	log(
		is_nested,
		"WARN: ",
		libescapes.TextColorBrightYellow,
		fmt.Sprintf(format, args...),
	)
}

func LogError(is_nested bool, format string, args ...any) {
	log(
		is_nested,
		"ERROR: ",
		libescapes.TextColorBrightRed,
		fmt.Sprintf(format, args...),
	)
}

func log(is_nested bool, prefix, color, body string) {
	fmt.Println(color + arrow(is_nested) + prefix + libescapes.ColorReset + body)
}

func arrow(is_nested bool) string {
	if is_nested {
		return " -> "
	}
	return "==> "
}

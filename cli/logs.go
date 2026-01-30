package cli

import (
	"fmt"
	"strings"
	"sync"

	libescapes "github.com/bbfh-dev/lib-ansi-escapes"
)

var mutex sync.Mutex

func LogDebug(nesting uint, format string, args ...any) {
	mutex.Lock()
	defer mutex.Unlock()
	if Main.Options.Debug {
		// Uses a different format
		fmt.Println(
			libescapes.TextColorWhite +
				arrow(nesting) +
				fmt.Sprintf(format, args...) +
				libescapes.ColorReset,
		)
	}
}

func LogInfo(nesting uint, format string, args ...any) {
	log(
		nesting,
		"",
		libescapes.TextColorBrightBlue,
		fmt.Sprintf(format, args...),
	)
}

func LogDone(nesting uint, format string, args ...any) {
	log(
		nesting,
		"DONE: ",
		libescapes.TextColorBrightGreen,
		fmt.Sprintf(format, args...),
	)
}

func LogWarn(nesting uint, format string, args ...any) {
	log(
		nesting,
		"WARN: ",
		libescapes.TextColorBrightYellow,
		fmt.Sprintf(format, args...),
	)
}

func LogError(nesting uint, format string, args ...any) {
	log(
		nesting,
		"ERROR: ",
		libescapes.TextColorBrightRed,
		fmt.Sprintf(format, args...),
	)
}

func log(nesting uint, prefix, color, body string) {
	mutex.Lock()
	defer mutex.Unlock()
	fmt.Println(color + arrow(nesting) + prefix + libescapes.ColorReset + body)
}

func arrow(nesting uint) string {
	switch nesting {
	case 0:
		return "==> "
	case 1:
		return " -> "
	}

	return strings.Repeat("  ", int(nesting)) + "-> "
}

func LogCached(nesting uint, format string, args ...any) {
	log(
		nesting,
		"CACHED: ",
		libescapes.TextColorBrightMagenta,
		fmt.Sprintf(format, args...),
	)
}

func ColorWord(word, color string) string {
	return color + word + libescapes.ColorReset
}

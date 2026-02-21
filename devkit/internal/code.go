package internal

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

func PathToResource(path string) string {
	fields := strings.Split(path, "/")

	// Convert into pack local space
	if index := slices.Index(fields, "data"); index != -1 {
		fields = fields[index+1:]
	} else if index := slices.Index(fields, "assets"); index != -1 {
		fields = fields[index+1:]
	}

	switch len(fields) {
	case 0, 1, 2:
		panic(fmt.Sprintf(
			"Invalid PathToResource(%q). Not enough directories to convert",
			path,
		))
	default:
		last := len(fields) - 1
		fields[last] = strings.TrimSuffix(
			fields[last],
			filepath.Ext(fields[last]),
		)
		return fields[0] + ":" + strings.Join(fields[2:], "/")
	}
}

func ResourceToPath(folder_name, resource string) string {
	parts := strings.SplitN(resource, ":", 2)
	if len(parts) == 1 {
		return ""
	}
	return filepath.Join(parts[0], folder_name, parts[1])
}

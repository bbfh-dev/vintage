package mcfunc

import (
	"strings"
	"sync"
)

var UsedNamespaces = map[string]byte{}
var Registry = map[string][]string{}

var mutex = sync.Mutex{}

func AddLine(path string, line string) error {
	mutex.Lock()
	defer mutex.Unlock()

	lines, ok := Registry[path]
	if !ok {
		Registry[path] = []string{}
	}

	collectNamespaces(line)
	Registry[path] = append(lines, line)
	return nil
}

func collectNamespaces(line string) {
	for field := range strings.FieldsSeq(line) {
		before, after, ok := strings.Cut(field, ":")
		if ok && after != "" && strings.ToLower(before) == before {
			UsedNamespaces[before] = 1
		}
	}
}

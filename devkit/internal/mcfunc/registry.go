package mcfunc

import (
	"strings"
	"sync"

	liblog "github.com/bbfh-dev/lib-log"
	"github.com/bbfh-dev/vintage/devkit/internal"
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
		liblog.Debug(3, "Created %q", internal.PathToResource(path))
	}

	collectNamespaces(line)
	Registry[path] = append(lines, line)
	return nil
}

func collectNamespaces(line string) {
	for field := range strings.FieldsSeq(line) {
		before, after, ok := strings.Cut(field, ":")
		if ok && after != "" && strings.ToLower(before) == before {
			UsedNamespaces[strings.TrimPrefix(before, "#")] = 1
		}
	}
}

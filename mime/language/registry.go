package language

import (
	"fmt"
	"sync"
)

var Registry = map[string][]string{}

var mutex = sync.Mutex{}

func Add(path string, lines []string) error {
	mutex.Lock()
	defer mutex.Unlock()

	if _, ok := Registry[path]; ok {
		return fmt.Errorf("function %q is already defined", path)
	}

	Registry[path] = lines
	return nil
}

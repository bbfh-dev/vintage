package templates

import (
	"os"
	"path/filepath"
	"strings"

	liberrors "github.com/bbfh-dev/lib-errors"
	"github.com/bbfh-dev/vintage/devkit/internal/drive"
)

type Custom struct {
	Root    string
	Program string
}

func NewCustom(root string, manifest *drive.JsonFile) (*Custom, error) {
	template := &Custom{
		Root:    root,
		Program: "",
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, liberrors.NewIO(err, root)
	}

	for entry := range drive.IterateFilesOnly(entries) {
		if strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())) != "call" {
			continue
		}

		template.Program = entry.Name()
		break
	}

	if template.Program == "" {
		return nil, &liberrors.DetailedError{
			Label:   liberrors.ERR_SYNTAX,
			Context: liberrors.DirContext{Path: root},
			Details: "Template has no program called 'call*'. Refer to documentation for help",
		}
	}

	return template, nil
}

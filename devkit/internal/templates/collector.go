package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	liberrors "github.com/bbfh-dev/lib-errors"
	"github.com/bbfh-dev/vintage/devkit/internal/drive"
	"github.com/tidwall/gjson"
)

type CollectorTemplate struct {
	Root      string
	Patterns  []string
	Collector string
}

func NewCollectorTemplate(root string, manifest *drive.JsonFile) (*CollectorTemplate, error) {
	template := &CollectorTemplate{
		Root:     root,
		Patterns: []string{},
	}

	if err := manifest.ExpectField("patterns", gjson.JSON); err != nil {
		return nil, &liberrors.DetailedError{
			Label:   liberrors.ERR_VALIDATE,
			Context: liberrors.DirContext{Path: root},
			Details: err.Error(),
		}
	}

	field_patterns := manifest.Get("patterns")
	for i, arg := range field_patterns.Array() {
		if arg.Type != gjson.String {
			return nil, newSyntaxError(
				root,
				fmt.Sprintf("field 'patterns[%d]' must be a string", i),
				arg,
			)
		}

		template.Patterns = append(template.Patterns, arg.String())
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, liberrors.NewIO(err, root)
	}

	for entry := range drive.IterateFilesOnly(entries) {
		if strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())) != "collect" {
			continue
		}

		template.Collector = entry.Name()
		break
	}

	if template.Collector == "" {
		return nil, &liberrors.DetailedError{
			Label:   liberrors.ERR_SYNTAX,
			Context: liberrors.DirContext{Path: root},
			Details: "Template has no program called 'collect*'. Refer to documentation for help",
		}
	}

	return template, nil
}

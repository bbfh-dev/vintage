package templates

import (
	"github.com/bbfh-dev/mime/mime/minecraft"
	"github.com/tidwall/gjson"
)

type Template struct {
	Dir       string
	Manifest  *minecraft.JsonFile
	Iterators map[string][]gjson.Result
}

func New(dir string, manifest *minecraft.JsonFile) *Template {
	return &Template{
		Dir:       dir,
		Manifest:  manifest,
		Iterators: map[string][]gjson.Result{},
	}
}

func (template *Template) Generate() error {
	return nil
}

// func (project *Project) generateFromTemplate(dir string, manifest *minecraft.JsonFile) error {
// 	cli.LogInfo(true, "Generating from %q", dir)
//
// 	iterators := map[string][]gjson.Result{}
//
// 	manifest_iterators := manifest.Get("iterators")
// 	if manifest_iterators.Exists() {
// 		if !manifest_iterators.IsObject() {
// 			return &liberrors.DetailedError{
// 				Label:   liberrors.ERR_VALIDATE,
// 				Context: liberrors.DirContext{Path: dir},
// 				Details: fmt.Sprintf(
// 					"field 'iterators' must be an object, but got %q instead",
// 					manifest_iterators.Type,
// 				),
// 			}
// 		}
//
// 		for _, key := range manifest_iterators.Get("@keys").Array() {
// 			iterators[key.String()] = manifest_iterators.Get(key.String()).Array()
// 		}
// 	}
//
// 	path := filepath.Join(dir, "definitions")
// 	entries, err := os.ReadDir(path)
// 	if err != nil {
// 		return liberrors.NewIO(err, path)
// 	}
//
// 	for entry := range internal.IterateFilesOnly(entries) {
// 		if strings.Contains(entry.Name(), "%[") {
// 			iterators := extractIteratorsFrom(entry.Name())
// 			continue
// 		}
//
// 		if err := project.generateTemplate(dir, entry.Name()); err != nil {
// 			return err
// 		}
// 	}
//
// 	return nil
// }

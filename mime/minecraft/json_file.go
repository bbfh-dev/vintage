package minecraft

import (
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
	"github.com/tidwall/sjson"
)

type JsonFile struct {
	Body []byte
}

func NewJsonFile(body []byte) *JsonFile {
	return &JsonFile{Body: body}
}

func (file *JsonFile) Get(path string) gjson.Result {
	return gjson.GetBytes(file.Body, path)
}

func (file *JsonFile) ExpectField(path string, kind gjson.Type) error {
	result := file.Get(path)
	if !result.Exists() {
		return fmt.Errorf(
			"missing field %q of type %q",
			path,
			kind.String(),
		)
	}
	if result.Type != kind {
		return fmt.Errorf(
			"field %q expected to be of type %q but got %q",
			path,
			kind.String(),
			result.Type.String(),
		)
	}

	return nil
}

func (file *JsonFile) Set(path string, value any) {
	var err error
	file.Body, err = sjson.SetBytes(file.Body, path, value)
	if err != nil {
		panic(
			fmt.Sprintf(
				"(Assertion fail) Failed setting inside of the json file: %s",
				err.Error(),
			),
		)
	}
}

var formattingOptions = &pretty.Options{
	Width:    80,
	Prefix:   "",
	Indent:   "\t",
	SortKeys: false,
}

func (file *JsonFile) Formatted() []byte {
	return pretty.PrettyOptions(file.Body, formattingOptions)
}

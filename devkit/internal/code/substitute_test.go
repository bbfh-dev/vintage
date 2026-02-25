package code_test

import (
	"testing"

	"github.com/bbfh-dev/vintage/devkit/internal/code"
	"github.com/bbfh-dev/vintage/devkit/internal/drive"
	"github.com/tidwall/gjson"
	"gotest.tools/assert"
)

func TestSubstituteString(t *testing.T) {
	env := code.NewEnv()
	env.Variables["test"] = gjson.Parse(`{"nested": {"within": 123}}`)
	result, err := code.SubstituteString("Hello %[test.nested.within]!", env)
	assert.NilError(t, err)
	assert.DeepEqual(t, result, "Hello 123!")

	env.Variables["test2"] = code.SimpleVariable("World")
	result, err = code.SubstituteString("Hello %[test2]!", env)
	assert.NilError(t, err)
	assert.DeepEqual(t, result, "Hello World!")
}

const SAMPLE_A = `{"test": "%[abc]", "value": [{"id": "%[abc.id]", "c": "Hello %[abc.zzz.c]!"}], "deleted": "%[unknown?]"}`

const SAMPLE_A_RESULT = `{"test": {"id": "example", "zzz": {"c": 123}}, "value": [{"id": "example", "c": "Hello 123!"}]}`

func TestSubstituteJsonFile(t *testing.T) {
	file := drive.NewJsonFile([]byte(SAMPLE_A))
	file_result := drive.NewJsonFile([]byte(SAMPLE_A_RESULT))

	env := code.NewEnv()
	env.Variables["abc"] = gjson.Parse(`{"id": "example", "zzz": {"c": 123}}`)

	err := code.SubstituteJsonFile(file, env)
	assert.NilError(t, err)
	assert.DeepEqual(t, file.Contents(), file_result.Contents())
}

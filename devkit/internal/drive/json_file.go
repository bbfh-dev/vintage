package drive

import (
	"fmt"
	"slices"
	"sync"

	liblog "github.com/bbfh-dev/lib-log"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
	"github.com/tidwall/sjson"
)

type JsonFilePool struct {
	mutex     sync.Mutex
	pool      []*JsonFile
	index     int
	chunkSize int
}

// Acquire returns a JsonFile with a copy of the given body.
// The pool grows automatically if more objects are needed.
func (pool *JsonFilePool) Acquire(body []byte) *JsonFile {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	// Grow the pool if needed
	if pool.index >= len(pool.pool) {
		newPool := make([]*JsonFile, len(pool.pool)+pool.chunkSize)
		copy(newPool, pool.pool)
		for i := len(pool.pool); i < len(newPool); i++ {
			newPool[i] = &JsonFile{}
		}
		pool.pool = newPool
	}

	file := pool.pool[pool.index]
	pool.index++

	file.Body = body
	return file
}

func NewJsonFilePool(initialSize, chunkSize int) *JsonFilePool {
	p := &JsonFilePool{
		pool:      make([]*JsonFile, initialSize),
		index:     0,
		chunkSize: chunkSize,
	}
	for i := range initialSize {
		p.pool[i] = &JsonFile{}
	}
	return p
}

var Pool = NewJsonFilePool(5000, 5000)

type JsonFile struct {
	Body []byte
}

func NewJsonFile(body []byte) *JsonFile {
	return Pool.Acquire(body)
}

func (file *JsonFile) Extension() string {
	return ".json"
}

func (file *JsonFile) Contents() []byte {
	return file.Formatted()
}

func (file *JsonFile) Clone() *JsonFile {
	return NewJsonFile(file.Body)
}

func (file *JsonFile) Get(path string) gjson.Result {
	return gjson.GetBytes(file.Body, path)
}

func (file *JsonFile) ExpectField(path string, kinds ...gjson.Type) error {
	result := file.Get(path)
	if !result.Exists() {
		return fmt.Errorf(
			"missing field %q of type %s",
			path,
			kinds,
		)
	}
	if slices.Contains(kinds, result.Type) {
		return nil
	}

	return fmt.Errorf(
		"field %q expected to be of type %s but got %q",
		path,
		kinds,
		result.Type.String(),
	)
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

func (file *JsonFile) Delete(path string) {
	var err error
	file.Body, err = sjson.DeleteBytes(file.Body, path)
	if err != nil {
		panic(
			fmt.Sprintf(
				"(Assertion fail) Failed deleting inside of the json file: %s",
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

// Merges top-level JSON keys.
// Overrides the value with [target] if merging isn't possible (string, number, etc.).
func (file *JsonFile) MergeWith(target *JsonFile) {
	for _, key := range target.Get("@keys").Array() {
		key := key.String()
		origin_value := file.Get(key)
		target_value := target.Get(key)

		if !origin_value.Exists() {
			file.Set(key, target_value.Value())
			continue
		}

		if origin_value.IsArray() {
			items := make([]any, 0, len(origin_value.Array())+len(target_value.Array()))
			for _, item := range origin_value.Array() {
				items = append(items, item.Value())
			}
			for _, item := range target_value.Array() {
				items = append(items, item.Value())
			}

			file.Set(key, items)
			continue
		}

		if origin_value.IsObject() {
			liblog.Debug(2, "WARN: Recursive merging is not implemented, overriding %q", key)
			continue
		}

		file.Set(key, target_value.Value())
	}
}

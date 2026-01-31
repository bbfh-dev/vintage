package internal

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"unicode"

	liberrors "github.com/bbfh-dev/lib-errors"
	"github.com/tidwall/gjson"
)

type Columns []string

type Rows []Columns

type Env struct {
	Iterators map[string]Columns
	Variables map[string]gjson.Result
}

func NewEnv() Env {
	return Env{
		Iterators: map[string]Columns{},
		Variables: map[string]gjson.Result{},
	}
}

func SimpleSubstitute(in string, env Env) (string, error) {
	var builder strings.Builder
	reader := bufio.NewReader(strings.NewReader(in))
	expect_bracket := false

	for {
		char, _, err := reader.ReadRune()
		if err != nil {
			return builder.String(), nil
		}
		if char == '%' {
			expect_bracket = true
			continue
		}
		if !expect_bracket {
			builder.WriteRune(char)
		}
		if expect_bracket {
			expect_bracket = false
			if char == '[' {
				str, err := reader.ReadString(']')
				if err != nil {
					return builder.String(), err
				}
				str = strings.TrimSuffix(str, "]")

				parts := strings.SplitN(str, ".", 2)
				key := parts[0]

				values, ok := env.Iterators[key]
				if !ok {
					value, ok := env.Variables[key]
					if ok {
						if value.Type != gjson.String {
							return "", fmt.Errorf(
								"simple subtitution only supports strings, got (%s) %q",
								value.Type.String(),
								value.String(),
							)
						}
						out := value.String()
						if len(parts) == 2 {
							switch parts[1] {
							case "to_file_name":
								var builder strings.Builder
								for _, char := range out {
									if unicode.IsLetter(char) || char == '_' || char == '+' {
										builder.WriteRune(char)
									}
									if unicode.IsSpace(char) {
										builder.WriteRune('_')
									}
								}
								out = builder.String()
							case "to_lower_case":
								out = strings.ToLower(out)
							case "to_upper_case":
								out = strings.ToUpper(out)
							case "length":
								out = fmt.Sprint(len(out))
							default:
								return "", fmt.Errorf(
									"unknown variable modifier %q from %q",
									parts[1],
									str,
								)
							}
						}
						builder.WriteString(out)
						continue
					}
					return "", fmt.Errorf("unknown variable %q", key)
				}
				if len(parts) == 2 {
					num, err := strconv.Atoi(parts[1])
					if err != nil {
						return "", err
					}
					if num >= len(values) {
						return "", fmt.Errorf("index %d out of range of %#v", num, values)
					}
					builder.WriteString(values[num])
					continue
				}
				builder.WriteString(values[0])
			} else {
				builder.WriteRune('%')
				builder.WriteRune(char)
			}
		}
	}
}

func SubstituteJsonFile(file *JsonFile, env Env) error {
	return SubstituteObject(file, env, "@this")
}

func SubstituteGenericFile(file *GenericFile, env Env) error {
	switch file.Ext {

	case ".json":
		json_file := NewJsonFile(file.Body)
		err := SubstituteJsonFile(json_file, env)
		if err != nil {
			return err
		}
		file.Body = json_file.Body

	case ".mcfunction", ".txt", ".md", ".py", ".js", ".ts", ".sh":
		body, err := SimpleSubstitute(string(file.Body), env)
		if err != nil {
			return err
		}
		file.Body = []byte(body)
	}

	return nil
}

func SubstituteObject(file *JsonFile, env Env, path string) error {
	for _, key := range file.Get(path + ".@keys").Array() {
		value := file.Get(join(path, key.String()))

		// Substitute the key
		new_key, err := SimpleSubstitute(key.String(), env)
		if err != nil {
			return err
		}
		if new_key != key.String() {
			file.Delete(join(path, key.String()))
			file.Set(join(path, new_key), value.Value())
		}

		switch value.Type {

		case gjson.Null, gjson.False, gjson.True, gjson.Number:
			// ignore

		case gjson.String:
			err := SubstituteString(file, env, join(path, new_key), value)
			if err != nil {
				return err
			}

		default:
			if value.IsArray() {
				err := SubstituteArray(file, env, join(path, new_key))
				if err != nil {
					return err
				}
				continue
			}

			err := SubstituteObject(file, env, join(path, new_key))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func SubstituteArray(file *JsonFile, env Env, path string) error {
	for i, value := range file.Get(path).Array() {
		switch value.Type {
		case gjson.Null, gjson.False, gjson.True, gjson.Number:
			// ignore

		case gjson.String:
			err := SubstituteString(file, env, join(path, fmt.Sprint(i)), value)
			if err != nil {
				return err
			}

		default:
			if value.IsArray() {
				err := SubstituteArray(file, env, join(path, fmt.Sprint(i)))
				if err != nil {
					return err
				}
				continue
			}

			err := SubstituteObject(file, env, join(path, fmt.Sprint(i)))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func SubstituteString(file *JsonFile, env Env, path string, value gjson.Result) error {
	variables := ExtractVariablesFrom(value.String())

	if !isSmartValue(value.String(), variables) {
		value, err := SimpleSubstitute(value.String(), env)
		if err != nil {
			return err
		}
		file.Set(path, value)
		return nil
	}

	parts := strings.Split(variables[0], ".")
	iterator, ok := env.Iterators[parts[0]]
	if ok {
		var err error
		num := 0
		if len(parts) > 1 {
			num, err = strconv.Atoi(parts[1])
			if err != nil {
				return err
			}
		}
		if num >= len(iterator) {
			return fmt.Errorf("index %d out of range of %#v", num, iterator)
		}
		file.Set(path, iterator[num])
		return nil
	}

	value, ok = env.Variables[variables[0]]
	if !ok {
		return fmt.Errorf("undefined variable %q", variables[0])
	}
	file.Set(path, value.Value())

	return nil
}

func join(path, item string) string {
	if path == "@this" {
		return item
	}
	return path + "." + item
}

func isSmartValue(value string, variables []string) bool {
	if len(variables) != 1 {
		return false
	}
	return len(value) == len("%[")+len(variables[0])+len("]")
}

func LoadTree(root string, dirs ...[2]string) (map[string]*GenericFile, error) {
	tree := map[string]*GenericFile{}
	var mutex sync.Mutex

	for _, entry := range dirs {
		dir := filepath.Join(root, entry[0])
		out_dir := entry[1]
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return err
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			path = strings.TrimPrefix(path, root+"/")
			path = filepath.Join(out_dir, path)

			mutex.Lock()
			tree[path] = NewGenericFile(filepath.Ext(path), data)
			mutex.Unlock()

			return nil
		})
		if err != nil {
			return nil, liberrors.NewIO(err, dir)
		}
	}

	return tree, nil
}

package code

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/bbfh-dev/vintage/cli"
	"github.com/bbfh-dev/vintage/devkit/internal/drive"
	"github.com/tidwall/gjson"
)

func SubstituteString(in string, env Env) (string, error) {
	var builder strings.Builder
	reader := bufio.NewReader(strings.NewReader(in))
	expect_bracket := false

	for {
		char, _, err := reader.ReadRune()
		if err != nil {
			return builder.String(), nil
		}

		switch {

		case char == '%':
			expect_bracket = true

		case !expect_bracket:
			builder.WriteRune(char)

		case char != '[':
			expect_bracket = false
			builder.WriteRune('%')
			builder.WriteRune(char)

		default:
			expect_bracket = false
			str, err := reader.ReadString(']')
			if err != nil {
				return builder.String(), err
			}
			str = strings.TrimSuffix(str, "]")

			parts := strings.SplitN(str, ".", 2)
			key := parts[0]

			if values, ok := env.Iterators[key]; ok {
				index := 0
				if len(parts) == 2 {
					index, err = strconv.Atoi(parts[1])
					if err != nil {
						return "", err
					}
					if index >= len(values) {
						return "", fmt.Errorf("index %d out of range of %#v", index, values)
					}
				}
				builder.WriteString(values[index])
				continue
			}

			value, ok := env.Variables[key]
			if !ok {
				return "", fmt.Errorf("unknown variable %q", key)
			}
			if len(parts) == 2 {
				value = Query(value, parts[1])
			}
			if !IsStringifiable(value) {
				if cli.Build.Options.ForceStringify {
					out := value.String()
					out = strings.ReplaceAll(out, "\t", "")
					out = strings.ReplaceAll(out, " ", "")
					out = strings.ReplaceAll(out, "\r", "")
					out = strings.ReplaceAll(out, "\n", "")
					builder.WriteString(out)
					continue
				}

				return "", fmt.Errorf(
					"simple subtitution only supports primitive datatypes, got (%s) %q",
					TypeOf(value),
					value,
				)
			}
			builder.WriteString(value.String())
		}
	}
}

func SubstituteSmartString(file *drive.JsonFile, env Env, path string, value Variable) error {
	variables := ExtractVariablesFrom(value.String())

	if !isSmartValue(value.String(), variables) {
		value, err := SubstituteString(value.String(), env)
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
		if len(parts) == 2 {
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

	value, ok = env.Variables[parts[0]]
	if !ok {
		return fmt.Errorf("undefined variable %q", variables[0])
	}
	if len(parts) == 2 {
		value = Query(value, parts[1])
	}
	file.Set(path, value.Value())

	return nil
}

func isSmartValue(value string, variables []string) bool {
	if len(variables) != 1 {
		return false
	}
	return len(value) == len("%[")+len(variables[0])+len("]")
}

func SubstituteJsonFile(file *drive.JsonFile, env Env) error {
	return SubstituteObject(file, env, "")
}

func SubstituteObject(file *drive.JsonFile, env Env, path string) error {
	for _, key := range file.Get(join(path, "@keys")).Array() {
		value := file.Get(join(path, key.String()))

		// Substitute the key
		new_key, err := SubstituteString(key.String(), env)
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
			err := SubstituteSmartString(file, env, join(path, new_key), value)
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

func SubstituteArray(file *drive.JsonFile, env Env, path string) error {
	for i, value := range file.Get(path).Array() {
		switch value.Type {
		case gjson.Null, gjson.False, gjson.True, gjson.Number:
			// ignore

		case gjson.String:
			err := SubstituteSmartString(file, env, join(path, fmt.Sprint(i)), value)
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

func join(path, item string) string {
	if path == "" {
		return item
	}
	return path + "." + item
}

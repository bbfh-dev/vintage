package code

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/bbfh-dev/vintage/cli"
	"github.com/bbfh-dev/vintage/devkit/internal/drive"
	"github.com/tidwall/gjson"
)

func SubstituteString(in string, env Env) (string, error) {
	var builder strings.Builder
	var err error
	builder.Grow(len(in))
	i := 0

	for i < len(in) {
		if in[i] != '%' {
			builder.WriteByte(in[i])
			i++
			continue
		}
		i++ // skip '%'

		if i >= len(in) {
			builder.WriteByte('%')
			break
		}
		if in[i] != '[' {
			builder.WriteByte('%')
			builder.WriteByte(in[i])
			i++
			continue
		}
		i++ // skip '['

		// find closing ']'
		start := i
		for i < len(in) && in[i] != ']' {
			i++
		}
		if i >= len(in) {
			return "", fmt.Errorf("unclosed placeholder")
		}
		placeholder := in[start:i]
		i++ // skip ']'

		before, after, ok := strings.Cut(placeholder, ".")
		var key, suffix string
		if !ok {
			key = placeholder
		} else {
			key = before
			suffix = after
		}

		if values, ok := env.Iterators[key]; ok {
			index := 0
			if suffix != "" {
				index, err = strconv.Atoi(suffix)
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
		if suffix != "" {
			value = Query(value, suffix)
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

	return builder.String(), nil
}

var ErrRemoveKey = errors.New("internal.remove_variable")

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

	is_optional := strings.HasSuffix(variables[0], "?")
	variables[0] = strings.TrimSuffix(variables[0], "?")

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
		if is_optional {
			return ErrRemoveKey
		}
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
	return SubstituteObject(file, env, "@this")
}

func SubstituteObject(file *drive.JsonFile, env Env, path string) error {
	obj := file.Get(path)
	if !obj.IsObject() {
		return nil
	}

	type changedKey struct {
		old, new string
	}
	var err error
	var changed_keys []changedKey

	obj.ForEach(func(key, value gjson.Result) bool {
		key_str := key.String()
		full_path := join(path, key_str)

		switch value.Type {

		case gjson.String:
			err = SubstituteSmartString(file, env, full_path, value)
			if err == ErrRemoveKey {
				file.Delete(full_path)
				err = nil
				return true
			}

		case gjson.JSON:
			if value.IsArray() {
				err = SubstituteArray(file, env, full_path)
			} else {
				err = SubstituteObject(file, env, full_path)
			}

		default:
			// skip
		}

		if err != nil {
			return false
		}

		new_key, sub_err := SubstituteString(key_str, env)
		if sub_err != nil {
			err = sub_err
			return false
		}

		if new_key != key_str {
			changed_keys = append(changed_keys, changedKey{key_str, new_key})
		}
		return true
	})

	for _, change := range changed_keys {
		old_full_path := join(path, change.old)
		new_full_path := join(path, change.new)
		value := file.Get(old_full_path).Value()
		file.Delete(old_full_path)
		file.Set(new_full_path, value)
	}

	return err
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
	if path == "@this" {
		return item
	}
	return path + "." + item
}

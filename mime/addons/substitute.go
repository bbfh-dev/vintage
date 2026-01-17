package addons

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/bbfh-dev/mime/mime/minecraft"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
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
						builder.WriteString(value.String())
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

func SubstituteFile(file *minecraft.JsonFile, env Env) error {
	body, err := SubstituteObject(string(file.Body), "@this", env)
	file.Body = []byte(body)
	return err
}

func SubstituteObject(body string, path string, env Env) (string, error) {
	for _, key := range gjson.Get(body, path+".@keys").Array() {
		if strings.HasPrefix(key.String(), "$") {
			path := path + "." + key.String()
			path = strings.TrimPrefix(path, "@this.")
			str, err := sjson.Delete(body, path)
			if err != nil {
				return body, err
			}
			body = str
			continue
		}
		str, err := SubstituteItem(body, path+"."+key.String(), env)
		if err != nil {
			return body, err
		}
		body = str
	}

	return body, nil
}

func SubstituteArray(body string, path string, env Env) (string, error) {
	for i := range gjson.Get(body, path).Array() {
		str, err := SubstituteItem(body, fmt.Sprintf("%s.%d", path, i), env)
		if err != nil {
			return body, err
		}
		body = str
	}

	return body, nil
}

func SubstituteItem(body string, path string, env Env) (string, error) {
	path = strings.TrimPrefix(path, "@this.")
	value := gjson.Get(body, path)

	switch {

	case value.IsObject():
		return SubstituteObject(body, path, env)

	case value.IsArray():
		return SubstituteArray(body, path, env)

	case value.Type == gjson.String:
		str, err := SubstituteString(value.String(), env)
		if err != nil {
			return body, err
		}
		if num, err := strconv.Atoi(str); err == nil {
			return sjson.Set(body, path, num)
		}
		return sjson.Set(body, path, str)
	}

	return body, nil
}

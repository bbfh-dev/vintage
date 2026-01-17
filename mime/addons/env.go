package addons

import "github.com/tidwall/gjson"

type Env struct {
	Iterators map[string][]string
	Variables map[string]gjson.Result
}

func NewEnv() Env {
	return Env{
		Iterators: map[string][]string{},
		Variables: map[string]gjson.Result{},
	}
}

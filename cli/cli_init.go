package cli

import (
	libparsex "github.com/bbfh-dev/lib-parsex/v3"
)

var Init struct {
	Options struct {
		Name        string `alt:"n" desc:"Specify the project name that will be used for exporting"`
		Minecraft   string `alt:"m" desc:"Specify the target Minecraft version. Use '-' to indicate version ranges, e.g. '1.20-1.21'" default:"1.21.11"`
		Version     string `alt:"v" desc:"Specify the project version using semantic versioning" default:"0.1.0-alpha"`
		Description string `alt:"d" desc:"Specify the project description"`
	}
	Args struct{}
}

var InitProgram = libparsex.Program{
	Name:        "init",
	Description: "Initialize a new Mime project",
	Options:     &Init.Options,
	Args:        &Init.Args,
	Commands:    []*libparsex.Program{},
	EntryPoint: func(raw_args []string) error {
		return nil
	},
}

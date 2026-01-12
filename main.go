package main

import (
	"os"

	libparsex "github.com/bbfh-dev/lib-parsex/v3"
	"github.com/bbfh-dev/mime/cli"
)

const VERSION = "0.1.0-alpha.1"

var MainProgram = libparsex.Program{
	Name:        "mime",
	Version:     VERSION,
	Description: "Minecraft data & resource pack processor designed to not significantly modify the Minecraft syntax while providing useful code generation",
	Options:     &cli.Main.Options,
	Args:        &cli.Main.Args,
	Commands:    []*libparsex.Program{},
	EntryPoint: func(raw_args []string) error {
		if cli.Main.Args.WorkDir != nil {
			if err := os.Chdir(*cli.Main.Args.WorkDir); err != nil {
				return err
			}
		}

		return nil
	},
}

func main() {
	err := libparsex.Run(&MainProgram, os.Args[1:])
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

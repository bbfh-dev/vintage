package main

import (
	"os"

	liberrors "github.com/bbfh-dev/lib-errors"
	libparsex "github.com/bbfh-dev/lib-parsex/v3"
	"github.com/bbfh-dev/vintage/cli"
	"github.com/bbfh-dev/vintage/devkit"
)

var MainProgram = libparsex.Program{
	Name:        "vintage",
	Version:     libparsex.GetVersion(),
	Description: "Minecraft data-driven vanilla data & resource pack development kit with minimum boilerplate and abstraction.",
	Options:     &cli.Main.Options,
	Args:        &cli.Main.Args,
	Commands: []*libparsex.Program{
		&cli.InitProgram,
		&BuildProgram,
	},
	EntryPoint: func(rawArgs []string) error {
		return libparsex.PrintHelpErr
	},
}

var BuildProgram = libparsex.Program{
	Name:        "build",
	Description: "Build the data & resource packs",
	Options:     &cli.Build.Options,
	Args:        &cli.Build.Args,
	Commands:    []*libparsex.Program{},
	EntryPoint:  devkit.Build,
}

func main() {
	err := libparsex.Run(&MainProgram, os.Args[1:])
	if err != nil {
		switch err := err.(type) {
		case *liberrors.DetailedError:
			err.Print(os.Stderr)
		default:
			os.Stderr.WriteString(err.Error())
		}

		os.Stderr.WriteString("\n")
		os.Exit(1)
	}
}

package main

import (
	"os"
	"runtime/debug"

	liberrors "github.com/bbfh-dev/lib-errors"
	libparsex "github.com/bbfh-dev/lib-parsex/v3"
	"github.com/bbfh-dev/mime/cli"
	"github.com/bbfh-dev/mime/devkit"
)

var Version = getVersion()

var MainProgram = libparsex.Program{
	Name:        "mime",
	Version:     Version,
	Description: "Minecraft data-driven vanilla data & resource pack development kit powered by pre-processors and generators with minimum boilerplate and setup",
	Options:     &cli.Main.Options,
	Args:        &cli.Main.Args,
	Commands: []*libparsex.Program{
		&cli.InitProgram,
	},
	EntryPoint: devkit.Main,
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

func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return "(devel)"
}

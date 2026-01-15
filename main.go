package main

import (
	"os"
	"path/filepath"

	libparsex "github.com/bbfh-dev/lib-parsex/v3"
	"github.com/bbfh-dev/mime/cli"
	"github.com/bbfh-dev/mime/mime"
	"github.com/bbfh-dev/mime/mime/errors"
	"github.com/bbfh-dev/mime/mime/minecraft"
)

const VERSION = "0.1.0-alpha.1"

var MainProgram = libparsex.Program{
	Name:        "mime",
	Version:     VERSION,
	Description: "Minecraft data & resource pack processor designed to be a useful tool for vanilla development rather than a new scripting language or ecosystem.",
	Options:     &cli.Main.Options,
	Args:        &cli.Main.Args,
	Commands: []*libparsex.Program{
		&cli.InitProgram,
	},
	EntryPoint: func(raw_args []string) error {
		if cli.Main.Args.WorkDir != nil {
			if err := os.Chdir(*cli.Main.Args.WorkDir); err != nil {
				return err
			}
		}

		mcmeta_body, err := os.ReadFile("pack.mcmeta")
		if err != nil {
			work_dir, _ := os.Getwd()
			return errors.NewError(errors.ERR_IO, work_dir, err.Error())
		}

		mcmeta := minecraft.NewPackMcmeta(mcmeta_body)
		if err := mcmeta.Validate(); err != nil {
			path, _ := filepath.Abs("pack.mcmeta")
			return errors.NewError(errors.ERR_META, path, err.Error())
		}

		project := mime.New(mcmeta)
		return project.Build()
	},
}

func main() {
	err := libparsex.Run(&MainProgram, os.Args[1:])
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

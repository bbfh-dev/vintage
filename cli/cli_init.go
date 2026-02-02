package cli

import (
	"os"
	"path/filepath"
	"strings"

	liberrors "github.com/bbfh-dev/lib-errors"
	liblog "github.com/bbfh-dev/lib-log"
	libparsex "github.com/bbfh-dev/lib-parsex/v3"
	"github.com/bbfh-dev/mime/devkit/minecraft"
)

var Init struct {
	Options struct {
		Name        string `alt:"n" desc:"Specify the project name that will be used for exporting" default:"untitled"`
		Minecraft   string `alt:"m" desc:"Specify the target Minecraft version. Use '-' to indicate version ranges, e.g. '1.20-1.21'" default:"1.21.11"`
		PackVersion string `alt:"v" desc:"Specify the project version using semantic versioning" default:"0.1.0-alpha"`
		Description string `alt:"d" desc:"Specify the project description"`
	}
	Args struct {
		WorkDir *string
	}
}

var InitProgram = libparsex.Program{
	Name:        "init",
	Description: "Initialize a new Mime project",
	Options:     &Init.Options,
	Args:        &Init.Args,
	Commands:    []*libparsex.Program{},
	EntryPoint: func(raw_args []string) error {
		if Init.Args.WorkDir != nil {
			if err := os.Chdir(*Init.Args.WorkDir); err != nil {
				return liberrors.NewIO(err, *Init.Args.WorkDir)
			}
		}

		if Main.Options.Debug {
			liblog.LogLevel = liblog.LEVEL_DEBUG
		}

		mcmeta_body, err := os.ReadFile("pack.mcmeta")
		if err != nil {
			liblog.Info(0, "Missing existing 'pack.mcmeta', so one will be created instead")
			mcmeta_body = []byte{}
		}

		mcmeta := minecraft.NewPackMcmeta(mcmeta_body)
		if value := Init.Options.Description; value != "" {
			mcmeta.File.Set("pack.description", value)
		}
		mcmeta.File.Set("meta.name", Init.Options.Name)
		if parts := strings.SplitN(Init.Options.Minecraft, "-", 2); len(parts) == 2 {
			mcmeta.File.Set("meta.minecraft.min", parts[0])
			mcmeta.File.Set("meta.minecraft.max", parts[1])
		} else {
			mcmeta.File.Set("meta.minecraft", Init.Options.Minecraft)
		}
		mcmeta.File.Set("meta.version", Init.Options.PackVersion)

		err = os.WriteFile("pack.mcmeta", mcmeta.File.Formatted(), os.ModePerm)
		if err != nil {
			work_dir, _ := os.Getwd()
			return liberrors.NewIO(err, filepath.Join(work_dir, "pack.mcmeta"))
		}

		liblog.Done(
			0,
			"Saved 'pack.mcmeta' for name=%q version=%q minecraft=%q",
			Init.Options.Name,
			Init.Options.PackVersion,
			Init.Options.Minecraft,
		)
		return nil
	},
}

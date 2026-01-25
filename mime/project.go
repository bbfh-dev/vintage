package mime

import (
	"github.com/bbfh-dev/mime/cli"
	"github.com/bbfh-dev/mime/mime/minecraft"
)

type Project struct {
	BuildDir string
	Meta     *minecraft.PackMcmeta
}

func New(mcmeta *minecraft.PackMcmeta) *Project {
	return &Project{
		BuildDir: cli.Main.Options.Output,
		Meta:     mcmeta,
	}
}

func (project *Project) Build() error {
	return nil
}

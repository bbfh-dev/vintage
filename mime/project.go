package mime

import "github.com/bbfh-dev/mime/mime/minecraft"

type Project struct {
	Meta *minecraft.PackMcmeta
}

func New(mcmeta *minecraft.PackMcmeta) *Project {
	return &Project{
		Meta: mcmeta,
	}
}

func (project *Project) Build() error {
	return nil
}

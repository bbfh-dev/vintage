package mime

import (
	"os"

	"github.com/bbfh-dev/mime/cli"
	"github.com/bbfh-dev/mime/mime/language/templates"
	"github.com/bbfh-dev/mime/mime/minecraft"
)

type Project struct {
	BuildDir         string
	Meta             *minecraft.PackMcmeta
	extraFilesToCopy []string
	isDataCached     bool
	isAssetsCached   bool
	inlineTemplates  []*templates.InlineTemplate
}

func New(mcmeta *minecraft.PackMcmeta) *Project {
	return &Project{
		BuildDir:         cli.Main.Options.Output,
		Meta:             mcmeta,
		extraFilesToCopy: []string{},
		isDataCached:     false,
		isAssetsCached:   false,
		inlineTemplates:  []*templates.InlineTemplate{},
	}
}

func (project *Project) Build() error {
	cli.LogInfo(
		false,
		"Building v%s for Minecraft %s",
		project.Meta.Version(),
		project.Meta.MinecraftFormatted(),
	)

	return Pipeline(
		project.detectPackIcon,
		project.checkIfCached(
			"data",
			"data_packs",
			project.getZipName("DP"),
			&project.isDataCached,
		),
		project.checkIfCached(
			"assets",
			"resource_packs",
			project.getZipName("RP"),
			&project.isAssetsCached,
		),
		project.loadTemplates,
		project.genDataPack,
		project.genResourcePack,
		project.depend(project.zipPacks, cli.Main.Options.Zip),
		project.depend(project.weldPacks, cli.Main.Options.Zip),
	)
}

func (project *Project) detectPackIcon() error {
	_, err := os.Stat("pack.png")
	if os.IsNotExist(err) {
		cli.LogWarn(false, "No pack icon found")
		return nil
	}

	cli.LogInfo(false, "Found 'pack.png'")
	project.extraFilesToCopy = append(project.extraFilesToCopy, "pack.png")
	return nil
}

func (project *Project) depend(call func() error, condition bool) func() error {
	if condition {
		return call
	}
	return nil
}

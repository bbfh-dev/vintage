package devkit

import (
	libescapes "github.com/bbfh-dev/lib-ansi-escapes"
	"github.com/bbfh-dev/mime/cli"
	"github.com/bbfh-dev/mime/devkit/internal"
	"github.com/bbfh-dev/mime/devkit/language"
	"github.com/bbfh-dev/mime/devkit/minecraft"
)

type Project struct {
	Meta     *minecraft.PackMcmeta
	BuildDir string

	extraFilesToCopy []string
	isDataCached     bool
	isAssetsCached   bool

	generatorTemplates map[string]*language.GeneratorTemplate
	inlineTemplates    map[string]*language.InlineTemplate
}

func New(mcmeta *minecraft.PackMcmeta) *Project {
	return &Project{
		Meta:     mcmeta,
		BuildDir: cli.Main.Options.Output,

		extraFilesToCopy: []string{},
		isDataCached:     false,
		isAssetsCached:   false,

		generatorTemplates: map[string]*language.GeneratorTemplate{},
		inlineTemplates:    map[string]*language.InlineTemplate{},
	}
}

func (project *Project) Build() error {
	cli.LogInfo(
		0,
		"Building %s for Minecraft %s",
		cli.ColorWord("v"+project.Meta.Version().String(), libescapes.TextColorBrightMagenta),
		cli.ColorWord(project.Meta.MinecraftFormatted(), libescapes.TextColorBrightBlue),
	)

	return internal.Pipeline(
		project.LogHeader("Preparing..."),
		project.DetectPackIcon,
		project.CheckIfCached(&project.isDataCached, FOLDER_DATA),
		project.CheckIfCached(&project.isAssetsCached, FOLDER_ASSETS),
		project.LoadTemplates,
		project.GenerateDataPack,
		project.GenerateResourcePack,
		project.GenerateFromTemplates,
		internal.If[internal.Task](cli.Main.Options.Zip, project.ZipPacks),
		internal.If[internal.Task](cli.Main.Options.Zip, project.WeldPacks),
	)
}

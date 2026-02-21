package devkit

import (
	liblog "github.com/bbfh-dev/lib-log"
	"github.com/bbfh-dev/vintage/cli"
	"github.com/bbfh-dev/vintage/devkit/internal/autolibs"
	"github.com/bbfh-dev/vintage/devkit/internal/mcfunc"
	"github.com/bbfh-dev/vintage/devkit/internal/pipeline"
	"github.com/bbfh-dev/vintage/devkit/internal/templates"
	"github.com/bbfh-dev/vintage/devkit/minecraft"
)

type Project struct {
	Meta     *minecraft.PackMcmeta
	BuildDir string

	extraFilesToCopy []string
	isDataCached     bool
	isAssetsCached   bool

	generatorTemplates map[string]*templates.Generator
	collectorTemplates map[string]*templates.Collector
	inlineTemplates    map[string]*templates.Inline

	libraries []*autolibs.Library
}

func New(mcmeta *minecraft.PackMcmeta) *Project {
	return &Project{
		Meta:     mcmeta,
		BuildDir: cli.Build.Options.Output,

		extraFilesToCopy: []string{},
		isDataCached:     false,
		isAssetsCached:   false,

		generatorTemplates: map[string]*templates.Generator{},
		collectorTemplates: map[string]*templates.Collector{},
		inlineTemplates:    map[string]*templates.Inline{},

		libraries: []*autolibs.Library{},
	}
}

func (project *Project) Build() error {
	liblog.Info(
		0,
		"Building %s for Minecraft %s",
		"v"+project.Meta.Version().String(),
		project.Meta.MinecraftFormatted(),
	)

	return pipeline.New(
		project.LogHeader("Preparing..."),
		project.DetectPackIcon,
		project.CheckIfCached(&project.isDataCached, FOLDER_DATA),
		project.CheckIfCached(&project.isAssetsCached, FOLDER_ASSETS),
		project.LoadTemplates,
		project.GenerateDataPack,
		project.GenerateResourcePack,
		project.GenerateFromTemplates,
		project.writeMcfunctions,
		// project.CollectFromTemplates,
		project.LoadAutoLibs,
		project.ManageAutoLibs,
		pipeline.If[pipeline.Task](cli.Build.Options.Zip).
			Then(project.ZipPacks),
		pipeline.If[pipeline.Task](cli.Build.Options.Zip).
			Then(project.WeldPacks),
	)
}

func Reset() {
	mcfunc.UsedNamespaces = map[string]byte{}
	mcfunc.Registry = map[string][]string{}
}

package minecraft

import (
	"errors"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
)

type PackMcmeta struct {
	File     *JsonFile
	Versions VersionRange
}

func NewPackMcmeta(body []byte) *PackMcmeta {
	return &PackMcmeta{
		File:     NewJsonFile(body),
		Versions: VersionRange{},
	}
}

// Checks whether all Mime-required fields are present
func (mcmeta *PackMcmeta) Validate() error {
	return errors.Join(
		mcmeta.File.ExpectField("meta.name", gjson.String),
		mcmeta.File.ExpectField("meta.minecraft", gjson.String),
		mcmeta.File.ExpectField("meta.version", gjson.String),
	)
}

func (mcmeta *PackMcmeta) FillVersion(formats map[string]Version) *PackMcmeta {
	// NOTE: This is not done in [NewPackMeta()] so that the file
	// can be validated first.
	mc_version := strings.Split(mcmeta.Minecraft().String(), "-")
	mcmeta.Versions = VersionRange{
		Min: formats[mc_version[0]],
		Max: formats[mc_version[len(mc_version)-1]],
	}
	return mcmeta
}

// Writes pack version into the file
func (mcmeta *PackMcmeta) SaveVersion() error {
	mcmeta.File.Set("pack.pack_format", mcmeta.Versions.Min.Value())

	switch mcmeta.Versions.Max.Flag {

	case USES_MIN_MAX_FORMAT:
		mcmeta.File.Set("pack.min_format", mcmeta.Versions.Min.Digits)
		mcmeta.File.Set("pack.max_format", mcmeta.Versions.Max.Digits)

	case USES_SUPPORTED_FORMATS:
		mcmeta.File.Set("pack.supported_formats.min_inclusive", mcmeta.Versions.Min.Value())
		mcmeta.File.Set("pack.supported_formats.max_inclusive", mcmeta.Versions.Max.Value())

	default:
		return fmt.Errorf(
			"Version %d does not support pack format ranges. Skipping...",
			mcmeta.Versions.Max.Digits,
		)
	}

	return nil
}

func (mcmeta *PackMcmeta) Name() gjson.Result {
	return mcmeta.File.Get("meta.name")
}

func (mcmeta *PackMcmeta) Minecraft() gjson.Result {
	return mcmeta.File.Get("meta.minecraft")
}

func (mcmeta *PackMcmeta) Version() gjson.Result {
	return mcmeta.File.Get("meta.version")
}

func (mcmeta *PackMcmeta) PrintableVersion() string {
	version := mcmeta.File.Get("meta.version")
	if !version.Exists() {
		return "UNKNOWN"
	}

	return strings.ReplaceAll(version.String(), ".", "-")
}

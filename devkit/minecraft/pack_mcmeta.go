package minecraft

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bbfh-dev/vintage/devkit/internal/drive"
	"github.com/tidwall/gjson"
)

// pack.mcmeta file util struct
type PackMcmeta struct {
	File     *drive.JsonFile
	Versions PackVersionRange
}

func NewPackMcmeta(body []byte) *PackMcmeta {
	return &PackMcmeta{
		File:     drive.NewJsonFile(body),
		Versions: PackVersionRange{},
	}
}

func (mcmeta *PackMcmeta) Clone() *PackMcmeta {
	return NewPackMcmeta(mcmeta.File.Body)
}

// Checks whether all fields required by Vintage are present
func (mcmeta *PackMcmeta) Validate() error {
	return errors.Join(
		mcmeta.File.ExpectField("meta.name", gjson.String),
		mcmeta.File.ExpectField("meta.minecraft", gjson.String, gjson.JSON),
		mcmeta.File.ExpectField("meta.version", gjson.String),
	)
}

// Sets mcmeta.Versions based on provided formats map
//
// NOTE: This is not done in [NewPackMeta()] so that the file
// can be validated first.
func (mcmeta *PackMcmeta) FillVersion(formats map[string]PackVersion) *PackMcmeta {
	mc_version := mcmeta.Minecraft()
	if mc_version[1] == "" {
		mc_version[1] = mc_version[0]
	}
	mcmeta.Versions = PackVersionRange{
		Min: formats[mc_version[0]],
		Max: formats[mc_version[1]],
	}
	return mcmeta
}

// Writes pack version into the in-memory file
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

// Returns a tuple of [min_version, max_version]
func (mcmeta *PackMcmeta) Minecraft() [2]string {
	field := mcmeta.File.Get("meta.minecraft")
	if !field.Exists() {
		return [2]string{}
	}

	if field.Type == gjson.String {
		return [2]string{field.String()}
	}

	out := [2]string{}
	if field := field.Get("min"); field.Exists() {
		out[0] = field.String()
	}
	if field := field.Get("max"); field.Exists() {
		out[1] = field.String()
	}

	return out
}

func (mcmeta *PackMcmeta) MinecraftFormatted() string {
	versions := mcmeta.Minecraft()
	if versions[1] == "" {
		return versions[0]
	}
	return fmt.Sprintf("%s — %s", versions[0], versions[1])
}

func (mcmeta *PackMcmeta) Version() gjson.Result {
	return mcmeta.File.Get("meta.version")
}

func (mcmeta *PackMcmeta) VersionFormatted() string {
	version := mcmeta.Version()
	if !version.Exists() {
		return "<?>"
	}
	return strings.ReplaceAll(version.String(), ".", "-")
}

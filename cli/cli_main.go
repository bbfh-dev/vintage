package cli

var Main struct {
	Options struct {
		Output string `alt:"o" desc:"Output directory relative to the pack working dir" default:"./build"`
		Zip    bool   `alt:"z" desc:"Export data & resource packs as .zip files"`
		Force  bool   `alt:"f" desc:"Force building even if the build directory looks off"`
		Debug  bool   `alt:"d" desc:"Print verbose debug information"`
	}
	Args struct {
		WorkDir *string
	}
}

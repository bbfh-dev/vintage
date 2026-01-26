package cli

var Main struct {
	Options struct {
		Output string `alt:"o" desc:"Output directory relative to the pack working dir" default:"./build"`
		Zip    bool   `alt:"z" desc:"Export data & resource packs as .zip files"`
		Debug  bool   `alt:"d" desc:"Print verbose debug information"`
		Force  bool   `alt:"f" desc:"Force build even if the project was cached"`
	}
	Args struct {
		WorkDir *string
	}
}

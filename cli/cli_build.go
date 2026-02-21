package cli

var Build struct {
	Options struct {
		Output           string `alt:"o" desc:"Output directory relative to the pack working dir" default:"./build"`
		Zip              bool   `alt:"z" desc:"Export data & resource packs as .zip files"`
		Debug            bool   `alt:"d" desc:"Print verbose debug information"`
		Force            bool   `alt:"f" desc:"Force build even if the project was cached"`
		DeleteUnusedLibs bool   `desc:"Delete unused automatic libraries rather than appending .disabled to file names"`
	}
	Args struct {
		WorkDir *string
	}
}

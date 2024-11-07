package colorpicker

import "go.wdy.de/nago/presentation/ui"

type Palette []ui.Color

// DefaultPalette contains 4x7 distinct colors in different harmonic octaves.
var DefaultPalette = Palette{
	// first octave
	"#1CCDFB",
	"#FD63D7",
	"#FBC83E",
	"#7CE513",
	"#FB5509",
	"#DCBEFF",
	"#FED8B1",
	// second octave
	"#1940BF",
	"#A800D6",
	"#FA991B",
	"#01BA6C",
	"#FA2C7F",
	"#FABED4",
	"#FEFAC8",
	// third octave
	"#21A7FF",
	"#8300C9",
	"#FDFF00",
	"#189A40",
	"#CC0E02",
	"#D88ED4",
	"#FDB77B",
	// fourth octave
	"#91C3FD",
	"#813EFF",
	"#FDDD6A",
	"#67C981",
	"#F61304",
	"#BE8EFF",
	"#FDD48B",
}

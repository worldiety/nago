package ui

import (
	"fmt"
	"go.wdy.de/nago/presentation/ora"
)

// Color specifies either a hex color like #rrggbb or #rrggbbaa or an internal custom color name.
type Color string

func (c Color) ora() ora.Color {
	return ora.Color(c)
}

// WithTransparency updates the alpha value part of the color (0-100), where 25% transparent means 75% opaque.
func (c Color) WithTransparency(a int8) Color {
	if len(c) == 8 {
		c = c[:len(c)-2]
	}

	// recalculate into 0-255 and invert
	ai := 255 - int(min(max((float64(a)/100*255), 0), 255))

	return Color(fmt.Sprintf("%s%02x", string(c), ai))
}

// WithBrightness recalculates the hex RGB value into HSL, set the given brightness (0-100) and returns
// the new hex RGB value.
func (c Color) WithBrightness(b int8) Color {
	return mustParseHSL(string(c)).Brightness(float64(b)).RGBHex()
}

const (
	// M0 is the source main color variable name.
	M0 Color = "M0"

	// M1 is a variable name usually used for the background.
	M1 Color = "M1"

	// M2 is a variable name usually used as the background for first level container.
	M2 Color = "M2"

	// M3 is a variable name usually used for a card bottom area.
	M3 Color = "M3"

	// M4 is a variable name usually used for a card body area.
	M4 Color = "M4"

	// M5 is a variable name usually used for Line / Dot on SC.
	M5 Color = "M5"

	// M6 is a variable name usually used for hovered containers.
	M6 Color = "M6"

	// M7 is a variable name usually used for Text or muted icons.
	M7 Color = "M7"

	// M8 is a variable name usually used for Text or icons.
	M8 Color = "M8"

	// M9 is a variable name usually used for card Top area.
	M9 Color = "M9"

	// A0 is the source accent color name.
	A0 Color = "A0"

	// A1 is a variable name usually used for card progress bars, H2 or Borders.
	A1 Color = "A1"

	// I0 is the source interactive color name.
	I0 Color = "I0"

	// I1 is a variable name usually used for buttons.
	I1 Color = "I1"

	// SE0 is the source error color variable name.
	SE0 Color = "SE0"

	// SW0 is the source warning color variable name.
	SW0 Color = "SW0"

	// SG0 is the source good color variable name.
	SG0 Color = "SG0"

	// SV0 is the source Informative color variable name.
	SV0 Color = "SV0"

	// SI0 is the source disabled input color variable name.
	SI0 Color = "SI0"

	// ST0 is the source disabled text color variable name.
	ST0 Color = "ST0"
)

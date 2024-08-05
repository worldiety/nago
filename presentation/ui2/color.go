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

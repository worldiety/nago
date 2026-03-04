// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package color

import (
	"fmt"
	"strings"

	"github.com/worldiety/material-color-utilities/hct"
	"go.wdy.de/nago/pkg/xcolor"
)

// Color specifies either a hex color like #rrggbb or #rrggbbaa or an internal custom color name.
// Optional opacity values of custom color names will be attached to the color name, separated by the `/` character,
// e.g. `M8/128` describes the custom color `M8` with opacity `128` (0-255)
type Color string

func (c Color) IsAbsolute() bool {
	return strings.HasPrefix(string(c), `#`)
}

// WithTransparency updates the alpha value part of the color (0-100), where 25% transparent means 75% opaque.
func (c Color) WithTransparency(a int8) Color {
	// recalculate into 0-255 and invert
	ai := 255 - int(min(max(float64(a)/100*255, 0), 255))

	if c.IsAbsolute() {
		if len(c) == 9 {
			co := c[:len(c)-2]

			return Color(fmt.Sprintf("%s%02x", string(co), ai))
		}
	}

	return Color(fmt.Sprintf("%s/%d", string(c), ai))
}

// WithoutTransparency updates the removes the alpha value of the color.
func (c Color) WithoutTransparency() Color {
	if c.IsAbsolute() {
		if len(c) == 9 {
			return c[:len(c)-2]
		}

		return c
	}

	return Color(strings.Split(string(c), "/")[0])
}

// WithChromaAndTone applies the given chroma and tone values on the actual hue value using the HCT colorspace.
func (c Color) WithChromaAndTone(chroma float64, tone float64) (Color, error) {
	cl, err := xcolor.ParseHex(string(c))
	if err != nil {
		return c, err
	}

	v := hct.FromInt(cl.Int())
	v.SetChroma(chroma)
	v.SetTone(tone)
	return Color(xcolor.Hex(xcolor.RGBA8FromInt(v.ToInt()))), nil
}

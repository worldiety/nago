package ora

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type hslColor struct {
	H float64 `json:"h"` // degree from 0 - 360
	S float64 `json:"s"` // percent from 0 to 100
	L float64 `json:"l"` // brightness percent from 0 to 100
}

func (c hslColor) RGBHex() Color {
	h := c.H
	s := c.S
	l := c.L
	r, g, b := hsbToRGB(h, s, l)

	return Color(fmt.Sprintf("#%02x%02x%02x", r, g, b))
}

func (c hslColor) Brightness(b float64) hslColor {
	c.L = b / 100.0
	return c
}

func mustParseHSL(hex string) hslColor {
	r, g, b, _ := hexToRGBA(hex)
	h, s, l := rgbToHSL(r, g, b)
	return hslColor{H: h, S: s, L: l}
}

func hexToRGBA(hex string) (r, g, b, a uint8) {
	if !strings.HasPrefix(hex, "#") {
		panic(fmt.Sprintf("unsupported hex color notation: %s", hex))
	}

	hex = hex[1:]
	switch len(hex) {
	case 8:
		//ok
	case 6:
		hex += "FF"
	default:
		panic(fmt.Sprintf("unsupported hex color notation: %s", hex))
	}

	rgba, _ := strconv.ParseUint(hex, 16, 32)
	r = uint8((rgba >> 24) & 0xFF)
	g = uint8((rgba >> 16) & 0xFF)
	b = uint8((rgba >> 8) & 0xFF)
	a = uint8(rgba & 0xFF)
	return
}

func round(x float64) float64 {
	return math.Round(x*1000) / 1000
}

func getMaxMin(a, b, c float64) (max, min float64) {
	if a > b {
		max = a
		min = b
	} else {
		max = b
		min = a
	}
	if c > max {
		max = c
	} else if c < min {
		min = c
	}
	return max, min
}

func rgbToHSL(r, g, b uint8) (h, s, l float64) {
	// convert uint32 pre-multiplied value to uint8
	// The r,g,b values are divided by 255 to change the range from 0..255 to 0..1:
	Rnot := float64(r) / 255
	Gnot := float64(g) / 255
	Bnot := float64(b) / 255
	Cmax, Cmin := getMaxMin(Rnot, Gnot, Bnot)
	Δ := Cmax - Cmin
	// Lightness calculation:
	l = (Cmax + Cmin) / 2
	// Hue and Saturation Calculation:
	if Δ == 0 {
		h = 0
		s = 0
	} else {
		switch Cmax {
		case Rnot:
			h = 60 * (math.Mod((Gnot-Bnot)/Δ, 6))
		case Gnot:
			h = 60 * (((Bnot - Rnot) / Δ) + 2)
		case Bnot:
			h = 60 * (((Rnot - Gnot) / Δ) + 4)
		}
		if h < 0 {
			h += 360
		}

		s = Δ / (1 - math.Abs((2*l)-1))
	}

	return h, round(s), round(l)
}

func hsbToRGB(h, s, l float64) (r int, g int, b int) {
	if h < 0 || h >= 360 ||
		s < 0 || s > 1 ||
		l < 0 || l > 1 {
		return 0, 0, 0
	}
	// When 0 ≤ h < 360, 0 ≤ s ≤ 1 and 0 ≤ l ≤ 1:
	C := (1 - math.Abs((2*l)-1)) * s
	X := C * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - (C / 2)
	var Rnot, Gnot, Bnot float64

	switch {
	case 0 <= h && h < 60:
		Rnot, Gnot, Bnot = C, X, 0
	case 60 <= h && h < 120:
		Rnot, Gnot, Bnot = X, C, 0
	case 120 <= h && h < 180:
		Rnot, Gnot, Bnot = 0, C, X
	case 180 <= h && h < 240:
		Rnot, Gnot, Bnot = 0, X, C
	case 240 <= h && h < 300:
		Rnot, Gnot, Bnot = X, 0, C
	case 300 <= h && h < 360:
		Rnot, Gnot, Bnot = C, 0, X
	}
	r = int(uint8(math.Round((Rnot + m) * 255)))
	g = int(uint8(math.Round((Gnot + m) * 255)))
	b = int(uint8(math.Round((Bnot + m) * 255)))
	return r, g, b
}

func hue2rgb(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1/6 {
		return p + (q-p)*6*t
	}
	if t < 1/2 {
		return q
	}
	if t < 2/3 {
		return p + (q-p)*(2/3-t)*6
	}
	return p
}

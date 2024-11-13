package xcolor

type YPbPr Vec4f

func (c YPbPr) RGBA() RGBA {
	r, g, b := yPbPrToRGB(float64(c[0]), float64(c[1]), float64(c[2]))
	res := RGBA8{r, g, b}.RGBA()
	res[3] = c[3]
	return res
}

// WithLuma sets the given luma value (0-1) and returns the new color.
func (c YPbPr) WithLuma(l float32) YPbPr {
	c[0] = l
	return c
}

func rgbToYPbPr(r, g, b uint8) (y, pb, pr float64) {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	y = 0.299*rf + 0.587*gf + 0.114*bf
	pb = -0.168736*rf - 0.331264*gf + 0.5*bf
	pr = 0.5*rf - 0.418688*gf - 0.081312*bf
	return
}

func yPbPrToRGB(y, pb, pr float64) (r, g, b uint8) {
	rf := y + 1.402*pr
	gf := y - 0.344136*pb - 0.714136*pr
	bf := y + 1.772*pb

	// Clamping the values to the range [0, 1]
	rf = clamp(rf, 0.0, 1.0)
	gf = clamp(gf, 0.0, 1.0)
	bf = clamp(bf, 0.0, 1.0)

	// Converting to uint8
	r = uint8(rf * 255.0)
	g = uint8(gf * 255.0)
	b = uint8(bf * 255.0)
	return
}

// Helper function to clamp values
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xcolor

type Color interface {
	RGBA() RGBA
}

type RGBA Vec4f

func (c RGBA) RGBA() RGBA {
	return c
}

func (c RGBA) YPbPr() YPbPr {
	rgb := c.RGBA8()
	y, pb, pr := rgbToYPbPr(rgb[0], rgb[1], rgb[2])
	return YPbPr{float32(y), float32(pb), float32(pr), c[3]}
}

func (c RGBA) RGBA8() RGBA8 {
	r := max(0, min(1, c[0]))
	g := max(0, min(1, c[1]))
	b := max(0, min(1, c[2]))
	a := max(0, min(1, c[3]))

	// Convert to 0-255 range
	ri := uint8(r * 255)
	gi := uint8(g * 255)
	bi := uint8(b * 255)
	ai := uint8(a * 255)

	return RGBA8{ri, gi, bi, ai}
}

func (c RGBA) LAB() LAB {
	rgba := c.RGBA8()

	cl, ca, cb := rgb2LAB(rgba[0], rgba[1], rgba[2])
	return LAB{float32(cl), float32(ca), float32(cb), c[3]}
}

func (c RGBA) YUV() YUV {
	rgb := c.RGBA8()
	y, u, v := rgbToYUV(rgb[0], rgb[1], rgb[2])
	return YUV{y, u, v, rgb[3]}
}

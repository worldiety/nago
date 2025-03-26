// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xcolor

type YUV Vec4i8

func (c YUV) RGBA() RGBA {
	r, g, b := yuvToRGB(c[0], c[1], c[2])
	return RGBA8{r, g, b, c[3]}.RGBA()
}

func (c YUV) WithLuma(l uint8) YUV {
	c[0] = l
	return c
}

func rgbToYUV(r, g, b uint8) (y, u, v uint8) {
	y = uint8(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))
	u = uint8(-0.14713*float64(r) - 0.28886*float64(g) + 0.436*float64(b) + 128)
	v = uint8(0.615*float64(r) - 0.51499*float64(g) - 0.10001*float64(b) + 128)
	return
}

func yuvToRGB(y, u, v uint8) (r, g, b uint8) {
	y1 := float64(y)
	u1 := float64(u) - 128
	v1 := float64(v) - 128

	r = uint8(y1 + 1.13983*v1)
	g = uint8(y1 - 0.39465*u1 - 0.58060*v1)
	b = uint8(y1 + 2.03211*u1)

	return
}

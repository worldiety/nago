// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package image

import (
	"image"
	"image/color"
	"math"
)

// shrink scales src down into dst with the same weights as draw.BiLinear of golang.org/x/image: a tent
// filter whose support is widened by the scale factor, so that every source pixel contributes. Unlike
// draw.BiLinear, which keeps a temporary buffer of dst width × src height × 32 bytes (about 470 MiB for a
// 4800×6400 image), it streams the source row by row and only holds the few destination rows the current
// source row contributes to.
func shrink(dst *image.RGBA, src image.Image) {
	sb, db := src.Bounds(), dst.Rect
	sw, sh, dw, dh := sb.Dx(), sb.Dy(), db.Dx(), db.Dy()
	if sw < dw || sh < dh || dw == 0 || dh == 0 {
		panic("shrink: dst must not be larger than src")
	}

	cols := tentTaps(dw, sw)
	rows := tentTaps(dh, sh)

	// the destination rows a source row contributes to, with their weight
	type share struct {
		row int
		w   float32
	}
	shares := make([][]share, sh)
	last := make([]int, dh) // the last source row of each destination row
	for y, t := range rows {
		for i, w := range t.w {
			shares[t.from+i] = append(shares[t.from+i], share{y, w})
		}
		last[y] = t.from + len(t.w) - 1
	}

	at := pixels(src)
	line := make([]float32, sw*4)  // the current source row
	hline := make([]float32, dw*4) // the current source row, scaled horizontally
	open := map[int][]float32{}    // the destination rows still receiving contributions
	var free [][]float32           // row buffers for reuse

	for sy := range sh {
		for sx := range sw {
			r, g, b, a := at(sb.Min.X+sx, sb.Min.Y+sy)
			line[sx*4+0], line[sx*4+1], line[sx*4+2], line[sx*4+3] = float32(r), float32(g), float32(b), float32(a)
		}

		for x, t := range cols {
			var r, g, b, a float32
			for i, w := range t.w {
				p := line[(t.from+i)*4:]
				r += w * p[0]
				g += w * p[1]
				b += w * p[2]
				a += w * p[3]
			}
			hline[x*4+0], hline[x*4+1], hline[x*4+2], hline[x*4+3] = r, g, b, a
		}

		for _, s := range shares[sy] {
			acc, ok := open[s.row]
			if !ok {
				if n := len(free); n > 0 {
					acc, free = free[n-1], free[:n-1]
					clear(acc)
				} else {
					acc = make([]float32, dw*4)
				}
				open[s.row] = acc
			}
			for i, v := range hline {
				acc[i] += s.w * v
			}

			if last[s.row] == sy {
				out := dst.Pix[dst.PixOffset(db.Min.X, db.Min.Y+s.row):]
				for i, v := range acc {
					out[i] = uint8(min(255, max(0, v+0.5)))
				}
				delete(open, s.row)
				free = append(free, acc)
			}
		}
	}
}

// taps are the normalized weights of the consecutive source pixels from which one destination pixel is
// mixed.
type taps struct {
	from int
	w    []float32
}

// tentTaps computes the weights of draw.BiLinear (see newDistrib in golang.org/x/image/draw) for scaling
// sw pixels down to dw pixels.
func tentTaps(dw, sw int) []taps {
	scale := float64(sw) / float64(dw)
	res := make([]taps, dw)
	for x := range res {
		center := (float64(x)+0.5)*scale - 0.5
		i := max(0, int(math.Floor(center-scale)))
		j := min(sw, int(math.Ceil(center+scale)))

		var t taps
		total := 0.0
		var ws []float64
		for c := i; c < j; c++ {
			w := 1 - math.Abs(center-float64(c))/scale
			if w <= 0 {
				if len(ws) == 0 {
					continue // leading zeros, start later
				}
				w = 0
			}
			if len(ws) == 0 {
				t.from = c
			}
			ws = append(ws, w)
			total += w
		}
		for len(ws) > 0 && ws[len(ws)-1] == 0 {
			ws = ws[:len(ws)-1]
		}
		t.w = make([]float32, len(ws))
		for k, w := range ws {
			t.w[k] = float32(w / total)
		}
		res[x] = t
	}
	return res
}

// pixels returns an accessor for the alpha-premultiplied 8-bit color of a pixel. The usual decoded formats
// are read directly, which avoids an interface allocation per pixel through [image.Image.At].
func pixels(src image.Image) func(x, y int) (r, g, b, a uint8) {
	switch img := src.(type) {
	case *image.RGBA:
		return func(x, y int) (r, g, b, a uint8) {
			p := img.Pix[img.PixOffset(x, y):]
			return p[0], p[1], p[2], p[3]
		}
	case *image.NRGBA:
		return func(x, y int) (r, g, b, a uint8) {
			p := img.Pix[img.PixOffset(x, y):]
			pa := uint16(p[3])
			return uint8(uint16(p[0]) * pa / 255), uint8(uint16(p[1]) * pa / 255), uint8(uint16(p[2]) * pa / 255), p[3]
		}
	case *image.YCbCr:
		return func(x, y int) (r, g, b, a uint8) {
			r, g, b = color.YCbCrToRGB(img.Y[img.YOffset(x, y)], img.Cb[img.COffset(x, y)], img.Cr[img.COffset(x, y)])
			return r, g, b, 0xff
		}
	case *image.Gray:
		return func(x, y int) (r, g, b, a uint8) {
			v := img.Pix[img.PixOffset(x, y)]
			return v, v, v, 0xff
		}
	default:
		return func(x, y int) (r, g, b, a uint8) {
			cr, cg, cb, ca := src.At(x, y).RGBA()
			return uint8(cr >> 8), uint8(cg >> 8), uint8(cb >> 8), uint8(ca >> 8)
		}
	}
}

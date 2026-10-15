// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package asset declares the events, which the host sends, when it has loaded an image or a texture, e.g. with
// canvas.LoadImage or three.LoadTexture. A program runs synchronously, so loading returns a handle right away,
// and the program learns about the result with a handler:
//
//	func OnLoaded(g *Game, e asset.Loaded) {
//		g.Ready = true
//	}
//
// Until then, drawing the image does nothing. The host loads only URLs of the application, e.g. its resources and
// uploads.
package asset

import "go.wdy.de/nago/pkg/gostrict/std/list"

// Loaded tells, that the image at the URL is ready. Width and Height are its size in pixels.
type Loaded struct {
	URL    string
	Width  float64
	Height float64
}

// Failed tells, that the image at the URL can not be loaded.
type Failed struct {
	URL    string
	Reason string
}

// Rect is a region of an image in pixels, e.g. a frame of a sprite sheet.
type Rect struct {
	X, Y, W, H float64
}

// Atlas describes the frames of a sprite sheet, e.g. for canvas.DrawImageRegion.
type Atlas struct {
	Frames list.List[Rect]
}

// Grid returns the atlas of a sheet with rows of frames of the same size, numbered from the left to the right and
// from the top to the bottom.
func Grid(cols, rows int32, w, h float64) Atlas {
	frames := list.New[Rect]()
	for y := range rows {
		for x := range cols {
			frames.Append(Rect{X: float64(x) * w, Y: float64(y) * h, W: w, H: h})
		}
	}
	return Atlas{Frames: frames}
}

// Frame returns the frame i or false, if the atlas has no such frame.
func (a Atlas) Frame(i int32) (Rect, bool) {
	return a.Frames.Get(i)
}

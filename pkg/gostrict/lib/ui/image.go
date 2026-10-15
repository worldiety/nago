// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	"go.wdy.de/nago/presentation/core"
	nui "go.wdy.de/nago/presentation/ui"
)

// ObjectFit declares how to layout an image in the according image view. If 0 or omitted, an automatic behavior is
// applied. This may treat SVG and raster formats differently.
type ObjectFit int32

const (
	// FitAuto is mostly undefined and used to not break the default behavior of the web-renderer.
	// For raster images like png or jpg this results in FitCover and for SVG it is FitFill.
	FitAuto ObjectFit = 0
	// FitFill fills the entire image view by stretching the pixel buffer.
	FitFill ObjectFit = 1
	// FitCover fills the entire image view, possibly zooming the image so much, that the resolution would
	// not be enough for a sharp image or to understand the image geometrics at all.
	FitCover ObjectFit = 2
	// FitContain is the inverse of FitCover, which keeps the aspect ratio and scales the image down, so that it can
	// be seen entirely. The image view adds transparent borders (letterboxing).
	FitContain ObjectFit = 3
	// FitNone maps the pixel buffer 1:1 into its render context.
	FitNone ObjectFit = 4
	// FitRepeat repeats the image in both directions to fill the entire image view with a 1:1 mapping of the
	// pixel buffer.
	FitRepeat ObjectFit = 5
)

// TImage is an image, see Image. It displays raster or vector images with support for light and dark mode variants,
// SVG embedding, object fit, custom colors, borders, padding, and accessibility.
type TImage struct{ ref js.Ref }

func (TImage) isView() {}

// Image returns an image with a default frame size of Auto x L160.
//
//gostrict:host
func Image() TImage {
	return TImage{ref: wrap(nui.Image())}
}

// ImageIcon returns an icon of the svg with the size L24 x L24. It is invisible, if the svg is empty.
//
//gostrict:host
func ImageIcon(svg SVG) TImage {
	return TImage{ref: wrap(nui.ImageIcon(core.SVG(svg)))}
}

// ImageIconAdaptive returns an icon with the size L24 x L24, which is onLight in the light mode and onDark in the
// dark mode. It is invisible, if onLight is empty.
//
//gostrict:host
func ImageIconAdaptive(onLight, onDark SVG) TImage {
	return TImage{ref: wrap(nui.ImageIconAdaptive(core.SVG(onLight), core.SVG(onDark)))}
}

// URI shows the image of the uri, e.g. a static resource or an image of a third-party domain.
//
//gostrict:host
func (c TImage) URI(uri URI) TImage {
	return TImage{ref: wrap(get[nui.TImage](c.ref).URI(core.URI(uri)))}
}

// URIAdaptive is like URI, but shows the image of light or of dark, depending on the mode.
//
//gostrict:host
func (c TImage) URIAdaptive(light, dark URI) TImage {
	return TImage{ref: wrap(get[nui.TImage](c.ref).URIAdaptive(core.URI(light), core.URI(dark)))}
}

// FillColor sets the fill color of an embedded SVG, which uses currentColor. Otherwise, it has no effect.
//
//gostrict:host
func (c TImage) FillColor(color Color) TImage {
	return TImage{ref: wrap(get[nui.TImage](c.ref).FillColor(nui.Color(color)))}
}

// StrokeColor sets the stroke color of an embedded SVG, which uses the stroke color. Otherwise, it has no effect.
//
//gostrict:host
func (c TImage) StrokeColor(color Color) TImage {
	return TImage{ref: wrap(get[nui.TImage](c.ref).StrokeColor(nui.Color(color)))}
}

// AccessibilityLabel sets a label for screen readers.
//
//gostrict:host
func (c TImage) AccessibilityLabel(label string) TImage {
	return TImage{ref: wrap(get[nui.TImage](c.ref).AccessibilityLabel(label).(nui.TImage))}
}

// Visible controls the visibility of the image; false hides it.
//
//gostrict:host
func (c TImage) Visible(b bool) TImage {
	return TImage{ref: wrap(get[nui.TImage](c.ref).Visible(b).(nui.TImage))}
}

// Padding sets the inner spacing around the image.
//
//gostrict:host
func (c TImage) Padding(padding Padding) TImage {
	return TImage{ref: wrap(get[nui.TImage](c.ref).Padding(padding.nago()).(nui.TImage))}
}

// Border sets the border styling of the image.
//
//gostrict:host
func (c TImage) Border(border Border) TImage {
	return TImage{ref: wrap(get[nui.TImage](c.ref).Border(border.nago()).(nui.TImage))}
}

// Frame sets the layout frame of the image, including size and positioning.
//
//gostrict:host
func (c TImage) Frame(frame Frame) TImage {
	return TImage{ref: wrap(get[nui.TImage](c.ref).Frame(frame.nago()).(nui.TImage))}
}

// WithFrame changes the layout frame with fn, which gets the current one.
//
//gostrict:host
func (c TImage) WithFrame(fn func(Frame) Frame) TImage {
	return TImage{ref: wrap(get[nui.TImage](c.ref).WithFrame(func(f nui.Frame) nui.Frame { return fn(frameFrom(f)).nago() }).(nui.TImage))}
}

// ObjectFit sets how the image is resized inside its frame, e.g. FitCover or FitContain.
//
//gostrict:host
func (c TImage) ObjectFit(fit ObjectFit) TImage {
	return TImage{ref: wrap(get[nui.TImage](c.ref).ObjectFit(nui.ObjectFit(fit)))}
}

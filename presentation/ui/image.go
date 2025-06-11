// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"bytes"
	"encoding/base64"
	"go.wdy.de/nago/application/image"
	httpimage "go.wdy.de/nago/application/image/http"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// ObjectFit declares how to layout an image in the according image view.
// If 0 or omitted, an automatic behavior is applied. This may treat SVG and raster formats differently.
type ObjectFit int

const (
	// FitAuto is mostly undefined and used to not break the default behavior of the web-renderer.
	// For raster images like png or jpg this results in FitCover and for SVG it is FitFill.
	FitAuto ObjectFit = 0
	// FitFill fills the entire image view by stretching the pixel buffer.
	FitFill ObjectFit = 1
	// FitCover fills the entire image view, possibly zooming the image so much, that the resolution would
	// not be enough for a sharp image or to understand the image geometrics at all.
	FitCover ObjectFit = 2
	// FitContain is the inverse of FitCover which means still respecting the aspect ratio but scaling the
	// image down so that it can be seen entirely and the image view adds transparent borders (letterboxing).
	FitContain ObjectFit = 3
	// FitNone maps the pixel buffer 1:1 into its render context.
	FitNone ObjectFit = 4
)

type TImage struct {
	lightUri, darkUri  proto.URI
	accessibilityLabel string
	invisible          bool
	border             proto.Border
	frame              Frame
	padding            proto.Padding
	svg                proto.SVG
	fillColor          proto.Color
	strokeColor        proto.Color
	objectFit          ObjectFit
	light, dark        []byte
}

func Image() TImage {
	return TImage{
		frame: Frame{}.Size(Auto, L160),
	}
}

// ImageIcon renders default with L24/L24 size and is invisible if svg is empty.
func ImageIcon(svg core.SVG) TImage {
	if svg.Empty() {
		return TImage{invisible: true}
	}

	return Image().Embed(svg).Frame(Frame{}.Size(L24, L24)).(TImage)
}

// URI can be used for static image resources which are not provided by
// the ui component itself. The source may be a hand written REST endpoint
// or even any third-party resource from a different domain.
// If you need optimized data access and caching policies, you have to use this
// way.
// See also [core.Window.AsURI] for an uncached dynamically delivered image resource.
func (c TImage) URI(uri core.URI) TImage {
	c.lightUri = proto.URI(uri)
	return c
}

// URIAdaptive is like [TImage.Embed] but picks whatever fits best.
func (c TImage) URIAdaptive(light, dark core.URI) TImage {
	c.lightUri = proto.URI(light)
	c.darkUri = proto.URI(dark)
	return c
}

func (c TImage) Adaptive(light, dark image.ID) TImage {
	return c.URIAdaptive(
		httpimage.URI(light, image.FitNone, 512, 512),
		httpimage.URI(dark, image.FitNone, 512, 512),
	)
}

// FillColor set the internal fill color value and is only applicable for embedded SVG images, which use fill=currentColor.
// Otherwise, a broken or no effect will be seen.
func (c TImage) FillColor(color Color) TImage {
	c.fillColor = color.ora()
	return c
}

// StrokeColor set the internal stroke color value and is only applicable for embedded SVG images, which use fill=strokeColor.
// Otherwise, a broken or no effect will be seen.
func (c TImage) StrokeColor(color Color) TImage {
	c.strokeColor = color.ora()
	return c
}

// Embed encodes the given buffer within the components attributes. This may be fine to
// load small images synchronously, but it may break the channel, the server or the frontend, if too large.
// Better use [application.Resource] for large static images. Embedding image data in the range of 100-200 byte
// is totally fine, though. The resource URI alone is already about 100 characters long.
// Usually, embedding SVGs in the range of 1-2KiB is also fine. To optimize render performance,
// the system uses a special caching technique. Important: due to caching, do not submit ever-changing SVGs, because
// the backend and the frontend may suffer from cache overflow. This will half the typical required bandwidth
// for icon heavy use cases. The larger the SVG, the better the effect.
// See also [TImage.EmbedAdaptive].
func (c TImage) Embed(buf []byte) TImage {
	c.light = buf
	return c
}

// EmbedAdaptive is like [TImage.Embed] but picks whatever fits best.
func (c TImage) EmbedAdaptive(light, dark []byte) TImage {
	c.light = light
	c.dark = dark
	return c
}

// AccessibilityLabel sets a label for screen readers. See also https://www.w3.org/WAI/tutorials/images/decision-tree/.
func (c TImage) AccessibilityLabel(label string) DecoredView {
	c.accessibilityLabel = label
	return c
}

func (c TImage) Visible(b bool) DecoredView {
	c.invisible = b
	return c
}

func (c TImage) Padding(padding Padding) DecoredView {
	c.padding = padding.ora()
	return c
}

func (c TImage) Border(border Border) DecoredView {
	c.border = border.ora()
	return c
}

func (c TImage) Frame(frame Frame) DecoredView {
	c.frame = frame
	return c
}

func (c TImage) WithFrame(fn func(Frame) Frame) DecoredView {
	c.frame = fn(c.frame)
	return c
}

func (c TImage) ObjectFit(fit ObjectFit) TImage {
	c.objectFit = fit
	return c
}

func (c TImage) Render(ctx core.RenderContext) core.RenderNode {
	// start of delayed encoding
	if c.light != nil || c.dark != nil {
		var buf []byte
		if ctx.Window().Info().ColorScheme == core.Dark {
			buf = c.dark
		} else {
			buf = c.light
		}

		if buf == nil {
			if c.light != nil {
				buf = c.light
			} else {
				buf = c.dark
			}
		}

		isSvg := bytes.Contains(buf[:min(len(buf), 200)], []byte("<svg"))
		if isSvg {
			c.svg = buf
			c.lightUri = ""
		} else {
			b64 := base64.StdEncoding.EncodeToString(buf)
			c.lightUri = proto.URI(`data:application/octet-stream;base64,` + b64)
		}

	}
	// end of delayed encoding

	svgData := c.svg

	myUri := c.lightUri
	if c.lightUri != "" && c.darkUri != "" {
		if ctx.Window().Info().ColorScheme == core.Dark {
			myUri = c.darkUri
		} else {
			myUri = c.lightUri
		}
	}

	return &proto.Img{
		Uri:                myUri,
		AccessibilityLabel: proto.Str(c.accessibilityLabel),
		Invisible:          proto.Bool(c.invisible),
		Border:             c.border,
		Frame:              c.frame.ora(),
		Padding:            c.padding,
		SVG:                svgData,
		FillColor:          c.fillColor,
		StrokeColor:        c.strokeColor,
		ObjectFit:          proto.ObjectFit(c.objectFit),
	}
}

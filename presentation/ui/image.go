package ui

import (
	"bytes"
	"encoding/base64"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

// we have observed several malfunctions during development, so it is not clear if this is due to the dev server.
const svgFrontendCacheSupport = false

type TImage struct {
	uri                ora.URI
	accessibilityLabel string
	invisible          bool
	border             ora.Border
	frame              ora.Frame
	padding            ora.Padding
	svg                ora.SVG
	fillColor          ora.Color
	strokeColor        ora.Color
}

func Image() TImage {
	return TImage{}
}

// URI can be used for static image resources which are not provided by
// the ui component itself. The source may be a hand written REST endpoint
// or even any third-party resource from a different domain.
// If you need optimized data access and caching policies, you have to use this
// way.
// See also [core.Window.AsURI] for an uncached dynamically delivered image resource.
func (c TImage) URI(uri core.URI) TImage {
	c.uri = ora.URI(uri)
	c.frame = Frame{}.Size(Auto, L160).ora()
	return c
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
func (c TImage) Embed(buf []byte) TImage {
	isSvg := bytes.Contains(buf[:min(len(buf), 100)], []byte("<svg"))
	if isSvg {
		c.svg = buf
		c.uri = ""
		return c
	}

	b64 := base64.StdEncoding.EncodeToString(buf)
	//c.uri = ora.URI(`data:image/svg+xml;base64,` + b64)
	c.uri = ora.URI(`data:application/octet-stream;base64,` + b64)
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
	c.frame = frame.ora()
	return c
}

func (c TImage) Render(ctx core.RenderContext) ora.Component {
	svgData := c.svg
	var cachePointer ora.Ptr
	if svgFrontendCacheSupport {
		ptr, created := ctx.Handle(c.svg)
		if ptr != 0 && !created {
			// if ptr is not nil and it has already been created, we can omit the data
			// because the client already knows how the data looks for the handle pointer.
			svgData = nil
			cachePointer = ptr
		}
	}

	return ora.Image{
		Type:               ora.ImageT,
		URI:                c.uri,
		AccessibilityLabel: c.accessibilityLabel,
		Invisible:          c.invisible,
		Border:             c.border,
		Frame:              c.frame,
		Padding:            c.padding,
		SVG:                svgData,
		CachedSVG:          cachePointer,
		FillColor:          c.fillColor,
		StrokeColor:        c.strokeColor,
	}
}

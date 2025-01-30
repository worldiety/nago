package ui

import (
	"bytes"
	"encoding/base64"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

type TImage struct {
	uri                proto.URI
	accessibilityLabel string
	invisible          bool
	border             proto.Border
	frame              proto.Frame
	padding            proto.Padding
	svg                proto.SVG
	fillColor          proto.Color
	strokeColor        proto.Color
	light, dark        []byte
}

func Image() TImage {
	return TImage{
		frame: Frame{}.Size(Auto, L160).ora(),
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
	c.uri = proto.URI(uri)
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
	c.frame = frame.ora()
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
			c.uri = ""
		}

		b64 := base64.StdEncoding.EncodeToString(buf)
		//c.uri = proto.URI(`data:image/svg+xml;base64,` + b64)
		c.uri = proto.URI(`data:application/octet-stream;base64,` + b64)
	}
	// end of delayed encoding

	svgData := c.svg

	return &proto.Img{
		Uri:                c.uri,
		AccessibilityLabel: proto.Str(c.accessibilityLabel),
		Invisible:          proto.Bool(c.invisible),
		Border:             c.border,
		Frame:              c.frame,
		Padding:            c.padding,
		SVG:                svgData,
		FillColor:          c.fillColor,
		StrokeColor:        c.strokeColor,
	}
}

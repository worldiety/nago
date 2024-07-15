package ui

import (
	"bytes"
	"encoding/base64"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type ImageView struct {
	uri                ora.URI
	accessibilityLabel string
	invisible          bool
	with               func(*ImageView)
	border             ora.Border
	frame              ora.Frame
	padding            ora.Padding
}

func Image(with func(img *ImageView)) *ImageView {
	return &ImageView{
		with: with,
	}
}

// URI can be used for static image resources which are not provided by
// the ui component itself. The source may be a hand written REST endpoint
// or even any third-party resource from a different domain.
// If you need optimized data access and caching policies, you have to use this
// way.
// See also [core.Window.AsURI] for an uncached dynamically delivered image resource.
func (c *ImageView) URI(uri ora.URI) {
	c.uri = uri
}

// Embed encodes the given buffer as a URI within the components attributes. This may be fine to
// load small images synchronously, but it may break the channel, the server or the frontend, if too large.
// Better use [application.Resource] for large static images. Embedding image data in the range of 100-200 byte
// is totally fine, though. The resource URI alone is already about 100 characters long.
func (c *ImageView) Embed(buf []byte) {
	b64 := base64.StdEncoding.EncodeToString(buf)
	if bytes.Contains(buf[:100], []byte("<svg")) {
		c.uri = ora.URI(`data:image/svg+xml;base64,` + b64)
	} else {
		c.uri = ora.URI(`data:application/octet-stream;base64,` + b64)
	}
}

// AccessibilityLabel sets a label for screen readers. See also https://www.w3.org/WAI/tutorials/images/decision-tree/.
func (c *ImageView) AccessibilityLabel(label string) {
	c.accessibilityLabel = label
}

func (c *ImageView) Visible(b bool) {
	c.invisible = b
}

func (c *ImageView) Padding() ora.Padding {
	return c.padding
}

func (c *ImageView) SetPadding(padding ora.Padding) {
	c.padding = padding
}

func (c *ImageView) Border(border ora.Border) {
	c.border = border
}

func (c *ImageView) Frame() ora.Frame {
	return c.frame
}

func (c *ImageView) SetFrame(frame ora.Frame) {
	c.frame = frame
}

func (c *ImageView) Render(ctx core.RenderContext) ora.Component {
	if c.with != nil {
		c.with(c)
	}

	return ora.Image{
		Type:               ora.ImageT,
		URI:                c.uri,
		AccessibilityLabel: c.accessibilityLabel,
		Invisible:          c.invisible,
		Border:             c.border,
		Frame:              c.frame,
		Padding:            c.padding,
	}
}

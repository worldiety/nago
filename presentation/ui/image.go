package ui

import (
	"encoding/base64"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"io"
)

type Image struct {
	id         ora.Ptr
	uri        *Shared[ora.URI]
	source     func() (io.Reader, error)
	caption    String
	properties []core.Property
}

func NewImage(with func(img *Image)) *Image {
	c := &Image{
		id:      nextPtr(),
		uri:     NewShared[ora.URI]("url"),
		caption: NewShared[string]("caption"),
	}

	c.properties = []core.Property{c.uri, c.caption}

	if with != nil {
		with(c)
	}

	return c
}

func (c *Image) ID() ora.Ptr {
	return c.id
}

// URI can be used for static image resources which are not provided by
// the ui component itself. The source may be a hand written REST endpoint
// or even any third-party resource from a different domain.
// If you need optimized data access and caching policies, you have to use this
// way.
// See also [core.Window.AsURI] for an uncached dynamically delivered image resource.
func (c *Image) URI() *Shared[ora.URI] {
	return c.uri
}

// SetDataURI encodes the given buffer as an URI with the embedded data image.
func (c *Image) SetDataURI(buf []byte) {
	b64 := base64.StdEncoding.EncodeToString(buf)
	c.URI().Set(ora.URI(`data:application/octet-stream;base64,` + b64))
}

func (c *Image) Caption() String {
	return c.caption
}

func (c *Image) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *Image) Render() ora.Component {
	return c.render()
}

func (c *Image) render() ora.Image {
	return ora.Image{
		Ptr:     c.id,
		Type:    ora.ImageT,
		URI:     c.uri.render(),
		Caption: c.caption.render(),
	}
}

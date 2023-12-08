package ui

import (
	"go.wdy.de/nago/container/slice"
	"io"
)

type Image struct {
	id            CID
	url           String
	source        func() (io.Reader, error)
	downloadToken String
	caption       String
	properties    slice.Slice[Property]
}

func NewImage(with func(img *Image)) *Image {
	c := &Image{
		id:            nextPtr(),
		url:           NewShared[string]("url"),
		downloadToken: NewShared[string]("downloadToken"),
		caption:       NewShared[string]("caption"),
	}

	c.downloadToken.Set(nextToken())
	c.properties = slice.Of[Property](c.url, c.downloadToken, c.caption)

	if with != nil {
		with(c)
	}

	return c
}

func (c *Image) ID() CID {
	return c.id
}

func (c *Image) Type() string {
	return "Image"
}

func (c *Image) Properties() slice.Slice[Property] {
	return c.properties
}

// Source sets a dynamic stream provider as a data source for this image.
// Note, that the callback is always called from a different thread
// to allows concurrent image loading. Internally, each page and component instance
// gets a new unique token, so that any http caching becomes useless.
// TODO: we could create an api which just configures a pipeline to the repo, so that this can be called safely from the websocket event handler, but it makes dynamic creations in the domain harder
func (c *Image) Source(open func() (io.Reader, error)) {
	c.source = open
	if open == nil {
		c.url.Set("")
		return
	}

	c.url.Set("/api/v1/download")
}

// URL can be used for static image resources which are not provided by
// the ui component itself. The source may be a hand written REST endpoint
// or even any third-party resource from a different domain.
// If you need optimized data access and caching policies, you have to use this
// way.
// See also [Image.Source] for an uncached dynamically delivered image.
func (c *Image) URL() String {
	return c.url
}

func (c *Image) Caption() String {
	return c.caption
}

func (c *Image) DownloadSource() func() (io.Reader, error) {
	return c.source
}

func (c *Image) DownloadToken() DownloadToken {
	return DownloadToken(c.downloadToken.Get())
}

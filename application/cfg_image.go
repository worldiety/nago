package application

import (
	"go.wdy.de/nago/image"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data/json"
)

type Images struct {
	SrcSets      image.Repository
	Blobs        blob.Store
	LoadBestFit  image.LoadBestFit
	CreateSrcSet image.CreateSrcSet
	LoadSrcSet   image.LoadSrcSet
}

// Images returns the default Images backend, including use case implementations, repositories and http endpoints.
// The default
func (c *Configurator) Images() Images {
	if c.images == nil {
		imageSrcSetRepo := json.NewSloppyJSONRepository[image.SrcSet](c.EntityStore(".nago.img.set"))
		imageBlobs := c.FileStore(".nago.img.blob")
		imgBestFit := image.NewLoadBestFit(imageSrcSetRepo, imageBlobs)
		imgCreateSrcSet := image.NewCreateSrcSet(image.Options{}, imageSrcSetRepo, imageBlobs)
		loadSrcSet := image.NewLoadSrcSet(imageSrcSetRepo)

		c.images = &Images{
			SrcSets:      imageSrcSetRepo,
			Blobs:        imageBlobs,
			LoadBestFit:  imgBestFit,
			CreateSrcSet: imgCreateSrcSet,
			LoadSrcSet:   loadSrcSet,
		}
	}

	return *c.images
}

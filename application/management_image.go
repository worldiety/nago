package application

import (
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/std"
)

type ImageManagement struct {
	repo     image.Repository
	blobs    blob.Store
	UseCases image.UseCases
}

// ImageManagement returns the default Images backend, including use case implementations, repositories and http endpoints.
func (c *Configurator) ImageManagement() (ImageManagement, error) {
	if c.imageManagement == nil {
		imageSrcSetRepo := json.NewSloppyJSONRepository[image.SrcSet](std.Must(c.EntityStore(".nago.img.set")))
		imageBlobs := std.Must(c.FileStore(".nago.img.blob"))

		c.imageManagement = &ImageManagement{
			repo:     imageSrcSetRepo,
			blobs:    imageBlobs,
			UseCases: image.NewUseCases(imageSrcSetRepo, imageBlobs),
		}
	}

	return *c.imageManagement, nil
}

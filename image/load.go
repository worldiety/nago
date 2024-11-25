package image

import (
	"context"
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/std"
	"io"
)

// LoadBestFit calculates which available image resolution fits best and returns it.
// TODO it is unclear how to handle authenticated, authorized and public use cases over the same API and endpoint
type LoadBestFit func(user auth.Subject, img ID, fit ObjectFit, width, height int) (std.Option[io.ReadCloser], error)

func NewLoadBestFit(repo Repository, imageBlobs blob.Store) LoadBestFit {
	return func(user auth.Subject, imgID ID, fit ObjectFit, width, height int) (std.Option[io.ReadCloser], error) {
		// catch the non-sense
		if width < 32 {
			width = 32
		}

		if height < 32 {
			height = 32
		}

		// first, just try to find the source set
		optSrcSet, err := repo.FindByID(imgID)
		if err != nil {
			return std.None[io.ReadCloser](), fmt.Errorf("error on finding src set from repo: %w", err)
		}

		srcSet := optSrcSet.UnwrapOr(SrcSet{})
		var imgBlobToLoad string
		img, ok := srcSet.Fit(fit, width, height)
		if ok {
			imgBlobToLoad = string(img.Data)
		} else {
			// oops, try image just directly from blob store
			imgBlobToLoad = string(imgID)
		}

		optRaw, err := imageBlobs.NewReader(context.Background(), imgBlobToLoad)
		if err != nil {
			return std.None[io.ReadCloser](), fmt.Errorf("error on loading image blob '%s': %w", imgBlobToLoad, err)
		}

		return optRaw, nil
	}
}

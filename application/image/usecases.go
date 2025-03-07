package image

import "go.wdy.de/nago/pkg/blob"

type UseCases struct {
	LoadBestFit  LoadBestFit
	CreateSrcSet CreateSrcSet
	LoadSrcSet   LoadSrcSet
}

func NewUseCases(imageSrcSetRepo Repository, imageBlobs blob.Store) UseCases {

	imgBestFit := NewLoadBestFit(imageSrcSetRepo, imageBlobs)
	imgCreateSrcSet := NewCreateSrcSet(Options{}, imageSrcSetRepo, imageBlobs)
	loadSrcSet := NewLoadSrcSet(imageSrcSetRepo)

	return UseCases{
		LoadBestFit:  imgBestFit,
		CreateSrcSet: imgCreateSrcSet,
		LoadSrcSet:   loadSrcSet,
	}
}

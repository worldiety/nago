// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

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

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package image

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/blob"
	"io"
)

type OpenReader func(user permission.Auditable, id ID) (option.Opt[io.ReadCloser], error)
type UseCases struct {
	LoadBestFit  LoadBestFit
	CreateSrcSet CreateSrcSet
	LoadSrcSet   LoadSrcSet
	OpenReader   OpenReader
}

func NewUseCases(imageSrcSetRepo Repository, imageBlobs blob.Store) UseCases {

	imgBestFit := NewLoadBestFit(imageSrcSetRepo, imageBlobs)
	imgCreateSrcSet := NewCreateSrcSet(Options{}, imageSrcSetRepo, imageBlobs)
	loadSrcSet := NewLoadSrcSet(imageSrcSetRepo)

	return UseCases{
		LoadBestFit:  imgBestFit,
		CreateSrcSet: imgCreateSrcSet,
		LoadSrcSet:   loadSrcSet,
		OpenReader:   NewOpenReader(imageBlobs),
	}
}

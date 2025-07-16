// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"fmt"
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data/json"
)

type ImageManagement struct {
	repo     image.Repository
	blobs    blob.Store
	UseCases image.UseCases
}

// ImageManagement returns the default Images backend, including use case implementations, repositories and http endpoints.
func (c *Configurator) ImageManagement() (ImageManagement, error) {
	if c.imageManagement == nil {
		stores, err := c.Stores()
		if err != nil {
			return ImageManagement{}, err
		}

		optSetStore, err := stores.Get(".nago.img.set") // we cannot migrate store names, thus keep the old for existing data
		if err != nil {
			return ImageManagement{}, fmt.Errorf("could not get nago img set store: %w", err)
		}

		var setStore blob.Store
		if optSetStore.IsSome() {
			setStore = optSetStore.Unwrap()
		} else {
			store, err := stores.Open("nago.img.set", blob.OpenStoreOptions{Type: blob.EntityStore})
			if err != nil {
				return ImageManagement{}, fmt.Errorf("could not open nago img set store: %w", err)
			}

			setStore = store
		}

		imageSrcSetRepo := json.NewSloppyJSONRepository[image.SrcSet](setStore) // we cannot migrate store names, thus keep the old for existing data

		optblobStore, err := stores.Get(".nago.img.blob")
		if err != nil {
			return ImageManagement{}, fmt.Errorf("could not get nago img blob store: %w", err)
		}

		var imageBlobs blob.Store
		if optblobStore.IsSome() {
			imageBlobs = optblobStore.Unwrap()
		} else {
			store, err := stores.Open("nago.img.blob", blob.OpenStoreOptions{Type: blob.FileStore})
			if err != nil {
				return ImageManagement{}, fmt.Errorf("could not open nago img blob store: %w", err)
			}

			imageBlobs = store
		}

		c.imageManagement = &ImageManagement{
			repo:     imageSrcSetRepo,
			blobs:    imageBlobs,
			UseCases: image.NewUseCases(imageSrcSetRepo, imageBlobs),
		}
	}

	return *c.imageManagement, nil
}

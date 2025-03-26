// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package backup

import (
	"time"
)

type Index struct {
	CreatedAt time.Time `json:"createdAt"`
	Stores    []Store   `json:"stores,omitempty"`
}

func (i *Index) Size() int64 {
	count := int64(0)
	for _, store := range i.Stores {
		for _, blob := range store.Blobs {
			count += blob.Size
		}
	}

	return count
}

func (i *Index) Count() int64 {
	count := int64(0)
	for _, store := range i.Stores {
		for range store.Blobs {
			count++
		}
	}

	return count
}

type Stereotype string

const (
	StereotypeDocument Stereotype = "document"
	StereotypeBlob     Stereotype = "blob"
)

type Store struct {
	Name       string     `json:"name"`
	Stereotype Stereotype `json:"stereotype"`
	Blobs      []Blob     `json:"blobs"`
}

type Blob struct {
	ID     string `json:"id"`
	Size   int64  `json:"size"`
	Sha256 string `json:"sha256"`
	Path   string `json:"path"`
	/*LastMod   time.Time   `json:"lastMod,omitempty"`
	CreatedAt time.Time   `json:"createdAt,omitempty"`
	Mode      fs.FileMode `json:"mode,omitempty"`*/
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package signature

import (
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/pkg/blob"
	"sync"
)

func NewSignUnqualified(mutex *sync.Mutex, openImgReader image.OpenReader, repo Repository, stores blob.Stores, idx *inMemoryIndex) SignUnqualified {
	return func(signData AnonSignData) (ID, error) {
		return signUnqualified(mutex, openImgReader, repo, stores, idx, "", signData)
	}
}

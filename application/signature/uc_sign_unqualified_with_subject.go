// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package signature

import (
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob"
	"sync"
)

func NewSignUnqualifiedWithSubject(mutex *sync.Mutex, openImgReader image.OpenReader, repo Repository, stores blob.Stores, idx *inMemoryIndex) SignUnqualifiedWithSubject {
	return func(subject user.Subject, cdata SignData) (ID, error) {
		if !subject.Valid() {
			return "", user.InvalidSubjectErr
		}

		return signUnqualified(mutex, openImgReader, repo, stores, idx, subject.ID(), AnonSignData{
			Firstname: subject.Firstname(),
			Lastname:  subject.Lastname(),
			Email:     user.Email(subject.Email()),
			SignData:  cdata,
		})
	}
}

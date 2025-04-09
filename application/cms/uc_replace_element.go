// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cms

import (
	"go.wdy.de/nago/auth"
	"os"
	"sync"
)

func NewReplaceElement(mutex *sync.Mutex, repo Repository) ReplaceElement {
	return func(subject auth.Subject, id ID, elem Element) error {
		if err := subject.AuditResource(repo.Name(), string(id), PermUpdateElement); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		optDoc, err := repo.FindByID(id)
		if err != nil {
			return err
		}

		if optDoc.IsNone() {
			return os.ErrNotExist
		}

		doc := optDoc.Unwrap().IntoModel()
		_, ok := doc.ElementByID(elem.Identity())
		if !ok {
			return os.ErrNotExist
		}

		doc.Replace(elem)

		return repo.Save(doc.IntoPersistence())
	}
}

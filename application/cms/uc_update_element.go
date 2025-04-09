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

func NewUpdateElement(mutex *sync.Mutex, repo Repository) UpdateElement {
	return func(subject auth.Subject, id ID, eid EID, mutator func(elem Element) Element) error {
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
		elem, ok := doc.ElementByID(eid)
		if !ok {
			return os.ErrNotExist
		}

		elem = mutator(elem)
		doc.Replace(elem)

		return repo.Save(doc.IntoPersistence())
	}
}

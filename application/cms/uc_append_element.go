// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cms

import (
	"os"
	"sync"

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

func NewAppendElement(mutex *sync.Mutex, repo Repository) AppendElement {
	return func(subject auth.Subject, id ID, parent EID, elem Element) error {
		if err := subject.AuditResource(rebac.Namespace(repo.Name()), rebac.Instance(id), PermAppendElement); err != nil {
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
		myParent, ok := doc.ElementByID(parent)
		if !ok {
			return os.ErrNotExist
		}

		elem.SetIdentity(data.RandIdent[EID]())
		myParent.Append(elem)

		return repo.Save(doc.IntoPersistence())
	}
}

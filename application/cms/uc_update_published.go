// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cms

import (
	"fmt"
	"sync"

	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
)

func NewUpdatePublished(mutex *sync.Mutex, repo Repository) UpdatePublished {
	return func(subject auth.Subject, id ID, published bool) error {
		if err := subject.AuditResource(rebac.Namespace(repo.Name()), rebac.Instance(id), PermUpdatePublished); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		optDoc, err := repo.FindByID(id)
		if err != nil {
			return err
		}

		if optDoc.IsNone() {
			return fmt.Errorf("cannot change slug: not found: %s", id)
		}

		doc := optDoc.Unwrap()
		if doc.Published == published {
			// nothing changed
			return nil
		}

		doc.Published = published
		return repo.Save(doc)
	}
}

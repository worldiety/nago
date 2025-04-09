// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cms

import (
	"go.wdy.de/nago/auth"
	"golang.org/x/text/language"
	"os"
	"sync"
)

func NewUpdateTitle(mutex *sync.Mutex, repo Repository) UpdateTitle {
	return func(subject auth.Subject, id ID, lang language.Tag, title string) error {
		if err := subject.AuditResource(repo.Name(), string(id), PermUpdateTitle); err != nil {
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

		doc := optDoc.Unwrap()
		if doc.Title == nil {
			doc.Title[lang] = title
		} else {
			if doc.Title[lang] == title {
				// nothing changed
				return nil
			}
		}

		doc.Title[lang] = title

		return repo.Save(doc)
	}
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cms

import (
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/std/concurrent"
	"sync"
)

func NewUpdateSlug(mutex *sync.Mutex, slugs *concurrent.RWMap[Slug, ID], repo Repository) UpdateSlug {
	return func(subject auth.Subject, id ID, slug Slug) error {
		if err := subject.AuditResource(repo.Name(), string(id), PermUpdateSlug); err != nil {
			return err
		}

		if err := slug.Validate(); err != nil {
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
		if doc.Slug == slug {
			// nothing changed
			return nil
		}

		if other, ok := slugs.Get(slug); ok {
			return std.NewLocalizedError("Slug nicht angepasst", fmt.Sprintf("Der Slug %s ist bereits in Verwendung von Seite %s", slug, other))
		}

		oldSlug := doc.Slug
		doc.Slug = slug
		if err := repo.Save(doc); err != nil {
			return err
		}

		slugs.Delete(oldSlug)
		slugs.Put(slug, doc.ID)

		return nil
	}
}

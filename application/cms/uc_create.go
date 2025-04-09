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
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std/concurrent"
	"golang.org/x/text/language"
	"sync"
	"time"
)

func NewCreate(mutex *sync.Mutex, slugs *concurrent.RWMap[Slug, ID], repo Repository) Create {
	return func(subject auth.Subject, d CreationData) (ID, error) {
		if err := subject.Audit(PermCreate); err != nil {
			return "", err
		}

		mutex.Lock()
		defer mutex.Unlock()

		if id, ok := slugs.Get(d.Slug); ok {
			return "", fmt.Errorf("slug already exists: %s.%s", id, d.Slug)
		}

		id := data.RandIdent[ID]()
		optDoc, err := repo.FindByID(id)
		if err != nil {
			return "", err
		}

		if optDoc.IsSome() {
			return "", fmt.Errorf("doc already exists: %s", id)
		}

		err = repo.Save(PDocument{
			ID:          id,
			Slug:        d.Slug,
			LastUpdated: time.Now(),
			Title:       LocStr{language.Und: d.Title},
			Published:   d.Published,
			Body: &PVStack{
				ID: data.RandIdent[EID](),
			},
		})

		if err != nil {
			return "", err
		}

		slugs.Put(d.Slug, id)

		return id, nil
	}
}

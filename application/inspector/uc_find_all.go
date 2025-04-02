// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package inspector

import (
	"go.wdy.de/nago/application/backup"
	"go.wdy.de/nago/auth"
	"slices"
	"strings"
)

func NewFindAll(p backup.Persistence) FindAll {
	return func(subject auth.Subject) ([]Store, error) {
		if err := subject.Audit(PermDataInspector); err != nil {
			return nil, err
		}

		var res []Store

		// file stores
		for storeName, err := range p.FileStores() {
			r := Store{
				Name:       storeName,
				Stereotype: backup.StereotypeBlob,
			}
			if err != nil {
				r.Error = err
				res = append(res, r)
				continue
			}

			store, err := p.FileStore(storeName)
			if err != nil {
				r.Error = err
				res = append(res, r)
				continue
			}

			r.Store = store

			res = append(res, r)
		}

		// entity stores
		for storeName, err := range p.EntityStores() {
			r := Store{
				Name:       storeName,
				Stereotype: backup.StereotypeDocument,
			}
			if err != nil {
				r.Error = err
				res = append(res, r)
				continue
			}

			store, err := p.EntityStore(storeName)
			if err != nil {
				r.Error = err
				res = append(res, r)
				continue
			}

			r.Store = store

			res = append(res, r)
		}

		slices.SortFunc(res, func(a, b Store) int {
			return strings.Compare(a.Name, b.Name)
		})

		return res, nil
	}
}

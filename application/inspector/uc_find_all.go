// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package inspector

import (
	"fmt"
	"go.wdy.de/nago/application/backup"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"os"
	"slices"
	"strings"
)

func NewFindAll(p blob.Stores) FindAll {
	return func(subject auth.Subject) ([]Store, error) {
		if err := subject.Audit(PermDataInspector); err != nil {
			return nil, err
		}

		var res []Store

		for storeName, err := range p.All() {
			r := Store{
				Name: storeName,
			}

			if err != nil {
				r.Error = err
				res = append(res, r)
				continue
			}

			optStat, err := p.Stat(storeName)
			if err != nil {
				r.Error = err
				res = append(res, r)
				continue
			}

			if optStat.IsNone() {
				r.Error = os.ErrNotExist
				res = append(res, r)
				continue
			}

			stat := optStat.Unwrap()
			switch stat.Type {
			case blob.FileStore:
				r.Stereotype = backup.StereotypeBlob
			case blob.EntityStore:
				r.Stereotype = backup.StereotypeDocument
			default:
				r.Error = fmt.Errorf("unknown blob store type: %v", stat.Type)
				res = append(res, r)
				continue
			}

			store, err := p.Open(storeName, blob.OpenStoreOptions{Type: stat.Type})
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

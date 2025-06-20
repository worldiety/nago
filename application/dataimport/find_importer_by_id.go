// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataimport

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/dataimport/importer"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std/concurrent"
)

func NewFindImporterByID(m *concurrent.RWMap[importer.ID, importer.Importer]) FindImporterByID {
	return func(subject auth.Subject, id importer.ID) (option.Opt[importer.Importer], error) {
		if err := subject.Audit(PermFindImporters); err != nil {
			return option.Opt[importer.Importer]{}, err
		}

		i, ok := m.Get(id)
		if !ok {
			return option.None[importer.Importer](), nil
		}

		return option.Some(i), nil
	}
}

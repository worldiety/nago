// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cms

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std/concurrent"
	"os"
)

func NewFindBySlug(slugs *concurrent.RWMap[Slug, ID], repo Repository) FindBySlug {
	return func(subject auth.Subject, slug Slug) (option.Opt[*Document], error) {
		id, ok := slugs.Get(slug)
		if !ok {
			return option.None[*Document](), nil
		}

		optDoc, err := repo.FindByID(id)
		if err != nil {
			return option.None[*Document](), err
		}

		if optDoc.IsNone() {
			return option.None[*Document](), os.ErrNotExist
		}

		if err := subject.AuditResource(repo.Name(), string(optDoc.Unwrap().ID), PermFindAll); err != nil {
			return option.None[*Document](), err
		}

		doc := optDoc.Unwrap().IntoModel()

		return option.Some(doc), nil

	}
}

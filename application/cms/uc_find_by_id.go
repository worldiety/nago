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
	"os"
)

func NewFindByID(repo Repository) FindByID {
	return func(subject auth.Subject, id ID) (option.Opt[*Document], error) {
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

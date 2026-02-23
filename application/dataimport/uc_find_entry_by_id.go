// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataimport

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
)

func NewFindEntryByID(repoEntry EntryRepository) FindEntryByID {
	return func(subject auth.Subject, id Key) (option.Opt[Entry], error) {
		if err := subject.AuditResource(rebac.Namespace(repoEntry.Name()), rebac.Instance(id), PermFindEntryByID); err != nil {
			return option.Opt[Entry]{}, err
		}

		return repoEntry.FindByID(id)
	}
}

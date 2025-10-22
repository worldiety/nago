// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package conversation

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/auth"
)

func NewFindByID(repo Repository) FindByID {
	return func(subject auth.Subject, id ID) (option.Opt[Conversation], error) {
		optConv, err := repo.FindByID(id)
		if err != nil {
			return option.None[Conversation](), err
		}

		if optConv.IsNone() {
			return option.None[Conversation](), nil
		}

		conversation := optConv.Unwrap()

		if subject.ID() != conversation.CreatedBy && !subject.HasResourcePermission(repo.Name(), string(conversation.ID), PermFindAll) {
			return option.None[Conversation](), subject.Audit(PermFindAll)
		}

		return optConv, nil
	}
}

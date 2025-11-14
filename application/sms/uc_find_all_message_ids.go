// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package sms

import (
	"iter"
	"slices"

	"go.wdy.de/nago/application/sms/message"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xslices"
)

func NewFindAllMessageIDs(repo message.Repository) FindAllMessageIDs {
	return func(subject auth.Subject) iter.Seq2[message.ID, error] {
		return func(yield func(message.ID, error) bool) {
			if !subject.HasPermission(PermFindAllMessageIDs) {
				return
			}

			idents, err := xslices.Collect2(repo.Identifiers())
			if err != nil {
				yield("", err)
				return
			}

			slices.Reverse(idents)
			for _, id := range idents {
				if !yield(id, nil) {
					return
				}
			}
		}
	}
}

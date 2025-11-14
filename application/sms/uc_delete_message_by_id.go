// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package sms

import (
	"go.wdy.de/nago/application/sms/message"
	"go.wdy.de/nago/auth"
)

func NewDeleteMessageByID(repo message.Repository) DeleteMessageByID {
	return func(subject auth.Subject, id message.ID) error {
		if err := subject.Audit(PermDeleteMessageByID); err != nil {
			return err
		}

		return repo.DeleteByID(id)
	}
}

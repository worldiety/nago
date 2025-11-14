// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package sms

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/sms/message"
	"go.wdy.de/nago/auth"
)

func NewFindMessageByID(repo message.Repository) FindMessageByID {
	return func(subject auth.Subject, id message.ID) (option.Opt[message.SMS], error) {
		if err := subject.Audit(PermFindByID); err != nil {
			return option.Opt[message.SMS]{}, err
		}

		return repo.FindByID(id)
	}
}

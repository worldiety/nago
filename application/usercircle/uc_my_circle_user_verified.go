// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package usercircle

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"sync"
)

func NewMyCircleUserVerified(mutex *sync.Mutex, repo Repository, users user.UseCases) MyCircleUserVerified {
	return func(subject auth.Subject, circleId ID, usrId user.ID, verified bool) error {
		mutex.Lock()
		defer mutex.Unlock()

		_, usr, err := myCircleAndUser(repo, users.FindByID, subject, circleId, usrId)
		if err != nil {
			return err
		}

		return users.UpdateVerification(user.SU(), usr.ID, verified)
	}
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"fmt"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"time"
)

func NewLogin(sessions Repository, authenticate user.AuthenticateByPassword) Login {
	return func(id ID, login user.Email, password user.Password) (bool, error) {
		// first install the session
		optSession, err := sessions.FindByID(id)
		if err != nil {
			return false, fmt.Errorf("sessions.FindByID failed: %w", err)
		}

		var session Session
		if optSession.IsNone() {
			session.ID = id
			session.CreatedAt = time.Now()
		} else {
			session = optSession.Unwrap()
		}

		// try to authenticate
		optUsr, err := authenticate(login, password)
		if err != nil {
			return false, fmt.Errorf("auhentication failed: %w", err)
		}

		if optUsr.IsNone() {
			return false, nil
		}

		session.User = std.Some(optUsr.Unwrap().ID)
		session.AuthenticatedAt = time.Now()

		if err := sessions.Save(session); err != nil {
			return false, fmt.Errorf("sessions.Save failed: %w", err)
		}

		return true, nil
	}
}

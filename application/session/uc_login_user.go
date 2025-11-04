// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"fmt"
	"time"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std"
)

func NewLoginUser(bus events.Bus, sessions Repository) LoginUser {
	return func(id ID, usr user.ID) error {
		// first install the session
		optSession, err := sessions.FindByID(id)
		if err != nil {
			return fmt.Errorf("sessions.FindByID failed: %w", err)
		}

		var session Session
		if optSession.IsNone() {
			session.ID = id
			session.CreatedAt = time.Now()
		} else {
			session = optSession.Unwrap()
		}

		session.User = std.Some(usr)
		session.AuthenticatedAt = time.Now()

		if err := sessions.Save(session); err != nil {
			return fmt.Errorf("sessions.Save failed: %w", err)
		}

		bus.Publish(Authenticated{
			Session: id,
			User:    session.User.Unwrap(),
		})

		return nil
	}
}

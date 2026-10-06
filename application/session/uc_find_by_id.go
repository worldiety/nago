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
	"go.wdy.de/nago/logging"
	"go.wdy.de/nago/pkg/std"
	"log/slog"
	"time"
)

func NewFindByID(sessions Repository) FindByID {
	return func(id ID) (std.Option[Session], error) {
		optSession, err := sessions.FindByID(id)
		if err != nil {
			return std.None[Session](), fmt.Errorf("failed to find session: %w", err)
		}

		if optSession.IsNone() {
			return std.None[Session](), nil
		}

		session := optSession.Unwrap()

		if session.User.IsNone() {
			return std.None[Session](), nil
		}

		if expired(session, time.Now()) {
			slog.Error("session expired for user", "sessionID", logging.Secret(string(session.ID)), "user", session.User)
			session.User = std.None[user.ID]()
			session.AuthenticatedAt = time.Time{}
			if err := sessions.Save(session); err != nil {
				return std.None[Session](), fmt.Errorf("failed to save expired session: %w", err)
			}

			return std.None[Session](), nil
		}

		return std.Some(session), nil
	}
}

// sessionLifetime is the time after the authentication, after which a session loses its user.
// TODO make this configurable and perhaps add another deadline for short sessions?
const sessionLifetime = 90 * 24 * time.Hour

// expired reports whether the session has a user whose authentication is too old.
func expired(session Session, now time.Time) bool {
	return session.User.IsSome() && now.Sub(session.AuthenticatedAt) > sessionLifetime
}

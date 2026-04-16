// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package http

import (
	"log/slog"
	"net/http"

	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

type SubjectHandlerFunc func(w http.ResponseWriter, r *http.Request, subject auth.Subject)

// NewSubjectHandlerFunc creates a new SubjectHandlerFunc which always injects a subject. It is an invalid anon user or
// a valid user subject authenticated by its cookie. Eventually we will also support token based
// authentication.
func NewSubjectHandlerFunc(findSession session.FindUserSessionByID, subjectFromUser user.SubjectFromUser, newAnon user.GetAnonUser, fn SubjectHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := r.Cookie("wdy-ora-access")
		if cookie == nil {
			fn(w, r, newAnon())
			return
		}

		sessionID := session.ID(cookie.Value)
		if sessionID == "" {
			fn(w, r, newAnon())
			return
		}

		s := findSession(sessionID)
		optUsr := s.User()
		if optUsr.IsNone() {
			fn(w, r, newAnon())
			return
		}

		optSubject, err := subjectFromUser(nil, optUsr.Unwrap())
		if err != nil {
			slog.Error("failed to load subject for user", "userID", optUsr.Unwrap(), "error", err)
			http.Error(w, "failed to load subject for user", http.StatusInternalServerError)
			return
		}

		if optSubject.IsNone() {
			fn(w, r, newAnon())
			return
		}

		fn(w, r, optSubject.Unwrap())
	}
}

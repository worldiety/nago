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
//
// It prefers the session cookie of https, but falls back to the one of plain http, which another application of the
// same site may have set. Use [go.wdy.de/nago/application.Configurator.HandleFuncSubject] instead, which knows
// whether the request came through https.
func NewSubjectHandlerFunc(findSession session.FindUserSessionByID, subjectFromUser user.SubjectFromUser, newAnon user.GetAnonUser, fn SubjectHandlerFunc) http.HandlerFunc {
	return NewSessionSubjectHandlerFunc(func(r *http.Request) session.ID {
		for _, name := range []string{"__Host-wdy-ora-access", "wdy-ora-access"} {
			if cookie, err := r.Cookie(name); err == nil && cookie.Value != "" {
				return session.ID(cookie.Value)
			}
		}

		return ""
	}, findSession, subjectFromUser, newAnon, fn)
}

// NewSessionSubjectHandlerFunc is like [NewSubjectHandlerFunc], but takes the session id of the request from the
// given function. An empty id is anonymous.
func NewSessionSubjectHandlerFunc(sessionOf func(r *http.Request) session.ID, findSession session.FindUserSessionByID, subjectFromUser user.SubjectFromUser, newAnon user.GetAnonUser, fn SubjectHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := sessionOf(r)
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

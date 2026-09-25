// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"fmt"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xtime"
)

// NewResolve returns a [Resolve] use case.
//
// It runs under the session's keyed lock like [Append]. The lock is only held while the continued run is
// actually working; a run that suspends again is persisted and releases it, so no goroutine ever waits for a
// user.
func NewResolve(locks *locker, repo Repository) Resolve {
	return func(subject auth.Subject, id ID, opts ResolveOptions) (Session, error) {
		if opts.Run.Completions == nil {
			return Session{}, fmt.Errorf("session: ResolveOptions.Run.Completions must not be nil")
		}

		defer locks.lock(id)()

		session, err := loadPending(subject, repo, id, opts.Revision)
		if err != nil {
			return Session{}, err
		}

		if opts.Run.Model != "" && opts.Run.Model != session.Model {
			return Session{}, fmt.Errorf("session: a pending run cannot continue on another model (%s instead of %s)", opts.Run.Model, session.Model)
		}

		system := session.System
		if opts.Run.System != "" {
			system = opts.Run.System
		}

		base := completion.Options{
			Model:       session.Model,
			System:      system,
			Messages:    session.Messages,
			MaxTokens:   opts.Run.MaxTokens,
			Temperature: opts.Run.Temperature,
		}

		out, rerr := completion.Continue(subject, opts.Run.Completions, runOptions(base, opts.Run), *session.Pending, opts.Resolutions)
		if rerr != nil {
			// A mismatching decision changed nothing; everything else may already have executed an approved
			// call, which must be kept.
			return Session{}, persistFailure(repo, session, session.Model, out, rerr)
		}

		return saveOutcome(repo, session, session.Model, out)
	}
}

// NewDismiss returns a [Dismiss] use case.
func NewDismiss(locks *locker, repo Repository) Dismiss {
	return func(subject auth.Subject, id ID, revision int) (Session, error) {
		defer locks.lock(id)()

		session, err := loadPending(subject, repo, id, revision)
		if err != nil {
			return Session{}, err
		}

		history, err := completion.Dismiss(session.Messages, *session.Pending)
		if err != nil {
			return Session{}, err
		}

		session.Messages = history
		session.Pending = nil
		session.UpdatedAt = xtime.Now()
		if err := repo.Save(session); err != nil {
			return Session{}, fmt.Errorf("cannot persist session: %w", err)
		}

		return session, nil
	}
}

// loadPending loads a session which waits on the given pending revision, auditing like [Append].
func loadPending(subject auth.Subject, repo Repository, id ID, revision int) (Session, error) {
	optSession, err := repo.FindByID(id)
	if err != nil {
		return Session{}, fmt.Errorf("cannot load session: %w", err)
	}

	// A denied audit is reported like a missing session so foreign sessions are not revealed.
	if optSession.IsNone() || subject.AuditResource(Namespace, rebacInstance(id), PermAppend) != nil {
		return Session{}, fmt.Errorf("session %q does not exist", id)
	}

	session := optSession.Unwrap()
	if session.Pending == nil || session.PendingRevision != revision {
		return Session{}, ErrNoPendingDecision
	}

	return session, nil
}

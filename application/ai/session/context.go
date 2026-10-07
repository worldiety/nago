// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"context"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
)

type ctxKey struct{}

type ctxValue struct {
	root    ID
	running ID
}

// IDOf returns the conversation a tool runs in, from the context of the subject it is invoked with, i.e.
// IDOf(subject.Context()). A sub-agent runs in a child session, for which the root session is returned, which is
// the conversation the user sees and deletes. It reports false outside of a persisted session, e.g. in a chat
// without history.
func IDOf(ctx context.Context) (ID, bool) {
	v, ok := ctx.Value(ctxKey{}).(ctxValue)
	return v.root, ok
}

// RunningIDOf is like [IDOf], but returns the session which runs the tool, which is a child session for a
// sub-agent.
func RunningIDOf(ctx context.Context) (ID, bool) {
	v, ok := ctx.Value(ctxKey{}).(ctxValue)
	return v.running, ok
}

// withSession returns the subject, whose context tells the tools of a run the session, see [IDOf].
func withSession(repo Repository, subject auth.Subject, session Session) auth.Subject {
	ctx := subject.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	return user.WithContext(subject, context.WithValue(ctx, ctxKey{}, ctxValue{root: rootOf(repo, session), running: session.ID}))
}

// maxSessionDepth bounds the walk to the root session, in case of a broken chain of parents.
const maxSessionDepth = 16

// rootOf returns the root session of a child session, or the session itself.
func rootOf(repo Repository, session Session) ID {
	root := session.ID
	parent := session.ParentID
	for range maxSessionDepth {
		if parent == "" {
			break
		}

		root = parent
		optParent, err := repo.FindByID(parent)
		if err != nil || optParent.IsNone() {
			break
		}

		parent = optParent.Unwrap().ParentID
	}

	return root
}

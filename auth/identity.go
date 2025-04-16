// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package auth

import (
	"context"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/user"
)

type Subject = user.Subject

type contextKey string

const subjectKey contextKey = "authSubject"

func OneOf(subject Subject, permissions ...permission.ID) bool {
	for _, permission := range permissions {
		if subject.HasPermission(permission) {
			return true
		}
	}

	return false
}

func WithSubject(ctx context.Context, subject Subject) context.Context {
	return context.WithValue(ctx, subjectKey, subject)
}

func FromContext(ctx context.Context) (Subject, bool) {
	subject, ok := ctx.Value(subjectKey).(Subject)
	return subject, ok
}

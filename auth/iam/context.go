// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package iam

import (
	"context"
	"go.wdy.de/nago/auth"
)

type subjectKey string

const subjKey subjectKey = "iam.subject"

// FromContext returns always a subject from the given context. If no subject is available, returns
// an invalid subject.
func FromContext(ctx context.Context) auth.Subject {
	if subj, ok := ctx.Value(subjKey).(auth.Subject); ok {
		return subj
	}

	return auth.InvalidSubject{}
}

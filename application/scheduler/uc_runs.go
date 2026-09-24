// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler

import (
	"iter"

	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xiter"
	"go.wdy.de/nago/pkg/xslices"
)

// NewListRuns requires [PermStatus], because run statistics are status information and contain no log content.
func NewListRuns(m *Manager) ListRuns {
	return func(subject auth.Subject, id ID) iter.Seq2[Run, error] {
		if err := subject.Audit(PermStatus); err != nil {
			return xiter.WithError[Run](err)
		}

		runs, err := m.Runs(id)
		if err != nil {
			return xiter.WithError[Run](err)
		}

		return xslices.Values2[[]Run, Run, error](runs)
	}
}

// NewViewRunLog requires [PermViewLogs], so that existing role assignments keep working.
func NewViewRunLog(m *Manager) ViewRunLog {
	return func(subject auth.Subject, id ID, run RunID, query LogQuery) (LogPage, error) {
		if err := subject.Audit(PermViewLogs); err != nil {
			return LogPage{}, err
		}

		return m.RunLog(id, run, query)
	}
}

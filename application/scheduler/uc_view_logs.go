// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xiter"
	"go.wdy.de/nago/pkg/xslices"
	"iter"
)

func NewViewLogs(m *Manager) ViewLogs {
	return func(subject auth.Subject, id ID) iter.Seq2[LogEntry, error] {
		if err := subject.Audit(PermViewLogs); err != nil {
			return xiter.WithError[LogEntry](err)
		}

		return xslices.Values2[[]LogEntry, LogEntry, error](m.Logs(id))
	}
}

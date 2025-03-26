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

func NewListSchedulers(m *Manager) ListSchedulers {
	return func(subject auth.Subject) iter.Seq2[Options, error] {
		if err := subject.Audit(PermListSchedulers); err != nil {
			return xiter.WithError[Options](err)
		}

		tmp := m.Scheduler()
		return xslices.Values2[[]Options, Options, error](tmp)
	}
}

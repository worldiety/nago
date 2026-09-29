// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler

import (
	"go.wdy.de/nago/auth"
)

func NewRemove(m *Manager) Remove {
	return func(subject auth.Subject, id ID) error {
		if err := subject.Audit(PermRemove); err != nil {
			return err
		}

		return m.Remove(id)
	}
}

func NewReconfigure(m *Manager) Reconfigure {
	return func(subject auth.Subject, opts Options) error {
		if err := subject.Audit(PermConfigure); err != nil {
			return err
		}

		return m.Reconfigure(opts)
	}
}

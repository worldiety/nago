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

func NewConfigure(m *Manager) Configure {
	return func(subject auth.Subject, opts Options) error {
		if err := subject.Audit(PermConfigure); err != nil {
			return err
		}

		return m.Configure(opts)
	}
}

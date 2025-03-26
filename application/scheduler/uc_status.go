// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler

import (
	"errors"
	"go.wdy.de/nago/auth"
)

func NewStatus(m *Manager) Status {
	return func(subject auth.Subject, id ID) (StatusResult, error) {
		if err := subject.Audit(PermStatus); err != nil {
			return StatusResult{}, err
		}

		opts, ok := m.Options(id)
		if !ok {
			return StatusResult{}, errors.New("no options found")
		}

		var status StatusResult
		status.LastCompletedAt = m.LastCompletedAt(id)
		status.State = m.State(id)
		status.NextPlannedAt = m.NextPlannedAt(id)
		status.LastError = m.LastError(id)
		status.LastStartedAt = m.LastStartedAt(id)
		status.Options = opts

		return status, nil
	}
}

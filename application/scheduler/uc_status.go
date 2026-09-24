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
	"go.wdy.de/nago/pkg/std"
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

		runStats, runs, err := m.Stats(id)
		if err != nil {
			return StatusResult{}, err
		}

		status.Stats = runStats
		if len(runs) > 0 {
			status.LastRun = std.Some(runs[0])
		}

		status.Settings = opts.Defaults
		if m.settingsRepo != nil {
			optSettings, err := m.settingsRepo.FindByID(id)
			if err != nil {
				return StatusResult{}, err
			}

			if optSettings.IsSome() {
				status.Settings = optSettings.Unwrap()
				status.CustomSettings = true
			}
		}

		return status, nil
	}
}

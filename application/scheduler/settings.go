// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler

import (
	"go.wdy.de/nago/pkg/data"
	"time"
)

type Settings struct {
	ID         ID            `json:"id" visible:"false"`
	StartDelay time.Duration `json:"startDelay,omitempty" label:"Verzögerung nach Systemstart"`
	PauseTime  time.Duration `json:"pauseTime,omitempty" label:"Verzögerung zwischen den Ausführungen"`
	Disabled   bool          `json:"disabled,omitempty" label:"Deaktiviert"`
	CronHour   int           `json:"cronHour,omitempty" label:"Cron Hour"`
	CronMinute int           `json:"cronMinute,omitempty" label:"Cron Minute"`
}

func (s Settings) Identity() ID {
	return s.ID
}

func (s Settings) WithIdentity(id ID) Settings {
	s.ID = id
	return s
}

type SettingsRepository data.Repository[Settings, ID]

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

	// CronWeekdays restricts a [Cron] scheduler to these weekdays. Empty means every day. It must not be combined
	// with CronMonthDays.
	CronWeekdays []time.Weekday `json:"cronWeekdays,omitempty" visible:"false"`
	// CronMonthDays restricts a [Cron] scheduler to these days of the month (1 to 31 or [LastDayOfMonth]). A month
	// without a selected day is skipped. Empty means every day.
	CronMonthDays []int `json:"cronMonthDays,omitempty" visible:"false"`
	// CatchUp runs a durable scheduler once after its start, if at least one due elapsed, while its manager was not
	// running, see [Options.Durable].
	CatchUp bool `json:"catchUp,omitempty" visible:"false"`
	// KeepRunLogs is the amount of runs whose log file is kept. Zero means [KeepRunLogs]. Logs of runs older than
	// [RunRetention] are deleted in any case.
	KeepRunLogs int `json:"keepRunLogs,omitempty" visible:"false"`
}

func (s Settings) Identity() ID {
	return s.ID
}

func (s Settings) WithIdentity(id ID) Settings {
	s.ID = id
	return s
}

type SettingsRepository data.Repository[Settings, ID]

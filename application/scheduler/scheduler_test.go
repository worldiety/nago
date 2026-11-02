// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler

import (
	"testing"
	"time"
)

func TestNextCronAt(t *testing.T) {
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}

	newYork, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}

	date := func(loc *time.Location, month time.Month, day, hour, minute int) time.Time {
		return time.Date(2026, month, day, hour, minute, 0, 0, loc)
	}

	tests := []struct {
		name     string
		now      time.Time
		settings Settings
		want     time.Time
	}{
		{"later today", date(berlin, 6, 10, 2, 0), Settings{CronHour: 3}, date(berlin, 6, 10, 3, 0)},
		{"tomorrow", date(berlin, 6, 10, 10, 0), Settings{CronHour: 3, CronMinute: 15}, date(berlin, 6, 11, 3, 15)},
		{"next month", date(berlin, 6, 30, 10, 0), Settings{CronHour: 3}, date(berlin, 7, 1, 3, 0)},
		{"pause", date(berlin, 6, 10, 10, 0), Settings{CronHour: 3, PauseTime: 4 * time.Hour}, date(berlin, 6, 10, 11, 0)},

		// the daylight saving time ends at 2026-11-01 02:00 in New York
		{"tomorrow is 25h long", date(newYork, 10, 31, 4, 0), Settings{CronHour: 3}, date(newYork, 11, 1, 3, 0)},
		{"today is 25h long", date(newYork, 11, 1, 0, 30), Settings{CronHour: 3}, date(newYork, 11, 1, 3, 0)},

		// the daylight saving time begins at 2026-03-29 02:00 and ends at 2026-10-25 03:00 in Berlin
		{"tomorrow is 23h long", date(berlin, 3, 28, 4, 0), Settings{CronHour: 3}, date(berlin, 3, 29, 3, 0)},
		{"today is 23h long", date(berlin, 3, 29, 0, 30), Settings{CronHour: 3}, date(berlin, 3, 29, 3, 0)},
		{"skipped time", date(berlin, 3, 29, 0, 30), Settings{CronHour: 2, CronMinute: 30}, date(berlin, 3, 29, 3, 30)},
		{"pause today is 25h long", date(berlin, 10, 25, 5, 0), Settings{CronHour: 6, PauseTime: 6 * time.Hour}, date(berlin, 10, 25, 6, 0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nextCronAt(tt.now, tt.settings); !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

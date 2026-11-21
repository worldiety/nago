// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler_test

import (
	"testing"
	"time"

	"go.wdy.de/nago/application/scheduler"
)

// TestN29InvalidPlansAreRefused checks, that Configure and Update refuse a plan with weekdays together with days of
// the month, a day of the month 0 or 32, or catch up without Durable, without an effect, next to valid plans which
// differ in only that value. The clock is Tuesday 2027-06-01 21:00.
//
// Zusagen: R-SCHEDULE-DEFINE, R-SCHEDULE-PLAN-TIME, R-SCHEDULE-CATCH-UP, R-DEC-SCHEDULE-NAGO
func TestN29InvalidPlansAreRefused(t *testing.T) {
	tests := []struct {
		name  string
		mods  []func(*scheduler.Options)
		valid bool
		next  time.Time
	}{
		{"Wochentage zusammen mit Tagen des Monats", mods(weekdays(time.Wednesday), monthDays(1)), false, time.Time{}},
		{"Gegenzeile nur Wochentage", mods(weekdays(time.Wednesday)), true, bln(2027, 6, 2, 6, 0)},
		{"Gegenzeile nur Tage des Monats", mods(monthDays(1)), true, bln(2027, 7, 1, 6, 0)},
		{"Tag des Monats 0", mods(monthDays(0)), false, time.Time{}},
		{"Tag des Monats 32", mods(monthDays(32)), false, time.Time{}},
		{"Gegenzeile Tage 1, 31 und letzter Tag", mods(monthDays(1, 31, scheduler.LastDayOfMonth)), true, bln(2027, 6, 30, 6, 0)},
		{"Nachholen ohne Durable", mods(catchUp, durable(false)), false, time.Time{}},
		{"Gegenzeile Nachholen mit Durable", mods(catchUp), true, bln(2027, 6, 2, 6, 0)},
	}

	for _, tt := range tests {
		t.Run("Configure: "+tt.name, func(t *testing.T) {
			w := newWorld(t, bln(2027, 6, 1, 21, 0))
			p := newProbe(w.clock)
			err := w.m.Configure(cronOpts(job, p.run, 6, 0, tt.mods...))
			if tt.valid {
				if err != nil {
					t.Fatalf("gültiger Plan abgelehnt: %v", err)
				}

				w.ids = append(w.ids, job)
				w.waitPlanned(job, tt.next)
				return
			}

			if err == nil {
				t.Fatal("ungültiger Plan angenommen")
			}

			if w.m.Has(job) {
				t.Fatal("der abgelehnte Scheduler ist konfiguriert")
			}
		})

		t.Run("Update: "+tt.name, func(t *testing.T) {
			w := newWorld(t, bln(2027, 6, 1, 21, 0))
			p := newProbe(w.clock)
			w.configure(cronOpts(job, p.run, 22, 0))
			w.waitPlanned(job, bln(2027, 6, 1, 22, 0))

			err := w.m.Update(cronOpts(job, p.run, 6, 0, tt.mods...))
			if tt.valid {
				if err != nil {
					t.Fatalf("gültiger Plan abgelehnt: %v", err)
				}

				w.waitPlanned(job, tt.next)
				return
			}

			if err == nil {
				t.Fatal("ungültiger Plan angenommen")
			}

			// the plan before stays
			mustSettle(t)
			w.waitPlanned(job, bln(2027, 6, 1, 22, 0))
			if opts, ok := w.m.Options(job); !ok || opts.Defaults.CronHour != 22 || len(opts.Defaults.CronMonthDays) != 0 || opts.Defaults.CatchUp {
				t.Fatalf("Optionen unverändert erwartet: %+v", opts.Defaults)
			}

			if runs := w.runs(job); len(runs) != 0 {
				t.Fatalf("kein Lauf erwartet:%s", describeRuns(runs))
			}

			p.none(t)
		})
	}
}

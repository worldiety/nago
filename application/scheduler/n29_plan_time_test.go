// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler_test

import (
	"context"
	"testing"
	"time"

	"go.wdy.de/nago/application/scheduler"
)

func mods(m ...func(*scheduler.Options)) []func(*scheduler.Options) {
	return m
}

// TestN29PlanTimeDue checks the table of R-SCHEDULE-PLAN-TIME and the extensions of R-DEC-SCHEDULE-NAGO: a time
// on chosen weekdays or days of the month (1 to 31, last day), a zone per scheduler and the first of two equal wall
// clock times when the daylight saving time ends. Each row configures the scheduler at "now" and observes the
// planned due. Rows with several dues move the clock to each due, see exactly one run with this due and then the
// next due.
//
// Zusagen: R-SCHEDULE-PLAN-TIME, R-DEC-SCHEDULE-NAGO
func TestN29PlanTimeDue(t *testing.T) {
	tests := []struct {
		name         string
		now          time.Time
		hour, minute int
		mods         []func(*scheduler.Options)
		dues         []time.Time
	}{
		{"täglich 22:00 um 21:59", bln(2027, 6, 1, 21, 59), 22, 0, nil,
			[]time.Time{bln(2027, 6, 1, 22, 0), bln(2027, 6, 2, 22, 0)}},
		{"täglich 22:00 um 22:00:30", time.Date(2027, 6, 1, 22, 0, 30, 0, berlin), 22, 0, nil,
			[]time.Time{bln(2027, 6, 2, 22, 0)}},
		{"Montag und Freitag 06:00 am Samstag", bln(2027, 6, 5, 8, 0), 6, 0, mods(weekdays(time.Monday, time.Friday)),
			[]time.Time{bln(2027, 6, 7, 6, 0), bln(2027, 6, 11, 6, 0)}},
		{"am 1. 06:00", bln(2027, 1, 15, 10, 0), 6, 0, mods(monthDays(1)),
			[]time.Time{bln(2027, 2, 1, 6, 0), bln(2027, 3, 1, 6, 0)}},
		{"am 31. 06:00 lässt April und Juni aus", bln(2027, 4, 1, 10, 0), 6, 0, mods(monthDays(31)),
			[]time.Time{bln(2027, 5, 31, 6, 0), bln(2027, 7, 31, 6, 0)}},
		{"letzter Tag 23:00", bln(2027, 2, 10, 10, 0), 23, 0, mods(monthDays(scheduler.LastDayOfMonth)),
			[]time.Time{bln(2027, 2, 28, 23, 0), bln(2027, 3, 31, 23, 0)}},
		{"am 30. und letzter Tag 23:00: im April genau ein Lauf", bln(2027, 4, 29, 10, 0), 23, 0, mods(monthDays(30, scheduler.LastDayOfMonth)),
			[]time.Time{bln(2027, 4, 30, 23, 0), bln(2027, 5, 30, 23, 0), bln(2027, 5, 31, 23, 0)}},
		// 02:30 does not exist on 2027-03-28, time.Date normalizes it to 03:30 MESZ, i.e. 01:30 UTC
		{"täglich 02:30 am Tag ohne 02:30", bln(2027, 3, 27, 12, 0), 2, 30, nil,
			[]time.Time{utc(2027, 3, 28, 1, 30), bln(2027, 3, 29, 2, 30)}},
		// 02:30 exists twice on 2027-10-31: the first is 02:30 MESZ, i.e. 00:30 UTC (time.Date would choose MEZ)
		{"täglich 02:30 am Tag mit zwei 02:30: die erste, genau ein Lauf", bln(2027, 10, 30, 12, 0), 2, 30, nil,
			[]time.Time{utc(2027, 10, 31, 0, 30), bln(2027, 11, 1, 2, 30)}},
		{"täglich 22:00 in Europe/Berlin bei einer Uhr in UTC", utc(2027, 6, 1, 12, 0), 22, 0, nil,
			[]time.Time{utc(2027, 6, 1, 20, 0)}},
		{"Gegenzeile: täglich 22:00 in der Zone UTC", utc(2027, 6, 1, 12, 0), 22, 0, mods(zone(time.UTC)),
			[]time.Time{utc(2027, 6, 1, 22, 0)}},
		{"Gegenzeile ohne die neuen Schalter: täglich 22:00 in time.Local wie bisher", time.Date(2027, 6, 1, 12, 0, 0, 0, time.Local), 22, 0, mods(oldStyle),
			[]time.Time{time.Date(2027, 6, 1, 22, 0, 0, 0, time.Local)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld(t, tt.now)
			p := newProbe(w.clock)
			w.configure(cronOpts(job, p.run, tt.hour, tt.minute, tt.mods...))

			for i, due := range tt.dues {
				w.waitPlanned(job, due)
				if i == len(tt.dues)-1 {
					break
				}

				w.clock.set(t, due)
				inv := p.next(t)
				if !inv.hasInfo || inv.info.Trigger != scheduler.TriggerPlan || !inv.info.Due.Equal(due) || !inv.at.Equal(due) {
					t.Fatalf("Lauf zu %v: Beginn %v, RunInfo %+v (vorhanden %v)", due, inv.at, inv.info, inv.hasInfo)
				}

				runs := w.ended(job, i+1)
				if r := runs[0]; !r.Due.Equal(due) || !r.StartedAt.Equal(due) || r.Trigger != scheduler.TriggerPlan || r.Outcome != scheduler.OutcomeSucceeded {
					t.Fatalf("Lauf zu %v:%s", due, describeRuns(runs))
				}
			}

			p.none(t)
		})
	}
}

// TestN29RunNumberPerDayInZone checks, that the consecutive number of a run per day counts the days of the zone of
// the scheduler. 21:30 and 22:30 UTC on 2027-06-01 are one day in UTC, but two days in Europe/Berlin.
//
// Zusagen: R-DEC-SCHEDULE-NAGO
func TestN29RunNumberPerDayInZone(t *testing.T) {
	first, second := utc(2027, 6, 1, 21, 30), utc(2027, 6, 1, 22, 30)
	local := []int{1, 1}
	if first.In(time.Local).YearDay() == second.In(time.Local).YearDay() {
		local = []int{1, 2}
	}

	tests := []struct {
		name string
		mod  func(*scheduler.Options)
		nrs  []int
	}{
		{"Europe/Berlin: zwei Tage", zone(berlin), []int{1, 1}},
		{"UTC: ein Tag", zone(time.UTC), []int{1, 2}},
		{"Gegenzeile ohne die neuen Schalter: Tage von time.Local", oldStyle, local},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld(t, first)
			p := newProbe(w.clock)
			// a manual scheduler has no due, thus only the runs of StartNow count
			manual := func(o *scheduler.Options) { o.Kind = scheduler.Manual }
			w.configure(intervalOpts(job, p.run, 30*day, tt.mod, manual))

			w.mustStartNow(job, "Anna")
			w.ended(job, 1)
			w.clock.set(t, second)
			w.mustStartNow(job, "Anna")
			runs := w.ended(job, 2)

			if got := []int{runs[1].Nr, runs[0].Nr}; got[0] != tt.nrs[0] || got[1] != tt.nrs[1] {
				t.Fatalf("Nummern %v, erwartet %v:%s", got, tt.nrs, describeRuns(runs))
			}
		})
	}
}

// TestN29RunInfo checks the due and the trigger, which the runner receives and which the run keeps.
//
// Zusagen: R-SCHEDULE-INPUT, R-SCHEDULE-LOG, R-DEC-SCHEDULE-NAGO
func TestN29RunInfo(t *testing.T) {
	if info, ok := scheduler.RunInfoFrom(context.Background()); ok {
		t.Fatalf("Gegenzeile ohne Lauf: RunInfo %+v", info)
	}

	tests := []struct {
		name    string
		trigger func(w *world)
		want    scheduler.RunInfo
	}{
		{"nach Plan", func(w *world) { w.fire(job, bln(2027, 6, 1, 22, 0)) },
			scheduler.RunInfo{Due: bln(2027, 6, 1, 22, 0), Trigger: scheduler.TriggerPlan}},
		{"von Hand", func(w *world) { w.clock.set(w.t, bln(2027, 6, 1, 21, 13)); w.mustStartNow(job, "Anna") },
			scheduler.RunInfo{Due: bln(2027, 6, 1, 21, 13), Trigger: scheduler.TriggerManual, By: "Anna"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld(t, bln(2027, 6, 1, 21, 0))
			p := newProbe(w.clock)
			w.configure(cronOpts(job, p.run, 22, 0))
			w.waitPlanned(job, bln(2027, 6, 1, 22, 0))

			tt.trigger(w)
			inv := p.next(t)
			if !inv.hasInfo || !inv.info.Due.Equal(tt.want.Due) || inv.info.Trigger != tt.want.Trigger || inv.info.By != tt.want.By {
				t.Fatalf("RunInfo %+v (vorhanden %v), erwartet %+v", inv.info, inv.hasInfo, tt.want)
			}

			runs := w.ended(job, 1)
			if r := runs[0]; !r.Due.Equal(tt.want.Due) || r.Trigger != tt.want.Trigger || r.By != tt.want.By {
				t.Fatalf("Lauf erwartet %+v:%s", tt.want, describeRuns(runs))
			}
		})
	}
}

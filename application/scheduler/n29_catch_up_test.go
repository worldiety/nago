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

// TestN29CatchUp checks the table of R-SCHEDULE-CATCH-UP. Each row builds its history on a first manager, ends it
// in an orderly way and starts a second manager on the same stores. The row "pausiert, noch pausiert" is missing,
// because the interface names no way to configure a scheduler as paused (Configure and then Pause races with the
// catch up). The row of a run, which a restart cancelled, is in TestN29RestartClosesOpenRun.
//
// Zusagen: R-SCHEDULE-CATCH-UP, R-SCHEDULE-PLAN-INTERVAL, R-SCHEDULE-PLAN-TIME, R-SCHEDULE-PAUSE,
// R-SCHEDULE-OVERLAP, R-SCHEDULE-RETENTION, R-SCHEDULE-INPUT, R-DEC-SCHEDULE-NAGO
func TestN29CatchUp(t *testing.T) {
	daily := func(mods ...func(*scheduler.Options)) func(p *probe) scheduler.Options {
		return func(p *probe) scheduler.Options { return cronOpts(job, p.run, 22, 0, mods...) }
	}

	hourly := func(mods ...func(*scheduler.Options)) func(p *probe) scheduler.Options {
		return func(p *probe) scheduler.Options { return intervalOpts(job, p.run, time.Hour, mods...) }
	}

	weekly := func(mods ...func(*scheduler.Options)) func(p *probe) scheduler.Options {
		return func(p *probe) scheduler.Options { return intervalOpts(job, p.run, 7*day, mods...) }
	}

	// last run 2027-11-20 22:00, end of the process 2027-11-21 21:00
	ranYesterday := func(w *world, p *probe) {
		w.fire(job, bln(2027, 11, 20, 22, 0))
		p.next(w.t)
		w.ended(job, 1)
		w.waitPlanned(job, bln(2027, 11, 21, 22, 0))
		w.clock.set(w.t, bln(2027, 11, 21, 21, 0))
	}

	// last end 21:00, end of the process 21:30
	endedAt21 := func(w *world, p *probe) {
		w.fire(job, bln(2027, 11, 21, 21, 0))
		p.next(w.t)
		w.ended(job, 1)
		w.waitPlanned(job, bln(2027, 11, 21, 22, 0))
		w.clock.set(w.t, bln(2027, 11, 21, 21, 30))
	}

	// last end Monday 08:00 of a manual run, end of the process Saturday 09:00
	endedMonday := func(w *world, p *probe) {
		w.waitPlanned(job, bln(2027, 11, 22, 7, 0))
		w.clock.set(w.t, bln(2027, 11, 15, 8, 0))
		w.mustStartNow(job, "Anna")
		p.next(w.t)
		w.ended(job, 1)
		w.waitPlanned(job, bln(2027, 11, 22, 8, 0))
		w.clock.set(w.t, bln(2027, 11, 20, 9, 0))
	}

	tests := []struct {
		name    string
		begin   time.Time
		opts    func(p *probe) scheduler.Options
		history func(w *world, p *probe) // ends with the clock at the end of the first process
		start   time.Time                // start of the second process
		due     time.Time                // due of the run caught up; zero: no run
		endAt   time.Time                // if set, the run caught up ends at this time
		next    time.Time                // planned due afterward
		every   time.Duration            // if set, the planned due is this interval after the end of the run caught up
	}{
		{
			name:    "täglich 22:00, an: ein Lauf nachgeholt bis 23:01, dann 22.11. 22:00",
			begin:   bln(2027, 11, 20, 21, 0),
			opts:    daily(catchUp),
			history: ranYesterday,
			start:   bln(2027, 11, 21, 23, 0),
			due:     bln(2027, 11, 21, 22, 0),
			next:    bln(2027, 11, 22, 22, 0),
		},
		{
			name:    "täglich 22:00, aus: kein Lauf, nächster 22.11. 22:00",
			begin:   bln(2027, 11, 20, 21, 0),
			opts:    daily(),
			history: ranYesterday,
			start:   bln(2027, 11, 21, 23, 0),
			next:    bln(2027, 11, 22, 22, 0),
		},
		{
			name:  "täglich 22:00, an, drei Fälligkeiten verpasst: genau ein Lauf mit der spätesten",
			begin: bln(2027, 11, 17, 21, 0),
			opts:  daily(catchUp),
			history: func(w *world, p *probe) {
				w.fire(job, bln(2027, 11, 17, 22, 0))
				p.next(w.t)
				w.ended(job, 1)
				w.waitPlanned(job, bln(2027, 11, 18, 22, 0))
				w.clock.set(w.t, bln(2027, 11, 18, 21, 0))
			},
			start: bln(2027, 11, 21, 23, 0),
			due:   bln(2027, 11, 21, 22, 0),
			next:  bln(2027, 11, 22, 22, 0),
		},
		{
			name:  "täglich 22:00, an, pausiert 21:00, fortgesetzt 21:10: ein Lauf nachgeholt",
			begin: bln(2027, 11, 20, 21, 0),
			opts:  daily(catchUp),
			history: func(w *world, p *probe) {
				w.fire(job, bln(2027, 11, 20, 22, 0))
				p.next(w.t)
				w.ended(job, 1)
				w.waitPlanned(job, bln(2027, 11, 21, 22, 0))
				w.clock.set(w.t, bln(2027, 11, 21, 21, 0))
				if err := w.m.Pause(job); err != nil {
					w.t.Fatal(err)
				}

				w.clock.set(w.t, bln(2027, 11, 21, 21, 10))
				if err := w.m.Resume(job); err != nil {
					w.t.Fatal(err)
				}

				w.waitPlanned(job, bln(2027, 11, 21, 22, 0))
				w.clock.set(w.t, bln(2027, 11, 21, 21, 30))
			},
			start: bln(2027, 11, 21, 23, 0),
			due:   bln(2027, 11, 21, 22, 0),
			next:  bln(2027, 11, 22, 22, 0),
		},
		{
			name:    "alle 60 Minuten, an: ein Lauf nachgeholt, der nächste 60 Minuten nach seinem Ende",
			begin:   bln(2027, 11, 21, 20, 0),
			opts:    hourly(catchUp),
			history: endedAt21,
			start:   bln(2027, 11, 21, 23, 0),
			due:     bln(2027, 11, 21, 22, 0),
			endAt:   bln(2027, 11, 21, 23, 5),
			every:   time.Hour,
		},
		{
			name:    "alle 60 Minuten, aus: erster Lauf 60 Minuten nach dem Start",
			begin:   bln(2027, 11, 21, 20, 0),
			opts:    hourly(),
			history: endedAt21,
			start:   bln(2027, 11, 21, 23, 0),
			next:    bln(2027, 11, 22, 0, 0),
		},
		{
			name:  "alle 60 Minuten, an, pausiert 10:30, fortgesetzt 12:00: nachgeholt für 13:00, keiner für 11:00",
			begin: bln(2027, 11, 21, 9, 0),
			opts:  hourly(catchUp),
			history: func(w *world, p *probe) {
				w.fire(job, bln(2027, 11, 21, 10, 0))
				p.next(w.t)
				w.ended(job, 1)
				w.waitPlanned(job, bln(2027, 11, 21, 11, 0))
				w.clock.set(w.t, bln(2027, 11, 21, 10, 30))
				if err := w.m.Pause(job); err != nil {
					w.t.Fatal(err)
				}

				w.clock.set(w.t, bln(2027, 11, 21, 12, 0))
				if err := w.m.Resume(job); err != nil {
					w.t.Fatal(err)
				}

				w.waitPlanned(job, bln(2027, 11, 21, 13, 0))
				if runs := w.runs(job); len(runs) != 1 {
					w.t.Fatalf("kein Lauf während der Pause erwartet:%s", describeRuns(runs))
				}

				w.clock.set(w.t, bln(2027, 11, 21, 12, 10))
			},
			start: bln(2027, 11, 21, 23, 0),
			due:   bln(2027, 11, 21, 13, 0),
			every: time.Hour,
		},
		{
			name:  "täglich 22:00, an, Lauf von Hand 21:50 bis 22:10, Fälligkeit 22:00 entfiel: kein Lauf",
			begin: bln(2027, 11, 20, 21, 0),
			opts:  daily(catchUp),
			history: func(w *world, p *probe) {
				w.fire(job, bln(2027, 11, 20, 22, 0))
				p.next(w.t)
				w.ended(job, 1)
				w.waitPlanned(job, bln(2027, 11, 21, 22, 0))
				w.clock.set(w.t, bln(2027, 11, 21, 21, 50))
				p.hold.Store(true)
				w.mustStartNow(job, "Anna")
				inv := p.next(w.t)
				w.clock.set(w.t, bln(2027, 11, 21, 22, 0))
				w.waitPlanned(job, bln(2027, 11, 22, 22, 0))
				p.none(w.t)
				w.clock.set(w.t, bln(2027, 11, 21, 22, 10))
				p.hold.Store(false)
				inv.finish <- nil
				w.ended(job, 2)
				w.clock.set(w.t, bln(2027, 11, 21, 22, 20))
			},
			start: bln(2027, 11, 21, 22, 30),
			next:  bln(2027, 11, 22, 22, 0),
		},
		{
			name:    "alle 7 Tage, an: kein Lauf beim Start, nächster Montag 08:00",
			begin:   bln(2027, 11, 15, 7, 0),
			opts:    weekly(catchUp),
			history: endedMonday,
			start:   bln(2027, 11, 20, 10, 0),
			next:    bln(2027, 11, 22, 8, 0),
		},
		{
			name:    "alle 7 Tage, aus: kein Lauf beim Start, nächster Montag 08:00",
			begin:   bln(2027, 11, 15, 7, 0),
			opts:    weekly(),
			history: endedMonday,
			start:   bln(2027, 11, 20, 10, 0),
			next:    bln(2027, 11, 22, 8, 0),
		},
		{
			name:  "am 31. 06:00, an, Lauf vom 31.05. nach 30 Tagen gelöscht: nachgeholt für 31.07.",
			begin: bln(2027, 5, 31, 5, 0),
			opts: func(p *probe) scheduler.Options {
				return cronOpts(job, p.run, 6, 0, monthDays(31), catchUp)
			},
			history: func(w *world, p *probe) {
				w.fire(job, bln(2027, 5, 31, 6, 0))
				p.next(w.t)
				w.ended(job, 1)
				w.waitPlanned(job, bln(2027, 7, 31, 6, 0))
				w.clock.set(w.t, bln(2027, 7, 30, 0, 0))
				var runs []scheduler.Run
				waitFor(w.t, func() string {
					return "Prämisse: der Lauf vom 31.05. ist nach 30 Tagen ohne weiteren Lauf gelöscht, vorhanden:" + describeRuns(runs)
				}, func() bool {
					runs = w.runs(job)
					return len(runs) == 0
				})
				w.waitPlanned(job, bln(2027, 7, 31, 6, 0))
			},
			start: bln(2027, 8, 1, 0, 0),
			due:   bln(2027, 7, 31, 6, 0),
			next:  bln(2027, 8, 31, 6, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld(t, tt.begin)
			p := newProbe(w.clock)
			w.configure(tt.opts(p))
			tt.history(w, p)
			p.none(t)
			before := len(w.runs(job))

			w.restartAt(tt.start)
			p2 := newProbe(w.clock)
			p2.hold.Store(!tt.endAt.IsZero())
			w.configure(tt.opts(p2))
			limit := tt.start.Add(time.Minute)

			if tt.due.IsZero() {
				// the loop waits for its next due; up to 60 s after the start nothing runs
				w.waitPlanned(job, tt.next)
				w.clock.set(t, limit)
				w.waitPlanned(job, tt.next)
				if runs := w.runs(job); len(runs) != before {
					t.Fatalf("kein Lauf nach dem Start erwartet:%s", describeRuns(runs))
				}

				p2.none(t)
				return
			}

			inv := w.awaitCatchUp(p2, limit)
			if !inv.hasInfo || inv.info.Trigger != scheduler.TriggerCatchUp || !inv.info.Due.Equal(tt.due) || inv.at.After(limit) {
				t.Fatalf("nachgeholter Lauf um %v mit %+v (vorhanden %v), erwartet fällig %v bis %v", inv.at, inv.info, inv.hasInfo, tt.due, limit)
			}

			if !tt.endAt.IsZero() {
				w.clock.set(t, tt.endAt)
				inv.finish <- nil
			}

			runs := w.ended(job, before+1)
			if r := runs[0]; r.Trigger != scheduler.TriggerCatchUp || !r.Due.Equal(tt.due) {
				t.Fatalf("nachgeholter Lauf mit Fälligkeit %v erwartet:%s", tt.due, describeRuns(runs))
			}

			// exactly one run caught up, then the plan continues
			next := tt.next
			if tt.every > 0 {
				next = runs[0].CompletedAt.Add(tt.every)
				if !tt.endAt.IsZero() && !runs[0].CompletedAt.Equal(tt.endAt) {
					t.Fatalf("Ende %v erwartet:%s", tt.endAt, describeRuns(runs))
				}
			}

			w.waitPlanned(job, next)
			if runs := w.runs(job); len(runs) != before+1 {
				t.Fatalf("genau ein nachgeholter Lauf erwartet:%s", describeRuns(runs))
			}

			p2.none(t)
		})
	}
}

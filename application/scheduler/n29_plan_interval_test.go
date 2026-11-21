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

// TestN29PlanIntervalNoRunWhenConfigured checks, that configuring a durable interval starts no run, but plans the
// first due after the interval. Without the new switches, the scheduler runs immediately as before.
//
// Zusagen: R-SCHEDULE-PLAN-INTERVAL, R-DEC-SCHEDULE-NAGO
func TestN29PlanIntervalNoRunWhenConfigured(t *testing.T) {
	t0 := bln(2027, 6, 1, 10, 0)
	tests := []struct {
		name string
		mod  func(*scheduler.Options)
		runs int
	}{
		{"Durable: kein Lauf beim Anlegen", durable(true), 0},
		{"Gegenzeile ohne die neuen Schalter: sofortiger Lauf wie bisher", oldStyle, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld(t, t0)
			p := newProbe(w.clock)
			w.configure(intervalOpts(job, p.run, time.Hour, tt.mod))

			if tt.runs == 1 {
				if inv := p.next(t); !inv.at.Equal(t0) {
					t.Fatalf("sofortiger Lauf um %v, erwartet %v", inv.at, t0)
				}

				w.ended(job, 1)
			}

			w.waitPlanned(job, t0.Add(time.Hour))
			if runs := w.runs(job); len(runs) != tt.runs {
				t.Fatalf("%d Läufe erwartet:%s", tt.runs, describeRuns(runs))
			}

			p.none(t)
		})
	}
}

// TestN29NextDueAfterRun checks, that with the plan "Abstand" the next due is the interval after the end of the
// previous run, also of a manual one, while a manual run does not move the due of the plan "Uhrzeit". The next due
// then runs exactly once.
//
// Zusagen: R-SCHEDULE-PLAN-INTERVAL, R-SCHEDULE-RUN-NOW, R-DEC-SCHEDULE-NAGO
func TestN29NextDueAfterRun(t *testing.T) {
	t0 := bln(2027, 6, 1, 10, 0)
	tests := []struct {
		name   string
		opts   func(p *probe) scheduler.Options
		manual bool      // a manual run from 10:30 to 10:40 instead of the planned run at 11:00
		end    time.Time // end of the first run
		next   time.Time
	}{
		{"Abstand: nach dem Ende eines Laufs nach Plan, nicht nach seinem Beginn",
			func(p *probe) scheduler.Options { return intervalOpts(job, p.run, time.Hour) },
			false, bln(2027, 6, 1, 11, 20), bln(2027, 6, 1, 12, 20)},
		{"Abstand: ein Lauf von Hand zählt als voriger Lauf",
			func(p *probe) scheduler.Options { return intervalOpts(job, p.run, time.Hour) },
			true, bln(2027, 6, 1, 10, 40), bln(2027, 6, 1, 11, 40)},
		{"Gegenzeile Uhrzeit: ein Lauf von Hand verschiebt die Fälligkeit nicht",
			func(p *probe) scheduler.Options { return cronOpts(job, p.run, 11, 0) },
			true, bln(2027, 6, 1, 10, 40), bln(2027, 6, 1, 11, 0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld(t, t0)
			p := newProbe(w.clock)
			p.hold.Store(true)
			w.configure(tt.opts(p))
			w.waitPlanned(job, bln(2027, 6, 1, 11, 0))

			if tt.manual {
				w.clock.set(t, bln(2027, 6, 1, 10, 30))
				w.mustStartNow(job, "Anna")
			} else {
				w.clock.set(t, bln(2027, 6, 1, 11, 0))
			}

			inv := p.next(t)
			p.hold.Store(false)
			w.clock.set(t, tt.end)
			inv.finish <- nil
			w.ended(job, 1)

			w.fire(job, tt.next)
			if inv := p.next(t); inv.info.Trigger != scheduler.TriggerPlan || !inv.info.Due.Equal(tt.next) {
				t.Fatalf("Lauf zu %v mit %+v", tt.next, inv.info)
			}

			runs := w.ended(job, 2)
			if r := runs[0]; r.Trigger != scheduler.TriggerPlan || !r.Due.Equal(tt.next) || !r.StartedAt.Equal(tt.next) {
				t.Fatalf("Lauf zu %v:%s", tt.next, describeRuns(runs))
			}

			p.none(t)
		})
	}
}

// TestN29SetReference checks, that SetReference adds a link to the reference of the plan "Abstand": a later
// reference moves the due, an earlier one does not. The reference survives a restart.
//
// Zusagen: R-SCHEDULE-PLAN-INTERVAL, R-SCHEDULE-CATCH-UP, R-DEC-SCHEDULE-NAGO
func TestN29SetReference(t *testing.T) {
	ref := func(h, m int) *time.Time { r := bln(2027, 6, 1, h, m); return &r }
	tests := []struct {
		name string
		ref  *time.Time
		next time.Time
	}{
		{"Gegenzeile ohne weiteren Bezug: Ende des Laufs", nil, bln(2027, 6, 1, 11, 0)},
		{"späterer Bezug zählt", ref(10, 15), bln(2027, 6, 1, 11, 15)},
		{"früherer Bezug zählt nicht", ref(9, 45), bln(2027, 6, 1, 11, 0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld(t, bln(2027, 6, 1, 9, 30))
			p := newProbe(w.clock)
			opts := intervalOpts(job, p.run, time.Hour)
			w.configure(opts)
			w.waitPlanned(job, bln(2027, 6, 1, 10, 30))

			w.clock.set(t, bln(2027, 6, 1, 10, 0))
			w.mustStartNow(job, "Anna")
			p.next(t)
			w.ended(job, 1)
			w.waitPlanned(job, bln(2027, 6, 1, 11, 0))

			w.clock.set(t, bln(2027, 6, 1, 10, 15))
			if tt.ref != nil {
				if err := w.m.SetReference(job, *tt.ref); err != nil {
					t.Fatal(err)
				}
			}

			w.waitPlanned(job, tt.next)

			// the reference is kept in the store of the runs
			w.restartAt(bln(2027, 6, 1, 10, 30))
			w.configure(opts)
			w.waitPlanned(job, tt.next)
			if runs := w.runs(job); len(runs) != 1 {
				t.Fatalf("kein Lauf beim Start erwartet:%s", describeRuns(runs))
			}

			p.none(t)
		})
	}
}

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

// TestN29PauseDropsDue checks, that a due while paused is dropped without a run.
//
// Zusagen: R-SCHEDULE-PAUSE, R-DEC-SCHEDULE-NAGO
func TestN29PauseDropsDue(t *testing.T) {
	tests := []struct {
		name  string
		pause bool
		runs  int
	}{
		{"pausiert: die Fälligkeit entfällt", true, 0},
		{"Gegenzeile ohne Pause: Lauf zur Fälligkeit", false, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld(t, bln(2027, 6, 1, 21, 0))
			p := newProbe(w.clock)
			w.configure(cronOpts(job, p.run, 22, 0))
			w.waitPlanned(job, bln(2027, 6, 1, 22, 0))

			if tt.pause {
				if err := w.m.Pause(job); err != nil {
					t.Fatal(err)
				}

				w.clock.set(t, bln(2027, 6, 1, 22, 0))
				w.clock.set(t, bln(2027, 6, 1, 22, 30))
				if err := w.m.Resume(job); err != nil {
					t.Fatal(err)
				}
			} else {
				w.fire(job, bln(2027, 6, 1, 22, 0))
				p.next(t)
				w.ended(job, 1)
			}

			w.waitPlanned(job, bln(2027, 6, 2, 22, 0))
			if runs := w.runs(job); len(runs) != tt.runs {
				t.Fatalf("%d Läufe erwartet:%s", tt.runs, describeRuns(runs))
			}

			p.none(t)
		})
	}
}

// TestN29PauseKeepsRunningRun checks, that Pause does not end a running run, unlike Stop.
//
// Zusagen: R-SCHEDULE-PAUSE, R-DEC-SCHEDULE-NAGO
func TestN29PauseKeepsRunningRun(t *testing.T) {
	tests := []struct {
		name     string
		action   func(m *scheduler.Manager) error
		canceled bool
		outcome  scheduler.Outcome
	}{
		{"Pause: der Lauf läuft zu Ende", func(m *scheduler.Manager) error { return m.Pause(job) }, false, scheduler.OutcomeSucceeded},
		{"Gegenzeile Stop: bricht ihn ab wie bisher", func(m *scheduler.Manager) error { return m.Stop(job) }, true, scheduler.OutcomeCanceled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld(t, bln(2027, 6, 1, 21, 0))
			p := newProbe(w.clock)
			p.hold.Store(true)
			w.configure(cronOpts(job, p.run, 22, 0))
			w.fire(job, bln(2027, 6, 1, 22, 0))
			inv := p.next(t)

			if err := tt.action(w.m); err != nil {
				t.Fatal(err)
			}

			if tt.canceled {
				waitFor(t, func() string { return "der Kontext des Laufs endet" }, func() bool { return inv.ctx.Err() != nil })
			} else {
				if err := inv.ctx.Err(); err != nil {
					t.Fatalf("der Lauf wurde beendet: %v", err)
				}

				if runs := w.runs(job); len(runs) != 1 || runs[0].Outcome != scheduler.OutcomeRunning {
					t.Fatalf("laufender Lauf erwartet:%s", describeRuns(runs))
				}

				w.clock.set(t, bln(2027, 6, 1, 22, 5))
				inv.finish <- nil
			}

			runs := w.ended(job, 1)
			if runs[0].Outcome != tt.outcome {
				t.Fatalf("Ergebnis %d erwartet:%s", tt.outcome, describeRuns(runs))
			}
		})
	}
}

// TestN29ResumeWithoutImmediateRun checks, that Resume starts no run, although a due passed while paused, and plans
// the next due from the time of Resume.
//
// Zusagen: R-SCHEDULE-PAUSE, R-SCHEDULE-PLAN-INTERVAL, R-SCHEDULE-CATCH-UP, R-DEC-SCHEDULE-NAGO
func TestN29ResumeWithoutImmediateRun(t *testing.T) {
	tests := []struct {
		name  string
		opts  func(p *probe) scheduler.Options
		pause bool
		next  time.Time
		runs  int
	}{
		{"Abstand: kein Lauf beim Fortsetzen, Abstand ab dem Fortsetzen",
			func(p *probe) scheduler.Options { return intervalOpts(job, p.run, time.Hour) },
			true, bln(2027, 6, 1, 13, 0), 0},
		{"Uhrzeit: kein Lauf beim Fortsetzen, nächste Uhrzeit",
			func(p *probe) scheduler.Options { return cronOpts(job, p.run, 11, 0) },
			true, bln(2027, 6, 2, 11, 0), 0},
		{"Gegenzeile Abstand ohne Pause: Lauf zur Fälligkeit",
			func(p *probe) scheduler.Options { return intervalOpts(job, p.run, time.Hour) },
			false, bln(2027, 6, 1, 12, 0), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld(t, bln(2027, 6, 1, 10, 0))
			p := newProbe(w.clock)
			w.configure(tt.opts(p))
			w.waitPlanned(job, bln(2027, 6, 1, 11, 0))

			if tt.pause {
				w.clock.set(t, bln(2027, 6, 1, 10, 30))
				if err := w.m.Pause(job); err != nil {
					t.Fatal(err)
				}

				w.clock.set(t, bln(2027, 6, 1, 12, 0))
				if err := w.m.Resume(job); err != nil {
					t.Fatal(err)
				}
			} else {
				w.fire(job, bln(2027, 6, 1, 11, 0))
				p.next(t)
				w.ended(job, 1)
			}

			w.waitPlanned(job, tt.next)
			if runs := w.runs(job); len(runs) != tt.runs {
				t.Fatalf("%d Läufe erwartet:%s", tt.runs, describeRuns(runs))
			}

			p.none(t)
		})
	}
}

// TestN29UpdateKeepsRunningRun checks, that Update replaces the runner without ending a running run, which ends
// with the runner it began with, unlike Reconfigure, and that the next run uses the new runner.
//
// Zusagen: R-SCHEDULE-DEFINE, R-DEC-SCHEDULE-NAGO
func TestN29UpdateKeepsRunningRun(t *testing.T) {
	tests := []struct {
		name     string
		action   func(m *scheduler.Manager, opts scheduler.Options) error
		canceled bool
	}{
		{"Update: der Lauf läuft zu Ende", func(m *scheduler.Manager, opts scheduler.Options) error { return m.Update(opts) }, false},
		{"Gegenzeile Reconfigure: bricht ihn ab wie bisher", func(m *scheduler.Manager, opts scheduler.Options) error { return m.Reconfigure(opts) }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld(t, bln(2027, 6, 1, 21, 0))
			p1, p2 := newProbe(w.clock), newProbe(w.clock)
			p1.hold.Store(true)
			w.configure(cronOpts(job, p1.run, 22, 0))
			w.waitPlanned(job, bln(2027, 6, 1, 22, 0))
			w.mustStartNow(job, "Anna")
			inv := p1.next(t)

			if err := tt.action(w.m, cronOpts(job, p2.run, 22, 0)); err != nil {
				t.Fatal(err)
			}

			if tt.canceled {
				waitFor(t, func() string { return "der Kontext des Laufs endet" }, func() bool { return inv.ctx.Err() != nil })
				w.ended(job, 1)
			} else {
				if err := inv.ctx.Err(); err != nil {
					t.Fatalf("der Lauf wurde beendet: %v", err)
				}

				w.clock.set(t, bln(2027, 6, 1, 21, 5))
				inv.finish <- nil
				if runs := w.ended(job, 1); runs[0].Outcome != scheduler.OutcomeSucceeded {
					t.Fatalf("gelungener Lauf erwartet:%s", describeRuns(runs))
				}
			}

			w.mustStartNow(job, "Anna")
			p2.next(t)
			p1.none(t)
			w.ended(job, 2)
		})
	}
}

// TestN29UpdatePlan checks, that Update changes the plan from the next due on, also between "Abstand" and
// "Uhrzeit", and starts no run. The changes happen at the time of the reference, so that the result does not depend
// on whether Update itself counts as a reference.
//
// Zusagen: R-SCHEDULE-DEFINE, R-SCHEDULE-PLAN-INTERVAL, R-SCHEDULE-PLAN-TIME, R-DEC-SCHEDULE-NAGO
func TestN29UpdatePlan(t *testing.T) {
	t0 := bln(2027, 6, 1, 10, 0)
	tests := []struct {
		name   string
		before func(p *probe) scheduler.Options
		first  time.Time
		after  func(p *probe) scheduler.Options
		next   time.Time
	}{
		{"Abstand 60 auf 30 Minuten",
			func(p *probe) scheduler.Options { return intervalOpts(job, p.run, time.Hour) }, bln(2027, 6, 1, 11, 0),
			func(p *probe) scheduler.Options { return intervalOpts(job, p.run, 30*time.Minute) }, bln(2027, 6, 1, 10, 30)},
		{"Uhrzeit 22:00 auf Abstand 60 Minuten",
			func(p *probe) scheduler.Options { return cronOpts(job, p.run, 22, 0) }, bln(2027, 6, 1, 22, 0),
			func(p *probe) scheduler.Options { return intervalOpts(job, p.run, time.Hour) }, bln(2027, 6, 1, 11, 0)},
		{"Abstand 60 Minuten auf Uhrzeit 14:00",
			func(p *probe) scheduler.Options { return intervalOpts(job, p.run, time.Hour) }, bln(2027, 6, 1, 11, 0),
			func(p *probe) scheduler.Options { return cronOpts(job, p.run, 14, 0) }, bln(2027, 6, 1, 14, 0)},
		{"Uhrzeit 22:00 auf 20:00",
			func(p *probe) scheduler.Options { return cronOpts(job, p.run, 22, 0) }, bln(2027, 6, 1, 22, 0),
			func(p *probe) scheduler.Options { return cronOpts(job, p.run, 20, 0) }, bln(2027, 6, 1, 20, 0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld(t, t0)
			p := newProbe(w.clock)
			w.configure(tt.before(p))
			w.waitPlanned(job, tt.first)

			if err := w.m.Update(tt.after(p)); err != nil {
				t.Fatal(err)
			}

			w.waitPlanned(job, tt.next)
			if runs := w.runs(job); len(runs) != 0 {
				t.Fatalf("kein Lauf durch Update erwartet:%s", describeRuns(runs))
			}

			p.none(t)
			w.clock.set(t, tt.next)
			if inv := p.next(t); inv.info.Trigger != scheduler.TriggerPlan || !inv.info.Due.Equal(tt.next) {
				t.Fatalf("Lauf zu %v erwartet: %+v", tt.next, inv.info)
			}

			if runs := w.ended(job, 1); !runs[0].Due.Equal(tt.next) {
				t.Fatalf("genau ein Lauf zu %v erwartet:%s", tt.next, describeRuns(runs))
			}
		})
	}
}

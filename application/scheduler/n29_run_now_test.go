// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler_test

import (
	"errors"
	"testing"
	"time"

	"go.wdy.de/nago/application/scheduler"
)

// TestN29StartNow checks, that StartNow begins a manual run without waiting for it, and that it is refused without
// a run and without waiting while a run is running. The runner holds each run, thus a StartNow which waited for
// a run would never return.
//
// Zusagen: R-SCHEDULE-RUN-NOW, R-SCHEDULE-OVERLAP, R-SCHEDULE-PAUSE, R-DEC-SCHEDULE-NAGO
func TestN29StartNow(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(w *world, p *probe) []*invocation // returns the held runs
		wantErr error
		runs    int // runs in the store afterward
	}{
		{"ohne Lauf: gelingt ohne zu warten",
			func(w *world, p *probe) []*invocation { return nil }, nil, 1},
		{"während eines Laufs von Hand: abgelehnt",
			func(w *world, p *probe) []*invocation {
				w.mustStartNow(job, "Bea")
				return []*invocation{p.next(w.t)}
			}, scheduler.ErrRunning, 1},
		{"während eines Laufs nach Plan: abgelehnt",
			func(w *world, p *probe) []*invocation {
				w.fire(job, bln(2027, 6, 1, 22, 0))
				return []*invocation{p.next(w.t)}
			}, scheduler.ErrRunning, 1},
		{"pausiert: gelingt",
			func(w *world, p *probe) []*invocation {
				if err := w.m.Pause(job); err != nil {
					w.t.Fatal(err)
				}
				return nil
			}, nil, 1},
		{"Gegenzeile nach dem Ende eines Laufs von Hand: gelingt",
			func(w *world, p *probe) []*invocation {
				w.mustStartNow(job, "Bea")
				p.next(w.t).finish <- nil
				w.ended(job, 1)
				// the newest run is the one which began later
				w.clock.set(w.t, bln(2027, 6, 1, 21, 1))
				return nil
			}, nil, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld(t, bln(2027, 6, 1, 21, 0))
			p := newProbe(w.clock)
			p.hold.Store(true)
			w.configure(cronOpts(job, p.run, 22, 0))
			w.waitPlanned(job, bln(2027, 6, 1, 22, 0))
			held := tt.setup(w, p)

			err := w.startNow(job, "Anna")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("StartNow: %v, erwartet %v", err, tt.wantErr)
			}

			if tt.wantErr == nil {
				now := w.clock.Now()
				inv := p.next(t)
				if !inv.hasInfo || inv.info.Trigger != scheduler.TriggerManual || inv.info.By != "Anna" || !inv.info.Due.Equal(now) {
					t.Fatalf("RunInfo %+v (vorhanden %v), erwartet von Hand durch Anna, fällig %v", inv.info, inv.hasInfo, now)
				}

				runs := w.started(job, tt.runs)
				if r := runs[0]; r.Trigger != scheduler.TriggerManual || r.By != "Anna" || !r.Due.Equal(now) || r.Outcome != scheduler.OutcomeRunning {
					t.Fatalf("laufender Lauf von Hand erwartet:%s", describeRuns(runs))
				}

				held = append(held, inv)
			} else {
				p.none(t)
				if runs := w.runs(job); len(runs) != tt.runs {
					t.Fatalf("kein weiterer Lauf erwartet:%s", describeRuns(runs))
				}
			}

			for _, inv := range held {
				inv.finish <- nil
			}

			w.ended(job, tt.runs)
		})
	}
}

// TestN29DueDuringManualRunIsDropped checks, that a due of the plan "Uhrzeit" during a manual run is dropped
// without a run and the plan continues with the next due.
//
// Zusagen: R-SCHEDULE-OVERLAP, R-DEC-SCHEDULE-NAGO
func TestN29DueDuringManualRunIsDropped(t *testing.T) {
	tests := []struct {
		name    string
		manual  bool
		trigger scheduler.Trigger
	}{
		{"Lauf von Hand über die Fälligkeit: kein Lauf nach Plan", true, scheduler.TriggerManual},
		{"Gegenzeile ohne Lauf von Hand: Lauf zur Fälligkeit", false, scheduler.TriggerPlan},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld(t, bln(2027, 6, 1, 21, 0))
			p := newProbe(w.clock)
			w.configure(cronOpts(job, p.run, 22, 0))
			w.waitPlanned(job, bln(2027, 6, 1, 22, 0))

			if tt.manual {
				w.clock.set(t, bln(2027, 6, 1, 21, 50))
				p.hold.Store(true)
				w.mustStartNow(job, "Anna")
				inv := p.next(t)
				p.hold.Store(false)

				w.clock.set(t, bln(2027, 6, 1, 22, 0))
				w.waitPlanned(job, bln(2027, 6, 2, 22, 0))
				p.none(t)
				if runs := w.runs(job); len(runs) != 1 {
					t.Fatalf("nur der Lauf von Hand erwartet:%s", describeRuns(runs))
				}

				w.clock.set(t, bln(2027, 6, 1, 22, 10))
				inv.finish <- nil
			} else {
				w.fire(job, bln(2027, 6, 1, 22, 0))
				p.next(t)
			}

			runs := w.ended(job, 1)
			if runs[0].Trigger != tt.trigger {
				t.Fatalf("Auslöser %d erwartet:%s", tt.trigger, describeRuns(runs))
			}

			w.waitPlanned(job, bln(2027, 6, 2, 22, 0))
			p.none(t)
			if runs := w.runs(job); len(runs) != 1 {
				t.Fatalf("genau ein Lauf erwartet:%s", describeRuns(runs))
			}
		})
	}
}

// TestN29CancelRun checks, that CancelRun ends only the running run with its own outcome and the name of the person,
// while the loop continues with the next due, and that it is refused without a run.
//
// Zusagen: R-SCHEDULE-RUN-NOW, R-DEC-SCHEDULE-NAGO
func TestN29CancelRun(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(w *world, p *probe) *invocation
		wantErr error
		by      string    // who started the run
		next    time.Time // the next due, which must still run
	}{
		{"Lauf von Hand",
			func(w *world, p *probe) *invocation { w.mustStartNow(job, "Anna"); return p.next(w.t) },
			nil, "Anna", bln(2027, 6, 1, 22, 0)},
		{"Lauf nach Plan",
			func(w *world, p *probe) *invocation { w.fire(job, bln(2027, 6, 1, 22, 0)); return p.next(w.t) },
			nil, "", bln(2027, 6, 2, 22, 0)},
		{"Gegenzeile ohne Lauf: abgelehnt",
			func(w *world, p *probe) *invocation { return nil },
			scheduler.ErrNotRunning, "", bln(2027, 6, 1, 22, 0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld(t, bln(2027, 6, 1, 21, 0))
			p := newProbe(w.clock)
			p.hold.Store(true)
			w.configure(cronOpts(job, p.run, 22, 0))
			w.waitPlanned(job, bln(2027, 6, 1, 22, 0))
			inv := tt.setup(w, p)
			p.hold.Store(false)

			done := make(chan error, 1)
			go func() { done <- w.m.CancelRun(job, "Bea") }()
			var err error
			select {
			case err = <-done:
			case <-time.After(waitLimit):
				t.Fatal("CancelRun kehrt nicht zurück")
			}

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CancelRun: %v, erwartet %v", err, tt.wantErr)
			}

			n := 0
			if inv != nil {
				n = 1
				waitFor(t, func() string { return "der Kontext des Laufs endet" }, func() bool { return inv.ctx.Err() != nil })
				runs := w.ended(job, 1)
				if r := runs[0]; r.Outcome != scheduler.OutcomeCanceled || r.CanceledBy != "Bea" || r.By != tt.by || r.Outcome.Failed() {
					t.Fatalf("abgebrochen von Bea erwartet:%s", describeRuns(runs))
				}
			} else if runs := w.runs(job); len(runs) != 0 {
				t.Fatalf("kein Lauf erwartet:%s", describeRuns(runs))
			}

			// the loop has not been stopped
			w.fire(job, tt.next)
			if inv := p.next(t); inv.info.Trigger != scheduler.TriggerPlan || !inv.info.Due.Equal(tt.next) {
				t.Fatalf("Lauf nach Plan zu %v erwartet: %+v", tt.next, inv.info)
			}

			runs := w.ended(job, n+1)
			if r := runs[0]; r.Outcome != scheduler.OutcomeSucceeded || r.CanceledBy != "" {
				t.Fatalf("folgender Lauf gelungen erwartet:%s", describeRuns(runs))
			}
		})
	}
}

// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler_test

import (
	"fmt"
	"testing"
	"time"

	"go.wdy.de/nago/application/scheduler"
)

// The windows after the end of a run: a due or StartNow, which arrives while the scheduler still deals with the end
// of the previous run, must neither be lost nor refused. Each window is forced by the test clock or by the store of
// the runs and then held, until every goroutine waits (settle), instead of being hoped for by repetition.

// beginManualRun configures "täglich 22:00" at 21:00 and holds a manual run of Anna.
func beginManualRun(t *testing.T) (*world, *probe, *invocation) {
	t.Helper()
	w := newWorld(t, bln(2027, 6, 1, 21, 0))
	p := newProbe(w.clock)
	w.configure(cronOpts(job, p.run, 22, 0))
	w.waitPlanned(job, bln(2027, 6, 1, 22, 0))
	p.hold.Store(true)
	w.mustStartNow(job, "Anna")
	inv := p.next(t)
	p.hold.Store(false)
	return w, p, inv
}

func manualEnd(r scheduler.Run) bool {
	return r.By == "Anna" && r.Outcome != scheduler.OutcomeRunning
}

// expectDue checks, after every goroutine waits, that the due 22:00 ran exactly once, if the manual run ended before
// it, and that the plan continues with the next day.
func expectDue(t *testing.T, w *world, p *probe) {
	t.Helper()
	due, next := bln(2027, 6, 1, 22, 0), bln(2027, 6, 2, 22, 0)
	mustSettle(t)

	var manual scheduler.Run
	var plan []scheduler.Run
	for _, r := range w.runs(job) {
		switch r.Trigger {
		case scheduler.TriggerManual:
			manual = r
		case scheduler.TriggerPlan:
			plan = append(plan, r)
		}
	}

	if manual.CompletedAt.IsZero() || manual.Outcome != scheduler.OutcomeSucceeded {
		t.Fatalf("der Lauf von Hand ist gelungen zu Ende:%s", describeRuns(w.runs(job)))
	}

	switch {
	case manual.CompletedAt.Before(due):
		runs := w.ended(job, 2)
		if len(plan) > 1 || runs[0].Trigger != scheduler.TriggerPlan || !runs[0].Due.Equal(due) {
			t.Fatalf("genau ein Lauf zur Fälligkeit %v nach dem Ende des Laufs von Hand erwartet:%s", due, describeRuns(runs))
		}

		if inv := p.next(t); inv.info.Trigger != scheduler.TriggerPlan || !inv.info.Due.Equal(due) {
			t.Fatalf("Lauf zur Fälligkeit %v erwartet: %+v", due, inv.info)
		}
	case len(plan) > 1:
		// the clock reached the due before the end of the manual run was read; the due may run once or not at all
		t.Fatalf("höchstens ein Lauf zur Fälligkeit erwartet:%s", describeRuns(w.runs(job)))
	default:
		for range plan {
			p.next(t)
		}
	}

	w.waitPlanned(job, next)
	p.none(t)
}

// TestN29WindowsAfterManualRun checks the three windows of e4d60b1ca, each next to its counter row without the window:
// (i) the loop has read the clock after the end of a manual run, but not yet armed its timer, when the clock reaches
// the due; (ii) the end of the run is stored, but the scheduler has not finished with it, when the due arrives;
// (iii) StartNow in the same window.
//
// Zusagen: R-SCHEDULE-OVERLAP, R-SCHEDULE-RUN-NOW, R-SCHEDULE-PLAN-TIME, R-DEC-SCHEDULE-NAGO
func TestN29WindowsAfterManualRun(t *testing.T) {
	due := bln(2027, 6, 1, 22, 0)

	t.Run("(i) die Uhr erreicht die Fälligkeit innerhalb eines Aufrufs von Now nach dem Ende des Laufs von Hand", func(t *testing.T) {
		// each call of Now after the end of the manual run is a row: the caller still reads 21:00, then the clock is 22:00
		const most = 50
		for n := 1; ; n++ {
			if n > most {
				t.Fatalf("mehr als %d Aufrufe von Now nach dem Ende eines Laufs von Hand", most)
			}

			last := false
			t.Run(fmt.Sprintf("%d. Aufruf", n), func(t *testing.T) {
				w, p, inv := beginManualRun(t)
				w.clock.armNow(n, due)
				inv.finish <- nil
				mustSettle(t)
				if !w.clock.disarm() {
					last = true
					t.Skipf("nach dem Ende des Laufs von Hand gibt es nur %d Aufrufe von Now", n-1)
				}

				expectDue(t, w, p)
			})

			if last {
				break
			}
		}
	})

	t.Run("(ii) Fälligkeit, wenn das Ende des Laufs gespeichert ist und der Scheduler noch nicht fertig damit", func(t *testing.T) {
		w, p, inv := beginManualRun(t)
		done := w.onceStored(manualEnd, func() {
			w.clock.move(due)
			if !settle() {
				t.Error("die Goroutinen kommen im Fenster nicht zur Ruhe")
			}
		})

		inv.finish <- nil
		select {
		case <-done:
		case <-time.After(waitLimit):
			t.Fatal("das Ende des Laufs von Hand wurde nicht gespeichert")
		}

		expectDue(t, w, p)
	})

	t.Run("Gegenzeile (i, ii) ohne Fenster: die Uhr erreicht die Fälligkeit, wenn alles wartet", func(t *testing.T) {
		w, p, inv := beginManualRun(t)
		inv.finish <- nil
		w.ended(job, 1)
		mustSettle(t)
		w.fire(job, due)
		expectDue(t, w, p)
	})

	tests := []struct {
		name   string
		window bool
	}{
		{"(iii) StartNow, wenn das Ende des Laufs gespeichert ist und der Scheduler noch nicht fertig damit", true},
		{"Gegenzeile (iii) ohne Fenster: StartNow, wenn alles wartet", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, p, inv := beginManualRun(t)
			result := make(chan error, 1)
			if tt.window {
				done := w.onceStored(manualEnd, func() {
					go func() { result <- w.m.StartNow(job, "Bea") }()
					if !settle() {
						t.Error("die Goroutinen kommen im Fenster nicht zur Ruhe")
					}
				})

				inv.finish <- nil
				select {
				case <-done:
				case <-time.After(waitLimit):
					t.Fatal("das Ende des Laufs von Hand wurde nicht gespeichert")
				}
			} else {
				inv.finish <- nil
				w.ended(job, 1)
				mustSettle(t)
				go func() { result <- w.m.StartNow(job, "Bea") }()
			}

			select {
			case err := <-result:
				if err != nil {
					t.Fatalf("StartNow nach dem Ende des Laufs von Hand: %v", err)
				}
			case <-time.After(waitLimit):
				t.Fatal("StartNow kehrt nicht zurück")
			}

			if inv := p.next(t); inv.info.Trigger != scheduler.TriggerManual || inv.info.By != "Bea" {
				t.Fatalf("Lauf von Hand durch Bea erwartet: %+v", inv.info)
			}

			runs := w.ended(job, 2)
			bea := 0
			for _, r := range runs {
				if r.By == "Bea" && r.Trigger == scheduler.TriggerManual && r.Outcome == scheduler.OutcomeSucceeded {
					bea++
				}
			}

			if bea != 1 {
				t.Fatalf("genau ein gelungener Lauf von Bea erwartet:%s", describeRuns(runs))
			}

			p.none(t)
		})
	}
}

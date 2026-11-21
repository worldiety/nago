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

// TestN29RestartClosesOpenRun checks, that a durable scheduler closes a run without end, which a crashed process
// left in the store, as canceled by "Neustart" and does not repeat it, also with catch up. Without the new switches
// the run stays open as before.
//
// Zusagen: R-SCHEDULE-RESTART, R-SCHEDULE-CATCH-UP, R-DEC-SCHEDULE-NAGO
func TestN29RestartClosesOpenRun(t *testing.T) {
	tests := []struct {
		name    string
		mods    []func(*scheduler.Options)
		start   time.Time
		outcome scheduler.Outcome
		by      string
	}{
		{"Durable: abgebrochen: Neustart", nil, bln(2027, 6, 1, 22, 5), scheduler.OutcomeCanceled, "Neustart"},
		{"Durable mit Nachholen: abgebrochen: Neustart und nicht wiederholt", mods(catchUp), bln(2027, 6, 1, 23, 0), scheduler.OutcomeCanceled, "Neustart"},
		{"Gegenzeile ohne Durable und Quiet: der Lauf bleibt offen wie bisher", mods(durable(false), quiet(false)), bln(2027, 6, 1, 22, 5), scheduler.OutcomeRunning, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld(t, bln(2027, 6, 1, 21, 0))
			p := newProbe(w.clock)
			p.hold.Store(true)
			w.configure(cronOpts(job, p.run, 22, 0, tt.mods...))
			w.fire(job, bln(2027, 6, 1, 22, 0))
			p.next(t)
			open := w.started(job, 1)[0]

			c := w.crashAt(tt.start)
			p2 := newProbe(c.clock)
			c.configure(cronOpts(job, p2.run, 22, 0, tt.mods...))

			var runs []scheduler.Run
			waitFor(t, func() string { return "Lauf ohne Ende abgeschlossen:" + describeRuns(runs) }, func() bool {
				runs = c.runs(job)
				return len(runs) == 1 && runs[0].Outcome == tt.outcome && runs[0].CanceledBy == tt.by
			})

			if r := runs[0]; r.ID != open.ID || r.Outcome.Failed() || (tt.outcome == scheduler.OutcomeCanceled && r.CompletedAt.IsZero()) {
				t.Fatalf("Lauf %s abgeschlossen erwartet:%s", open.ID, describeRuns(runs))
			}

			// up to 60 s after the start nothing runs
			c.waitPlanned(job, bln(2027, 6, 2, 22, 0))
			c.clock.set(t, tt.start.Add(time.Minute))
			c.waitPlanned(job, bln(2027, 6, 2, 22, 0))
			if runs := c.runs(job); len(runs) != 1 {
				t.Fatalf("kein Lauf nach dem Start erwartet:%s", describeRuns(runs))
			}

			p2.none(t)
		})
	}
}

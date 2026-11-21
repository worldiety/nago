// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler_test

import (
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"go.wdy.de/nago/application/scheduler"
	"go.wdy.de/nago/application/user"
)

// TestN29KeepRunLogs checks the amount of kept logs per scheduler: the log of each older run is deleted on disk,
// while its statistics stay. Zero keeps scheduler.KeepRunLogs as before.
//
// Zusagen: R-SCHEDULE-RETENTION, R-DEC-SCHEDULE-NAGO
func TestN29KeepRunLogs(t *testing.T) {
	tests := []struct {
		name    string
		keep    int
		runs    int
		withLog int
	}{
		{"2 aufbewahren, 2 Läufe: beide Protokolle bleiben", 2, 2, 2},
		{"2 aufbewahren, 3 Läufe: der älteste verliert sein Protokoll", 2, 3, 2},
		{"Gegenzeile 0: KeepRunLogs wie bisher", 0, scheduler.KeepRunLogs + 1, scheduler.KeepRunLogs},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t0 := bln(2027, 6, 1, 10, 0)
			w := newWorld(t, t0)
			p := newProbe(w.clock)
			w.configure(intervalOpts(job, p.run, 30*day, keepLogs(tt.keep)))

			var created []scheduler.Run // oldest first
			for i := range tt.runs {
				w.clock.set(t, t0.Add(time.Duration(i)*time.Minute))
				w.mustStartNow(job, "Anna")
				r := w.ended(job, i+1)[0]
				if r.LogFile == "" || r.Entries == 0 {
					t.Fatalf("Prämisse: der neueste Lauf hat ein Protokoll mit Einträgen:%s", describeRuns([]scheduler.Run{r}))
				}

				created = append(created, r)
			}

			kept := func(i int) bool { return i >= tt.runs-tt.withLog }
			waitFor(t, func() string {
				var s string
				for i, r := range created {
					s += fmt.Sprintf("\n  %s: Datei vorhanden %v, erwartet %v", r.ID, fileExists(filepath.Join(w.logDir, r.LogFile)), kept(i))
				}
				return "Protokolle auf der Platte:" + s
			}, func() bool {
				for i, r := range created {
					if fileExists(filepath.Join(w.logDir, r.LogFile)) != kept(i) {
						return false
					}
				}
				return true
			})

			runs := w.runs(job)
			if len(runs) != tt.runs {
				t.Fatalf("die Kennzahlen aller Läufe bleiben:%s", describeRuns(runs))
			}

			for i, r := range created {
				page, err := w.m.RunLog(job, r.ID, scheduler.LogQuery{MinLevel: slog.LevelDebug})
				if kept(i) {
					if err != nil || page.Total == 0 {
						t.Fatalf("Protokoll von %s erwartet: %v %+v", r.ID, err, page)
					}
					continue
				}

				if !errors.Is(err, scheduler.ErrRunLogNotAvailable) {
					t.Fatalf("Protokoll von %s gelöscht erwartet: %v", r.ID, err)
				}

				// begin, end, trigger, outcome and the amount of entries stay
				now := runs[len(runs)-1-i]
				if now.ID != r.ID || !now.StartedAt.Equal(r.StartedAt) || !now.CompletedAt.Equal(r.CompletedAt) || now.Trigger != r.Trigger || now.Outcome != r.Outcome || now.Entries != r.Entries {
					t.Fatalf("Kennzahlen von %s unverändert erwartet:%s\nvorher:%s", r.ID, describeRuns([]scheduler.Run{now}), describeRuns([]scheduler.Run{r}))
				}
			}
		})
	}
}

// TestN29RunsExpireWithoutFurtherRun checks, that a run older than 30 days is deleted at the latest 24 hours later
// without a further run, also while paused, and that a younger one stays.
//
// Zusagen: R-SCHEDULE-RETENTION, R-DEC-SCHEDULE-NAGO
func TestN29RunsExpireWithoutFurtherRun(t *testing.T) {
	tests := []struct {
		name  string
		pause bool
	}{
		{"aktiv, Abstand 30 Tage", false},
		{"pausiert", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t0 := bln(2027, 6, 1, 10, 0)
			w := newWorld(t, t0)
			p := newProbe(w.clock)
			w.configure(intervalOpts(job, p.run, 30*day))
			if tt.pause {
				if err := w.m.Pause(job); err != nil {
					t.Fatal(err)
				}
			}

			w.mustStartNow(job, "Anna")
			old := w.ended(job, 1)[0]
			w.clock.set(t, t0.Add(25*time.Hour))
			w.mustStartNow(job, "Anna")
			young := w.ended(job, 2)[0]
			if old.LogFile == "" {
				t.Fatalf("Prämisse: der ältere Lauf hat ein Protokoll:%s", describeRuns([]scheduler.Run{old}))
			}

			if !tt.pause {
				// the next due lies after the end of the test, thus no further run happens
				w.waitPlanned(job, young.CompletedAt.Add(30*day))
			}

			// 31 days after the older run, 29 days and 23 hours after the younger one
			w.clock.set(t, t0.Add(31*day))
			var runs []scheduler.Run
			waitFor(t, func() string { return "nur der jüngere Lauf bleibt:" + describeRuns(runs) }, func() bool {
				runs = w.runs(job)
				return len(runs) == 1 && runs[0].ID == young.ID
			})

			waitFor(t, func() string { return "Protokoll des älteren Laufs gelöscht: " + old.LogFile }, func() bool {
				return !fileExists(filepath.Join(w.logDir, old.LogFile))
			})

			if _, err := w.m.RunLog(job, old.ID, scheduler.LogQuery{}); !errors.Is(err, scheduler.ErrRunNotFound) {
				t.Fatalf("älterer Lauf nicht mehr vorhanden erwartet: %v", err)
			}

			p.next(t)
			p.next(t)
			p.none(t)
		})
	}
}

// TestN29PurgeRuns checks, that PurgeRuns deletes the runs and logs of one scheduler, but not those of another one
// whose id starts with the same characters.
//
// Zusagen: R-SCHEDULE-DELETE, R-DEC-SCHEDULE-NAGO
func TestN29PurgeRuns(t *testing.T) {
	const purged, other scheduler.ID = "n29.job1", "n29.job10"
	t0 := bln(2027, 6, 1, 10, 0)
	w := newWorld(t, t0)
	p := newProbe(w.clock)
	w.configure(intervalOpts(purged, p.run, 30*day))
	w.configure(intervalOpts(other, p.run, 30*day))

	for i := range 2 {
		w.clock.set(t, t0.Add(time.Duration(i)*time.Hour))
		w.mustStartNow(purged, "Anna")
		w.ended(purged, i+1)
		w.mustStartNow(other, "Anna")
		w.ended(other, i+1)
	}

	gone := w.runs(purged)
	kept := w.runs(other)
	for _, r := range slices.Concat(gone, kept) {
		if r.LogFile == "" || !fileExists(filepath.Join(w.logDir, r.LogFile)) {
			t.Fatalf("Prämisse: Protokoll vorhanden:%s", describeRuns([]scheduler.Run{r}))
		}
	}

	if err := w.m.PurgeRuns(purged); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		id   scheduler.ID
		runs []scheduler.Run
		left bool
	}{
		{"gelöscht: " + string(purged), purged, gone, false},
		{"Gegenzeile anderer Scheduler: " + string(other), other, kept, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := 0
			if tt.left {
				want = len(tt.runs)
			}

			if runs, err := w.m.Runs(tt.id); err != nil || len(runs) != want {
				t.Fatalf("%d Läufe erwartet: %v%s", want, err, describeRuns(runs))
			}

			if page, total, err := w.m.RunsPage(tt.id, 0, 10); err != nil || total != want || len(page) != want {
				t.Fatalf("RunsPage: %d von %d, %v, erwartet %d", len(page), total, err, want)
			}

			for _, r := range tt.runs {
				if fileExists(filepath.Join(w.logDir, r.LogFile)) != tt.left {
					t.Fatalf("Protokoll %s vorhanden erwartet %v", r.LogFile, tt.left)
				}

				_, err := w.m.RunLog(tt.id, r.ID, scheduler.LogQuery{})
				if tt.left && err != nil {
					t.Fatalf("Protokoll von %s erwartet: %v", r.ID, err)
				}

				if !tt.left && !errors.Is(err, scheduler.ErrRunNotFound) {
					t.Fatalf("Lauf %s gelöscht erwartet: %v", r.ID, err)
				}
			}
		})
	}
}

// TestN29PurgeRunsForgetsReference checks, that PurgeRuns also deletes the reference of the plan: after a restart
// the interval begins at the start like for a new scheduler instead of after the end of the last run.
//
// Zusagen: R-SCHEDULE-DELETE, R-SCHEDULE-PLAN-INTERVAL, R-DEC-SCHEDULE-NAGO
func TestN29PurgeRunsForgetsReference(t *testing.T) {
	tests := []struct {
		name  string
		purge bool
		next  time.Time
	}{
		{"PurgeRuns: Abstand nach dem Start", true, bln(2027, 6, 1, 21, 50)},
		{"Gegenzeile ohne PurgeRuns: Abstand nach dem Ende des letzten Laufs", false, bln(2027, 6, 1, 21, 30)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld(t, bln(2027, 6, 1, 20, 0))
			p := newProbe(w.clock)
			opts := intervalOpts(job, p.run, time.Hour)
			w.configure(opts)
			w.waitPlanned(job, bln(2027, 6, 1, 21, 0))
			w.clock.set(t, bln(2027, 6, 1, 20, 30))
			w.mustStartNow(job, "Anna")
			w.ended(job, 1)
			w.waitPlanned(job, bln(2027, 6, 1, 21, 30))
			w.clock.set(t, bln(2027, 6, 1, 20, 40))

			if tt.purge {
				if err := w.m.PurgeRuns(job); err != nil {
					t.Fatal(err)
				}
			}

			w.restartAt(bln(2027, 6, 1, 20, 50))
			w.configure(opts)
			w.waitPlanned(job, tt.next)
			p.next(t)
			p.none(t)
		})
	}
}

// TestN29RunsPage checks the pages of the runs, newest first, also across days and next to a scheduler whose id
// starts with the same characters.
//
// Zusagen: R-NFR-SCHEDULE-RUNS, R-SCHEDULE-PAGE, R-DEC-SCHEDULE-NAGO
func TestN29RunsPage(t *testing.T) {
	const id, other scheduler.ID = "n29.job1", "n29.job10"
	w := newWorld(t, bln(2027, 6, 1, 9, 0))
	p := newProbe(w.clock)
	w.configure(intervalOpts(id, p.run, 30*day))
	w.configure(intervalOpts(other, p.run, 30*day))

	starts := []time.Time{bln(2027, 6, 1, 10, 0), bln(2027, 6, 1, 11, 0), bln(2027, 6, 1, 23, 30), bln(2027, 6, 2, 0, 30), bln(2027, 6, 2, 10, 0)}
	others := 0
	for i, at := range starts {
		w.clock.set(t, at)
		w.mustStartNow(id, "Anna")
		w.ended(id, i+1)
		if i%2 == 0 {
			w.clock.set(t, at.Add(time.Minute))
			w.mustStartNow(other, "Anna")
			others++
			w.ended(other, others)
		}
	}

	tests := []struct {
		offset, limit int
		want          []time.Time
	}{
		{0, 2, []time.Time{starts[4], starts[3]}},
		{2, 2, []time.Time{starts[2], starts[1]}},
		{4, 2, []time.Time{starts[0]}},
		{5, 2, nil},
		{0, 10, []time.Time{starts[4], starts[3], starts[2], starts[1], starts[0]}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("ab %d höchstens %d", tt.offset, tt.limit), func(t *testing.T) {
			page, total, err := w.m.RunsPage(id, tt.offset, tt.limit)
			if err != nil {
				t.Fatal(err)
			}

			if total != len(starts) || len(page) != len(tt.want) {
				t.Fatalf("%d von %d, erwartet %d von %d:%s", len(page), total, len(tt.want), len(starts), describeRuns(page))
			}

			for i, r := range page {
				if r.Scheduler != id || !r.StartedAt.Equal(tt.want[i]) {
					t.Fatalf("Lauf %d: erwartet Beginn %v:%s", i, tt.want[i], describeRuns(page))
				}
			}
		})
	}
}

// TestN29Stats24h checks, that the statistics count the runs of the last 24 hours of the clock: a failed run falls
// out after 24 hours, while a younger one counts.
//
// Zusagen: R-NFR-SCHEDULE-RUNS, R-SCHEDULE-PAGE, R-DEC-SCHEDULE-NAGO
func TestN29Stats24h(t *testing.T) {
	t0 := bln(2027, 6, 1, 10, 0)
	w := newWorld(t, t0)
	p := newProbe(w.clock)
	w.configure(intervalOpts(job, p.run, 30*day))

	p.fail(errors.New("n29 gescheitert"))
	w.mustStartNow(job, "Anna")
	w.ended(job, 1)
	p.result.Store(nil)

	tests := []struct {
		name          string
		at            time.Time
		total, failed int
	}{
		{"nach 2 Stunden: beide Läufe, einer gescheitert", t0.Add(2 * time.Hour), 2, 1},
		{"nach 25 Stunden: der gescheiterte ist älter als 24 Stunden", t0.Add(25 * time.Hour), 2, 0},
	}

	// the rows build on each other, thus they run in order on the same world
	for i, tt := range tests {
		w.clock.set(t, tt.at)
		w.mustStartNow(job, "Anna")
		runs := w.ended(job, i+2)
		if runs[len(runs)-1].Outcome != scheduler.OutcomeFailed {
			t.Fatalf("%s: Prämisse: der erste Lauf ist gescheitert:%s", tt.name, describeRuns(runs))
		}

		stats, _, err := w.m.Stats(job)
		if err != nil || stats.Total24h != tt.total || stats.Failed24h != tt.failed {
			t.Errorf("%s: Stats %+v %v, erwartet %d Läufe, %d gescheitert", tt.name, stats, err, tt.total, tt.failed)
		}

		status, err := scheduler.NewStatus(w.m)(user.SU(), job)
		if err != nil || status.Stats.Total24h != tt.total || status.Stats.Failed24h != tt.failed {
			t.Errorf("%s: Status %+v %v, erwartet %d Läufe, %d gescheitert", tt.name, status.Stats, err, tt.total, tt.failed)
		}

		if err == nil && (status.LastRun.IsNone() || status.LastRun.Unwrap().ID != runs[0].ID) {
			t.Errorf("%s: letzter Lauf %s erwartet: %+v", tt.name, runs[0].ID, status.LastRun)
		}
	}
}

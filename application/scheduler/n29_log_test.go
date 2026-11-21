// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"go.wdy.de/nago/application/scheduler"
)

// captureHandler records each record of slog.Default as a line of message and attributes.
type captureHandler struct {
	mu    *sync.Mutex
	lines *[]string
	attrs []slog.Attr
}

func (h captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h captureHandler) Handle(_ context.Context, r slog.Record) error {
	var sb strings.Builder
	sb.WriteString(r.Message)
	for _, a := range h.attrs {
		fmt.Fprintf(&sb, " %s=%v", a.Key, a.Value)
	}

	r.Attrs(func(a slog.Attr) bool {
		fmt.Fprintf(&sb, " %s=%v", a.Key, a.Value)
		return true
	})

	h.mu.Lock()
	defer h.mu.Unlock()
	*h.lines = append(*h.lines, sb.String())
	return nil
}

func (h captureHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.attrs = append(slices.Clone(h.attrs), attrs...)
	return h
}

func (h captureHandler) WithGroup(string) slog.Handler { return h }

// captureDefault replaces slog.Default until the end of the test and returns the recorded lines.
func captureDefault(t *testing.T) func() []string {
	var mu sync.Mutex
	var lines []string
	old := slog.Default()
	slog.SetDefault(slog.New(captureHandler{mu: &mu, lines: &lines}))
	t.Cleanup(func() { slog.SetDefault(old) })
	return func() []string {
		mu.Lock()
		defer mu.Unlock()
		return slices.Clone(lines)
	}
}

// TestN29Quiet checks, that with Quiet no entry of a run, neither a message nor its attributes nor the error of the
// run, reaches slog.Default, while the log of the run keeps the entry. Without Quiet the entries are forwarded as
// before.
//
// Zusagen: R-SCHEDULE-LOG, R-DEC-SCHEDULE-NAGO
func TestN29Quiet(t *testing.T) {
	tests := []struct {
		name      string
		quiet     bool
		forwarded bool
	}{
		{"Quiet: nichts im Log des Servers", true, false},
		{"Gegenzeile ohne Quiet: weitergereicht wie bisher", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := captureDefault(t)
			w := newWorld(t, bln(2027, 6, 1, 21, 0))
			runner := func(ctx context.Context) error {
				scheduler.LoggerFrom(ctx).Info("n29-marker-info", "k", "n29-marker-attr")
				return errors.New("n29-marker-error")
			}

			w.configure(cronOpts(job, runner, 22, 0, quiet(tt.quiet)))
			w.mustStartNow(job, "Anna")
			w.ended(job, 1)
			w.fire(job, bln(2027, 6, 1, 22, 0))
			runs := w.ended(job, 2)
			w.waitPlanned(job, bln(2027, 6, 2, 22, 0))

			for _, r := range runs {
				if r.Outcome != scheduler.OutcomeFailed {
					t.Fatalf("Prämisse: der Lauf scheitert:%s", describeRuns(runs))
				}

				page, err := w.m.RunLog(job, r.ID, scheduler.LogQuery{MinLevel: slog.LevelDebug, Text: "n29-marker-info"})
				if err != nil || page.Total != 1 {
					t.Fatalf("Eintrag im Protokoll von %s erwartet: %v %+v", r.ID, err, page)
				}
			}

			var leaked []string
			for _, l := range lines() {
				if strings.Contains(l, "n29-marker") {
					leaked = append(leaked, l)
				}
			}

			if tt.forwarded {
				if !slices.ContainsFunc(leaked, func(l string) bool { return strings.Contains(l, "n29-marker-info") }) {
					t.Fatalf("weitergereichter Eintrag erwartet, im Log des Servers: %q", lines())
				}
			} else if len(leaked) > 0 {
				t.Fatalf("nichts im Log des Servers erwartet: %q", leaked)
			}
		})
	}
}

// TestN29LogNotWritable checks, that a run fails with „Protokoll nicht schreibbar“, if its log cannot be written,
// here because the log directory is a file. Without the new switches the run succeeds as before.
//
// Zusagen: R-SCHEDULE-LOG, R-DEC-SCHEDULE-NAGO
func TestN29LogNotWritable(t *testing.T) {
	tests := []struct {
		name    string
		broken  bool
		mod     func(*scheduler.Options)
		execute bool // ExecuteNow as before instead of StartNow
		outcome scheduler.Outcome
	}{
		{"Quiet und Durable: Protokoll nicht schreibbar", true, quiet(true), false, scheduler.OutcomeFailed},
		{"Gegenzeile Quiet und Durable: Protokoll schreibbar", false, quiet(true), false, scheduler.OutcomeSucceeded},
		{"Gegenzeile ohne die neuen Schalter: gelingt wie bisher", true, oldStyle, true, scheduler.OutcomeSucceeded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logDir := t.TempDir()
			if tt.broken {
				logDir = filepath.Join(logDir, "kein-ordner")
				if err := os.WriteFile(logDir, []byte("n29"), 0600); err != nil {
					t.Fatal(err)
				}
			}

			w := newWorldIn(t, bln(2027, 6, 1, 21, 0), logDir)
			p := newProbe(w.clock)
			w.configure(cronOpts(job, p.run, 22, 0, tt.mod))
			if tt.execute {
				if err := w.m.ExecuteNow(job); err != nil {
					t.Fatal(err)
				}
			} else {
				w.mustStartNow(job, "Anna")
			}

			runs := w.ended(job, 1)
			r := runs[0]
			if r.Outcome != tt.outcome {
				t.Fatalf("Ergebnis %d erwartet:%s", tt.outcome, describeRuns(runs))
			}

			if notWritable := strings.Contains(r.Error, "Protokoll nicht schreibbar"); notWritable != (tt.outcome == scheduler.OutcomeFailed) {
				t.Fatalf("Fehler „Protokoll nicht schreibbar“ erwartet %v:%s", tt.outcome == scheduler.OutcomeFailed, describeRuns(runs))
			}
		})
	}
}

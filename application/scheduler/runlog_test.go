// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
)

func newTestManager(t *testing.T, dir string) (*Manager, RunRepository) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	settings := json.NewSloppyJSONRepository[Settings, ID](mem.NewBlobStore("settings"))
	runs := json.NewSloppyJSONRepository[Run, RunID](mem.NewBlobStore("runs"))
	return NewManagerWithPersistence(ctx, settings, runs, dir), runs
}

func configureManual(t *testing.T, m *Manager, id ID, runner func(ctx context.Context) error) {
	t.Helper()
	if err := m.Configure(Options{ID: id, Name: string(id), Kind: Manual, Runner: runner}); err != nil {
		t.Fatal(err)
	}
}

func logFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	var res []string
	for _, e := range entries {
		res = append(res, e.Name())
	}
	return res
}

func TestRunLogIsWrittenToOwnFile(t *testing.T) {
	dir := t.TempDir()
	m, _ := newTestManager(t, dir)
	configureManual(t, m, "my.job", func(ctx context.Context) error {
		LoggerFrom(ctx).Info("hello", "count", 3)
		LoggerFrom(ctx).WithGroup("smtp").Warn("slow", "host", "a")
		return nil
	})

	if err := m.ExecuteNow("my.job"); err != nil {
		t.Fatal(err)
	}

	day := time.Now().Format(time.DateOnly)
	want := fmt.Sprintf("my.job-%s-00001.log", day)
	if files := logFiles(t, dir); !slices.Equal(files, []string{want}) {
		t.Fatalf("got files %v, want %v", files, want)
	}

	runs, err := m.Runs("my.job")
	if err != nil || len(runs) != 1 {
		t.Fatalf("runs=%v err=%v", runs, err)
	}

	run := runs[0]
	if run.Outcome != OutcomeSucceeded || run.Entries != 2 || run.Warnings != 1 || run.LogFile != want || run.Nr != 1 {
		t.Fatalf("unexpected run %+v", run)
	}

	page, err := m.RunLog("my.job", run.ID, LogQuery{MinLevel: slog.LevelDebug})
	if err != nil {
		t.Fatal(err)
	}

	if page.Total != 2 || page.Entries[0].Msg != "slow" || page.Entries[0].Values["smtp.host"] != "a" || page.Entries[1].Values["count"] != float64(3) {
		t.Fatalf("unexpected page %+v", page)
	}
}

func TestRunNumbersContinueAfterRestart(t *testing.T) {
	dir := t.TempDir()
	settings := json.NewSloppyJSONRepository[Settings, ID](mem.NewBlobStore("settings"))
	runs := json.NewSloppyJSONRepository[Run, RunID](mem.NewBlobStore("runs"))
	for i := 0; i < 2; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		m := NewManagerWithPersistence(ctx, settings, runs, dir)
		configureManual(t, m, "job", func(ctx context.Context) error { return nil })
		_ = m.ExecuteNow("job")
		cancel()
	}

	day := time.Now().Format(time.DateOnly)
	want := []string{fmt.Sprintf("job-%s-00001.log", day), fmt.Sprintf("job-%s-00002.log", day)}
	if files := logFiles(t, dir); !slices.Equal(files, want) {
		t.Fatalf("got %v, want %v", files, want)
	}
}

func TestOnlyNewestLogsAreKept(t *testing.T) {
	dir := t.TempDir()
	m, _ := newTestManager(t, dir)
	configureManual(t, m, "job", func(ctx context.Context) error {
		LoggerFrom(ctx).Info("x")
		return nil
	})

	for i := 0; i < KeepRunLogs+2; i++ {
		_ = m.ExecuteNow("job")
	}

	if files := logFiles(t, dir); len(files) != KeepRunLogs {
		t.Fatalf("expected %d files, got %v", KeepRunLogs, files)
	}

	runs, _ := m.Runs("job")
	if len(runs) != KeepRunLogs+2 {
		t.Fatalf("statistics must be kept, got %d runs", len(runs))
	}

	for i, run := range runs {
		hasLog := run.LogFile != ""
		if hasLog != (i < KeepRunLogs) {
			t.Fatalf("run %d (%s) hasLog=%v", i, run.ID, hasLog)
		}
	}

	if _, err := m.RunLog("job", runs[len(runs)-1].ID, LogQuery{}); !errors.Is(err, ErrRunLogNotAvailable) {
		t.Fatalf("expected ErrRunLogNotAvailable for pruned log, got %v", err)
	}

	if _, err := m.RunLog("job", "job#unknown", LogQuery{}); !errors.Is(err, ErrRunNotFound) {
		t.Fatalf("expected ErrRunNotFound, got %v", err)
	}
}

func TestOldRunsAreDeleted(t *testing.T) {
	dir := t.TempDir()
	m, repo := newTestManager(t, dir)
	configureManual(t, m, "job", func(ctx context.Context) error { return nil })

	old := Run{ID: "job#2000-01-01-00001", Scheduler: "job", Nr: 1, StartedAt: time.Now().Add(-RunRetention - time.Hour), Outcome: OutcomeSucceeded}
	if err := repo.Save(old); err != nil {
		t.Fatal(err)
	}

	_ = m.ExecuteNow("job")
	runs, _ := m.Runs("job")
	if len(runs) != 1 || runs[0].ID == old.ID {
		t.Fatalf("old run not deleted: %v", runs)
	}
}

func TestOutcomeAndLastError(t *testing.T) {
	m, _ := newTestManager(t, t.TempDir())
	configureManual(t, m, "fail", func(ctx context.Context) error { return errors.New("boom") })
	configureManual(t, m, "panic", func(ctx context.Context) error { panic("argh") })

	_ = m.ExecuteNow("fail")
	_ = m.ExecuteNow("panic")

	for id, want := range map[ID]Outcome{"fail": OutcomeFailed, "panic": OutcomePanicked} {
		runs, _ := m.Runs(id)
		if len(runs) != 1 || runs[0].Outcome != want || runs[0].Errors != 1 {
			t.Fatalf("%s: %+v", id, runs)
		}
		if m.LastError(id) == nil {
			t.Fatalf("%s: last error not set", id)
		}
	}

	st, _, _ := m.Stats("fail")
	if st.Total24h != 1 || st.Failed24h != 1 {
		t.Fatalf("stats %+v", st)
	}
}

func TestSafeFileName(t *testing.T) {
	for in, want := range map[string]string{
		"../../etc/passwd": "_.._etc_passwd",
		"a b/c":            "a_b_c",
		"nago.mail.outbox": "nago.mail.outbox",
		"..":               "_",
	} {
		if got := safeFileName(in); got != want {
			t.Errorf("%q: got %q want %q", in, got, want)
		}
	}
}

func seqOf(n int, level func(i int) slog.Level) func() iter.Seq[LogEntry] {
	return func() iter.Seq[LogEntry] {
		return func(yield func(LogEntry) bool) {
			for i := 0; i < n; i++ {
				if !yield(LogEntry{Level: level(i), Msg: fmt.Sprintf("m%d", i)}) {
					return
				}
			}
		}
	}
}

func TestPageEntries(t *testing.T) {
	info := func(int) slog.Level { return slog.LevelInfo }

	p := pageEntries(seqOf(0, info), LogQuery{})
	if p.Total != 0 || len(p.Entries) != 0 {
		t.Fatalf("%+v", p)
	}

	p = pageEntries(seqOf(250, info), LogQuery{Limit: 1000})
	if p.Total != 250 || len(p.Entries) != MaxLogPageSize || p.Entries[0].Msg != "m249" || p.Entries[MaxLogPageSize-1].Msg != fmt.Sprintf("m%d", 250-MaxLogPageSize) {
		t.Fatalf("first page: total=%d len=%d", p.Total, len(p.Entries))
	}

	p = pageEntries(seqOf(250, info), LogQuery{Offset: 220, Limit: 1000})
	if len(p.Entries) != 30 || p.Entries[0].Msg != "m29" || p.Entries[29].Msg != "m0" {
		t.Fatalf("last page: len=%d", len(p.Entries))
	}

	p = pageEntries(seqOf(250, info), LogQuery{Offset: 300})
	if p.Total != 250 || len(p.Entries) != 0 {
		t.Fatalf("beyond: %+v", p)
	}

	mixed := func(i int) slog.Level {
		if i%10 == 0 {
			return slog.LevelError
		}
		return slog.LevelDebug
	}

	p = pageEntries(seqOf(100, mixed), LogQuery{MinLevel: slog.LevelError})
	if p.Total != 10 || p.Entries[0].Msg != "m90" {
		t.Fatalf("level: %+v", p)
	}

	p = pageEntries(seqOf(100, mixed), LogQuery{MinLevel: slog.LevelDebug, Text: "M4"})
	if p.Total != 11 {
		t.Fatalf("text: %d", p.Total)
	}
}

func TestBrokenLinesAreSkipped(t *testing.T) {
	dir := t.TempDir()
	rs := newRunStore(json.NewSloppyJSONRepository[Run, RunID](mem.NewBlobStore("runs")), dir)
	content := `{"time":"2026-09-24T10:00:00Z","level":"INFO","msg":"ok"}` + "\n" + `{"time":"2026-09-24T10:0` + "\n"
	if err := os.WriteFile(filepath.Join(dir, "x.log"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	entries, err := rs.fileEntries(Run{ID: "x", LogFile: "x.log"})
	if err != nil {
		t.Fatal(err)
	}

	if got := slices.Collect(entries); len(got) != 1 || got[0].Msg != "ok" {
		t.Fatalf("%+v", got)
	}
}

func TestNewUseCasesStaysCompatible(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := NewManager(ctx, json.NewSloppyJSONRepository[Settings, ID](mem.NewBlobStore("settings")))
	configureManual(t, m, "job", func(ctx context.Context) error {
		LoggerFrom(ctx).Info("hello")
		return nil
	})

	if err := m.ExecuteNow("job"); err != nil {
		t.Fatal(err)
	}

	logs := m.Logs("job")
	if len(logs) != 1 || logs[0].Msg != "hello" {
		t.Fatalf("in-memory logs: %+v", logs)
	}

	runs, _ := m.Runs("job")
	if len(runs) != 1 || runs[0].LogFile != "" {
		t.Fatalf("runs: %+v", runs)
	}
}

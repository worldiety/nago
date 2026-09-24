// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

// runStore persists run statistics in a repository and the log of each run into its own file.
// If dir is empty, no log files are written.
type runStore struct {
	repo RunRepository
	dir  string
	now  func() time.Time
	mu   sync.Mutex
}

func newRunStore(repo RunRepository, dir string) *runStore {
	return &runStore{repo: repo, dir: dir, now: time.Now}
}

// safeFileName replaces everything which is not safe in a file name. Security: avoids path traversal.
func safeFileName(s string) string {
	var sb strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-', r == '_':
			sb.WriteRune(r)
		default:
			sb.WriteRune('_')
		}
	}

	res := strings.Trim(sb.String(), ".")
	if res == "" {
		res = "_"
	}

	return res
}

func runIDPrefix(id ID) RunID {
	return RunID(string(id) + "#")
}

// runs returns all runs of the given scheduler, newest first.
func (rs *runStore) runs(id ID) ([]Run, error) {
	var res []Run
	for run, err := range rs.repo.FindAllByPrefix(runIDPrefix(id)) {
		if err != nil {
			return nil, err
		}

		if run.Scheduler == id {
			res = append(res, run)
		}
	}

	slices.SortFunc(res, func(a, b Run) int {
		return b.StartedAt.Compare(a.StartedAt)
	})

	return res, nil
}

func (rs *runStore) begin(id ID) (Run, *runSink, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	now := rs.now()
	day := now.Format(time.DateOnly)
	dayPrefix := string(runIDPrefix(id)) + day + "-"

	nr := 0
	for run, err := range rs.repo.FindAllByPrefix(RunID(dayPrefix)) {
		if err != nil {
			return Run{}, nil, err
		}

		if run.Scheduler == id {
			nr = max(nr, run.Nr)
		}
	}
	nr++

	name := fmt.Sprintf("%s-%s-%05d", safeFileName(string(id)), day, nr)
	run := Run{
		ID:        RunID(fmt.Sprintf("%s%05d", dayPrefix, nr)),
		Scheduler: id,
		Nr:        nr,
		StartedAt: now,
		Outcome:   OutcomeRunning,
	}

	sink := &runSink{}
	if rs.dir != "" {
		if err := os.MkdirAll(rs.dir, 0700); err != nil {
			slog.Error("cannot create scheduler log directory", "dir", rs.dir, "err", err.Error())
		} else {
			file := name + ".log"
			f, err := os.OpenFile(filepath.Join(rs.dir, file), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
			if err != nil {
				slog.Error("cannot create scheduler log file", "file", file, "err", err.Error())
			} else {
				sink.file = f
				run.LogFile = file
			}
		}
	}

	if err := rs.repo.Save(run); err != nil {
		if sink.file != nil {
			_ = sink.file.Close()
		}
		return Run{}, nil, err
	}

	return run, sink, nil
}

func (rs *runStore) end(run Run, sink *runSink, runErr error) {
	entries, warns, errs := sink.close()

	run.CompletedAt = rs.now()
	run.Entries = entries
	run.Warnings = warns
	run.Errors = errs
	run.Outcome = OutcomeSucceeded
	if runErr != nil {
		run.Error = runErr.Error()
		var perr *PanicError
		switch {
		case errors.As(runErr, &perr):
			run.Outcome = OutcomePanicked
		case isCanceled(runErr):
			run.Outcome = OutcomeCanceled
		default:
			run.Outcome = OutcomeFailed
		}
	}

	rs.mu.Lock()
	defer rs.mu.Unlock()

	if err := rs.repo.Save(run); err != nil {
		slog.Error("cannot save scheduler run", "id", run.ID, "err", err.Error())
	}

	rs.prune(run.Scheduler)
}

// prune deletes all log files except the newest KeepRunLogs and all runs older than RunRetention.
func (rs *runStore) prune(id ID) {
	runs, err := rs.runs(id)
	if err != nil {
		slog.Error("cannot prune scheduler runs", "id", id, "err", err.Error())
		return
	}

	deadline := rs.now().Add(-RunRetention)
	withLog := 0
	for _, run := range runs {
		if run.StartedAt.Before(deadline) {
			rs.deleteLog(run)
			if err := rs.repo.DeleteByID(run.ID); err != nil {
				slog.Error("cannot delete scheduler run", "id", run.ID, "err", err.Error())
			}
			continue
		}

		if run.LogFile == "" {
			continue
		}

		withLog++
		if withLog <= KeepRunLogs {
			continue
		}

		rs.deleteLog(run)
		run.LogFile = ""
		if err := rs.repo.Save(run); err != nil {
			slog.Error("cannot save scheduler run", "id", run.ID, "err", err.Error())
		}
	}
}

func (rs *runStore) deleteLog(run Run) {
	if run.LogFile == "" || rs.dir == "" {
		return
	}

	if err := os.Remove(filepath.Join(rs.dir, filepath.Base(run.LogFile))); err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Error("cannot delete scheduler log file", "file", run.LogFile, "err", err.Error())
	}
}

// fileEntries reads the log file of the given run line by line. Broken lines (e.g. a partially written
// last line of a running execution) are skipped.
func (rs *runStore) fileEntries(run Run) (iter.Seq[LogEntry], error) {
	if run.LogFile == "" || rs.dir == "" {
		return nil, fmt.Errorf("run %s: %w", run.ID, ErrRunLogNotAvailable)
	}

	path := filepath.Join(rs.dir, filepath.Base(run.LogFile))
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("run %s: %w: %w", run.ID, ErrRunLogNotAvailable, err)
	}

	return func(yield func(LogEntry) bool) {
		f, err := os.Open(path)
		if err != nil {
			slog.Error("cannot open scheduler log file", "file", path, "err", err.Error())
			return
		}
		defer f.Close()

		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
		for sc.Scan() {
			var e fileEntry
			if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
				continue
			}

			if !yield(e.intoEntry()) {
				return
			}
		}
	}, nil
}

// fileEntry is the persistence model of a single JSON log line.
type fileEntry struct {
	Time   time.Time      `json:"time"`
	Level  slog.Level     `json:"level"`
	Msg    string         `json:"msg"`
	Values map[string]any `json:"attrs,omitempty"`
}

func (e fileEntry) intoEntry() LogEntry {
	return LogEntry{Level: e.Level, Time: e.Time, Msg: e.Msg, Values: e.Values}
}

// runSink receives the log entries of a single run.
type runSink struct {
	mu       sync.Mutex
	file     *os.File
	entries  int
	warnings int
	errors   int
	closed   bool
}

func (s *runSink) write(e LogEntry) {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	s.entries++
	switch {
	case e.Level >= slog.LevelError:
		s.errors++
	case e.Level >= slog.LevelWarn:
		s.warnings++
	}

	if s.file == nil {
		return
	}

	buf, err := json.Marshal(fileEntry{Time: e.Time, Level: e.Level, Msg: e.Msg, Values: jsonSafe(e.Values)})
	if err != nil {
		buf, _ = json.Marshal(fileEntry{Time: e.Time, Level: e.Level, Msg: e.Msg, Values: map[string]any{"marshalError": err.Error()}})
	}

	buf = append(buf, '\n')
	if _, err := s.file.Write(buf); err != nil {
		slog.Error("cannot write scheduler log file", "file", s.file.Name(), "err", err.Error())
	}
}

func (s *runSink) close() (entries, warnings, errs int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.closed && s.file != nil {
		_ = s.file.Close()
	}

	s.closed = true
	return s.entries, s.warnings, s.errors
}

// jsonSafe converts values which cannot be marshalled (e.g. errors) into strings.
func jsonSafe(values map[string]any) map[string]any {
	if len(values) == 0 {
		return nil
	}

	res := make(map[string]any, len(values))
	for k, v := range values {
		switch t := v.(type) {
		case error:
			res[k] = t.Error()
		case fmt.Stringer:
			res[k] = t.String()
		default:
			if _, err := json.Marshal(v); err != nil {
				res[k] = fmt.Sprint(v)
			} else {
				res[k] = v
			}
		}
	}

	return res
}

// pageEntries returns the requested page of the given entries (in chronological order) newest first.
// The source is iterated twice, so that at most one page is held in memory.
func pageEntries(src func() iter.Seq[LogEntry], q LogQuery) LogPage {
	limit := min(max(q.Limit, 1), MaxLogPageSize)
	if q.Limit <= 0 {
		limit = MaxLogPageSize
	}
	offset := max(q.Offset, 0)
	text := strings.ToLower(strings.TrimSpace(q.Text))

	match := func(e LogEntry) bool {
		if e.Level < q.MinLevel {
			return false
		}

		if text == "" {
			return true
		}

		if strings.Contains(strings.ToLower(e.Msg), text) {
			return true
		}

		for k, v := range e.Values {
			if strings.Contains(strings.ToLower(k+"="+fmt.Sprint(v)), text) {
				return true
			}
		}

		return false
	}

	total := 0
	for e := range src() {
		if match(e) {
			total++
		}
	}

	// newest first: page [offset, offset+limit) corresponds to chronological [total-offset-limit, total-offset)
	from := max(total-offset-limit, 0)
	to := total - offset
	page := LogPage{Total: total, Offset: offset}
	if to <= 0 {
		return page
	}

	i := 0
	for e := range src() {
		if !match(e) {
			continue
		}

		if i >= from && i < to {
			page.Entries = append(page.Entries, e)
		}

		i++
		if i >= to {
			break
		}
	}

	// newest first, also if concurrent writers interleaved the file order within the page
	slices.Reverse(page.Entries)
	slices.SortStableFunc(page.Entries, func(a, b LogEntry) int {
		return b.Time.Compare(a.Time)
	})

	return page
}

// stats aggregates runs (newest first) of the last 24 hours relative to now.
func stats(runs []Run, now time.Time) RunStats {
	var res RunStats
	var sum time.Duration
	var completed int
	for i, run := range runs {
		if i == 0 || (res.LastDuration == 0 && run.Outcome != OutcomeRunning) {
			if run.Outcome != OutcomeRunning {
				res.LastDuration = run.Duration()
			}
		}

		if run.StartedAt.Before(now.Add(-24 * time.Hour)) {
			continue
		}

		res.Total24h++
		if run.Outcome.Failed() {
			res.Failed24h++
		}

		if run.Outcome != OutcomeRunning {
			sum += run.Duration()
			completed++
		}
	}

	if completed > 0 {
		res.AvgDuration = sum / time.Duration(completed)
	}

	return res
}

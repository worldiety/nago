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
	"strconv"
	"strings"
	"sync"
	"time"
)

// runStore persists run statistics in a repository and the log of each run into its own file.
// If dir is empty, no log files are written.
//
// The key of a run contains its day and its number (see [RunID]), so that the store finds the runs of a period,
// a page or the oldest runs by their keys, without reading all runs.
type runStore struct {
	repo  RunRepository
	dir   string
	clock Clock
	mu    sync.Mutex
	// floors is the oldest run per scheduler, which may still have a log file. A missing entry means, that it is
	// unknown, e.g. after a restart of a scheduler without [Options.Durable].
	floors map[ID]RunID
	// epochs are incremented by purge, so that the late end of a purged run does not save it again.
	epochs  map[ID]uint64
	indexes logIndexCache
}

func newRunStore(repo RunRepository, dir string) *runStore {
	return newRunStoreWith(repo, dir, nil)
}

func newRunStoreWith(repo RunRepository, dir string, clock Clock) *runStore {
	return &runStore{
		repo:    repo,
		dir:     dir,
		clock:   clockOrSystem(clock),
		floors:  map[ID]RunID{},
		epochs:  map[ID]uint64{},
		indexes: logIndexCache{files: map[string]*logIndex{}},
	}
}

func (rs *runStore) now() time.Time {
	return rs.clock.Now()
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

// stateKey is the key of the plan state of a durable scheduler, see [DurableState]. It is no run key.
func stateKey(id ID) RunID {
	return RunID(string(id) + "#state")
}

// runKeySuffix returns the part of a run key after the last '#', e.g. 2026-09-24-00001.
func runKeySuffix(key RunID) string {
	s := string(key)
	return s[strings.LastIndexByte(s, '#')+1:]
}

// isRunKeySuffix accepts a day and a number, e.g. 2026-09-24-00001.
func isRunKeySuffix(s string) bool {
	if len(s) < len("2006-01-02-00001") || s[10] != '-' {
		return false
	}

	for i := 0; i < 10; i++ {
		c := s[i]
		if i == 4 || i == 7 {
			if c != '-' {
				return false
			}
			continue
		}

		if c < '0' || c > '9' {
			return false
		}
	}

	return isDigits(s[11:])
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}

	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}

	return true
}

// keyDay returns the day of a run key in the format of [time.DateOnly].
func keyDay(key RunID) string {
	return runKeySuffix(key)[:10]
}

func keyNr(key RunID) int {
	n, _ := strconv.Atoi(runKeySuffix(key)[11:])
	return n
}

// compareRunKeys orders the keys of one scheduler by day and number.
func compareRunKeys(a, b RunID) int {
	if c := strings.Compare(keyDay(a), keyDay(b)); c != 0 {
		return c
	}

	return keyNr(a) - keyNr(b)
}

// atOrAfterFloor reports, if the run may still have a log file. The empty floor is before every run.
func atOrAfterFloor(key, floor RunID) bool {
	return floor == "" || compareRunKeys(key, floor) >= 0
}

// logFileName returns the name of the log file of the run with the given key.
func logFileName(id ID, key RunID) string {
	return safeFileName(string(id)) + "-" + runKeySuffix(key) + ".log"
}

// keys returns the keys of the runs of the given scheduler, oldest first, without reading the runs.
func (rs *runStore) keys(id ID) ([]RunID, error) {
	prefix := string(runIDPrefix(id))
	var res []RunID
	sorted := true
	for key, err := range rs.repo.IdentifiersByPrefix(RunID(prefix)) {
		if err != nil {
			return nil, err
		}

		if !strings.HasPrefix(string(key), prefix) || !isRunKeySuffix(string(key)[len(prefix):]) {
			// e.g. the plan state or the run of another scheduler, whose id starts with this one
			continue
		}

		if n := len(res); n > 0 && compareRunKeys(res[n-1], key) > 0 {
			sorted = false
		}

		res = append(res, key)
	}

	if !sorted {
		// only numbers with more than 5 digits break the lexicographic order
		slices.SortFunc(res, compareRunKeys)
	}

	return res, nil
}

func (rs *runStore) load(key RunID) (Run, bool, error) {
	opt, err := rs.repo.FindByID(key)
	if err != nil {
		return Run{}, false, err
	}

	if opt.IsNone() {
		return Run{}, false, nil
	}

	return opt.Unwrap(), true, nil
}

// runs returns all runs of the given scheduler, newest first.
func (rs *runStore) runs(id ID) ([]Run, error) {
	keys, err := rs.keys(id)
	if err != nil {
		return nil, err
	}

	res := make([]Run, 0, len(keys))
	for run, err := range rs.repo.FindAllByID(slices.Values(keys)) {
		if err != nil {
			return nil, err
		}

		res = append(res, run)
	}

	slices.SortFunc(res, func(a, b Run) int {
		return b.StartedAt.Compare(a.StartedAt)
	})

	return res, nil
}

// page returns the runs of the given page, newest first, and the amount of all runs. It reads only the runs of
// the page.
func (rs *runStore) page(id ID, offset, limit int) ([]Run, int, error) {
	keys, err := rs.keys(id)
	if err != nil {
		return nil, 0, err
	}

	total := len(keys)
	offset = max(offset, 0)
	if limit <= 0 || offset >= total {
		return nil, total, nil
	}

	hi := total - offset
	lo := max(hi-limit, 0)
	res := make([]Run, 0, hi-lo)
	for i := hi - 1; i >= lo; i-- {
		run, ok, err := rs.load(keys[i])
		if err != nil {
			return nil, 0, err
		}

		if ok {
			res = append(res, run)
		}
	}

	return res, total, nil
}

// maxRunsBeyondWindow bounds the runs, which recent reads beyond the last 24 hours to find the latest completed run.
const maxRunsBeyondWindow = 10

// recent returns the runs of the last 24 hours and at least the latest completed run, newest first. It reads only
// the runs of the days of the last 24 hours.
func (rs *runStore) recent(id ID, loc *time.Location) ([]Run, error) {
	keys, err := rs.keys(id)
	if err != nil {
		return nil, err
	}

	cutoffDay := rs.now().Add(-24 * time.Hour).In(loc).Format(time.DateOnly)
	var res []Run
	completed := false
	beyond := 0
	for i := len(keys) - 1; i >= 0; i-- {
		if keyDay(keys[i]) < cutoffDay {
			if completed || beyond >= maxRunsBeyondWindow {
				break
			}
			beyond++
		}

		run, ok, err := rs.load(keys[i])
		if err != nil {
			return nil, err
		}

		if !ok {
			continue
		}

		res = append(res, run)
		if run.Outcome != OutcomeRunning {
			completed = true
		}
	}

	slices.SortStableFunc(res, func(a, b Run) int {
		return b.StartedAt.Compare(a.StartedAt)
	})

	return res, nil
}

// newest returns the newest run of the given scheduler.
func (rs *runStore) newest(id ID) (Run, bool, error) {
	runs, _, err := rs.page(id, 0, 1)
	if err != nil || len(runs) == 0 {
		return Run{}, false, err
	}

	return runs[0], true, nil
}

// oldestStart returns the start of the oldest run of the given scheduler or the zero time.
func (rs *runStore) oldestStart(id ID) (time.Time, error) {
	keys, err := rs.keys(id)
	if err != nil || len(keys) == 0 {
		return time.Time{}, err
	}

	run, ok, err := rs.load(keys[0])
	if err != nil || !ok {
		return time.Time{}, err
	}

	return run.StartedAt, nil
}

// loadState returns the plan state of a durable scheduler.
func (rs *runStore) loadState(id ID) (DurableState, bool, error) {
	run, ok, err := rs.load(stateKey(id))
	if err != nil || !ok || run.Durable == nil {
		return DurableState{}, false, err
	}

	return *run.Durable, true, nil
}

func (rs *runStore) saveState(id ID, state DurableState) error {
	return rs.repo.Save(Run{ID: stateKey(id), Scheduler: id, Outcome: OutcomeSucceeded, Durable: &state})
}

// setFloor sets the oldest run, which may still have a log file.
func (rs *runStore) setFloor(id ID, floor RunID) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.floors[id] = floor
}

// closeUnfinished ends the newest runs without an end, e.g. after a crash, as cancelled by a restart. It returns
// the newest run, if any.
func (rs *runStore) closeUnfinished(id ID) (Run, bool, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	keys, err := rs.keys(id)
	if err != nil {
		return Run{}, false, err
	}

	var newest Run
	found := false
	for i := len(keys) - 1; i >= 0; i-- {
		run, ok, err := rs.load(keys[i])
		if err != nil {
			return Run{}, false, err
		}

		if !ok {
			continue
		}

		unfinished := run.Outcome == OutcomeRunning
		if unfinished {
			run.Outcome = OutcomeCanceled
			run.CanceledBy = CanceledByRestart
			if err := rs.repo.Save(run); err != nil {
				return Run{}, false, err
			}
		}

		if !found {
			newest, found = run, true
		}

		// only the newest runs can have no end, because runs never overlap
		if !unfinished {
			break
		}
	}

	return newest, found, nil
}

// begin persists a new run, which starts now, and opens its log file. If strictLog is set and the log file
// cannot be created, the returned sink has failed, see [Options.Quiet].
func (rs *runStore) begin(id ID, info RunInfo, now time.Time, loc *time.Location, strictLog bool) (Run, *runSink, uint64, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	day := now.In(loc).Format(time.DateOnly)
	dayPrefix := string(runIDPrefix(id)) + day + "-"

	nr := 0
	for key, err := range rs.repo.IdentifiersByPrefix(RunID(dayPrefix)) {
		if err != nil {
			return Run{}, nil, 0, err
		}

		if suffix := string(key)[len(dayPrefix):]; isDigits(suffix) {
			n, _ := strconv.Atoi(suffix)
			nr = max(nr, n)
		}
	}
	nr++

	key := RunID(fmt.Sprintf("%s%05d", dayPrefix, nr))
	run := Run{
		ID:        key,
		Scheduler: id,
		Nr:        nr,
		StartedAt: now,
		Outcome:   OutcomeRunning,
		Trigger:   info.Trigger,
		By:        info.By,
		Due:       info.Due,
	}

	sink := &runSink{strict: strictLog}
	if rs.dir != "" {
		file := logFileName(id, key)
		if err := os.MkdirAll(rs.dir, 0700); err != nil {
			slog.Error("cannot create scheduler log directory", "dir", rs.dir, "err", err.Error())
			sink.failed = strictLog
		} else {
			path := filepath.Join(rs.dir, file)
			f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
			if err != nil {
				slog.Error("cannot create scheduler log file", "file", file, "err", err.Error())
				sink.failed = strictLog
			} else {
				sink.file = f
				sink.path = path
				if fi, err := f.Stat(); err == nil {
					sink.info = fi
				}
				run.LogFile = file
				rs.indexes.forget(path)
			}
		}
	}

	if err := rs.repo.Save(run); err != nil {
		if sink.file != nil {
			_ = sink.file.Close()
		}
		return Run{}, nil, 0, err
	}

	return run, sink, rs.epochs[id], nil
}

// end persists the end of the given run and prunes the runs of its scheduler. It returns the start of its oldest
// remaining run and the oldest run which may still have a log file.
func (rs *runStore) end(run Run, sink *runSink, epoch uint64, keep int, loc *time.Location) (time.Time, RunID, bool) {
	entries, warns, errs := sink.close()
	run.Entries = entries
	run.Warnings = warns
	run.Errors = errs

	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.epochs[run.Scheduler] != epoch {
		// the runs have been purged while this run was in progress
		rs.deleteLog(run)
		if err := rs.repo.DeleteByID(run.ID); err != nil {
			slog.Error("cannot delete scheduler run", "id", run.ID, "err", err.Error())
		}
		return time.Time{}, "", false
	}

	if err := rs.repo.Save(run); err != nil {
		slog.Error("cannot save scheduler run", "id", run.ID, "err", err.Error())
	}

	return rs.prune(run.Scheduler, keep, loc)
}

// pruneRuns is prune with the lock.
func (rs *runStore) pruneRuns(id ID, keep int, loc *time.Location) (time.Time, RunID, bool) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	return rs.prune(id, keep, loc)
}

// prune deletes the runs older than [RunRetention] together with their log files and the log files of all runs
// except the newest keep. It reads the keys, the runs of the day which reaches the retention and the runs whose
// log file it deletes, not all runs. It returns the start of the oldest remaining run (zero, if there is none)
// and the oldest run which may still have a log file. It requires the lock.
func (rs *runStore) prune(id ID, keep int, loc *time.Location) (oldest time.Time, floor RunID, floorKnown bool) {
	keys, err := rs.keys(id)
	if err != nil {
		slog.Error("cannot prune scheduler runs", "id", id, "err", err.Error())
		return time.Time{}, "", false
	}

	knownFloor, known := rs.floors[id]
	hasLog := func(key RunID) bool {
		return !known || atOrAfterFloor(key, knownFloor)
	}

	deadline := rs.now().Add(-RunRetention)
	cutoffDay := deadline.In(loc).Format(time.DateOnly)
	expired := 0
	for ; expired < len(keys); expired++ {
		key := keys[expired]
		day := keyDay(key)
		if day > cutoffDay {
			break
		}

		if day == cutoffDay {
			// the runs of a day start in the order of their numbers
			run, ok, err := rs.load(key)
			if err != nil {
				slog.Error("cannot read scheduler run", "id", key, "err", err.Error())
				break
			}

			if ok && !run.StartedAt.Before(deadline) {
				oldest = run.StartedAt
				break
			}
		}

		if hasLog(key) {
			rs.removeLogFile(id, key)
		}
	}

	if expired > 0 {
		if err := rs.repo.DeleteAllByID(slices.Values(keys[:expired])); err != nil {
			slog.Error("cannot delete scheduler runs", "id", id, "err", err.Error())
			return time.Time{}, "", false
		}

		keys = keys[expired:]
	}

	if oldest.IsZero() && len(keys) > 0 {
		if run, ok, err := rs.load(keys[0]); err == nil && ok {
			oldest = run.StartedAt
		}
	}

	keepFrom := max(len(keys)-max(keep, 1), 0)
	from := 0
	if known && knownFloor != "" {
		from, _ = slices.BinarySearchFunc(keys, knownFloor, compareRunKeys)
	}

	for i := from; i < keepFrom; i++ {
		run, ok, err := rs.load(keys[i])
		if err != nil {
			slog.Error("cannot read scheduler run", "id", keys[i], "err", err.Error())
			return oldest, "", false
		}

		if !ok || run.LogFile == "" {
			continue
		}

		rs.deleteLog(run)
		run.LogFile = ""
		if err := rs.repo.Save(run); err != nil {
			slog.Error("cannot save scheduler run", "id", run.ID, "err", err.Error())
		}
	}

	floor = ""
	if keepFrom < len(keys) {
		floor = keys[keepFrom]
	}

	rs.floors[id] = floor
	return oldest, floor, true
}

// purge deletes all runs, log files and the plan state of the given scheduler.
func (rs *runStore) purge(id ID) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.epochs[id]++

	keys, err := rs.keys(id)
	if err != nil {
		return err
	}

	floor, known := rs.floors[id]
	for _, key := range keys {
		if !known || atOrAfterFloor(key, floor) {
			rs.removeLogFile(id, key)
		}
	}

	if err := rs.repo.DeleteAllByID(slices.Values(append(keys, stateKey(id)))); err != nil {
		return err
	}

	rs.floors[id] = ""
	return nil
}

func (rs *runStore) deleteLog(run Run) {
	if run.LogFile == "" || rs.dir == "" {
		return
	}

	path := filepath.Join(rs.dir, filepath.Base(run.LogFile))
	rs.indexes.forget(path)
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Error("cannot delete scheduler log file", "file", run.LogFile, "err", err.Error())
	}
}

// removeLogFile deletes the log file of the run with the given key, if it exists.
func (rs *runStore) removeLogFile(id ID, key RunID) {
	if rs.dir == "" {
		return
	}

	rs.deleteLog(Run{LogFile: logFileName(id, key)})
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

// pageLog returns a page of the log file of the given run. Without a text filter, it reads only the lines of the
// page and the part of the file, which has been written since the last call, see [logIndexCache].
func (rs *runStore) pageLog(run Run, q LogQuery) (LogPage, error) {
	if strings.TrimSpace(q.Text) != "" {
		entries, err := rs.fileEntries(run)
		if err != nil {
			return LogPage{}, err
		}

		return pageEntries(func() iter.Seq[LogEntry] { return entries }, q), nil
	}

	if run.LogFile == "" || rs.dir == "" {
		return LogPage{}, fmt.Errorf("run %s: %w", run.ID, ErrRunLogNotAvailable)
	}

	path := filepath.Join(rs.dir, filepath.Base(run.LogFile))
	f, err := os.Open(path)
	if err != nil {
		return LogPage{}, fmt.Errorf("run %s: %w: %w", run.ID, ErrRunLogNotAvailable, err)
	}
	defer f.Close()

	lines, err := rs.indexes.lines(f, path)
	if err != nil {
		return LogPage{}, fmt.Errorf("run %s: %w", run.ID, err)
	}

	return pageIndexed(f, lines, q), nil
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
	path     string
	info     os.FileInfo
	strict   bool // a failed write fails the run, see [Options.Quiet]
	failed   bool
	onFail   func()
	entries  int
	warnings int
	errors   int
	closed   bool
}

func (s *runSink) setOnFail(fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.onFail = fn
}

func (s *runSink) hasFailed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.failed
}

func (s *runSink) write(e LogEntry) {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed || s.failed {
		return
	}

	if s.file != nil {
		buf, err := json.Marshal(fileEntry{Time: e.Time, Level: e.Level, Msg: e.Msg, Values: jsonSafe(e.Values)})
		if err != nil {
			buf, _ = json.Marshal(fileEntry{Time: e.Time, Level: e.Level, Msg: e.Msg, Values: map[string]any{"marshalError": err.Error()}})
		}

		buf = append(buf, '\n')
		_, err = s.file.Write(buf)
		if err == nil && s.strict {
			// a deleted file still accepts writes
			if cur, serr := os.Stat(s.path); serr != nil || s.info == nil || !os.SameFile(s.info, cur) {
				err = errors.New("log file is gone")
			}
		}

		if err != nil {
			if s.strict {
				s.failed = true
				if s.onFail != nil {
					s.onFail()
				}
				return
			}

			slog.Error("cannot write scheduler log file", "file", s.file.Name(), "err", err.Error())
		}
	}

	s.entries++
	switch {
	case e.Level >= slog.LevelError:
		s.errors++
	case e.Level >= slog.LevelWarn:
		s.warnings++
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

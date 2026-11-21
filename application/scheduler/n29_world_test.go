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
	"os"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	_ "time/tzdata"

	"go.wdy.de/nago/application/scheduler"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
)

// The N29 acceptance tests drive the scheduler only through its public surface. Time is controlled by a fake
// [scheduler.Clock]: a test never waits for a duration, it waits for a result (a planned due, a timer, a run in the
// store) and moves the clock only to a timer which the scheduler has armed before.

// waitLimit bounds each wait for a result. No test relies on it, it only turns a missing result into a failure.
const waitLimit = 10 * time.Second

const job scheduler.ID = "n29.job"

var berlin = mustZone("Europe/Berlin")

func mustZone(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}

	return loc
}

// bln returns a wall clock time in Europe/Berlin. An ambiguous wall clock time is resolved by time.Date, thus
// the tests write such instants in UTC.
func bln(year int, month time.Month, day, hour, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, berlin)
}

func utc(year int, month time.Month, day, hour, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
}

const day = 24 * time.Hour

// waitFor polls cond until it holds and fails after waitLimit with the description of what.
func waitFor(t testing.TB, what func() string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(waitLimit)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("nicht eingetreten innerhalb von %v: %s", waitLimit, what())
		}

		time.Sleep(time.Millisecond)
	}
}

// fakeClock implements scheduler.Clock. Its time only moves by set, which fires all timers that are due.
type fakeClock struct {
	mu      sync.Mutex
	now     time.Time
	pending []*fakeTimer
}

type fakeTimer struct {
	clock    *fakeClock
	ch       chan time.Time
	deadline time.Time
	done     bool
}

func newFakeClock(now time.Time) *fakeClock {
	return &fakeClock{now: now}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) NewTimer(d time.Duration) scheduler.Timer {
	c.mu.Lock()
	defer c.mu.Unlock()

	tm := &fakeTimer{clock: c, ch: make(chan time.Time, 1), deadline: c.now.Add(d)}
	if d <= 0 {
		tm.done = true
		tm.ch <- c.now
		return tm
	}

	c.pending = append(c.pending, tm)
	return tm
}

func (tm *fakeTimer) C() <-chan time.Time {
	return tm.ch
}

func (tm *fakeTimer) Stop() bool {
	c := tm.clock
	c.mu.Lock()
	defer c.mu.Unlock()

	if tm.done {
		return false
	}

	tm.done = true
	c.pending = slices.DeleteFunc(c.pending, func(p *fakeTimer) bool { return p == tm })
	return true
}

// set moves the clock forward to `to` and fires each pending timer whose deadline is not after `to`, earliest
// first.
func (c *fakeClock) set(t testing.TB, to time.Time) {
	t.Helper()
	c.mu.Lock()
	defer c.mu.Unlock()

	if to.Before(c.now) {
		t.Fatalf("die Uhr läuft nicht rückwärts: %v vor %v", to, c.now)
	}

	c.now = to
	slices.SortFunc(c.pending, func(a, b *fakeTimer) int { return a.deadline.Compare(b.deadline) })
	var keep []*fakeTimer
	for _, tm := range c.pending {
		if tm.deadline.After(to) {
			keep = append(keep, tm)
			continue
		}

		tm.done = true
		tm.ch <- tm.deadline
	}

	c.pending = keep
}

func (c *fakeClock) deadlines() []time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()

	var res []time.Time
	for _, tm := range c.pending {
		res = append(res, tm.deadline)
	}

	slices.SortFunc(res, func(a, b time.Time) int { return a.Compare(b) })
	return res
}

func (c *fakeClock) hasTimer(at time.Time) bool {
	return slices.ContainsFunc(c.deadlines(), func(d time.Time) bool { return d.Equal(at) })
}

// earliest returns the deadline of the earliest pending timer.
func (c *fakeClock) earliest() (time.Time, bool) {
	d := c.deadlines()
	if len(d) == 0 {
		return time.Time{}, false
	}

	return d[0], true
}

// advance waits until a timer is armed for `to` and then moves the clock there.
func (c *fakeClock) advance(t testing.TB, to time.Time) {
	t.Helper()
	waitFor(t, func() string { return fmt.Sprintf("Zeitgeber auf %v, vorhanden %v", to, c.deadlines()) }, func() bool {
		return c.hasTimer(to)
	})
	c.set(t, to)
}

// world is a manager on stores, which survive a restart of the manager, like the data directory of a process.
type world struct {
	t        *testing.T
	clock    *fakeClock
	settings *mem.BlobStore
	runStore *mem.BlobStore
	logDir   string
	m        *scheduler.Manager
	cancel   context.CancelFunc
	ids      []scheduler.ID
}

func newWorld(t *testing.T, now time.Time) *world {
	return newWorldIn(t, now, t.TempDir())
}

func newWorldIn(t *testing.T, now time.Time, logDir string) *world {
	w := &world{
		t:        t,
		clock:    newFakeClock(now),
		settings: mem.NewBlobStore("n29.settings"),
		runStore: mem.NewBlobStore("n29.runs"),
		logDir:   logDir,
	}

	w.boot()
	return w
}

// boot starts a new manager on the stores of the world.
func (w *world) boot() {
	ctx, cancel := context.WithCancel(context.Background())
	w.t.Cleanup(cancel)
	w.cancel = cancel
	w.ids = nil
	w.m = scheduler.NewManagerWith(ctx, scheduler.ManagerOptions{
		Settings: json.NewSloppyJSONRepository[scheduler.Settings, scheduler.ID](w.settings),
		Runs:     json.NewSloppyJSONRepository[scheduler.Run, scheduler.RunID](w.runStore),
		LogDir:   w.logDir,
		Clock:    w.clock,
	})
}

func (w *world) configure(opts scheduler.Options) {
	w.t.Helper()
	if err := w.m.Configure(opts); err != nil {
		w.t.Fatalf("Configure %s: %v", opts.ID, err)
	}

	w.ids = append(w.ids, opts.ID)
}

// shutdown ends the manager like an orderly end of the process and waits until each loop has ended.
func (w *world) shutdown() {
	w.t.Helper()
	w.cancel()
	for _, id := range w.ids {
		waitFor(w.t, func() string {
			return fmt.Sprintf("Schleife von %s endet nach dem Beenden, Zustand %v", id, w.m.State(id))
		}, func() bool {
			return w.m.State(id) == scheduler.Stopped
		})
	}
}

// restartAt ends the manager, moves the clock to `at` and starts a new manager on the same stores.
func (w *world) restartAt(at time.Time) {
	w.t.Helper()
	w.shutdown()
	w.clock.set(w.t, at)
	w.boot()
}

// crashAt returns a world on copies of the stores and of the log directory as they are now, like a process which
// died at this moment and starts again at `at`. The crashed manager keeps its originals and its clock, which does
// not move anymore, until the end of the test.
func (w *world) crashAt(at time.Time) *world {
	w.t.Helper()
	logDir := w.t.TempDir()
	if err := os.CopyFS(logDir, os.DirFS(w.logDir)); err != nil {
		w.t.Fatalf("Protokolle kopieren: %v", err)
	}

	c := &world{
		t:        w.t,
		clock:    newFakeClock(at),
		settings: copyStore(w.t, w.settings),
		runStore: copyStore(w.t, w.runStore),
		logDir:   logDir,
	}

	c.boot()
	return c
}

func copyStore(t testing.TB, src *mem.BlobStore) *mem.BlobStore {
	t.Helper()
	dst := mem.NewBlobStore(src.Name())
	for key, err := range src.List(context.Background(), blob.ListOptions{}) {
		if err != nil {
			t.Fatal(err)
		}

		if buf, ok := src.Load(key); ok {
			dst.Store(key, buf)
		}
	}

	return dst
}

// runs returns the runs of id, newest first.
func (w *world) runs(id scheduler.ID) []scheduler.Run {
	w.t.Helper()
	runs, err := w.m.Runs(id)
	if err != nil {
		w.t.Fatalf("Runs %s: %v", id, err)
	}

	return runs
}

func describeRuns(runs []scheduler.Run) string {
	if len(runs) == 0 {
		return " keine"
	}

	var sb strings.Builder
	for _, r := range runs {
		fmt.Fprintf(&sb, "\n  %s Nr %d Beginn %v Ende %v Ergebnis %d Fehler %q Auslöser %d von %q fällig %v abgebrochen von %q Protokoll %q Einträge %d",
			r.ID, r.Nr, r.StartedAt, r.CompletedAt, r.Outcome, r.Error, r.Trigger, r.By, r.Due, r.CanceledBy, r.LogFile, r.Entries)
	}

	return sb.String()
}

// started waits until id has exactly n runs, of which the newest may still run, and returns them newest first.
func (w *world) started(id scheduler.ID, n int) []scheduler.Run {
	w.t.Helper()
	var runs []scheduler.Run
	waitFor(w.t, func() string { return fmt.Sprintf("%d Läufe von %s, vorhanden:%s", n, id, describeRuns(runs)) }, func() bool {
		runs = w.runs(id)
		return len(runs) == n
	})

	return runs
}

// ended waits until id has exactly n runs which have all ended and returns them newest first.
func (w *world) ended(id scheduler.ID, n int) []scheduler.Run {
	w.t.Helper()
	var runs []scheduler.Run
	waitFor(w.t, func() string {
		return fmt.Sprintf("%d beendete Läufe von %s, vorhanden:%s", n, id, describeRuns(runs))
	}, func() bool {
		runs = w.runs(id)
		if len(runs) != n {
			return false
		}

		for _, r := range runs {
			if r.Outcome == scheduler.OutcomeRunning || r.CompletedAt.IsZero() {
				return false
			}
		}

		return true
	})

	return runs
}

// waitPlanned waits until the loop of id has planned due and waits for it with a timer.
func (w *world) waitPlanned(id scheduler.ID, due time.Time) {
	w.t.Helper()
	waitFor(w.t, func() string {
		return fmt.Sprintf("%s geplant auf %v mit Zeitgeber; geplant ist %v, Zeitgeber %v, Läufe:%s",
			id, due, w.m.NextPlannedAt(id), w.clock.deadlines(), describeRuns(w.runs(id)))
	}, func() bool {
		return w.m.NextPlannedAt(id).Equal(due) && w.clock.hasTimer(due)
	})
}

// fire waits until due is planned for id and moves the clock there.
func (w *world) fire(id scheduler.ID, due time.Time) {
	w.t.Helper()
	w.waitPlanned(id, due)
	w.clock.set(w.t, due)
}

// startNow calls StartNow and fails, if it does not return while the run is still running.
func (w *world) startNow(id scheduler.ID, by string) error {
	w.t.Helper()
	done := make(chan error, 1)
	go func() { done <- w.m.StartNow(id, by) }()
	select {
	case err := <-done:
		return err
	case <-time.After(waitLimit):
		w.t.Fatalf("StartNow(%s, %q) kehrt nicht zurück", id, by)
		return nil
	}
}

// mustStartNow is startNow which must succeed.
func (w *world) mustStartNow(id scheduler.ID, by string) {
	w.t.Helper()
	if err := w.startNow(id, by); err != nil {
		w.t.Fatalf("StartNow(%s, %q): %v", id, by, err)
	}
}

// awaitCatchUp waits for the next invocation of p and moves the clock meanwhile only to armed timers which are
// due until limit.
func (w *world) awaitCatchUp(p *probe, limit time.Time) *invocation {
	w.t.Helper()
	deadline := time.Now().Add(waitLimit)
	for {
		select {
		case inv := <-p.calls:
			return inv
		default:
		}

		if time.Now().After(deadline) {
			w.t.Fatalf("kein Lauf bis %v; jetzt %v, Zeitgeber %v, Läufe:%s", limit, w.clock.Now(), w.clock.deadlines(), describeRuns(w.runs(job)))
		}

		if d, ok := w.clock.earliest(); ok && !d.After(limit) {
			w.clock.set(w.t, d)
		}

		time.Sleep(time.Millisecond)
	}
}

// invocation is a single call of a probe runner.
type invocation struct {
	info    scheduler.RunInfo
	hasInfo bool
	ctx     context.Context
	at      time.Time
	finish  chan error
}

// probe is a runner which reports each call. A held call runs until the test sends its result to finish or
// until its context ends.
type probe struct {
	now    func() time.Time
	hold   atomic.Bool
	result atomic.Pointer[error]
	calls  chan *invocation
}

func newProbe(clock *fakeClock) *probe {
	return &probe{now: clock.Now, calls: make(chan *invocation, 100)}
}

func (p *probe) run(ctx context.Context) error {
	info, ok := scheduler.RunInfoFrom(ctx)
	inv := &invocation{info: info, hasInfo: ok, ctx: ctx, at: p.now(), finish: make(chan error, 1)}
	scheduler.LoggerFrom(ctx).Info("probe")
	p.calls <- inv

	if p.hold.Load() {
		select {
		case err := <-inv.finish:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if err := p.result.Load(); err != nil {
		return *err
	}

	return nil
}

// fail makes each call, which is not held, return err.
func (p *probe) fail(err error) {
	p.result.Store(&err)
}

// next waits for the next call.
func (p *probe) next(t testing.TB) *invocation {
	t.Helper()
	select {
	case inv := <-p.calls:
		return inv
	case <-time.After(waitLimit):
		t.Fatal("der Runner wurde nicht aufgerufen")
		return nil
	}
}

// none fails, if a call is pending.
func (p *probe) none(t testing.TB) {
	t.Helper()
	select {
	case inv := <-p.calls:
		t.Fatalf("unerwarteter Aufruf des Runners um %v mit %+v", inv.at, inv.info)
	default:
	}
}

func cronOpts(id scheduler.ID, run func(context.Context) error, hour, minute int, mods ...func(*scheduler.Options)) scheduler.Options {
	o := scheduler.Options{
		ID:       id,
		Name:     string(id),
		Kind:     scheduler.Cron,
		Defaults: scheduler.Settings{ID: id, CronHour: hour, CronMinute: minute},
		Runner:   run,
		Location: berlin,
		Durable:  true,
		Quiet:    true,
	}

	for _, mod := range mods {
		mod(&o)
	}

	return o
}

func intervalOpts(id scheduler.ID, run func(context.Context) error, every time.Duration, mods ...func(*scheduler.Options)) scheduler.Options {
	o := scheduler.Options{
		ID:       id,
		Name:     string(id),
		Kind:     scheduler.Schedule,
		Defaults: scheduler.Settings{ID: id, PauseTime: every},
		Runner:   run,
		Location: berlin,
		Durable:  true,
		Quiet:    true,
	}

	for _, mod := range mods {
		mod(&o)
	}

	return o
}

func catchUp(o *scheduler.Options) {
	o.Defaults.CatchUp = true
}

func durable(on bool) func(*scheduler.Options) {
	return func(o *scheduler.Options) { o.Durable = on }
}

func quiet(on bool) func(*scheduler.Options) {
	return func(o *scheduler.Options) { o.Quiet = on }
}

func zone(loc *time.Location) func(*scheduler.Options) {
	return func(o *scheduler.Options) { o.Location = loc }
}

func weekdays(days ...time.Weekday) func(*scheduler.Options) {
	return func(o *scheduler.Options) { o.Defaults.CronWeekdays = days }
}

func monthDays(days ...int) func(*scheduler.Options) {
	return func(o *scheduler.Options) { o.Defaults.CronMonthDays = days }
}

func keepLogs(n int) func(*scheduler.Options) {
	return func(o *scheduler.Options) { o.Defaults.KeepRunLogs = n }
}

// oldStyle removes every new switch, so that the scheduler behaves like nago 0a4135245.
func oldStyle(o *scheduler.Options) {
	o.Durable = false
	o.Quiet = false
	o.Location = nil
	o.Defaults.CatchUp = false
	o.Defaults.KeepRunLogs = 0
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

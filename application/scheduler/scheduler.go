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
	"log/slog"
	"runtime/debug"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.wdy.de/nago/logging"
)

type LogEntry struct {
	Level  slog.Level
	Time   time.Time
	Msg    string
	Values map[string]any
}

type Scheduler struct {
	externalCtx     context.Context
	ctx             context.Context
	settingsRepo    SettingsRepository
	state           atomic.Pointer[State]
	lastError       atomic.Pointer[error]
	id              ID
	optsMutex       sync.RWMutex // protects opts, which [Scheduler.Update] replaces
	opts            Options
	clock           Clock
	cancel          func()
	logs            []LogEntry
	logsMutex       sync.Mutex
	singleRunMutex  sync.Mutex // held during a run, from its begin until its end has been persisted
	lastStartedAt   atomic.Pointer[time.Time]
	lastCompletedAt atomic.Pointer[time.Time]
	nextPlannedAt   atomic.Pointer[time.Time]
	launchMutex     sync.Mutex                 // protects ctx and cancel and serializes the (re)launch of the loop
	loop            atomic.Pointer[loopHandle] // the current loop of Launch or nil
	wake            chan struct{}              // signals changed settings to the loop
	runs            *runStore
	sink            atomic.Pointer[runSink]
	currentRun      atomic.Pointer[Run]

	// runMutex protects the run in progress and the state changes of runs. runChanged is signalled, when a run
	// begins, its runner returns or it has ended.
	runMutex       sync.Mutex
	runChanged     *sync.Cond
	active         *activeRun
	activeReturned bool   // the runner of the active run has returned, the run only persists its end
	runGen         uint64 // incremented by each run, so that the end of a run changes the state only if no run followed

	// planMutex protects the plan state, which a durable scheduler persists.
	planMutex sync.Mutex
	plan      DurableState
	removed   bool // removed from its manager, so that a late end of a run does not persist anything

	onRunEnd func(oldest time.Time) // tells the manager the start of the oldest remaining run
}

// activeRun is the run in progress.
type activeRun struct {
	mu              sync.Mutex
	cancel          context.CancelCauseFunc
	returned        bool // the runner has returned, the run cannot be cancelled anymore
	cancelRequested bool
	canceledBy      string
	logFailed       bool

	gen      uint64
	async    bool // begun by StartNow
	info     RunInfo
	opts     Options
	keep     int
	run      Run
	sink     *runSink
	epoch    uint64
	beginErr error
}

// maxMemoryLogs limits the in-memory log buffer of the current run. The complete log is persisted in the run log file.
const maxMemoryLogs = 1000

// loopHandle identifies a single loop started by [Scheduler.Launch]. The loop is alive, as long as its context
// has not been cancelled. Each Launch creates a new handle, so that an old loop which ends late cannot affect a
// newer one.
type loopHandle struct {
	ctx context.Context
}

// alive returns true, if the loop exists and has not been cancelled.
func (h *loopHandle) alive() bool {
	return h != nil && h.ctx.Err() == nil
}

func NewScheduler(ctx context.Context, opts Options, settingsRepo SettingsRepository) *Scheduler {
	return newScheduler(ctx, opts, settingsRepo, nil, nil)
}

func newScheduler(ctx context.Context, opts Options, settingsRepo SettingsRepository, clock Clock, runs *runStore) *Scheduler {
	s := &Scheduler{
		externalCtx:  ctx,
		ctx:          ctx,
		cancel:       func() {},
		id:           opts.ID,
		opts:         opts,
		clock:        clockOrSystem(clock),
		settingsRepo: settingsRepo,
		wake:         make(chan struct{}, 1),
		runs:         runs,
	}
	s.runChanged = sync.NewCond(&s.runMutex)

	var zeroTime time.Time
	s.lastStartedAt.Store(&zeroTime)
	s.lastCompletedAt.Store(&zeroTime)
	s.nextPlannedAt.Store(&zeroTime)
	state := Stopped
	s.state.Store(&state)

	return s
}

// options returns the current options, see [Scheduler.update].
func (s *Scheduler) options() Options {
	s.optsMutex.RLock()
	defer s.optsMutex.RUnlock()

	return s.opts
}

func (s *Scheduler) setOptions(opts Options) {
	s.optsMutex.Lock()
	defer s.optsMutex.Unlock()

	s.opts = opts
}

func (s *Scheduler) LastError() error {
	err := s.lastError.Load()
	if err == nil {
		return nil
	}

	return *err
}

func (s *Scheduler) State() State {
	return *s.state.Load()
}

// Destroy cancels the current context, which stops the loop and cancels a run in progress. It does not wait for
// the loop or the run to end, but [Scheduler.Looping] returns false immediately.
func (s *Scheduler) Destroy() {
	s.launchMutex.Lock()
	defer s.launchMutex.Unlock()

	s.cancel()
}

// ResetContext cancels the current context, including a running loop, and replaces it with a new one.
func (s *Scheduler) ResetContext() {
	s.launchMutex.Lock()
	defer s.launchMutex.Unlock()

	s.resetContext()
}

// resetContext requires the launchMutex.
func (s *Scheduler) resetContext() {
	s.cancel()

	ctx := logging.WithContext(s.externalCtx, slog.New(slogHandler{sched: s}))

	myCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.ctx = myCtx
}

// liveContext returns the current context or a new one, if the current one has been cancelled, e.g. by Stop.
func (s *Scheduler) liveContext() context.Context {
	s.launchMutex.Lock()
	defer s.launchMutex.Unlock()

	if s.ctx.Err() != nil {
		s.resetContext()
	}

	return s.ctx
}

// Launch starts a new loop. A previous loop is cancelled but not awaited and its late end does not affect the new
// loop.
func (s *Scheduler) Launch() {
	s.launchMutex.Lock()
	s.resetContext()

	// the loop keeps its own context, so that a later ResetContext, e.g. by ExecuteNow on a stopped scheduler,
	// cannot revive a cancelled loop
	ctx := s.ctx
	h := &loopHandle{ctx: ctx}
	s.loop.Store(h)
	s.launchMutex.Unlock()

	// setState changes the state only while this loop is the current one, so that a cancelled loop which still
	// completes its run does not overwrite the state of a newer loop, and never while a run is in progress
	setState := func(state State) {
		s.runMutex.Lock()
		defer s.runMutex.Unlock()

		if s.loop.Load() == h && h.alive() && s.active == nil {
			s.state.Store(&state)
		}
	}

	durable := s.options().Durable

	go func() {
		defer func() {
			// Only the current loop clears the handle and sets the state. The mutex ensures that a concurrent Launch
			// happens either before, so that the CompareAndSwap fails, or afterward, so that the new loop sets its
			// state later.
			s.launchMutex.Lock()
			defer s.launchMutex.Unlock()

			if s.loop.CompareAndSwap(h, nil) {
				s.runMutex.Lock()
				if s.active == nil {
					state := Stopped
					s.state.Store(&state)
				}
				s.runMutex.Unlock()
			}
		}()

		if durable {
			s.durableLoop(ctx, setState)
			return
		}

		s.classicLoop(ctx, setState)
	}()
}

// classicLoop is the loop of the schedulers without [Options.Durable].
func (s *Scheduler) classicLoop(ctx context.Context, setState func(State)) {
loop:
	for {
		settings, err := s.loadSettings()
		if err != nil {
			slog.Error("service looper failed to load settings", "id", s.id, "err", err.Error())
			return
		}

		opts := s.options()
		if settings.Disabled || opts.Kind == Manual || s.isSuspended() {
			setState(Disabled)
			// wait the config-reload time or until the settings have been changed or exit early on cancel
			if _, alive := s.sleep(ctx, time.Minute); !alive {
				slog.Info("service shutdown due to context signal")
				return
			}

			continue
		}

		setState(Paused)
		// wait the start-delay or exit early on cancel. Changed settings restart the delay.
		woken, alive := s.sleep(ctx, settings.StartDelay)
		if !alive {
			slog.Info("service shutdown due to context signal")
			return
		}

		if woken {
			continue
		}

		// perform the actual work execution

		opts = s.options()
		pauseTime := settings.PauseTime
		var deadline time.Time // the end of the pause, if it differs from the default

		if opts.Kind != Cron {
			if next, later := s.laterThanNow(opts, settings); later {
				// a reference, e.g. Resume, defers the next run of a Schedule
				deadline = next
			} else if err := s.executePlanned(ctx, RunInfo{Due: s.clock.Now(), Trigger: TriggerPlan}); err != nil {
				s.logLoopError("service looper failed to run", err)
			}
		}

		switch opts.Kind {
		case OneShot, Manual:
			var zeroT time.Time
			s.nextPlannedAt.Store(&zeroT)
			return
		case Schedule:
			// do nothing, go ahead and sleep
			nextPlannedAt := s.clock.Now().Add(settings.PauseTime)
			if !deadline.IsZero() {
				nextPlannedAt = deadline
			}
			s.nextPlannedAt.Store(&nextPlannedAt)
		case Cron:
			nextPlannedAt, err := nextCronDue(s.clock.Now(), opts, settings)
			if err != nil {
				slog.Error("service looper failed to plan", "id", s.id, "err", err.Error())
				return
			}
			pauseTime = nextPlannedAt.Sub(s.clock.Now())
			s.nextPlannedAt.Store(&nextPlannedAt)
		}

		// wait the pause-delay or exit early on cancel
		setState(Paused)
		// do not schedule faster than 1 second, everything else is probably a configuration mistake
		if deadline.IsZero() {
			deadline = s.clock.Now().Add(max(pauseTime, time.Second))
		}
		for {
			woken, alive := s.sleepUntil(ctx, deadline)
			if !alive {
				slog.Info("service shutdown due to context signal", "id", s.id)
				return
			}

			if !woken {
				break
			}

			// the settings have been changed while pausing
			if s.options().Kind != Schedule {
				continue loop // a cron plan is recalculated from scratch
			}

			settings, err = s.loadSettings()
			if err != nil || settings.Disabled || s.isSuspended() {
				continue loop
			}

			// keep the distance to the last run, but with the new pause time
			deadline = later(s.LastCompletedAt(), s.reference()).Add(max(settings.PauseTime, time.Second))
			s.nextPlannedAt.Store(&deadline)
		}

		if s.options().Kind == Cron {
			if err := s.executePlanned(ctx, RunInfo{Due: s.NextPlannedAt(), Trigger: TriggerPlan}); err != nil {
				s.logLoopError("service looper cron failed to run", err)
			}
		}
	}
}

// laterThanNow returns the next run of a [Schedule], if a reference (see [DurableState.Reference]) defers it
// beyond now. Without a reference, a Schedule runs at the begin of its loop, as before.
func (s *Scheduler) laterThanNow(opts Options, settings Settings) (time.Time, bool) {
	ref := s.reference()
	if opts.Kind != Schedule || ref.IsZero() {
		return time.Time{}, false
	}

	next := later(ref, s.LastCompletedAt()).Add(max(settings.PauseTime, time.Second))
	return next, next.After(s.clock.Now())
}

func (s *Scheduler) logLoopError(msg string, err error) {
	if s.options().Quiet {
		// the error may contain data of the run, which must not go into the log of the server
		slog.Error(msg, "id", s.id)
		return
	}

	slog.Error(msg, "id", s.id, "err", err.Error())
}

// durableLoop is the loop of a scheduler with [Options.Durable]. Its dues follow from its persisted plan state,
// see [DurableState], and its start never runs anything except a catch up.
func (s *Scheduler) durableLoop(ctx context.Context, setState func(State)) {
	start := s.clock.Now()
	from := start           // a cron due is searched from here
	var startBase time.Time // the start, if a due of a Schedule elapsed while the manager was not running
	first := true

	for {
		if ctx.Err() != nil {
			return
		}

		settings, err := s.loadSettings()
		if err != nil {
			slog.Error("service looper failed to load settings", "id", s.id, "err", err.Error())
			return
		}

		opts := s.options()
		if s.isSuspended() || settings.Disabled || opts.Kind == Manual {
			// nothing is caught up, while paused or disabled
			first = false
			s.storeNextPlanned(time.Time{})
			setState(Disabled)
			if !s.waitWake(ctx) {
				return
			}

			continue
		}

		var due time.Time
		trigger := TriggerPlan
		if first {
			first = false
			if d, missed := s.missedDue(opts, settings, start); missed && settings.CatchUp {
				due, trigger = d, TriggerCatchUp
			} else if opts.Kind == Schedule {
				// without catching up, a due which elapsed while the manager was not running is the pause
				// time after the start
				if n, err := s.nextDue(opts, settings, from, time.Time{}); err == nil && n.Before(start) {
					startBase = start
				}
			}
		}

		if trigger == TriggerPlan {
			due, err = s.nextDue(opts, settings, from, startBase)
			if err != nil {
				slog.Error("service looper failed to plan", "id", s.id, "err", err.Error())
				s.storeNextPlanned(time.Time{})
				setState(Disabled)
				if !s.waitWake(ctx) {
					return
				}

				continue
			}

			s.storeNextPlanned(due)
			setState(Paused)
			woken, alive := s.sleepUntil(ctx, due)
			if !alive {
				return
			}

			if woken {
				// a pause, a changed plan or the end of a manual run: plan again
				continue
			}
		} else {
			s.storeNextPlanned(s.clock.Now())
		}

		s.fire(ctx, due, trigger)
		from = s.clock.Now()
	}
}

// missedDue returns the latest due, which elapsed while the manager was not running, see [Settings.CatchUp].
func (s *Scheduler) missedDue(opts Options, settings Settings, start time.Time) (time.Time, bool) {
	p := s.planState()
	switch opts.Kind {
	case Cron:
		ref := later(p.LastDue, p.Reference)
		d, err := newCronPlan(settings, opts.Location).prev(start)
		if err != nil || ref.IsZero() || !d.After(ref) {
			return time.Time{}, false
		}

		return d, true
	case Schedule:
		ref := later(p.LastEnd, p.Reference)
		if ref.IsZero() {
			return time.Time{}, false
		}

		d := ref.Add(interval(settings))
		if !d.Before(start) || !d.After(p.LastDue) {
			// not elapsed or already handled, e.g. by a run which a restart cancelled
			return time.Time{}, false
		}

		return d, true
	default:
		return time.Time{}, false
	}
}

// nextDue returns the next due of a durable scheduler. A due in the past is due immediately.
func (s *Scheduler) nextDue(opts Options, settings Settings, from, startBase time.Time) (time.Time, error) {
	p := s.planState()
	switch opts.Kind {
	case Cron:
		f := later(from, p.Reference)
		if !p.LastDue.Before(f) {
			f = p.LastDue.Add(time.Nanosecond)
		}

		return nextCronDue(f, opts, settings)
	case Schedule:
		n := later(later(p.LastEnd, p.Reference), startBase).Add(interval(settings))
		if !n.After(p.LastDue) {
			// the due has been handled, e.g. omitted during a manual run, whose end plans again
			n = p.LastDue.Add(interval(settings))
		}

		return n, nil
	default:
		return time.Time{}, fmt.Errorf("kind %d has no plan", opts.Kind)
	}
}

// fire starts a run at the given due, unless the scheduler has been paused meanwhile. A due while a run is in
// progress, e.g. a manual one, is omitted.
func (s *Scheduler) fire(ctx context.Context, due time.Time, trigger Trigger) {
	if ctx.Err() != nil {
		return
	}

	if !s.lockUnlessRunning() {
		// a run is in progress, e.g. a manual one: the due is omitted
		s.planMutex.Lock()
		if !s.plan.Paused && due.After(s.plan.LastDue) {
			s.plan.LastDue = due
			s.persistPlanLocked()
		}
		s.planMutex.Unlock()
		return
	}

	s.planMutex.Lock()
	if s.plan.Paused {
		s.planMutex.Unlock()
		s.singleRunMutex.Unlock()
		return
	}

	// the due is handled from now on, also if a restart cancels its run
	if due.After(s.plan.LastDue) {
		s.plan.LastDue = due
		s.persistPlanLocked()
	}
	s.planMutex.Unlock()

	a, runCtx := s.beginRun(ctx, RunInfo{Due: due, Trigger: trigger}, false)
	_ = s.execRun(a, runCtx)
}

// lockUnlessRunning acquires the singleRunMutex for a due or StartNow and returns true, or returns false while the
// runner of a run is running. A run whose runner has returned, which only persists its end, or a run which is just
// beginning holds the mutex only briefly: it is awaited, so that it is decided by whether a runner runs, not by the
// scheduling of the goroutines. The caller must not hold the planMutex, which the end of a run needs.
func (s *Scheduler) lockUnlessRunning() bool {
	s.runMutex.Lock()
	defer s.runMutex.Unlock()

	for {
		if s.singleRunMutex.TryLock() {
			return true
		}

		if s.active != nil && !s.activeReturned {
			return false
		}

		// the run is beginning or persisting its end, both signal runChanged
		s.runChanged.Wait()
	}
}

func interval(settings Settings) time.Duration {
	return max(settings.PauseTime, time.Second)
}

func later(a, b time.Time) time.Time {
	if b.After(a) {
		return b
	}

	return a
}

// nextCronAt returns the next planned execution after now. The cron time is a wall clock time in the location of
// now, thus it stays the same on days with 23 or 25 hours, when the daylight saving time begins or ends. A cron
// time, which does not exist on such a day, is normalized like by [time.Date], e.g. 02:30 to 03:30, and an
// ambiguous one is planned once. The pause time is a duration between the executions, starting at the cron time.
func nextCronAt(now time.Time, settings Settings) time.Time {
	at := func(day int) time.Time {
		return time.Date(now.Year(), now.Month(), day, settings.CronHour, settings.CronMinute, 0, 0, now.Location())
	}

	nextPlannedAt := at(now.Day())
	if settings.PauseTime > 0 {
		for !nextPlannedAt.After(now) {
			nextPlannedAt = nextPlannedAt.Add(settings.PauseTime)
		}
	} else if nextPlannedAt.Before(now) {
		nextPlannedAt = at(now.Day() + 1)
	}

	return nextPlannedAt
}

// loadSettings returns the persisted settings or the defaults.
func (s *Scheduler) loadSettings() (Settings, error) {
	return s.loadSettingsFor(s.options())
}

func (s *Scheduler) loadSettingsFor(opts Options) (Settings, error) {
	optSettings, err := s.settingsRepo.FindByID(s.id)
	if err != nil {
		return Settings{}, err
	}

	if optSettings.IsSome() {
		return optSettings.Unwrap(), nil
	}

	return opts.Defaults, nil
}

// sleep waits for d. It returns woken=true, if the settings have been changed in the meantime (see [Scheduler.Wake])
// and alive=false, if the scheduler has been cancelled.
func (s *Scheduler) sleep(ctx context.Context, d time.Duration) (woken, alive bool) {
	if d <= 0 {
		select {
		case <-ctx.Done():
			return false, false
		case <-s.wake:
			return true, true
		default:
			return false, true
		}
	}

	timer := s.clock.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false, false
	case <-s.wake:
		return true, true
	case <-timer.C():
		return false, true
	}
}

// sleepUntil waits until the clock reaches deadline, like sleep. Its timer is armed with a duration, thus it checks
// the clock again after arming the timer: a clock which passed the deadline meanwhile would otherwise fire the timer
// that much later.
func (s *Scheduler) sleepUntil(ctx context.Context, deadline time.Time) (woken, alive bool) {
	d := deadline.Sub(s.clock.Now())
	if d <= 0 {
		return s.sleep(ctx, 0)
	}

	timer := s.clock.NewTimer(d)
	defer timer.Stop()

	if !s.clock.Now().Before(deadline) {
		return s.sleep(ctx, 0)
	}

	select {
	case <-ctx.Done():
		return false, false
	case <-s.wake:
		return true, true
	case <-timer.C():
		return false, true
	}
}

// waitWake waits without a timer until the scheduler is woken. It returns false, if the scheduler has been
// cancelled.
func (s *Scheduler) waitWake(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case <-s.wake:
		return true
	}
}

// Looping returns true, if the loop of the scheduler has been launched and not yet stopped. A stopped loop is not
// looping anymore, even if its goroutine has not yet noticed its cancellation. In contrast, the [State] may be
// Stopped or Running while a stopped scheduler is executed manually.
func (s *Scheduler) Looping() bool {
	return s.loop.Load().alive()
}

// Wake makes the loop reload its settings immediately instead of after the current delay.
func (s *Scheduler) Wake() {
	select {
	case s.wake <- struct{}{}:
	default:
		// a wake up is already pending
	}
}

// executePlanned runs a due of the loop of a scheduler without [Options.Durable]. It waits for a run of
// [Scheduler.ExecuteNow], as before, but a due during a run of [Scheduler.StartNow] is omitted. A cancelled context,
// e.g. of a loop which has been stopped while its timer fired, is not executed at all.
func (s *Scheduler) executePlanned(ctx context.Context, info RunInfo) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if s.isSuspended() {
		return nil
	}

	// like lockUnlessRunning, only the runner of a manual run omits the due, not its begin or the persisting of its end
	s.runMutex.Lock()
	manual := s.active != nil && s.active.async && !s.activeReturned
	s.runMutex.Unlock()
	if manual {
		return nil
	}

	return s.execute(ctx, info)
}

// execute runs the runner once with the given context and blocks until it completes. It waits for a run in
// progress. A cancelled context is not executed at all.
func (s *Scheduler) execute(ctx context.Context, info RunInfo) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.singleRunMutex.Lock()
	a, runCtx := s.beginRun(ctx, info, false)
	return s.execRun(a, runCtx)
}

// ExecuteNow runs the runner immediately and blocks until it completes. It waits for a run of the loop which is
// in progress.
func (s *Scheduler) ExecuteNow() error {
	return s.execute(s.liveContext(), RunInfo{Trigger: TriggerManual})
}

// StartNow begins a manual run and returns without waiting for it. While the runner of a run is running, it returns
// [ErrRunning] without a run. A paused or stopped scheduler can be started as well.
func (s *Scheduler) StartNow(by string) error {
	ctx := s.liveContext()
	if !s.lockUnlessRunning() {
		return ErrRunning
	}

	a, runCtx := s.beginRun(ctx, RunInfo{Trigger: TriggerManual, By: by}, true)
	go func() {
		_ = s.execRun(a, runCtx)
	}()

	return nil
}

// CancelRun cancels the run in progress, without stopping the loop. The run ends as [OutcomeCanceled] with by as
// [Run.CanceledBy]. Without a run in progress, it returns [ErrNotRunning].
func (s *Scheduler) CancelRun(by string) error {
	s.runMutex.Lock()
	a := s.active
	s.runMutex.Unlock()

	if a == nil {
		return ErrNotRunning
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.returned {
		return ErrNotRunning
	}

	if !a.cancelRequested {
		a.cancelRequested = true
		a.canceledBy = by
	}

	a.cancel(context.Canceled)
	return nil
}

// beginRun begins a run, while the caller holds the singleRunMutex, which execRun releases.
func (s *Scheduler) beginRun(parent context.Context, info RunInfo, async bool) (*activeRun, context.Context) {
	opts := s.options()
	settings, err := s.loadSettingsFor(opts)
	if err != nil {
		settings = opts.Defaults
	}

	now := s.clock.Now()
	if info.Trigger == TriggerManual || info.Due.IsZero() {
		info.Due = now
	}

	runCtx, cancel := context.WithCancelCause(parent)
	a := &activeRun{cancel: cancel, async: async, info: info, opts: opts, keep: keepRunLogs(settings)}

	s.runMutex.Lock()
	s.runGen++
	a.gen = s.runGen
	s.active = a
	s.activeReturned = false
	state := Running
	s.state.Store(&state)
	s.runChanged.Broadcast()
	s.runMutex.Unlock()

	s.ClearLogs()
	s.lastStartedAt.Store(&now)

	if info.Trigger == TriggerManual && (async || opts.Durable) {
		// the start of the last manual run is a reference of the plan; ExecuteNow keeps the behavior of a
		// scheduler without Durable as before
		s.planMutex.Lock()
		if now.After(s.plan.Reference) {
			s.plan.Reference = now
			s.persistPlanLocked()
		}
		s.planMutex.Unlock()
	}

	if s.runs == nil {
		a.beginErr = errors.New("no run store")
	} else {
		run, sink, epoch, err := s.runs.begin(s.id, info, now, locationOf(opts), opts.Quiet)
		if err != nil {
			slog.Error("cannot begin scheduler run", "id", s.id, "err", err.Error())
			a.beginErr = err
		} else {
			a.run, a.sink, a.epoch = run, sink, epoch
			sink.setOnFail(func() {
				a.mu.Lock()
				a.logFailed = true
				a.mu.Unlock()
				cancel(ErrLogNotWritable)
			})
			s.currentRun.Store(&run)
			s.sink.Store(sink)
		}
	}

	return a, withRunInfo(runCtx, info)
}

// execRun calls the runner of a begun run and ends it. It releases the singleRunMutex.
func (s *Scheduler) execRun(a *activeRun, ctx context.Context) error {
	var err error
	if a.sink != nil && a.sink.hasFailed() {
		// the log of a quiet scheduler cannot be written, so the runner is not called
		a.mu.Lock()
		a.logFailed = true
		a.mu.Unlock()
	} else {
		err = callRunner(a.opts, ctx)
	}

	interrupted := ctx.Err() != nil
	a.mu.Lock()
	a.returned = true
	canceled, canceledBy, logFailed := a.cancelRequested, a.canceledBy, a.logFailed
	a.mu.Unlock()

	s.runMutex.Lock()
	s.activeReturned = true
	s.runChanged.Broadcast()
	s.runMutex.Unlock()

	return s.endRun(a, err, interrupted, canceled, canceledBy, logFailed)
}

func callRunner(opts Options, ctx context.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if !opts.Quiet {
				debug.PrintStack()
			}
			err = &PanicError{Trace: string(debug.Stack()), Cause: fmt.Errorf("recovered from panic: %v", r)}
		}
	}()

	return opts.Runner(ctx)
}

// endRun persists the end of a run and releases the singleRunMutex. A run of a durable scheduler, which has been
// interrupted by the end of its manager, ends as cancelled by [CanceledByRestart].
func (s *Scheduler) endRun(a *activeRun, err error, interrupted, canceled bool, canceledBy string, logFailed bool) error {
	opts := a.opts
	run := a.run
	shutdown := opts.Durable && interrupted && s.externalCtx.Err() != nil

	if logFailed {
		err = ErrLogNotWritable
	}

	if err != nil {
		run.Error = err.Error()
	}

	var perr *PanicError
	switch {
	case logFailed:
		run.Outcome = OutcomeFailed
	case canceled:
		run.Outcome = OutcomeCanceled
		run.CanceledBy = canceledBy
	case shutdown:
		run.Outcome = OutcomeCanceled
		run.CanceledBy = CanceledByRestart
	case err == nil:
		run.Outcome = OutcomeSucceeded
	case errors.As(err, &perr):
		run.Outcome = OutcomePanicked
	case isCanceled(err):
		run.Outcome = OutcomeCanceled
	default:
		run.Outcome = OutcomeFailed
	}

	if err != nil {
		if !opts.Quiet {
			s.Error("service looper encountered error", "err", err.Error())
		}
		s.lastError.Store(&err)
	}

	s.sink.Store(nil)
	s.currentRun.Store(nil)

	now := s.clock.Now()
	var oldest time.Time
	floor, floorKnown := RunID(""), false
	if a.beginErr == nil {
		run.CompletedAt = now
		oldest, floor, floorKnown = s.runs.end(run, a.sink, a.epoch, a.keep, locationOf(opts))
	}

	s.planEnd(opts, now, floor, floorKnown)
	s.lastCompletedAt.Store(&now)

	s.runMutex.Lock()
	s.active = nil
	s.runMutex.Unlock()

	s.singleRunMutex.Unlock()

	// without a loop, e.g. a stopped scheduler executed manually, nothing is paused
	s.runMutex.Lock()
	s.runChanged.Broadcast()
	if s.runGen == a.gen && s.active == nil {
		state := Stopped
		if s.Looping() {
			state = Paused
			if opts.Durable && s.isSuspended() {
				state = Disabled
			}
		}
		s.state.Store(&state)
	}
	s.runMutex.Unlock()

	if opts.Durable || a.async {
		// e.g. a manual run is the previous run of a Schedule
		s.Wake()
	}

	if s.onRunEnd != nil {
		s.onRunEnd(oldest)
	}

	return err
}

// planEnd keeps the end of a run in the plan state.
func (s *Scheduler) planEnd(opts Options, end time.Time, floor RunID, floorKnown bool) {
	s.planMutex.Lock()
	defer s.planMutex.Unlock()

	s.plan.LastEnd = end
	if floorKnown {
		s.plan.LogFloor = floor
	}

	if opts.Durable && opts.Kind == Cron {
		// a due during the run has been omitted
		if settings, err := s.loadSettingsFor(opts); err == nil && settings.PauseTime <= 0 {
			if d, err := newCronPlan(settings, opts.Location).prev(end.Add(time.Nanosecond)); err == nil && d.After(s.plan.LastDue) {
				s.plan.LastDue = d
			}
		}
	}

	s.persistPlanLocked()
}

// persistPlanLocked saves the plan state of a durable scheduler. It requires the planMutex.
func (s *Scheduler) persistPlanLocked() {
	if s.removed || s.runs == nil || !s.options().Durable {
		return
	}

	if err := s.runs.saveState(s.id, s.plan); err != nil {
		slog.Error("cannot save scheduler state", "id", s.id, "err", err.Error())
	}
}

func (s *Scheduler) planState() DurableState {
	s.planMutex.Lock()
	defer s.planMutex.Unlock()

	return s.plan
}

func (s *Scheduler) isSuspended() bool {
	s.planMutex.Lock()
	defer s.planMutex.Unlock()

	return s.plan.Paused
}

func (s *Scheduler) reference() time.Time {
	s.planMutex.Lock()
	defer s.planMutex.Unlock()

	return s.plan.Reference
}

func (s *Scheduler) storeNextPlanned(t time.Time) {
	s.nextPlannedAt.Store(&t)
}

// Pause omits every due until [Scheduler.Resume]. A run in progress is not cancelled.
func (s *Scheduler) Pause() {
	s.planMutex.Lock()
	if !s.plan.Paused {
		s.plan.Paused = true
		s.persistPlanLocked()
	}
	s.planMutex.Unlock()

	s.Wake()
}

// Resume ends a pause. Its time becomes a reference of the plan, so that it does not start a run immediately.
func (s *Scheduler) Resume() {
	s.planMutex.Lock()
	if s.plan.Paused {
		s.plan.Paused = false
		s.plan.Reference = later(s.plan.Reference, s.clock.Now())
		s.persistPlanLocked()
	}
	s.planMutex.Unlock()

	s.Wake()
}

// addReference adds a reference of the plan.
func (s *Scheduler) addReference(t time.Time) {
	s.planMutex.Lock()
	if t.After(s.plan.Reference) {
		s.plan.Reference = t
		s.persistPlanLocked()
	}
	s.planMutex.Unlock()

	s.Wake()
}

// update replaces the options without cancelling a run in progress. A changed plan applies from the next due on
// and its time becomes a reference of the plan.
func (s *Scheduler) update(opts Options) {
	old := s.options()
	oldSettings, errOld := s.loadSettingsFor(old)
	s.setOptions(opts)
	newSettings, errNew := s.loadSettingsFor(opts)
	if errOld != nil || errNew != nil || planOf(old, oldSettings) != planOf(opts, newSettings) {
		s.addReference(s.clock.Now())
		return
	}

	s.Wake()
}

// CurrentRun returns the currently executing run, if any.
func (s *Scheduler) CurrentRun() (Run, bool) {
	r := s.currentRun.Load()
	if r == nil {
		return Run{}, false
	}

	return *r, true
}

func isCanceled(err error) bool {
	return errors.Is(err, context.Canceled)
}

func (s *Scheduler) Info(msg string, args ...any) {
	if !s.options().Quiet {
		slog.Info(msg, args...)
	}
	s.logLevel(slog.LevelInfo, msg, args...)
}

func (s *Scheduler) Error(msg string, args ...any) {
	if !s.options().Quiet {
		slog.Error(msg, args...)
	}
	s.logLevel(slog.LevelError, msg, args...)
}

func (s *Scheduler) Warn(msg string, args ...any) {
	if !s.options().Quiet {
		slog.Warn(msg, args...)
	}
	s.logLevel(slog.LevelWarn, msg, args...)
}

func (s *Scheduler) Debug(msg string, args ...any) {
	if !s.options().Quiet {
		slog.Debug(msg, args...)
	}
	s.logLevel(slog.LevelDebug, msg, args...)
}

func (s *Scheduler) Logs() []LogEntry {
	s.logsMutex.Lock()
	defer s.logsMutex.Unlock()

	return slices.Clone(s.logs)
}

func (s *Scheduler) LastStartedAt() time.Time {
	return *s.lastStartedAt.Load()
}

func (s *Scheduler) LastCompletedAt() time.Time {
	return *s.lastCompletedAt.Load()
}

func (s *Scheduler) NextPlannedAt() time.Time {
	return *s.nextPlannedAt.Load()
}

func (s *Scheduler) ClearLogs() {
	s.logsMutex.Lock()
	defer s.logsMutex.Unlock()
	clear(s.logs)
	s.logs = s.logs[:0]
}

func (s *Scheduler) logLevel(level slog.Level, msg string, args ...any) {
	if len(args)%2 != 0 {
		slog.Error("invalid arguments in log level")
		debug.PrintStack()
		args = nil
	}

	var tmp map[string]any
	if len(args) > 0 {
		tmp = make(map[string]any, len(args)/2)
		for i := 0; i < len(args); i += 2 {
			if v, ok := args[i].(string); ok && v != "" {
				tmp[v] = args[i+1]
			} else {
				tmp[fmt.Sprint(args[i])] = args[i+1]
			}
		}
	}

	s.record(LogEntry{Level: level, Time: s.clock.Now(), Msg: msg, Values: tmp})
}

// record appends the entry to the in-memory buffer and to the log file of the current run.
func (s *Scheduler) record(e LogEntry) {
	s.logsMutex.Lock()
	if len(s.logs) >= maxMemoryLogs {
		s.logs = slices.Delete(s.logs, 0, len(s.logs)-maxMemoryLogs+1)
	}
	s.logs = append(s.logs, e)
	s.logsMutex.Unlock()

	s.sink.Load().write(e)
}

type PanicError struct {
	Trace string
	Cause error
}

func (e *PanicError) Error() string {
	return fmt.Sprintf("panic: %v", e.Cause)
}

func (e *PanicError) Unwrap() error {
	return e.Cause
}

func LoggerFrom(ctx context.Context) *slog.Logger {
	return logging.FromContext(ctx)
}

type slogHandler struct {
	sched  *Scheduler
	attrs  []slog.Attr
	groups []string
}

func (s slogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (s slogHandler) Handle(ctx context.Context, record slog.Record) error {
	values := make(map[string]any, len(s.attrs)+record.NumAttrs())
	for _, attr := range s.attrs {
		putAttr(values, "", attr)
	}

	prefix := groupPrefix(s.groups)
	record.Attrs(func(attr slog.Attr) bool {
		putAttr(values, prefix, attr)
		return true
	})

	if len(values) == 0 {
		values = nil
	}

	// slog stamps the system time, an inserted clock rules instead
	t := record.Time
	if _, system := s.sched.clock.(systemClock); !system || t.IsZero() {
		t = s.sched.clock.Now()
	}

	s.sched.record(LogEntry{Level: record.Level, Time: t, Msg: record.Message, Values: values})

	if s.sched.options().Quiet {
		return nil
	}

	// keep forwarding into the default logger, as before
	if def := slog.Default().Handler(); def.Enabled(ctx, record.Level) {
		r := record.Clone()
		r.AddAttrs(slog.String("scheduler", string(s.sched.id)))
		_ = slog.Default().Handler().WithAttrs(s.attrs).Handle(ctx, r)
	}

	return nil
}

func groupPrefix(groups []string) string {
	if len(groups) == 0 {
		return ""
	}

	return strings.Join(groups, ".") + "."
}

func putAttr(dst map[string]any, prefix string, attr slog.Attr) {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return
	}

	if attr.Value.Kind() == slog.KindGroup {
		p := prefix
		if attr.Key != "" {
			p = prefix + attr.Key + "."
		}

		for _, a := range attr.Value.Group() {
			putAttr(dst, p, a)
		}

		return
	}

	dst[prefix+attr.Key] = attr.Value.Any()
}

func (s slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	prefix := groupPrefix(s.groups)
	tmp := slices.Clone(s.attrs)
	for _, a := range attrs {
		a.Key = prefix + a.Key
		tmp = append(tmp, a)
	}

	return slogHandler{sched: s.sched, attrs: tmp, groups: s.groups}
}

func (s slogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return s
	}

	return slogHandler{sched: s.sched, attrs: s.attrs, groups: append(slices.Clone(s.groups), name)}
}

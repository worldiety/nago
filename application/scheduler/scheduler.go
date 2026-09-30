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
	opts            Options
	cancel          func()
	logs            []LogEntry
	logsMutex       sync.Mutex
	singleRunMutex  sync.Mutex
	lastStartedAt   atomic.Pointer[time.Time]
	lastCompletedAt atomic.Pointer[time.Time]
	nextPlannedAt   atomic.Pointer[time.Time]
	launchMutex     sync.Mutex                 // protects ctx and cancel and serializes the (re)launch of the loop
	loop            atomic.Pointer[loopHandle] // the current loop of Launch or nil
	wake            chan struct{}              // signals changed settings to the loop
	runs            *runStore
	sink            atomic.Pointer[runSink]
	currentRun      atomic.Pointer[Run]
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
	s := &Scheduler{
		externalCtx:  ctx,
		ctx:          ctx,
		cancel:       func() {},
		opts:         opts,
		settingsRepo: settingsRepo,
		wake:         make(chan struct{}, 1),
	}

	var zeroTime time.Time
	s.lastStartedAt.Store(&zeroTime)
	s.lastCompletedAt.Store(&zeroTime)
	s.nextPlannedAt.Store(&zeroTime)
	state := Stopped
	s.state.Store(&state)

	return s
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
	// completes its run does not overwrite the state of a newer loop
	setState := func(state State) {
		if s.loop.Load() == h && h.alive() {
			s.state.Store(&state)
		}
	}

	go func() {
		defer func() {
			// Only the current loop clears the handle and sets the state. The mutex ensures that a concurrent Launch
			// happens either before, so that the CompareAndSwap fails, or afterward, so that the new loop sets its
			// state later.
			s.launchMutex.Lock()
			defer s.launchMutex.Unlock()

			if s.loop.CompareAndSwap(h, nil) {
				state := Stopped
				s.state.Store(&state)
			}
		}()

	loop:
		for {
			settings, err := s.loadSettings()
			if err != nil {
				slog.Error("service looper failed to load settings", "id", s.opts.ID, "err", err.Error())
				return
			}

			if settings.Disabled || s.opts.Kind == Manual {
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

			if s.opts.Kind != Cron {
				if err := s.execute(ctx); err != nil {
					slog.Error("service looper failed to run", "id", s.opts.ID, "err", err.Error())
				}
			}

			pauseTime := settings.PauseTime

			switch s.opts.Kind {
			case OneShot, Manual:
				var zeroT time.Time
				s.nextPlannedAt.Store(&zeroT)
				return
			case Schedule:
				// do nothing, go ahead and sleep
				nextPlannedAt := time.Now().Add(settings.PauseTime)
				s.nextPlannedAt.Store(&nextPlannedAt)
			case Cron:
				nextPlannedAt := nextCronAt(time.Now(), settings)
				pauseTime = time.Until(nextPlannedAt)
				s.nextPlannedAt.Store(&nextPlannedAt)
			}

			// wait the pause-delay or exit early on cancel
			setState(Paused)
			// do not schedule faster than 1 second, everything else is probably a configuration mistake
			deadline := time.Now().Add(max(pauseTime, time.Second))
			for {
				woken, alive := s.sleep(ctx, time.Until(deadline))
				if !alive {
					slog.Info("service shutdown due to context signal", "id", s.opts.ID)
					return
				}

				if !woken {
					break
				}

				// the settings have been changed while pausing
				if s.opts.Kind != Schedule {
					continue loop // a cron plan is recalculated from scratch
				}

				settings, err = s.loadSettings()
				if err != nil || settings.Disabled {
					continue loop
				}

				// keep the distance to the last run, but with the new pause time
				deadline = s.LastCompletedAt().Add(max(settings.PauseTime, time.Second))
				s.nextPlannedAt.Store(&deadline)
			}

			if s.opts.Kind == Cron {
				if err := s.execute(ctx); err != nil {
					slog.Error("service looper cron failed to run", "id", s.opts.ID, "err", err.Error())
				}
			}
		}
	}()
}

// nextCronAt returns the next planned execution after now.
func nextCronAt(now time.Time, settings Settings) time.Time {
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	nextPlannedAt := startOfDay.
		Add(time.Duration(settings.CronHour) * time.Hour).
		Add(time.Duration(settings.CronMinute) * time.Minute)

	if settings.PauseTime > 0 {
		for !nextPlannedAt.After(now) {
			nextPlannedAt = nextPlannedAt.Add(settings.PauseTime)
		}
	} else if nextPlannedAt.Before(now) {
		nextPlannedAt = nextPlannedAt.Add(24 * time.Hour)
	}

	return nextPlannedAt
}

// loadSettings returns the persisted settings or the defaults.
func (s *Scheduler) loadSettings() (Settings, error) {
	optSettings, err := s.settingsRepo.FindByID(s.opts.ID)
	if err != nil {
		return Settings{}, err
	}

	if optSettings.IsSome() {
		return optSettings.Unwrap(), nil
	}

	return s.opts.Defaults, nil
}

// sleep waits for d. It returns woken=true, if the settings have been changed in the meantime (see [Scheduler.Wake])
// and alive=false, if the scheduler has been cancelled.
func (s *Scheduler) sleep(ctx context.Context, d time.Duration) (woken, alive bool) {
	timer := time.NewTimer(max(d, 0))
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false, false
	case <-s.wake:
		return true, true
	case <-timer.C:
		return false, true
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

// execute runs the runner once with the given context and tracks its start. A cancelled context, e.g. of a loop
// which has been stopped while its timer fired, is not executed at all.
func (s *Scheduler) execute(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return s.protectExec(func() error {
		return s.opts.Runner(ctx)
	})
}

func (s *Scheduler) protectExec(fn func() error) (err error) {
	s.singleRunMutex.Lock()
	defer s.singleRunMutex.Unlock()

	s.ClearLogs()
	run, sink, beginErr := s.beginRun()
	defer func() {
		if err != nil {
			s.logError(err)
		}

		if beginErr == nil {
			s.sink.Store(nil)
			s.currentRun.Store(nil)
			s.runs.end(run, sink, err)
		}
	}()

	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			err = &PanicError{Trace: string(debug.Stack()), Cause: fmt.Errorf("recovered from panic: %v", r)}
		}
	}()

	defer func() {
		// without a loop, e.g. a stopped scheduler executed manually, nothing is paused
		state := Stopped
		if s.Looping() {
			state = Paused
		}
		s.state.Store(&state)

		doneAt := time.Now()
		s.lastCompletedAt.Store(&doneAt)
	}()

	startedAt := time.Now()
	s.lastStartedAt.Store(&startedAt)
	state := Running
	s.state.Store(&state)

	err = fn()
	return
}

// ExecuteNow runs the runner immediately and blocks until it completes. It waits for a run of the loop which is
// in progress.
func (s *Scheduler) ExecuteNow() error {
	s.launchMutex.Lock()
	if s.ctx.Err() != nil {
		s.resetContext()
	}
	ctx := s.ctx
	s.launchMutex.Unlock()

	return s.execute(ctx)
}

func (s *Scheduler) beginRun() (Run, *runSink, error) {
	if s.runs == nil {
		return Run{}, nil, errors.New("no run store")
	}

	run, sink, err := s.runs.begin(s.opts.ID)
	if err != nil {
		slog.Error("cannot begin scheduler run", "id", s.opts.ID, "err", err.Error())
		return Run{}, nil, err
	}

	s.currentRun.Store(&run)
	s.sink.Store(sink)
	return run, sink, nil
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

func (s *Scheduler) logError(err error) {
	s.Error("service looper encountered error", "err", err.Error())
	s.lastError.Store(&err)
}

func (s *Scheduler) Info(msg string, args ...any) {
	slog.Info(msg, args...)
	s.logLevel(slog.LevelInfo, msg, args...)
}

func (s *Scheduler) Error(msg string, args ...any) {
	slog.Error(msg, args...)
	s.logLevel(slog.LevelError, msg, args...)
}

func (s *Scheduler) Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
	s.logLevel(slog.LevelWarn, msg, args...)
}

func (s *Scheduler) Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
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

	s.record(LogEntry{Level: level, Time: time.Now(), Msg: msg, Values: tmp})
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

	t := record.Time
	if t.IsZero() {
		t = time.Now()
	}

	s.sched.record(LogEntry{Level: record.Level, Time: t, Msg: record.Message, Values: values})

	// keep forwarding into the default logger, as before
	if def := slog.Default().Handler(); def.Enabled(ctx, record.Level) {
		r := record.Clone()
		r.AddAttrs(slog.String("scheduler", string(s.sched.opts.ID)))
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

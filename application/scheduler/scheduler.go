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
	launchMutex     sync.Mutex
	runs            *runStore
	sink            atomic.Pointer[runSink]
	currentRun      atomic.Pointer[Run]
}

// maxMemoryLogs limits the in-memory log buffer of the current run. The complete log is persisted in the run log file.
const maxMemoryLogs = 1000

func NewScheduler(ctx context.Context, opts Options, settingsRepo SettingsRepository) *Scheduler {
	s := &Scheduler{
		externalCtx:  ctx,
		ctx:          ctx,
		cancel:       func() {},
		opts:         opts,
		settingsRepo: settingsRepo,
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

func (s *Scheduler) Destroy() {
	s.cancel()
}

func (s *Scheduler) ResetContext() {
	s.launchMutex.Lock()
	defer s.launchMutex.Unlock()

	s.cancel()

	ctx := logging.WithContext(s.externalCtx, slog.New(slogHandler{sched: s}))

	myCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.ctx = myCtx
}

func (s *Scheduler) Launch() {
	s.ResetContext()

	go func() {
		defer func() {
			state := Stopped
			s.state.Store(&state)
		}()

		for {
			optSettings, err := s.settingsRepo.FindByID(s.opts.ID)
			if err != nil {
				slog.Error("service looper failed to load settings", "id", s.opts.ID, "err", err.Error())
				return
			}

			settings := s.opts.Defaults
			if optSettings.IsSome() {
				settings = optSettings.Unwrap()
			}

			if settings.Disabled || s.opts.Kind == Manual {
				state := Disabled
				s.state.Store(&state)
				// wait the config-reload time or exit early on cancel
				select {
				case <-s.ctx.Done():
					slog.Info("service shutdown due to context signal")
					return
				case <-time.After(time.Minute):
					continue
				}
			} else {
				state := Paused
				s.state.Store(&state)
				// wait the start-delay or exit early on cancel
				select {
				case <-s.ctx.Done():
					slog.Info("service shutdown due to context signal")
					return
				case <-time.After(settings.StartDelay):
				}

				// perform the actual work execution

				if s.opts.Kind != Cron {
					startedAt := time.Now()
					s.lastStartedAt.Store(&startedAt)
					err := s.protectExec(func() error {
						return s.opts.Runner(s.ctx)
					})

					if err != nil {
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
					now := time.Now()
					startOfDay := time.Date(
						now.Year(), now.Month(), now.Day(),
						0, 0, 0, 0,
						now.Location(),
					)

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

					pauseTime = nextPlannedAt.Sub(now)
					s.nextPlannedAt.Store(&nextPlannedAt)
				}

				// wait the pause-delay or exit early on cancel
				state = Paused
				s.state.Store(&state)
				select {
				case <-s.ctx.Done():
					slog.Info("service shutdown due to context signal", "id", s.opts.ID)
					return
					// do not schedule faster than 1 second, everything else is probably a configuration mistake
				case <-time.After(max(pauseTime, time.Second)):

					if s.opts.Kind == Cron {
						startedAt := time.Now()
						s.lastStartedAt.Store(&startedAt)
						err := s.protectExec(func() error {
							return s.opts.Runner(s.ctx)
						})

						if err != nil {
							slog.Error("service looper cron failed to run", "id", s.opts.ID, "err", err.Error())
						}

					}

					continue
				}
			}

		}
	}()
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
		state := Paused
		s.state.Store(&state)

		doneAt := time.Now()
		s.lastCompletedAt.Store(&doneAt)
	}()

	state := Running
	s.state.Store(&state)

	err = fn()
	return
}

func (s *Scheduler) ExecuteNow() error {
	if s.ctx.Err() != nil {
		s.ResetContext()
	}
	return s.protectExec(func() error {
		return s.opts.Runner(s.ctx)
	})
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

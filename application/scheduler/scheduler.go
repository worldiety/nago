// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler

import (
	"context"
	"fmt"
	"go.wdy.de/nago/logging"
	"log/slog"
	"runtime/debug"
	"slices"
	"sync"
	"sync/atomic"
	"time"
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
}

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
						s.logError(err)
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

					if nextPlannedAt.Before(now) {
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
							s.logError(err)
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
	s.ClearLogs()

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
	s.logsMutex.Lock()
	defer s.logsMutex.Unlock()

	// TODO remove this and make me real slog handler

	if len(args)%2 != 0 {
		slog.Error("invalid arguments in log level")
		debug.PrintStack()
		args = nil
	}
	var tmp map[string]any
	if len(args) > 0 {
		tmp = make(map[string]any, len(args))
		for i := 0; i < len(args); i += 2 {
			k := args[i]
			if v, ok := k.(string); ok && v != "" {
				tmp[v] = args[i+1]
			} else {
				tmp[fmt.Sprint(s.logs[i])] = args[i+1]
			}

		}
	}

	s.logs = append(s.logs, LogEntry{
		Level:  level,
		Time:   time.Now(),
		Msg:    msg,
		Values: tmp,
	})

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
	sched *Scheduler
	attrs []slog.Attr
	group string
}

func (s slogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (s slogHandler) Handle(ctx context.Context, record slog.Record) error {
	var allAttr []slog.Attr
	for _, attr := range s.attrs {
		allAttr = append(allAttr, attr)
	}

	record.Attrs(func(attr slog.Attr) bool {
		allAttr = append(allAttr, attr)
		return true
	})

	// TODO remove me and make a real handler
	tmp := ""
	for _, a := range allAttr {
		tmp = tmp + a.String() + " "
	}

	switch record.Level {
	case slog.LevelDebug:
		s.sched.Debug(record.Message, "attr", tmp)
	case slog.LevelInfo:
		s.sched.Info(record.Message, "attr", tmp)
	case slog.LevelWarn:
		s.sched.Warn(record.Message, "attr", tmp)
	default:
		s.sched.Error(record.Message, "attr", tmp)

	}

	return nil
}

func (s slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	s.attrs = append(s.attrs, attrs...)
	return s
}

func (s slogHandler) WithGroup(name string) slog.Handler {
	s.group = name
	return s
}

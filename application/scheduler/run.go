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
	"log/slog"
	"time"

	"go.wdy.de/nago/pkg/data"
)

const (
	// MaxLogPageSize is the upper bound of log entries returned by a single [ViewRunLog] call.
	MaxLogPageSize = 50
	// KeepRunLogs is the amount of log files kept per scheduler. Older log files are deleted.
	KeepRunLogs = 5
	// RunRetention is the duration for which run statistics are kept.
	RunRetention = 30 * 24 * time.Hour
)

var (
	// ErrRunLogNotAvailable is returned by [ViewRunLog], if the log of a run has already been deleted,
	// see [KeepRunLogs], or has never been written.
	ErrRunLogNotAvailable = errors.New("run log not available")
	// ErrRunNotFound is returned by [ViewRunLog], if the run is unknown, e.g. because it is older than [RunRetention].
	ErrRunNotFound = errors.New("run not found")
	// ErrRunning is returned by [Manager.StartNow], while a run of the scheduler is in progress.
	ErrRunning = errors.New("scheduler is running")
	// ErrNotRunning is returned by [Manager.CancelRun], if no run of the scheduler is in progress.
	ErrNotRunning = errors.New("scheduler is not running")
	// ErrLogNotWritable is the error of a run of a [Options.Quiet] scheduler, whose log could not be written.
	ErrLogNotWritable = errors.New("Protokoll nicht schreibbar")
)

// CanceledByRestart is [Run.CanceledBy] of a run of a durable scheduler, which was in progress, when its manager
// stopped, or which had no end, when the scheduler was configured again, see [Options.Durable].
const CanceledByRestart = "Neustart"

// Trigger tells, why a run started.
type Trigger int

const (
	// TriggerPlan is a run at a due of the plan.
	TriggerPlan Trigger = iota
	// TriggerCatchUp is a run which catches up a due, which elapsed while the manager was not running, see
	// [Settings.CatchUp].
	TriggerCatchUp
	// TriggerManual is a run by [Manager.StartNow] or [Manager.ExecuteNow].
	TriggerManual
)

// RunInfo describes the current run to its runner, see [RunInfoFrom].
type RunInfo struct {
	// Due is the due of the plan, for a caught up run the latest elapsed due and for a manual run its start.
	Due     time.Time
	Trigger Trigger
	// By is the name of the person, who started a manual run.
	By string
}

type runInfoKey struct{}

// RunInfoFrom returns the information about the run, which the runner executes with the given context.
func RunInfoFrom(ctx context.Context) (RunInfo, bool) {
	if ctx == nil {
		return RunInfo{}, false
	}

	info, ok := ctx.Value(runInfoKey{}).(RunInfo)
	return info, ok
}

func withRunInfo(ctx context.Context, info RunInfo) context.Context {
	return context.WithValue(ctx, runInfoKey{}, info)
}

// RunID identifies a single execution of a scheduler, e.g. my.scheduler#2026-09-24-00001.
type RunID string

type Outcome int

const (
	OutcomeRunning Outcome = iota
	OutcomeSucceeded
	OutcomeFailed
	OutcomePanicked
	OutcomeCanceled
)

func (o Outcome) Failed() bool {
	return o == OutcomeFailed || o == OutcomePanicked
}

// Run is the persisted statistic of a single scheduler execution.
type Run struct {
	ID          RunID     `json:"id"`
	Scheduler   ID        `json:"scheduler"`
	Nr          int       `json:"nr"` // consecutive number per scheduler and day, starting at 1
	StartedAt   time.Time `json:"startedAt"`
	CompletedAt time.Time `json:"completedAt,omitzero"`
	Outcome     Outcome   `json:"outcome"`
	Error       string    `json:"error,omitempty"`
	Entries     int       `json:"entries"`
	Warnings    int       `json:"warnings"`
	Errors      int       `json:"errors"`
	// LogFile is the base name of the log file within the log directory. Empty, if no log is available (anymore).
	LogFile string `json:"logFile,omitempty"`
	// Trigger tells, why the run started.
	Trigger Trigger `json:"trigger,omitempty"`
	// By is the name of the person, who started a manual run.
	By string `json:"by,omitempty"`
	// Due is the due of the plan, see [RunInfo.Due].
	Due time.Time `json:"due,omitzero"`
	// CanceledBy is the name of the person, who cancelled the run by [Manager.CancelRun], or [CanceledByRestart].
	CanceledBy string `json:"canceledBy,omitempty"`
	// Durable is only set at the record, which keeps the state of a durable scheduler in the run repository next
	// to its runs (see [Options.Durable]). This record is no run and never returned as one.
	Durable *DurableState `json:"durable,omitempty"`
}

// DurableState is what a durable scheduler keeps over restarts, independent of its runs, which are deleted after
// [RunRetention], see [Options.Durable].
type DurableState struct {
	// LastDue is the latest handled due: a due which ran (also if it was cancelled) or which was omitted, because
	// a run was in progress.
	LastDue time.Time `json:"lastDue,omitzero"`
	// LastEnd is the end of the latest run, also of a manual one.
	LastEnd time.Time `json:"lastEnd,omitzero"`
	// Reference is the latest of the first configuration, the last [Manager.Resume], the last change of the plan,
	// [Manager.SetReference] and the start of the last manual run.
	Reference time.Time `json:"reference,omitzero"`
	// Paused is true between [Manager.Pause] and [Manager.Resume].
	Paused bool `json:"paused,omitempty"`
	// LogFloor is the oldest run which may still have a log file, so that deleting old logs reads only the runs
	// whose log it deletes.
	LogFloor RunID `json:"logFloor,omitempty"`
}

func (r Run) Identity() RunID {
	return r.ID
}

func (r Run) WithIdentity(id RunID) Run {
	r.ID = id
	return r
}

// Duration returns the execution time or the time since start, if still running.
func (r Run) Duration() time.Duration {
	if r.CompletedAt.IsZero() {
		if r.Outcome == OutcomeRunning {
			return time.Since(r.StartedAt)
		}
		return 0
	}

	return r.CompletedAt.Sub(r.StartedAt)
}

type RunRepository data.Repository[Run, RunID]

// RunStats aggregates the runs of the last 24 hours.
type RunStats struct {
	Total24h     int
	Failed24h    int
	AvgDuration  time.Duration
	LastDuration time.Duration
}

// LogQuery selects a page of log entries. Entries are always returned newest first.
type LogQuery struct {
	Offset int
	// Limit is clamped to [1, MaxLogPageSize].
	Limit int
	// MinLevel filters entries. Use [slog.LevelDebug] to include everything. The zero value is [slog.LevelInfo].
	MinLevel slog.Level
	// Text is a case-insensitive substring filter applied to the message and the values.
	Text string
}

type LogPage struct {
	Entries []LogEntry
	// Total is the amount of entries which match the query.
	Total  int
	Offset int
}

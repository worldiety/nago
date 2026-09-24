// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler

import (
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
)

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

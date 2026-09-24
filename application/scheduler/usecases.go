// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler

import (
	"context"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"iter"
	"time"
)

type ID string

type State int

const (
	Stopped State = iota
	Running
	Disabled
	Paused
)

type Kind int

const (
	OneShot Kind = iota
	Schedule
	Manual
	Cron
)

type CustomAction struct {
	Title  string
	Action func(ctx context.Context)
}
type Options struct {
	ID          ID
	Name        string
	Description string
	Kind        Kind
	Defaults    Settings
	Runner      func(context.Context) error
	Actions     []CustomAction
}

type StatusResult struct {
	State           State
	LastStartedAt   time.Time
	LastCompletedAt time.Time
	NextPlannedAt   time.Time
	LastError       error
	Options         Options

	// LastRun is the newest known run, which may still be running.
	LastRun std.Option[Run]
	// Stats aggregates the runs of the last 24 hours.
	Stats RunStats
	// Settings are the effective settings, i.e. the persisted ones or the defaults.
	Settings Settings
	// CustomSettings is true, if persisted settings replace the defaults.
	CustomSettings bool
}
type Status func(subject auth.Subject, id ID) (StatusResult, error)

// Configure introduces a new system level service. It is not intended, that these services are configured by
// end users. Usually, a developer defines the schedulers at build time.
type Configure func(subject auth.Subject, opts Options) error

type ViewLogs func(subject auth.Subject, id ID) iter.Seq2[LogEntry, error]
type ExecuteNow func(subject auth.Subject, id ID) error

// ListRuns returns the known runs of a scheduler, newest first. Statistics are kept for [RunRetention],
// logs only for the newest [KeepRunLogs] runs.
type ListRuns func(subject auth.Subject, id ID) iter.Seq2[Run, error]

// ViewRunLog returns a single page of at most [MaxLogPageSize] log entries of the given run, newest first.
type ViewRunLog func(subject auth.Subject, id ID, run RunID, query LogQuery) (LogPage, error)

type ListSchedulers func(subject auth.Subject) iter.Seq2[Options, error]

type Stop func(subject auth.Subject, id ID) error
type Start func(subject auth.Subject, id ID) error

type FindSettingsByID func(subject auth.Subject, id ID) (std.Option[Settings], error)
type UpdateSettings func(subject auth.Subject, settings Settings) error
type DeleteSettingsByID func(subject auth.Subject, id ID) error

type UseCases struct {
	Configure          Configure
	ViewLogs           ViewLogs
	Status             Status
	ExecuteNow         ExecuteNow
	ListSchedulers     ListSchedulers
	Stop               Stop
	Start              Start
	FindSettingsByID   FindSettingsByID
	UpdateSettings     UpdateSettings
	DeleteSettingsByID DeleteSettingsByID
	ListRuns           ListRuns
	ViewRunLog         ViewRunLog
}

// NewUseCases creates the use cases without persistent run statistics and without log files.
func NewUseCases(ctx context.Context, settingsRepo SettingsRepository) UseCases {
	return NewUseCasesWithRuns(ctx, settingsRepo, nil, "")
}

// NewUseCasesWithRuns creates the use cases which persist run statistics into runRepo and the log of each
// run into its own file within logDir, see [NewManagerWithPersistence].
func NewUseCasesWithRuns(ctx context.Context, settingsRepo SettingsRepository, runRepo RunRepository, logDir string) UseCases {
	m := NewManagerWithPersistence(ctx, settingsRepo, runRepo, logDir)
	return UseCases{
		Configure:          NewConfigure(m),
		ViewLogs:           NewViewLogs(m),
		Status:             NewStatus(m),
		ExecuteNow:         NewExecuteNow(m),
		ListSchedulers:     NewListSchedulers(m),
		Stop:               NewStop(m),
		Start:              NewStart(m),
		FindSettingsByID:   NewFindSettingsByID(settingsRepo),
		UpdateSettings:     NewUpdateSettings(settingsRepo),
		DeleteSettingsByID: NewDeleteSettingsByID(settingsRepo),
		ListRuns:           NewListRuns(m),
		ViewRunLog:         NewViewRunLog(m),
	}
}

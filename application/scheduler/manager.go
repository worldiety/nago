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
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
	"iter"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"
)

type Manager struct {
	ctx          context.Context
	mutex        sync.Mutex
	services     map[ID]*Scheduler
	settingsRepo SettingsRepository
	runs         *runStore
}

// NewManager creates a manager which keeps run statistics only in memory and writes no log files.
func NewManager(ctx context.Context, settingsRepo SettingsRepository) *Manager {
	return NewManagerWithPersistence(ctx, settingsRepo, nil, "")
}

// NewManagerWithPersistence creates a manager which persists run statistics into the given repository and
// writes the log of each run into its own file within logDir. If runRepo is nil, an in-memory repository is used.
// If logDir is empty, no log files are written.
func NewManagerWithPersistence(ctx context.Context, settingsRepo SettingsRepository, runRepo RunRepository, logDir string) *Manager {
	if runRepo == nil {
		runRepo = json.NewSloppyJSONRepository[Run, RunID](mem.NewBlobStore("nago.scheduler.runs"))
	}

	return &Manager{ctx: ctx, services: make(map[ID]*Scheduler), settingsRepo: settingsRepo, runs: newRunStore(runRepo, logDir)}
}

func (m *Manager) Configure(opts Options) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.services[opts.ID]; ok {
		return fmt.Errorf("service with id %s already exists", opts.ID)
	}

	if opts.Runner == nil {
		return fmt.Errorf("runner is required")
	}

	s := NewScheduler(m.ctx, opts, m.settingsRepo)
	s.runs = m.runs
	m.services[opts.ID] = s
	s.Launch()

	return nil
}

func (m *Manager) Start(id ID) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	s, ok := m.services[id]
	if !ok {
		return fmt.Errorf("service with id %s not found", id)
	}

	if s.State() == Running {
		slog.Info("service already started")
		return nil
	}

	s.Launch()
	return nil
}

func (m *Manager) Stop(id ID) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	s, ok := m.services[id]
	if !ok {
		return fmt.Errorf("service with id %s not found", id)
	}

	if s.State() == Stopped {
		slog.Info("service already stopped")
		return nil
	}

	s.Destroy()
	return nil
}

func (m *Manager) LastStartedAt(id ID) time.Time {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s, ok := m.services[id]
	if !ok {
		return time.Time{}
	}

	return s.LastStartedAt()
}

func (m *Manager) LastCompletedAt(id ID) time.Time {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s, ok := m.services[id]
	if !ok {
		return time.Time{}
	}

	return s.LastCompletedAt()
}

func (m *Manager) NextPlannedAt(id ID) time.Time {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s, ok := m.services[id]
	if !ok {
		return time.Time{}
	}

	return s.NextPlannedAt()
}

func (m *Manager) Logs(id ID) []LogEntry {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s, ok := m.services[id]
	if !ok {
		return nil
	}

	tmp := s.Logs()
	slices.Reverse(tmp)
	return tmp
}

func (m *Manager) LastError(id ID) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	s, ok := m.services[id]
	if !ok {
		return nil
	}

	return s.LastError()
}

func (m *Manager) State(id ID) State {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	s, ok := m.services[id]
	if !ok {
		return Stopped
	}

	return s.State()
}

func (m *Manager) ExecuteNow(id ID) error {
	m.mutex.Lock()

	s, ok := m.services[id]
	if !ok {
		m.mutex.Unlock()
		return fmt.Errorf("service with id %s not found", id)
	}

	m.mutex.Unlock()

	// ensure, that we execute without the manager lock
	return s.ExecuteNow()
}

func (m *Manager) Options(id ID) (Options, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s, ok := m.services[id]
	if !ok {
		return Options{}, false
	}

	return s.opts, true
}

// Scheduler returns the configured options for all schedulers sorted by name ascending.
func (m *Manager) Scheduler() []Options {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	tmp := make([]Options, 0, len(m.services))
	for _, s := range m.services {
		tmp = append(tmp, s.opts)
	}

	slices.SortFunc(tmp, func(a, b Options) int {
		return strings.Compare(a.Name, b.Name)
	})

	return tmp
}

func (m *Manager) scheduler(id ID) (*Scheduler, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s, ok := m.services[id]
	return s, ok
}

// Runs returns the known runs of the given scheduler, newest first.
func (m *Manager) Runs(id ID) ([]Run, error) {
	if _, ok := m.scheduler(id); !ok {
		return nil, fmt.Errorf("service with id %s not found", id)
	}

	return m.runs.runs(id)
}

// Stats aggregates the runs of the last 24 hours.
func (m *Manager) Stats(id ID) (RunStats, []Run, error) {
	runs, err := m.Runs(id)
	if err != nil {
		return RunStats{}, nil, err
	}

	return stats(runs, m.runs.now()), runs, nil
}

// RunLog returns a single page of the log of the given run.
func (m *Manager) RunLog(id ID, runID RunID, q LogQuery) (LogPage, error) {
	s, ok := m.scheduler(id)
	if !ok {
		return LogPage{}, fmt.Errorf("service with id %s not found", id)
	}

	optRun, err := m.runs.repo.FindByID(runID)
	if err != nil {
		return LogPage{}, err
	}

	if optRun.IsNone() || optRun.Unwrap().Scheduler != id {
		return LogPage{}, fmt.Errorf("run %s: %w", runID, ErrRunNotFound)
	}

	run := optRun.Unwrap()

	// without a log file, only the current run is available from memory
	if run.LogFile == "" || m.runs.dir == "" {
		if cur, ok := s.CurrentRun(); ok && cur.ID == run.ID {
			logs := s.Logs()
			return pageEntries(func() iter.Seq[LogEntry] { return slices.Values(logs) }, q), nil
		}
	}

	entries, err := m.runs.fileEntries(run)
	if err != nil {
		return LogPage{}, err
	}

	return pageEntries(func() iter.Seq[LogEntry] { return entries }, q), nil
}

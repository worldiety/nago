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
}

func NewManager(ctx context.Context, settingsRepo SettingsRepository) *Manager {
	return &Manager{ctx: ctx, services: make(map[ID]*Scheduler), settingsRepo: settingsRepo}
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

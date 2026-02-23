// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package migration

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"reflect"
	"sync"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/pkg/xtime"
)

// Version is sorted alphabetically and executed in ascending order by default. Even though the version may
// be an arbitrary string, it is strongly recommended to use the factory [NewVersion].
type Version string

func NewVersion(year int, month time.Month, day int, hour, minute int, name string) Version {
	return Version(fmt.Sprintf("%04d%02d%02d%02d%02d_%s", year, month, day, hour, minute, name))
}

// Migration is the common interface for nago orchestrated data migrations. Once deployed and applied, a Migration
// implementation must be treated as immutable. We cannot calculate a stable hash sum based on the transformation a
// migration applies (as it would be possible with sql-based migrations).
type Migration interface {
	Version() Version
	Migrate(ctx context.Context) error
}

type Status struct {
	Version     Version                `json:"version"`
	Installed   bool                   `json:"installed,omitempty"`
	Error       string                 `json:"error,omitempty"`
	InstalledAt xtime.UnixMilliseconds `json:"installedAt,omitempty"`
	Script      string                 `json:"script,omitempty"`
}

func (s Status) Identity() Version {
	return s.Version
}

type Options struct {
	Context   context.Context
	Immediate bool // if true, the migration is applied immediately, independent of any order. You should normally don't do that.
}

type Migrations struct {
	repo       data.Repository[Status, Version]
	migrations concurrent.RWMap[Version, Migration]
	mutex      sync.Mutex
}

func NewMigrations(repo data.Repository[Status, Version]) *Migrations {
	return &Migrations{repo: repo}
}

// Declare registers a migration implementation but does not necessarily apply it yet.
// See also Options and [Migrations.Apply].
func (m *Migrations) Declare(mg Migration, opts Options) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.migrations.Get(mg.Version()); ok {
		return fmt.Errorf("migration version %v already declared", mg.Version())
	}

	m.migrations.Put(mg.Version(), mg)

	optStat, err := m.repo.FindByID(mg.Version())
	if err != nil {
		return err
	}

	var stat Status
	if optStat.IsNone() {
		t := reflect.TypeOf(mg)
		if err := m.repo.Save(Status{
			Version: mg.Version(),
			Script:  fmt.Sprintf("%s.%s", t.PkgPath(), t.Name()),
		}); err != nil {
			return err
		}
	} else {
		stat = optStat.Unwrap()
	}

	if !stat.Installed && opts.Immediate {
		ctx := opts.Context
		if ctx == nil {
			ctx = context.Background()
		}

		return m.apply(ctx, mg)
	}

	return nil
}

func (m *Migrations) apply(ctx context.Context, mg Migration) error {
	optStat, err := m.repo.FindByID(mg.Version())
	if err != nil {
		return err
	}

	if optStat.IsNone() {
		return fmt.Errorf("migration version %v not declared", mg.Version())
	}

	stat := optStat.Unwrap()

	slog.Info("applying migration", "version", mg.Version(), "script", stat.Script)

	if err := mg.Migrate(ctx); err != nil {
		slog.Error("migration failed", "version", stat.Version, "script", stat.Script, "err", err.Error())
		stat.Installed = false
		stat.Error = err.Error()
		stat.InstalledAt = xtime.Now()
		if err := m.repo.Save(stat); err != nil {
			slog.Error("failed to save migration status", "version", stat.Version, "script", stat.Script, "err", err.Error())
		}

		return err
	}

	stat.Installed = true
	stat.InstalledAt = xtime.Now()
	stat.Error = ""

	if err := m.repo.Save(stat); err != nil {
		return err
	}

	slog.Info("migration applied successfully", "version", mg.Version(), "script", stat.Script)

	return nil
}

// ReApply re-applies a previously applied or failed migration immediately.
func (m *Migrations) ReApply(ctx context.Context, version Version) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	mg, ok := m.migrations.Get(version)
	if !ok {
		return fmt.Errorf("migration version %v not declared", version)
	}

	return m.apply(ctx, mg)
}

// Apply applies all declared migrations that are not yet applied.
func (m *Migrations) Apply(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// note that we are already in correct order due to repository guarantees
	for stat, err := range m.repo.All() {
		if err != nil {
			return err
		}

		if stat.Installed {
			continue
		}

		mg, ok := m.migrations.Get(stat.Version)
		if !ok {
			slog.Error("declared but not never applied migration found which is now missing, ignoring", "version", stat.Version, "script", stat.Script)
			continue
		}

		if err := m.apply(ctx, mg); err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrations) Versions() iter.Seq2[Version, error] {
	return m.repo.Identifiers()
}

func (m *Migrations) Status(version Version) (option.Opt[Status], error) {
	return m.repo.FindByID(version)
}

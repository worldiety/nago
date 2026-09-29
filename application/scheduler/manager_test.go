// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"go.wdy.de/nago/application/user"
)

// eventually polls cond, because the scheduler loop runs in its own goroutine.
func eventually(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("condition not met in time")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestRemoveAndConfigureAgain(t *testing.T) {
	m, _ := newTestManager(t, "")
	uc := UseCases{Remove: NewRemove(m), UpdateSettings: NewUpdateSettings(m, m.settingsRepo)}

	configureManual(t, m, "job", func(ctx context.Context) error { return nil })
	if err := uc.UpdateSettings(user.SU(), Settings{ID: "job", PauseTime: time.Hour}); err != nil {
		t.Fatal(err)
	}

	if err := uc.Remove(user.SU(), "job"); err != nil {
		t.Fatal(err)
	}

	if m.Has("job") {
		t.Fatal("job must be removed")
	}

	if opt, err := m.settingsRepo.FindByID("job"); err != nil || opt.IsSome() {
		t.Fatalf("settings must be removed: %v %v", opt, err)
	}

	// the id is free again
	configureManual(t, m, "job", func(ctx context.Context) error { return nil })
}

func TestReconfigureReplacesRunner(t *testing.T) {
	m, _ := newTestManager(t, "")
	var calls atomic.Int32
	configureManual(t, m, "job", func(ctx context.Context) error { calls.Add(1); return nil })

	if err := m.Reconfigure(Options{ID: "job", Kind: Manual, Runner: func(ctx context.Context) error {
		calls.Add(10)
		return nil
	}}); err != nil {
		t.Fatal(err)
	}

	if err := m.ExecuteNow("job"); err != nil {
		t.Fatal(err)
	}

	if calls.Load() != 10 {
		t.Fatalf("expected the new runner, got %d", calls.Load())
	}
}

func TestUpdateSettingsRequiresKnownID(t *testing.T) {
	m, _ := newTestManager(t, "")
	if err := NewUpdateSettings(m, m.settingsRepo)(user.SU(), Settings{ID: "unknown"}); err == nil {
		t.Fatal("expected an error for an unknown id")
	}
}

func TestUpdateSettingsApplyImmediately(t *testing.T) {
	m, _ := newTestManager(t, "")
	var calls atomic.Int32
	if err := m.Configure(Options{ID: "job", Kind: Schedule, Defaults: Settings{ID: "job", PauseTime: time.Hour}, Runner: func(ctx context.Context) error {
		calls.Add(1)
		return nil
	}}); err != nil {
		t.Fatal(err)
	}

	// the first run happens without delay, then the job pauses for an hour
	eventually(t, 5*time.Second, func() bool { return calls.Load() == 1 && m.State("job") == Paused })

	// the shorter pause applies to the current pause instead of after it
	if err := NewUpdateSettings(m, m.settingsRepo)(user.SU(), Settings{ID: "job", PauseTime: time.Second}); err != nil {
		t.Fatal(err)
	}

	eventually(t, 5*time.Second, func() bool { return calls.Load() >= 2 })

	// disabling applies immediately as well
	if err := NewUpdateSettings(m, m.settingsRepo)(user.SU(), Settings{ID: "job", PauseTime: time.Second, Disabled: true}); err != nil {
		t.Fatal(err)
	}

	eventually(t, 5*time.Second, func() bool { return m.State("job") == Disabled })
}

func TestExecuteNowTracksStartAndState(t *testing.T) {
	m, _ := newTestManager(t, "")
	configureManual(t, m, "job", func(ctx context.Context) error { return nil })

	if err := m.Stop("job"); err != nil {
		t.Fatal(err)
	}
	m.mutex.Lock()
	s := m.services["job"]
	m.mutex.Unlock()
	eventually(t, 5*time.Second, func() bool { return !s.Looping() })

	before := time.Now()
	if err := m.ExecuteNow("job"); err != nil {
		t.Fatal(err)
	}

	if m.LastStartedAt("job").Before(before) {
		t.Fatalf("last started at must be set, got %v", m.LastStartedAt("job"))
	}

	if m.State("job") != Stopped {
		t.Fatalf("a stopped job must stay stopped, got %v", m.State("job"))
	}
}

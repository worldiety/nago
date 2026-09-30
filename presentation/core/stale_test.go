// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"
)

// captureWarnings records the warnings of the default logger during the test.
func captureWarnings(t *testing.T) *bytes.Buffer {
	var buf bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn})))
	t.Cleanup(func() { slog.SetDefault(old) })
	return &buf
}

func TestDiscardStaleWarnsOnRepetition(t *testing.T) {
	logs := captureWarnings(t)
	s := &Scope{}

	// a double click is normal
	s.discardStale("callback", 1, 1)
	if logs.Len() != 0 {
		t.Fatalf("unexpected warning: %s", logs)
	}

	for range 10 {
		s.discardStale("callback", 1, 1)
	}

	if n := strings.Count(logs.String(), "level=WARN"); n != 1 {
		t.Fatalf("expected a single warning, got %d: %s", n, logs)
	}

	if s.StaleCalls() != 11 || !s.dirty {
		t.Fatalf("unexpected accounting: %d %v", s.StaleCalls(), s.dirty)
	}
}

func TestCheckRenderLoop(t *testing.T) {
	logs := captureWarnings(t)
	w := &scopeWindow{parent: &Scope{}}
	id := "my-state"
	w.lastDirtyState.Store(&id)

	// a long streak within a short time is not reported
	for range 2 * renderLoopCount {
		w.checkRenderLoop(true)
	}

	// a clean render resets the streak
	w.checkRenderLoop(false)
	w.checkRenderLoop(true)
	w.renderLoopSince = time.Now().Add(-2 * renderLoopDuration)
	for range renderLoopCount - 2 {
		w.checkRenderLoop(true)
	}

	if logs.Len() != 0 {
		t.Fatalf("unexpected warning: %s", logs)
	}

	for range 10 {
		w.checkRenderLoop(true)
	}

	if n := strings.Count(logs.String(), "level=WARN"); n != 1 || !strings.Contains(logs.String(), "my-state") {
		t.Fatalf("expected a single warning naming the state, got %d: %s", n, logs)
	}
}

// TestMarkStateDirtyIsMonotonic ensures that an older generation never overwrites a newer marker.
func TestMarkStateDirtyIsMonotonic(t *testing.T) {
	w := &scopeWindow{generation: 5, dirtyGeneration: -1}
	w.markStateDirty(nil)

	w.generation = 3
	w.markStateDirty(nil)

	if w.dirtyGeneration != 5 {
		t.Fatalf("expected 5, got %d", w.dirtyGeneration)
	}
}

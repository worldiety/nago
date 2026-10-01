// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gollama

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCtxMutexGivesUpWhenDone(t *testing.T) {
	var m ctxMutex
	if err := m.lock(context.Background()); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := m.lock(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected to give up, got %v", err)
	}

	m.unlock()
	if err := m.lock(context.Background()); err != nil {
		t.Fatal(err)
	}
	m.unlock()
}

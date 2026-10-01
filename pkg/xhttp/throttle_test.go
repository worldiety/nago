// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xhttp

import (
	"context"
	"errors"
	"testing"
	"time"
)

// A cancelled request gives up waiting for its slot and does not delay the others.
func TestThrottleGivesUpOnCancel(t *testing.T) {
	g := NewRequestGroup().RateLimit(1)
	if err := g.throttle(context.Background()); err != nil {
		t.Fatal(err)
	}

	// another request waits for its slot and holds the sequence meanwhile
	waiting := make(chan error, 1)
	go func() { waiting <- g.throttle(context.Background()) }()
	time.Sleep(20 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := g.throttle(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected the deadline, got %v", err)
	}
	if time.Since(start) > 500*time.Millisecond {
		t.Fatalf("the cancelled request waited %v", time.Since(start))
	}

	if err := <-waiting; err != nil {
		t.Fatal(err)
	}

	// a cancelled request consumed no slot: the next one waits for a single interval at most
	start = time.Now()
	if err := g.throttle(context.Background()); err != nil {
		t.Fatal(err)
	}
	if d := time.Since(start); d > 1500*time.Millisecond {
		t.Fatalf("the next request waited %v", d)
	}
}

// A request whose context is already done never takes a slot.
func TestThrottleRefusesDoneContext(t *testing.T) {
	g := NewRequestGroup().RateLimit(1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	for range 100 {
		if err := g.throttle(ctx); !errors.Is(err, context.Canceled) {
			t.Fatalf("expected a cancelled request to be refused, got %v", err)
		}
	}

	if !g.lastReqAt.IsZero() {
		t.Fatal("a refused request must not count as sent")
	}
}

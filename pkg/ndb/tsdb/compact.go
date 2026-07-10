// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"log/slog"
	"sync"
)

// compactor runs flush/compaction of columns off the write path. Columns are
// enqueued when their head crosses a threshold; a single background goroutine
// drains the queue so writers never block on the heavy chunk rewrite. Each
// column is processed at most once per pending signal (deduplicated).
type compactor struct {
	db *DB

	mu      sync.Mutex
	pending map[*Column]struct{}
	wake    chan struct{}
	done    chan struct{}
	stopped bool
	wg      sync.WaitGroup
}

func newCompactor(db *DB) *compactor {
	c := &compactor{
		db:      db,
		pending: make(map[*Column]struct{}),
		wake:    make(chan struct{}, 1),
		done:    make(chan struct{}),
	}
	c.wg.Add(1)
	go c.loop()
	return c
}

func (c *compactor) enqueue(col *Column) {
	c.mu.Lock()
	if c.stopped {
		c.mu.Unlock()
		return
	}
	c.pending[col] = struct{}{}
	c.mu.Unlock()
	select {
	case c.wake <- struct{}{}:
	default:
	}
}

func (c *compactor) loop() {
	defer c.wg.Done()
	for {
		select {
		case <-c.done:
			return
		case <-c.wake:
			c.drain()
		}
	}
}

func (c *compactor) drain() {
	for {
		c.mu.Lock()
		var col *Column
		for k := range c.pending {
			col = k
			delete(c.pending, k)
			break
		}
		c.mu.Unlock()
		if col == nil {
			return
		}
		col.mu.Lock()
		// re-check under lock: the head may have been reset by a concurrent
		// synchronous Flush.
		if col.needsFlush() {
			if err := col.flushLocked(); err != nil {
				slog.Error("tsdb: compaction failed", "bucket", col.bucket, "column", col.name, "err", err)
			}
		}
		col.mu.Unlock()
	}
}

// stop drains once more and terminates the background goroutine.
func (c *compactor) stop() {
	c.mu.Lock()
	if c.stopped {
		c.mu.Unlock()
		return
	}
	c.stopped = true
	c.mu.Unlock()
	close(c.done)
	c.wg.Wait()
}

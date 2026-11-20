// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.wdy.de/nago/pkg/blob"
)

// buildCheckpoint is the number of sessions, after which the build writes its cursor.
const buildCheckpoint = 500

// build derives the head of every session, see [Heads]. While the heads are invalid, it removes every head first and
// the invalid mark at its end. Otherwise, it continues behind the cursor of an earlier build, or does nothing after a
// finished one. It reads the sessions in the order of their ids, each one under its lock, so that a write of the same
// session waits. A session, which cannot be decoded, gets no head and is counted. Any other error ends the build
// without removing a head, so that it continues later behind its last cursor. At its end, it removes the heads of
// sessions, which no longer exist.
func (h *heads) build(ctx context.Context) error {
	invalid, err := h.exists(blob.InvalidKey)
	if err != nil {
		return err
	}

	var l layout
	if invalid {
		if err := h.clear(ctx); err != nil {
			return err
		}

		l = layout{Version: layoutVersion, State: stateBuilding}
		if err := h.put(keyLayout, l); err != nil {
			return err
		}
	} else {
		ok, err := h.get(keyLayout, &l)
		if err != nil {
			return err
		}

		if ok && l.State == stateDone {
			return nil
		}

		if !ok {
			l = layout{Version: layoutVersion, State: stateBuilding}
			if err := h.put(keyLayout, l); err != nil {
				return err
			}
		}
	}

	start := time.Now()
	from := l.Cursor
	derived := 0
	for id, err := range h.repo.Identifiers() {
		if err != nil {
			return fmt.Errorf("cannot list the sessions: %w", err)
		}

		if l.Cursor != "" && id <= l.Cursor {
			continue
		}

		if ctx.Err() != nil {
			return context.Cause(ctx)
		}

		skipped, err := h.derive(id)
		if err != nil {
			return fmt.Errorf("cannot read session %s: %w", id, err)
		}

		if skipped {
			l.Skipped++
		}

		l.Cursor = id
		derived++
		if derived%buildCheckpoint == 0 {
			if err := h.put(keyLayout, l); err != nil {
				return err
			}
		}
	}

	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	// so that a later build need not read them again
	if err := h.put(keyLayout, l); err != nil {
		return err
	}

	if err := h.dropOrphans(ctx); err != nil {
		return err
	}

	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	done := layout{Version: layoutVersion, State: stateDone, Skipped: l.Skipped}
	if err := h.put(keyLayout, done); err != nil {
		return err
	}

	if invalid {
		if ctx.Err() != nil {
			return context.Cause(ctx)
		}

		if err := h.del(blob.InvalidKey); err != nil {
			return err
		}
	}

	slog.Info("built the heads of the sessions", "store", h.store.Name(), "from", from, "derived", derived, "skipped", l.Skipped, "rebuilt", invalid, "duration", time.Since(start))
	return nil
}

// derive writes the head of the session under its lock. It reports a session, which cannot be decoded.
func (h *heads) derive(id ID) (skipped bool, err error) {
	release := h.lock(id)
	defer release()

	opt, err := h.repo.FindByID(id)
	if err != nil {
		if undecodable(err) {
			slog.Error("cannot decode a session, it gets no head", "session", id, "err", err)
			return true, nil
		}

		return false, err
	}

	if opt.IsNone() {
		// deleted meanwhile together with its heads
		return false, nil
	}

	return false, h.settle(opt.Unwrap())
}

// clear removes every head. The marks stay, they belong to writes.
func (h *heads) clear(ctx context.Context) error {
	for _, prefix := range []string{keyHeads, keyChildren} {
		keys, err := h.keys(prefix)
		if err != nil {
			return err
		}

		for _, key := range keys {
			if ctx.Err() != nil {
				return context.Cause(ctx)
			}

			if err := h.del(key); err != nil {
				return err
			}
		}
	}

	return nil
}

// dropOrphans removes every head, whose session does not exist, checked without reading the session.
func (h *heads) dropOrphans(ctx context.Context) error {
	for _, prefix := range []string{keyHeads, keyChildren} {
		keys, err := h.keys(prefix)
		if err != nil {
			return err
		}

		for _, key := range keys {
			if ctx.Err() != nil {
				return context.Cause(ctx)
			}

			if err := h.dropOrphan(key); err != nil {
				return err
			}
		}
	}

	return nil
}

func (h *heads) dropOrphan(key string) error {
	id := idOfKey(key)
	release := h.lock(id)
	defer release()

	exists, err := h.contentExists(id)
	if err != nil || exists {
		return err
	}

	return h.del(key)
}

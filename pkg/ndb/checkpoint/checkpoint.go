// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package checkpoint adds resumable, at-least-once cursor handling on top of
// [ndb.Tail] for side-effecting log consumers.
//
// It deliberately lives OUTSIDE evs.Projection: a projection is a resident,
// stateless read model that folds exclusively from the log and is freely
// rebuildable. Persisting a cursor would break that contract (a restart would
// resume from the delta with an empty in-memory map). Consumers with PERSISTENT
// side effects — mirrors, outboxes, external integrations — instead track how
// far they have durably processed the log and resume from there. That cursor is
// exactly what this package persists.
//
// # Delivery semantics
//
// At-least-once: the consumer commits the cursor AFTER its side effect (the
// fold) has been applied. A crash between the side effect and the commit
// re-processes a few events on restart. That is correct as long as the folds are
// idempotent and processed in Seq order ([ndb.Tail] guarantees the latter). It
// needs no transactions, outbox, or 2PC. Committing BEFORE the effect
// (at-most-once) would risk data loss and is intentionally not offered.
package checkpoint

import (
	"context"
	"iter"
	"sync"
	"sync/atomic"
	"time"

	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/ndb"
)

// Store persists the highest processed global [ndb.Seq] of a consumer so that a
// restart resumes from the delta instead of replaying the whole log.
//
// A Seq of 0 means "nothing committed yet" / "from the beginning".
type Store interface {
	// Load returns the highest committed Seq, or 0 if nothing has been committed
	// yet (start from the beginning).
	Load() (ndb.Seq, error)
	// Save durably records seq as the new high-water mark. It is called after the
	// consumer's fold, subject to the batching in [Options].
	Save(seq ndb.Seq) error
}

// Options configures the batching of checkpoint writes. The zero value is valid
// and applies a conservative default (see [DefaultSaveEvery] and
// [DefaultSaveInterval]) so a warm-up over a long log does not issue one blob
// write per event.
type Options struct {
	// SaveEvery flushes the cursor to the [Store] after this many committed
	// events. 0 selects [DefaultSaveEvery]. A value of 1 saves on every commit.
	SaveEvery uint64

	// SaveInterval flushes the cursor to the [Store] at most this often,
	// regardless of SaveEvery. 0 selects [DefaultSaveInterval]. A negative value
	// disables time-based flushing.
	SaveInterval time.Duration

	// Ctx, when non-nil, is forwarded to [ndb.TailOptions.Ctx] so the underlying
	// stream returns promptly on cancellation even while blocked on a silent live
	// edge. [Run] sets this from its ctx argument; a manual [Tail] caller may set
	// it to make the returned stream cancellable. nil = classic break-to-cancel.
	Ctx context.Context
}

const (
	// DefaultSaveEvery is the zero-value [Options.SaveEvery]: flush every 256
	// committed events.
	DefaultSaveEvery = 256
	// DefaultSaveInterval is the zero-value [Options.SaveInterval]: flush at most
	// once per second.
	DefaultSaveInterval = time.Second
)

// Committer records processed sequence numbers and flushes them to the [Store]
// according to the configured batching. It is NOT safe for concurrent use: call
// Commit and Flush from the single goroutine that ranges the stream.
type Committer struct {
	store Store

	saveEvery    uint64
	saveInterval time.Duration

	pending    ndb.Seq // highest Commit'd seq not yet persisted (0 = none pending)
	committed  ndb.Seq // highest seq durably Save'd
	sinceFlush uint64  // commits since the last flush
	lastFlush  time.Time
}

// Tail resumes from the [Store]'s committed cursor and follows the log via
// [ndb.Tail]. It returns the message stream (identical in shape to [ndb.Tail])
// plus a [Committer] the consumer must drive.
//
// This is the low-level API for consumers that need to control WHEN the cursor
// advances (e.g. commit once per batch of folded events). For the common case,
// prefer [Run], which owns the loop and commits after each handler call.
//
// Usage:
//
//	stream, cp, err := checkpoint.Tail(src, types, store, checkpoint.Options{})
//	if err != nil { ... }
//	defer cp.Flush()
//	for typeID, msg := range stream {
//	    fold(typeID, msg)      // idempotent, persistent side effect
//	    if err := cp.Commit(msg.Seq); err != nil { ... }
//	}
//
// Commit is called AFTER the fold to keep at-least-once semantics. Flush on exit
// persists the last batched Seq so the tail of a batch window is not lost (it
// would otherwise merely be re-processed on restart, which is still correct).
func Tail(src ndb.Followable, types []ndb.TypeID, store Store, opts Options) (iter.Seq2[ndb.TypeID, ndb.Message], *Committer, error) {
	from, err := store.Load()
	if err != nil {
		return nil, nil, err
	}

	saveEvery := opts.SaveEvery
	if saveEvery == 0 {
		saveEvery = DefaultSaveEvery
	}
	saveInterval := opts.SaveInterval
	if saveInterval == 0 {
		saveInterval = DefaultSaveInterval
	}

	cp := &Committer{
		store:        store,
		saveEvery:    saveEvery,
		saveInterval: saveInterval,
		committed:    from,
		lastFlush:    time.Now(),
	}

	// FromSeq is inclusive, so +1 skips the already-committed high-water mark.
	// from == 0 (nothing committed) maps to FromSeq 0 = from the beginning.
	fromSeq := from
	if fromSeq > 0 {
		fromSeq++
	}

	stream := ndb.Tail(src, types, ndb.TailOptions{FromSeq: fromSeq, Ctx: opts.Ctx})
	return stream, cp, nil
}

// Commit records seq as processed and flushes to the [Store] if the batching
// threshold (count or interval) is reached. It is monotonic: a seq at or below
// the highest already-committed value is ignored. Returns any error from the
// underlying [Store.Save].
func (c *Committer) Commit(seq ndb.Seq) error {
	if seq <= c.committed || seq <= c.pending {
		return nil
	}
	c.pending = seq
	c.sinceFlush++

	if c.sinceFlush >= c.saveEvery {
		return c.Flush()
	}
	if c.saveInterval > 0 && time.Since(c.lastFlush) >= c.saveInterval {
		return c.Flush()
	}
	return nil
}

// Flush persists the highest pending Seq to the [Store] immediately. It is a
// no-op if nothing new is pending. Call it before exiting the range loop so the
// tail of the current batch window is not left uncommitted.
func (c *Committer) Flush() error {
	if c.pending <= c.committed {
		c.sinceFlush = 0
		c.lastFlush = time.Now()
		return nil
	}
	if err := c.store.Save(c.pending); err != nil {
		return err
	}
	c.committed = c.pending
	c.sinceFlush = 0
	c.lastFlush = time.Now()
	return nil
}

// Committed returns the highest Seq that has been durably saved to the [Store].
func (c *Committer) Committed() ndb.Seq {
	return c.committed
}

// Handler processes a single message. It runs synchronously in the calling
// goroutine of [Consumer.Run], in strict ascending global Seq order.
//
// It MUST be idempotent: under at-least-once delivery a crash between the side
// effect and the cursor commit re-delivers a few events on restart. Return a
// non-nil error to stop the run BEFORE the cursor (and the watermark) advance
// past this Seq, so the message is re-delivered on the next run (treat the error
// as "retry later").
type Handler func(typeID ndb.TypeID, msg ndb.Message) error

// Consumer is the high-level, batteries-included driver over [Tail]: it resumes
// from a blob-backed cursor, follows the log, invokes a [Handler] per message
// and commits the cursor after each success (at-least-once).
//
// Beyond the fire-and-forget loop it offers a read-your-write watermark
// ([Consumer.WaitFor]), mirroring evs.Projection: after the write side appends
// an event and learns its global Seq, a caller can block until this consumer has
// applied that Seq's side effect before reading the sink. This replaces the
// WaitFor a projection-based mirror used to provide, without making the consumer
// stateful — the sink (e.g. the mirrored store) remains the sole state.
//
// # Watermark vs. cursor
//
// The watermark advances AFTER the handler applies a Seq's side effect, which is
// the exact read-your-write boundary. The persisted cursor advances separately
// and is batched (see [Options]); read-your-write needs "effect applied", not
// "cursor durably saved", so the two are deliberately decoupled.
//
// # Out of scope: event self-sufficiency
//
// Because a delta-resume never replays already-committed history, each event
// must carry everything its handler needs (e.g. a revoke must carry the full
// tuple, not a bare id that a prior grant would have supplied). That is a
// property of the event schema and the handler, not of this package.
type Consumer struct {
	src   ndb.Followable
	types []ndb.TypeID
	store Store
	opts  Options
	fn    Handler

	// processed is the highest global Seq whose handler has completed
	// successfully. It is the read-your-write watermark and advances contiguously
	// with Tail's delivery.
	processed atomic.Uint64
	waitMu    sync.Mutex
	waitCh    chan struct{}
}

// NewConsumer builds a [Consumer] that reads from src, resumes from the cursor
// stored under key in store (as an 8-byte big-endian Seq via [NewBlobStore]) and
// invokes fn for each selected message. Call [Consumer.Run] to start it.
func NewConsumer(src ndb.Followable, types []ndb.TypeID, store blob.Store, key string, opts Options, fn Handler) *Consumer {
	return &Consumer{
		src:    src,
		types:  types,
		store:  NewBlobStore(store, key),
		opts:   opts,
		fn:     fn,
		waitCh: make(chan struct{}),
	}
}

// Run follows the log and invokes the [Handler] for every selected message,
// committing the cursor after each success (at-least-once) and advancing the
// read-your-write watermark. It blocks until one of:
//
//   - ctx is cancelled: Run flushes the last batched cursor and returns ctx.Err().
//   - the handler returns an error: Run flushes what was committed so far and
//     returns that error WITHOUT committing/advancing the failing Seq, so it is
//     re-delivered on the next run.
//   - the stream ends: Run flushes and returns nil.
//
// Run must be called at most once per Consumer.
func (c *Consumer) Run(ctx context.Context) error {
	opts := c.opts
	opts.Ctx = ctx
	stream, cp, err := Tail(c.src, c.types, c.store, opts)
	if err != nil {
		return err
	}

	var handlerErr error
	for typeID, msg := range stream {
		if err := c.fn(typeID, msg); err != nil {
			handlerErr = err
			break // do NOT commit/advance this Seq: re-delivered on the next run
		}
		// Advance the watermark AFTER the side effect is applied: that is the
		// read-your-write boundary a WaitFor(seq) relies on. The cursor commit
		// below is mere (batched) persistence and independent of it.
		c.advance(msg.Seq)
		if err := cp.Commit(msg.Seq); err != nil {
			handlerErr = err
			break
		}
	}

	// Persist the tail of the current batch window regardless of why we stopped.
	if err := cp.Flush(); err != nil && handlerErr == nil {
		handlerErr = err
	}

	if handlerErr != nil {
		return handlerErr
	}
	return ctx.Err()
}

// Processed returns the highest global Seq whose handler has completed. It is 0
// before anything has been processed. Lock-free read.
func (c *Consumer) Processed() ndb.Seq {
	return ndb.Seq(c.processed.Load())
}

// WaitFor blocks until the consumer has applied the side effect of every event
// up to and including seq, or until ctx is done. It returns nil once
// Processed >= seq, or ctx.Err() if cancelled first. A seq of 0, or one already
// reached, returns immediately.
//
// This is the opt-in read-your-write primitive: pass the global Seq your write
// returned ([ndb.Envelope]/Append), then read the sink. It does not poll — it
// waits on a broadcast that fires each time the watermark advances.
func (c *Consumer) WaitFor(ctx context.Context, seq ndb.Seq) error {
	target := uint64(seq)
	for {
		if c.processed.Load() >= target {
			return nil
		}

		c.waitMu.Lock()
		ch := c.waitCh
		c.waitMu.Unlock()

		if c.processed.Load() >= target {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ch:
		}
	}
}

// advance moves the watermark forward and wakes any waiters. It is monotonic:
// Tail delivers in ascending Seq order, so this only ever moves forward.
func (c *Consumer) advance(seq ndb.Seq) {
	for {
		cur := c.processed.Load()
		if uint64(seq) <= cur {
			return
		}
		if c.processed.CompareAndSwap(cur, uint64(seq)) {
			break
		}
	}

	c.waitMu.Lock()
	old := c.waitCh
	c.waitCh = make(chan struct{})
	c.waitMu.Unlock()
	close(old)
}

// Run is a convenience wrapper over [NewConsumer] + [Consumer.Run] for the
// fire-and-forget case that does not need read-your-write. Callers that need
// [Consumer.WaitFor] must build a [Consumer] and hold on to it.
//
// Typical use — a persistent-side-effect consumer (mirror, outbox, integration):
//
//	err := checkpoint.Run(ctx, src, []ndb.TypeID{"OrderPlaced"},
//	    blobStore, "mirror.cursor", checkpoint.Options{},
//	    func(_ ndb.TypeID, msg ndb.Message) error {
//	        return mirror.Apply(msg) // idempotent
//	    })
func Run(
	ctx context.Context,
	src ndb.Followable,
	types []ndb.TypeID,
	store blob.Store,
	key string,
	opts Options,
	fn Handler,
) error {
	return NewConsumer(src, types, store, key, opts, fn).Run(ctx)
}

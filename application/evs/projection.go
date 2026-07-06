// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package evs

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"log/slog"
	"reflect"
	"sync"
	"sync/atomic"

	"go.wdy.de/nago/pkg/ndb"
)

// Source is the narrowest ndb capability a [Projection] needs: the append-only
// history to warm up from ([ndb.History]) plus the live edge to follow
// ([ndb.Notifier]), composed as [ndb.Followable]. A full [ndb.Messages]
// satisfies it automatically, but the projection deliberately depends only on
// this narrow contract so alternative engines with a reduced feature set remain
// usable and tests can supply a small fake.
//
// The two capabilities are composed by [ndb.Tail] into a single "replay then
// follow" stream, which is exactly the read path a projection consumes.
type Source = ndb.Followable

// Projection is an event-folded, resident read view: a map[K]S built by folding
// a selected set of event types, aggregate-crossing by design.
//
// It is the read-side counterpart to [Handler] and exists because the write side
// couples folding to a single aggregate root (an event's Evolve targets exactly
// one aggregate, see [Evt]); a read model must instead fold the same event into
// an arbitrary target, possibly under a different key. A projection therefore
// registers its folding at the target (via [Project]) rather than on the event.
//
// # Mechanics
//
// [Projection.Run] starts one goroutine that ranges [ndb.Tail]: it first replays
// the history (warm-up) and then follows live writes, both through the same code
// path and in strict ascending global [ndb.Seq] order. Each message is decoded
// in the same iteration step (the payload is only valid then, see [ndb.Message]),
// routed to the fold registered for its discriminator, and applied to map[K]*S.
// Reads are always current and resident; there is no on-the-fly replay per call.
//
// # Ordering
//
// The fold observes events in strict ascending global Seq order. That order is
// guaranteed by [ndb.Tail]. Per-key ordering (a given K sees its events in Seq
// order too) is a property of this implementation — a single goroutine folds
// sequentially — not of the engine; keep that in mind before ever sharding the
// fold across goroutines.
//
// # Rebuildability
//
// A projection folds exclusively from the log and keeps no persistence of its
// own, so it can be rebuilt from the source at any time simply by constructing
// it again and calling Run.
//
// # Consistency (read-your-write)
//
// Folding is asynchronous: after an event is appended there is a small window
// (typically sub-millisecond) before the tail goroutine has folded it, so a read
// is not read-your-write by default. For interactive UIs this is irrelevant (a
// re-render follows the event). When a caller must observe its own write it uses
// the global Seq the write returned — [Backend.Append] yields it as
// [Envelope.Sequence] — together with [Projection.WaitFor]:
//
//	env, _ := backend.Append(subject, evt)
//	_ = view.WaitFor(ctx, ndb.Seq(env.Sequence))
//	v, _ := view.Get(key) // now guaranteed to include env
//
// WaitFor blocks without polling and is opt-in; the fast-path reads Get/All stay
// lock-free and pay nothing for it.
//
// # Concurrency
//
// The tail goroutine is the sole writer of the state map; Get/All read it under
// an RWMutex and return a shallow copy of S. Treat the returned S as read-only:
// if S contains mutable reference fields (maps, slices, pointers), a reader could
// otherwise race the next fold — model S as a flat value or copy those fields in
// the fold.
type Projection[K ~string, S any] struct {
	src  Source
	opts ProjectionOptions

	// rules maps a registered discriminator to its decode+key+fold closure.
	// It is built entirely before Run by Project and read-only afterwards.
	rules map[ndb.TypeID]func(msg ndb.Message)
	types []ndb.TypeID

	mu    sync.RWMutex
	state map[K]*S

	// processedSeq is the highest global Seq the fold has accounted for
	// (folded, skipped as unwanted, or skipped on error). It advances
	// contiguously with Tail's watermark so WaitFor is exact.
	processedSeq atomic.Uint64
	// waitMu guards waitCh, the coalescing broadcast channel that is closed
	// (and replaced) every time processedSeq advances.
	waitMu sync.Mutex
	waitCh chan struct{}

	runOnce sync.Once
	stopFn  func()
}

// ProjectionOptions configures a [Projection]. The zero value is valid: warm-up
// starts at the very beginning and decode/fold errors are logged and skipped.
type ProjectionOptions struct {
	// FromSeq is the inclusive global Seq at which warm-up starts (0 = from the
	// very beginning). It exists for advanced cases (sub-projections, tests);
	// the default of 0 is what a full read model wants.
	FromSeq ndb.Seq

	// OnError is invoked when a message cannot be decoded, or when a fold panics
	// (the panic is recovered so it never kills the tail goroutine). In both
	// cases the offending event is skipped and folding continues. If nil, the
	// error is logged via slog at error level. OnError must not block or write
	// back into the source.
	OnError func(seq ndb.Seq, typeID ndb.TypeID, err error)
}

// Unit is the single fixed key type of a [Singleton]. It is exported (with one
// valid value, [TheUnit]) so a singleton fold can name its key type when
// registering with [Project]:
//
//	evs.Project(view, func(Posted) evs.Unit { return evs.TheUnit() }, fold)
//
// A Singleton is therefore literally a Projection with one entry and needs no
// second mechanism.
type Unit string

const theUnit Unit = ""

// Singleton is a [Projection] that folds all matching events into a single
// value, addressed by one fixed internal key. Use it for a global figure (a
// counter, a latest-snapshot, a rollup). Register folds with [Project] exactly
// as for a keyed projection, using a key function that returns [TheUnit]. Read
// it with [Value].
type Singleton[S any] = Projection[Unit, S]

// NewProjection creates a keyed projection reading from src. Register folds with
// [Project], then start it with [Projection.Run].
func NewProjection[K ~string, S any](src Source, opts ProjectionOptions) *Projection[K, S] {
	return &Projection[K, S]{
		src:    src,
		opts:   opts,
		rules:  make(map[ndb.TypeID]func(msg ndb.Message)),
		state:  make(map[K]*S),
		waitCh: make(chan struct{}),
	}
}

// NewSingleton creates a [Singleton] reading from src.
func NewSingleton[S any](src Source, opts ProjectionOptions) *Singleton[S] {
	return NewProjection[Unit, S](src, opts)
}

// TheUnit is the sole valid [Unit] value: the fixed key of a [Singleton]. A key
// function passed to [Project] for a singleton returns it.
//
//	evs.Project(view, func(Posted) evs.Unit { return evs.TheUnit() }, fold)
func TheUnit() Unit { return theUnit }

// Project registers decode + key + fold for event type E on projection p. It
// must be called before [Projection.Run] and is not safe for concurrent use.
//
// The decoder is derived from E via reflection and JSON; the ndb type id to read
// is E's discriminator (E{}.Discriminator()). key selects which K this event
// contributes to for THIS projection — the same event type may map to different
// keys in different projections (e.g. a Posted event carries both a thread id and
// a comment id). fold applies the decoded event to the target state, which is
// created on first touch (the zero value of S).
//
// fold is intentionally errorless to keep call sites terse; report and skip
// exceptional events via [ProjectionOptions.OnError]. A panic inside fold is
// recovered and routed to OnError as well, so a single bad event never stalls
// the projection.
func Project[K ~string, S any, E interface{ Discriminator() Discriminator }](
	p *Projection[K, S],
	key func(E) K,
	fold func(s *S, e E),
) {
	var proto E
	disc := proto.Discriminator()
	if err := disc.Validate(); err != nil {
		panic(fmt.Errorf("evs: event %T has invalid discriminator: %w", proto, err))
	}
	typeID := ndb.TypeID(disc)
	if !ndb.ValidTypeID(string(typeID)) {
		panic(fmt.Errorf("evs: discriminator %q is not a valid ndb type id", disc))
	}
	if _, dup := p.rules[typeID]; dup {
		panic(fmt.Errorf("evs: discriminator %q already registered on this projection", disc))
	}

	// rtype is the concrete struct type behind E (deref one pointer level so we
	// json.Unmarshal into an addressable value regardless of whether E is a
	// pointer or value type).
	rtype := reflect.TypeFor[E]()
	for rtype.Kind() == reflect.Ptr {
		rtype = rtype.Elem()
	}

	p.rules[typeID] = func(msg ndb.Message) {
		if msg.Encoding != ndb.EncodingRaw {
			p.reportError(msg.Seq, typeID, fmt.Errorf("unexpected payload encoding %d", msg.Encoding))
			return
		}

		rv := reflect.New(rtype)
		if err := json.Unmarshal(msg.Payload, rv.Interface()); err != nil {
			p.reportError(msg.Seq, typeID, fmt.Errorf("decode: %w", err))
			return
		}

		evt, ok := rv.Elem().Interface().(E)
		if !ok {
			// E is a pointer type: hand over the pointer instead of the value.
			evt, ok = rv.Interface().(E)
			if !ok {
				p.reportError(msg.Seq, typeID, fmt.Errorf("decoded value is not assignable to %T", proto))
				return
			}
		}

		k := key(evt)

		p.mu.Lock()
		s, exists := p.state[k]
		if !exists {
			s = new(S)
			p.state[k] = s
		}
		applyFold(p, msg.Seq, typeID, s, evt, fold)
		p.mu.Unlock()
	}
	p.types = append(p.types, typeID)
}

// applyFold runs fold under recover so a panic becomes a skipped event plus an
// OnError callback instead of a dead tail goroutine. It runs while p.mu is held.
func applyFold[K ~string, S any, E any](p *Projection[K, S], seq ndb.Seq, typeID ndb.TypeID, s *S, e E, fold func(*S, E)) {
	defer func() {
		if rec := recover(); rec != nil {
			p.reportError(seq, typeID, fmt.Errorf("fold panicked: %v", rec))
		}
	}()
	fold(s, e)
}

func (p *Projection[K, S]) reportError(seq ndb.Seq, typeID ndb.TypeID, err error) {
	if p.opts.OnError != nil {
		p.opts.OnError(seq, typeID, err)
		return
	}
	slog.Error("evs: projection skipped an event", "seq", seq, "type", typeID, "err", err)
}

// Run starts the tail goroutine that warms up from the log and then follows live
// writes, folding matching events into the state map. It returns a stop function
// that ends the goroutine (idempotent).
//
// Run itself is idempotent: the goroutine is started at most once, and every
// call returns the same stop function. Reads (Get/All/ProcessedSeq/WaitFor)
// issued before Run see an empty projection (ProcessedSeq == 0) rather than an
// error, and unblock/populate once Run has been called and folding proceeds.
func (p *Projection[K, S]) Run() (stop func()) {
	p.runOnce.Do(func() {
		done := make(chan struct{})
		var closeOnce sync.Once
		p.stopFn = func() { closeOnce.Do(func() { close(done) }) }

		go p.loop(done)
	})
	return p.stopFn
}

func (p *Projection[K, S]) loop(done <-chan struct{}) {
	for typeID, msg := range ndb.Tail(p.src, p.types, ndb.TailOptions{FromSeq: p.opts.FromSeq}) {
		select {
		case <-done:
			return
		default:
		}

		if rule, ok := p.rules[typeID]; ok {
			rule(msg)
		}

		// Advance the watermark AFTER folding this Seq so a WaitFor(seq) that
		// returns is guaranteed to see the folded state, not merely the arrival.
		p.advance(msg.Seq)

		select {
		case <-done:
			return
		default:
		}
	}
}

// advance moves the processed watermark forward and wakes any waiters. It is
// monotonic: out-of-order or duplicate seqs (never produced by Tail, but
// defensive) never move it backwards.
func (p *Projection[K, S]) advance(seq ndb.Seq) {
	for {
		cur := p.processedSeq.Load()
		if uint64(seq) <= cur {
			return
		}
		if p.processedSeq.CompareAndSwap(cur, uint64(seq)) {
			break
		}
	}

	// Wake everyone waiting on the current generation and open a fresh one.
	p.waitMu.Lock()
	old := p.waitCh
	p.waitCh = make(chan struct{})
	p.waitMu.Unlock()
	close(old)
}

// ProcessedSeq returns the highest global Seq the projection has folded (or
// deliberately skipped). It is 0 before anything has been processed. It is a
// lock-free read.
func (p *Projection[K, S]) ProcessedSeq() ndb.Seq {
	return ndb.Seq(p.processedSeq.Load())
}

// WaitFor blocks until the projection has processed all events up to and
// including seq, or until ctx is done. It returns nil once ProcessedSeq >= seq,
// or ctx.Err() if the context is cancelled first. A seq of 0, or a seq already
// reached, returns immediately.
//
// This is the opt-in read-your-write primitive: pass the global Seq your write
// returned, then read. It does not poll — it waits on a broadcast that fires
// each time the watermark advances.
func (p *Projection[K, S]) WaitFor(ctx context.Context, seq ndb.Seq) error {
	target := uint64(seq)
	for {
		if p.processedSeq.Load() >= target {
			return nil
		}

		// Snapshot the current generation BEFORE re-checking the watermark, so we
		// cannot miss an advance that happens between the check and the wait.
		p.waitMu.Lock()
		ch := p.waitCh
		p.waitMu.Unlock()

		if p.processedSeq.Load() >= target {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ch:
			// watermark advanced; loop and re-check
		}
	}
}

// Get returns a shallow copy of the state for key k, and whether it exists. The
// copy is served under a read lock; treat it as read-only (see the type-level
// note on mutable fields in S).
func (p *Projection[K, S]) Get(k K) (S, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	s, ok := p.state[k]
	if !ok {
		var zero S
		return zero, false
	}
	return *s, true
}

// All iterates over a point-in-time snapshot of every (key, state) pair. The
// snapshot is taken under a read lock and then yielded lock-free, so the fold
// may advance the projection while the caller ranges; the caller observes a
// consistent set as of the call. Each yielded S is a shallow copy (read-only).
func (p *Projection[K, S]) All() iter.Seq2[K, S] {
	p.mu.RLock()
	snap := make(map[K]S, len(p.state))
	for k, s := range p.state {
		snap[k] = *s
	}
	p.mu.RUnlock()

	return func(yield func(K, S) bool) {
		for k, s := range snap {
			if !yield(k, s) {
				return
			}
		}
	}
}

// Value returns the singleton's folded value and whether any event has been
// folded into it yet. It is the [Singleton] convenience over
// [Projection.Get]. (It is a free function, not a method, because Go does not
// allow methods on a generic alias type.)
func Value[S any](p *Singleton[S]) (S, bool) {
	return p.Get(theUnit)
}

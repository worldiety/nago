// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ndb

import (
	"context"
	"iter"
	"math"
)

// TailOptions configures [Tail].
type TailOptions struct {
	// FromSeq is the inclusive sequence number at which historical catch-up
	// starts. 0 means "from the very beginning".
	FromSeq Seq

	// Ctx, when non-nil, cancels the stream: once Ctx is done, Tail stops
	// ranging and returns, even while blocked waiting on the live edge for the
	// next write. It is the way to unblock a caller that is idle on a quiet tail.
	//
	// A nil Ctx preserves the classic behaviour: cancellation is done purely by
	// breaking out of the range loop, which is sufficient while actively
	// iterating (replay/catch-up) but cannot interrupt a wait on a silent live
	// edge. Prefer Ctx for long-lived followers; the zero value stays fully
	// backwards compatible.
	Ctx context.Context
}

// Tail returns a stream that first catches up the history from opts.FromSeq via
// [History.Replay] and then seamlessly continues live, delivering messages as
// they are written, until the consumer stops ranging.
//
// It is the engine-neutral composition of [Notifier.Subscribe] and
// [History.Replay]: every engine that provides those two capabilities (a
// [Followable], which any [Messages] satisfies) inherits this "replay then
// follow" behaviour without implementing it itself. Tail bridges the replay→live
// gap so no message is lost or delivered twice:
//
//   - It subscribes BEFORE replaying, so live writes are observed from the start.
//   - It then replays up to the current tip and follows the live edge, holding a
//     single global Seq watermark.
//   - Each write wakes a re-drain via Replay, which keeps the global ascending
//     Seq order intact even across event types.
//
// Tail is also a well-behaved subscriber under the synchronous [Notifier]
// contract: its subscription callback only flips a non-blocking wake signal, so
// a slow consumer of the Tail stream never stalls the writer.
//
// Messages are yielded in strict ascending global Seq order, exactly like
// [History.Replay]. types selects the event types (empty = all).
//
// Ordering and completeness are the guarantees; a gap-free run of sequence
// numbers is not. Sequence numbers are allowed to have permanent holes, because
// a message can be tombstoned via [Pruner.DeleteSeq], superseded by a later
// [Retained.Put] on the same type, or removed wholesale by [Pruner.DeleteType],
// and the Seq it occupied is never reissued. Tail delivers every message that
// exists, in order, and steps over the numbers of those that do not. (The same
// applies to Kafka offsets under log compaction, and for the same reason.)
//
// Telling a permanent hole from a write that is merely still in progress is not
// possible from a replay alone: both simply show up as a missing Seq. Tail
// therefore asks the engine, via the optional [Watermark] capability. At or
// below [Watermark.CommittedSeq] a gap is settled and is skipped; above it a
// gap may still be filled by an in-flight write, so Tail waits for the next
// write rather than delivering out of order or dropping the message.
//
// If the engine does not implement [Watermark], Tail cannot make that
// distinction and falls back to waiting on every gap. That is still safe — no
// message is lost or reordered — but a permanent hole then pauses delivery
// until a later append advances the log past it.
//
// Cancellation is done — as with Replay — by breaking out of the range loop; no
// context.Context is required for that, consistent with the rest of the ndb API.
// That suffices while actively iterating, but cannot interrupt a wait on a
// silent live edge (no more writes are coming). For long-lived followers that
// must return promptly on shutdown, pass [TailOptions.Ctx]; a nil Ctx keeps the
// classic break-to-cancel behaviour.
//
// The yielded Message.Payload follows the same lifetime rules as [History.Replay]
// (a view valid only for the current iteration step); clone it to retain it.
func Tail(m Followable, types []TypeID, opts TailOptions) iter.Seq2[TypeID, Message] {
	return func(yield func(TypeID, Message) bool) {
		// wake is a coalescing "something was written" signal with capacity 1.
		// We never depend on individual notifications — only on the fact that the
		// log advanced — so collapsing many notifications into one pending wake is
		// correct and cannot lose events: each drain() below replays up to the
		// current tip and thus picks up everything written so far, however many
		// notifications were coalesced. The non-blocking send also means the
		// synchronous Subscribe callback never blocks the writer.
		wake := make(chan struct{}, 1)
		signal := func() {
			select {
			case wake <- struct{}{}:
			default: // a wake is already pending; the upcoming drain will cover it
			}
		}

		unsubscribe := m.Subscribe(types, func(n Notification) {
			signal()
		})
		defer unsubscribe()

		// ctxDone is the cancellation signal. A nil Ctx yields a nil channel,
		// which blocks forever in a select — so the zero value behaves exactly
		// like the classic break-to-cancel Tail.
		var ctxDone <-chan struct{}
		if opts.Ctx != nil {
			ctxDone = opts.Ctx.Done()
		}

		// wanted reports whether a type is selected for delivery. An empty filter
		// means "all types".
		wantedSet := make(map[TypeID]bool, len(types))
		for _, t := range types {
			wantedSet[t] = true
		}
		wanted := func(t TypeID) bool {
			return len(wantedSet) == 0 || wantedSet[t]
		}

		// watermarkOf reports the engine's committed Seq and whether the engine
		// can supply one at all. Engines without the capability get a false,
		// which keeps the conservative wait-on-every-gap behaviour.
		wm, hasWatermark := m.(Watermark)
		watermarkOf := func() (Seq, bool) {
			if !hasWatermark {
				return 0, false
			}
			return wm.CommittedSeq(), true
		}

		// lastSeq is the global watermark: the highest Seq up to which everything
		// has been delivered (for the selected types), skipped as not selected, or
		// skipped as permanently absent.
		//
		// Strict global ordering plus completeness across late-appearing or
		// filtered-out types cannot both be decided from a filtered replay alone:
		// a gap might be a non-selected type (skip it) or a not-yet-visible event
		// of a selected type (must wait for it). To tell them apart, drain replays
		// ALL types and advances lastSeq across a Seq run from lastSeq+1, yielding
		// just the selected ones. The non-selected messages serve purely as proof
		// that the sequence advanced.
		//
		// A Seq that no type produces is either still being written or gone for
		// good; only the engine's committed watermark separates those two, see
		// [Watermark].
		var lastSeq Seq
		if opts.FromSeq > 0 {
			lastSeq = opts.FromSeq - 1
		}

		// drain yields all newly available SELECTED messages from lastSeq+1 up,
		// in strict ascending global Seq order, stepping over settled holes and
		// stopping at unsettled ones. Returns false if the consumer asked to stop.
		drain := func() bool {
			// Sample the watermark BEFORE starting the replay, and use that one
			// sample for the whole pass.
			//
			// The ordering is load-bearing. A replay observes the store as of the
			// moment it starts; the watermark keeps advancing as writers finish.
			// Reading it afterwards would let it cover a message that became
			// durable after this replay was already under way — a Seq reported as
			// settled that this pass genuinely cannot see. Tail would then class a
			// live write as a permanent hole and drop it. Sampling first makes the
			// watermark conservative by construction: anything at or below it was
			// durable before the replay began, so the replay must contain it, and
			// its absence really does mean it is gone.
			committed, hasCommitted := watermarkOf()

			expected := lastSeq + 1
			for typeID, msg := range m.Replay(nil, lastSeq+1, Seq(math.MaxUint64)) {
				if msg.Seq < expected {
					continue // already covered (defensive)
				}
				if msg.Seq > expected {
					// Hole at [expected, msg.Seq-1]: those sequence numbers are
					// allocated but not visible here.
					if !hasCommitted || committed < expected {
						// Either the engine cannot tell us, or the hole starts
						// above the committed boundary and may still be filled by
						// an in-flight write. Stop and let the next wake resume
						// from expected once it appears.
						break
					}

					// Everything up to committed is settled, so the hole from
					// expected onwards is permanent as far as committed reaches.
					// Skip only that far: the remainder, if any, is still
					// undecided and must be re-examined on the next pass.
					skipTo := msg.Seq - 1
					if committed < skipTo {
						skipTo = committed
					}
					lastSeq = skipTo
					expected = skipTo + 1

					if msg.Seq > expected {
						// the hole extends past the committed boundary; the rest
						// of it is not settled yet, so stop here
						break
					}
				}

				// in-order at expected
				if wanted(typeID) {
					if !yield(typeID, msg) {
						return false
					}
				}
				lastSeq = msg.Seq
				expected++
			}
			return true
		}

		// Phase 1+2: catch up history up to the current tip. Drain in a loop so
		// that anything written concurrently while we were replaying (and thus
		// flagged on wake) is picked up before we block.
		for {
			if !drain() {
				return
			}
			select {
			case <-ctxDone:
				return
			case <-wake:
				continue // more was written meanwhile; drain again
			default:
			}
			break
		}

		// Phase 3: follow the live edge. Every wake replays the newly available
		// range, which preserves global ordering. A done Ctx returns promptly
		// even though no further write would arrive to unblock the wait.
		for {
			select {
			case <-ctxDone:
				return
			case _, ok := <-wake:
				if !ok {
					return
				}
				if !drain() {
					return
				}
			}
		}
	}
}

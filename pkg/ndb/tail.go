// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ndb

import (
	"iter"
	"math"
)

// TailOptions configures [Tail].
type TailOptions struct {
	// FromSeq is the inclusive sequence number at which historical catch-up
	// starts. 0 means "from the very beginning".
	FromSeq Seq
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
// [History.Replay]. types selects the event types (empty = all). To uphold that
// order together with completeness, Tail only delivers a Seq once every smaller
// Seq is accounted for, so it briefly waits (until the next write) rather than
// deliver out of order when an event is allocated but not yet durable.
//
// Limitation: because Tail waits for missing sequence numbers, a tombstoned
// range (deleted via [Pruner]) that falls on the live edge can pause delivery
// until the next append advances the log past it. Tombstones in already-replayed
// history are handled normally.
//
// Cancellation is done — as with Replay — by breaking out of the range loop; no
// context.Context is required, consistent with the rest of the ndb API.
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

		// wanted reports whether a type is selected for delivery. An empty filter
		// means "all types".
		wantedSet := make(map[TypeID]bool, len(types))
		for _, t := range types {
			wantedSet[t] = true
		}
		wanted := func(t TypeID) bool {
			return len(wantedSet) == 0 || wantedSet[t]
		}

		// lastSeq is the global watermark: the highest Seq up to which everything
		// has been delivered (for the selected types) or skipped as not selected.
		//
		// Strict global ordering plus completeness across late-appearing or
		// filtered-out types cannot both be decided from a filtered replay alone:
		// a gap might be a non-selected type (skip it) or a not-yet-visible event
		// of a selected type (must wait for it). To tell them apart, drain replays
		// ALL types and advances lastSeq only across a CONTIGUOUS Seq run from
		// lastSeq+1, yielding just the selected ones. The non-selected messages
		// serve purely as proof that the sequence advanced without a hole. A real
		// hole (a Seq that no type produces yet) stops the pass; the next wake
		// retries once the missing event becomes durable.
		var lastSeq Seq
		if opts.FromSeq > 0 {
			lastSeq = opts.FromSeq - 1
		}

		// drain yields all newly available SELECTED messages whose Seq forms a
		// contiguous run from lastSeq+1, in strict ascending global Seq order.
		// Returns false if the consumer asked to stop.
		drain := func() bool {
			expected := lastSeq + 1
			for typeID, msg := range m.Replay(nil, lastSeq+1, Seq(math.MaxUint64)) {
				if msg.Seq < expected {
					continue // already covered (defensive)
				}
				if msg.Seq > expected {
					// Hole at [expected, msg.Seq-1]: those sequence numbers are
					// allocated but not visible here. We cannot tell a tombstone
					// from a not-yet-durable write, so stop conservatively and let
					// the next wake resume from expected once it appears.
					break
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
			case <-wake:
				continue // more was written meanwhile; drain again
			default:
			}
			break
		}

		// Phase 3: follow the live edge. Every wake replays the newly available
		// range, which preserves global ordering.
		for range wake {
			if !drain() {
				return
			}
		}
	}
}

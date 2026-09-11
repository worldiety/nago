// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package msgstore

import (
	"sync"
	"sync/atomic"
)

// inflightSet tracks the sequence numbers that have been allocated but whose
// write has not finished yet, so that the store can publish a committed
// watermark: the highest Seq below which nothing is still pending.
//
// The watermark exists because allocation and durability are two distinct
// moments. db.nextSeq.Add reserves a Seq; the message only becomes visible to
// Replay once the pwrite returns. In the window between the two, that Seq is
// simply absent from a replay — indistinguishable, from the outside, from a Seq
// that is gone for good (tombstoned by DeleteSeq, superseded by Put, or removed
// by DeleteType). A reader cannot tell those apart, and guessing either way is
// wrong: skipping loses a concurrent write, waiting stalls forever on a
// deletion. So the writer, which does know, publishes the boundary instead.
//
// This is the same device under different names across the field: LMAX
// Disruptor's availableBuffer feeding getHighestPublishedSequence, Kafka's
// last stable offset in the fetch response, PostgreSQL's xip_list/xmin in a
// snapshot, and CockroachDB's resolved timestamp. In all of them the producer
// side reports the boundary and the consumer side never infers it from holes in
// the data.
//
// Representation is a sorted slice rather than a map or a heap. The set holds
// one entry per writer that has allocated a Seq but not yet finished writing it,
// so it is bounded by the number of concurrently writing goroutines — in
// practice a handful. At that size a slice beats both a map (no hashing, no
// allocation, no iteration to find the minimum) and a heap; entries are appended
// in ascending order by construction, so the minimum is element zero.
type inflightSet struct {
	mu   sync.Mutex
	seqs []uint64 // ascending

	// maxEnded is the highest Seq that has been retired. It is the watermark
	// source whenever the set is empty, and is tracked here rather than read
	// back from nextSeq so that the watermark can never run ahead of a write
	// that has been allocated but not yet registered.
	maxEnded uint64
}

// allocate takes the next sequence number from counter and registers it as
// pending, as one atomic step. It must be paired with exactly one [inflightSet.end].
//
// Allocation and registration cannot be separated. If a writer incremented the
// counter first and registered afterwards, another writer could finish a later
// Seq entirely inside that window: the set would be empty, the watermark would
// jump past the unregistered Seq, and a follower would treat a write that is
// still in progress as a permanent hole and skip it. Holding the lock across
// both steps closes that window, and has the convenient side effect that the
// freshly allocated Seq is always the largest in the set — so the insert is a
// plain append and the slice stays sorted by construction.
func (s *inflightSet) allocate(counter *atomic.Uint64) uint64 {
	s.mu.Lock()
	seq := counter.Add(1) - 1
	s.seqs = append(s.seqs, seq)
	s.mu.Unlock()
	return seq
}

// end retires seq, either because its write became durable or because the write
// failed and the Seq is burned. Both cases are resolutions: the Seq will never
// appear later, so the watermark may advance past it.
func (s *inflightSet) end(seq uint64) {
	s.mu.Lock()
	for i, v := range s.seqs {
		if v == seq {
			s.seqs = append(s.seqs[:i], s.seqs[i+1:]...)
			break
		}
	}
	if seq > s.maxEnded {
		s.maxEnded = seq
	}
	s.mu.Unlock()
}

// committed returns the highest Seq for which no allocation is still pending.
// Every Seq less than or equal to it is either durably readable or permanently
// gone; no later write can ever fill one of those slots.
func (s *inflightSet) committed() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.seqs) == 0 {
		return s.maxEnded
	}
	// the smallest pending Seq is the first one we cannot vouch for
	return s.seqs[0] - 1
}

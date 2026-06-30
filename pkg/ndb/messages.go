// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ndb

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"iter"

	"github.com/worldiety/option"
)

// TypeID identifies an event type. It is the stable, engine-neutral identifier
// a caller uses to address a logical stream of messages, and it is meant to be a
// durable, human-meaningful name (e.g. a discriminator like "FormCreated") that
// never changes for a given logical type.
//
// It is a string rather than a number on purpose: a stable textual identifier
// removes any need for an external name→number mapping and the sequence-recycling
// hazards that come with it (a removed type's id can never be accidentally reused,
// because there is no id pool). An engine may use the TypeID directly as part of
// its physical layout (e.g. msgstore uses it as a directory name), so callers
// must keep it filesystem-safe; see [ValidTypeID].
type TypeID string

// ValidTypeID reports whether s is a well-formed [TypeID]: 1–255 characters from
// the set [A-Za-z0-9._-]. The restriction keeps a TypeID safe to use directly as
// a path segment and free of separators, traversal, or case-collision surprises.
func ValidTypeID(s string) bool {
	if len(s) == 0 || len(s) > 255 {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '.', r == '_', r == '-':
		default:
			return false
		}
	}
	return true
}

// Seq is a globally strict-monotonic sequence number assigned by the engine on
// write. It is unique and never reused across the whole store, even across
// different TypeIDs. A Seq of 0 is reserved: it denotes "no message" (e.g. an
// empty [option.Opt]) and is also used by engines as a tombstone marker. Valid
// messages always carry a Seq greater than 0.
//
// Seq is a distinct named type (not a bare uint64) on purpose: it documents the
// intent at call sites such as Replay(types, minSeq, maxSeq) and prevents
// accidental mixing with byte lengths or offsets.
type Seq uint64

// TraceID is an opaque 16-byte correlation id a writer attaches to a message,
// so that an event can be tied back to whatever the caller wants (a request,
// a workflow, a batch id, ...). The store treats it as opaque bytes and never
// interprets it; the caller decides what to put in it.
//
// It is a distinct named type rather than a bare [16]byte so that the intent is
// clear at call sites, the value is self-describing in logs (see [TraceID.String]),
// and it cannot be confused with other 16-byte arrays. The zero value is a valid
// "no trace" marker (see [TraceID.IsZero]).
type TraceID [16]byte

// NewTraceID returns a TraceID filled with random bytes from a cryptographically
// secure source. Callers that want their own identifier can construct a TraceID
// directly instead. It panics only if the system random source fails, which
// indicates a broken platform rather than a recoverable condition.
func NewTraceID() TraceID {
	var t TraceID
	if _, err := rand.Read(t[:]); err != nil {
		panic(fmt.Errorf("ndb: read random for TraceID: %w", err))
	}
	return t
}

// String returns the lowercase hex encoding of the TraceID (32 characters),
// which is convenient for logging.
func (t TraceID) String() string {
	return hex.EncodeToString(t[:])
}

// IsZero reports whether the TraceID is the zero value, i.e. no trace was set.
func (t TraceID) IsZero() bool {
	return t == TraceID{}
}

// Encoding describes how a [Message] payload is encoded on the wire. It is part
// of the neutral contract so that consumers can decide whether they must decode
// a payload before use, without depending on a specific engine package.
//
// The zero value [EncodingRaw] always means "payload bytes are verbatim".
// Additional encodings are engine-defined but share this namespace so that a
// payload read from one engine carries an unambiguous, self-describing marker.
type Encoding uint8

const (
	// EncodingRaw means the payload is stored verbatim, no transformation.
	EncodingRaw Encoding = 0
	// EncodingS2 means the payload is S2-compressed (Snappy-compatible).
	EncodingS2 Encoding = 1
)

// Message is the engine-neutral, read-side view of a single stored event.
//
// It is intentionally a plain value type (no interface, no pointers besides the
// payload slice header) so that it can be yielded through iterators and returned
// by value without forcing a heap allocation per message. This keeps high-volume
// replay on a zero-allocation hot path.
//
// Payload lifetime: when a Message is obtained from a streaming iterator
// ([History.Replay]), Payload is a zero-copy view into a reusable read buffer
// owned by the engine. It is only valid until the iterator advances to the next
// message. Callers that need to retain the payload beyond the current iteration
// step MUST copy it:
//
//	kept := slices.Clone(msg.Payload)
//
// Messages returned by point lookups ([Retained.Get]) own an independent Payload
// that is safe to retain.
//
// If Payload is compressed (see Encoding), it holds the compressed bytes and
// UncompressedLen reports the original size. Decoding is the caller's choice so
// that an engine can deliver payloads without spending CPU and memory bandwidth
// on consumers that forward bytes verbatim.
type Message struct {
	// Type is the event type this message belongs to.
	Type TypeID
	// Seq is the global, strict-monotonic sequence number of this message.
	Seq Seq
	// TimeNano is the wall-clock append time in unix nanoseconds.
	//
	// It is deliberately an int64 rather than a time.Time: a time.Time value
	// carries a location pointer and an optional monotonic reading, both of
	// which add per-message cost and heap-escape pressure on the replay hot
	// path. Convert explicitly at the edge when a time.Time is required:
	//
	//	t := time.Unix(0, msg.TimeNano)
	TimeNano int64
	// TraceID is an opaque correlation id (e.g. a UUID) supplied by the writer.
	TraceID TraceID
	// Encoding marks how Payload is encoded.
	Encoding Encoding
	// UncompressedLen is the original payload length in bytes before encoding.
	// For EncodingRaw it equals len(Payload).
	UncompressedLen uint32
	// Payload holds the (possibly encoded) payload bytes. See the type-level
	// documentation for its lifetime rules.
	Payload []byte
}

// IsTombstone reports whether this message slot has been soft-deleted. A
// tombstone carries a zero Seq and an empty payload; engines preserve the slot
// to keep the sequence space contiguous.
func (m *Message) IsTombstone() bool {
	return m.Seq == 0
}

// History is the event-sourcing capability: an append-only log with full,
// ordered replay. It models the classic "append events, replay a sequence
// range" workflow.
//
// An engine that only retains last-known values (see [Retained]) need not
// implement History, and vice versa. Splitting the capabilities keeps the
// contract honest about what a given engine actually supports and lets the
// storage layer evolve without forcing every engine to grow every method.
type History interface {
	// Append durably writes a new message for typeID and returns the globally
	// unique, strict-monotonic Seq assigned to it. Appends to different TypeIDs
	// may proceed concurrently; the assigned Seq still reflects a single global
	// order.
	//
	// payload is copied/encoded by the engine before Append returns, so the
	// caller may reuse the backing array immediately.
	Append(typeID TypeID, traceID TraceID, payload []byte) (Seq, error)

	// Replay yields all messages for the requested types whose Seq lies within
	// the inclusive range [minSeq, maxSeq], in strict ascending global Seq
	// order across all requested types (k-way merge). When types is empty, all
	// known types are included.
	//
	// The most recent, not-yet-finalized messages are always included so that
	// replay never silently omits the tail.
	//
	// Each yielded Message.Payload is a view into a reusable buffer and is only
	// valid for the duration of the current iteration step (see [Message]).
	Replay(types []TypeID, minSeq, maxSeq Seq) iter.Seq2[TypeID, Message]
}

// Retained is the last-known-value capability, analogous to MQTT retained
// messages: each Put overwrites the single retained value for a type, and Get
// reads it back in O(1). No history is kept.
type Retained interface {
	// Put writes or overwrites the single retained value for typeID and returns
	// the global Seq assigned to it, so consumers can detect updates by
	// comparing sequence numbers.
	Put(typeID TypeID, traceID TraceID, payload []byte) (Seq, error)

	// Get returns the current value for typeID, or an empty option if no value
	// has ever been written. The returned Message owns an independent Payload
	// that is safe to retain.
	//
	// Get also works for [History]-style types: it returns the most recently
	// appended message for the type.
	Get(typeID TypeID) (option.Opt[Message], error)
}

// TimeLookup is the optional inverse time index: it resolves a wall-clock
// instant to the smallest Seq at or after that instant. This is intended for
// rare, exploratory queries ("where were we at 09:00?"), not for the regular
// read path.
type TimeLookup interface {
	// SeqForTime returns the smallest Seq whose append time is >= tsNano (unix
	// nanoseconds).
	SeqForTime(tsNano int64) (Seq, error)
}

// Pruner is the optional deletion capability. Deletion is coarse on purpose:
// the strict-monotonic sequence space must stay intact, and freed sequence
// numbers must never be reissued.
type Pruner interface {
	// DeleteType removes all messages of typeID.
	DeleteType(typeID TypeID) error

	// DeleteSeq soft-deletes a single message, turning it into a tombstone
	// (zeroed Seq and payload) while preserving the slot. Returns an error if
	// no message with seq exists for typeID.
	DeleteSeq(typeID TypeID, seq Seq) error
}

// Notification is the live signal that a message was written. It deliberately
// carries no payload: fan-out stays cheap and allocation-light, and the payload
// lifetime pitfalls of [History.Replay] (a reused-buffer view) do not apply.
// A subscriber that needs the bytes loads them on demand via [Retained.Get] or
// [History.Replay].
type Notification struct {
	// Type is the event type the written message belongs to.
	Type TypeID
	// Seq is the global, strict-monotonic sequence number of the written message.
	Seq Seq
	// TimeNano is the wall-clock append time in unix nanoseconds.
	TimeNano int64
	// TraceID is the opaque correlation id the writer supplied.
	TraceID TraceID
}

// Notifier is the optional live-notification capability. It delivers only the
// live edge — messages written from the moment of subscription onwards.
// Historical catch-up is the caller's job via [History.Replay], or, more
// conveniently, via the engine-neutral [Tail] helper which composes Subscribe
// and Replay into a single "replay then follow" stream.
type Notifier interface {
	// Subscribe registers fn for new messages of the given types (empty = all
	// types). fn is invoked synchronously on the writer's goroutine, just after
	// the message becomes durable.
	//
	// Because delivery is synchronous and inline, fn MUST be fast and MUST NOT
	// block or write back into the same store: a slow fn delays the writer, and
	// writing back risks a deadlock. Do only trivial work (set a flag, signal a
	// worker, enqueue) and hand off anything expensive to another goroutine.
	//
	// The returned close function unsubscribes and is idempotent.
	Subscribe(types []TypeID, fn func(Notification)) (close func())
}

// Messages is the full-featured composition of all message capabilities plus
// [io.Closer]. It is a convenience for engines that implement everything (such
// as the default msgstore engine). Consumers should depend on the narrowest
// capability interface they actually need (e.g. accept a [History] when they
// only replay) so that alternative engines with a reduced feature set remain
// usable.
type Messages interface {
	History
	Retained
	TimeLookup
	Pruner
	Notifier
	io.Closer
}

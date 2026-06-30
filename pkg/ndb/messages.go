// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ndb

import (
	"io"
	"iter"

	"github.com/worldiety/option"
)

// TypeID identifies an event type. It is the stable, engine-neutral identifier
// a caller uses to address a logical stream of messages. How an engine maps a
// TypeID onto physical storage (a directory, a column family, a table, ...) is
// an implementation detail and must not leak through this contract.
type TypeID uint64

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
	TraceID [16]byte
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
	Append(typeID TypeID, traceID [16]byte, payload []byte) (Seq, error)

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
	Put(typeID TypeID, traceID [16]byte, payload []byte) (Seq, error)

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
	io.Closer
}

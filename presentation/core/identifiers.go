// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"crypto/rand"
	"encoding/binary"
	"slices"
	"sync/atomic"
	"time"

	"go.wdy.de/nago/presentation/proto"
)

// The pointers of a scope are exact within JavaScript numbers (2^53). State pointers stay below 2^32 and
// callback pointers above, which keeps the frequent state pointers short on the wire and tells both apart
// when debugging. Each space starts at a random base within the lower part of its range, see identifiers.
const (
	minStatePtr     = 1 << 16
	maxStateBase    = 1 << 27
	maxStatePtr     = 1 << 32
	minCallbackPtr  = 1 << 32
	maxCallbackBase = 1 << 40
	maxCallbackPtr  = 1 << 53
	maxSeqBase      = 1 << 31
)

// identifiers hands out the identifiers of a scope: the pointers of callbacks, states and asynchronous calls
// and the ids of file transfers. Each is unique within the scope and never reused, thus a request which refers
// to a former window of the scope never hits anything of the current one, see discardStale.
//
// Each scope starts at random bases. A browser keeps its scope id across a reconnect, and a scope which lost
// its connection is reaped after a while, so the browser may end up with a new scope of the same id while it
// still shows the tree of the former one. With fixed bases, a click on that tree would hit whatever the new
// scope has allocated under the same numbers.
type identifiers struct {
	// callback is the next callback pointer, only for the event loop. Callbacks are mounted while rendering,
	// thus the callbacks of one tree form a contiguous interval, see callbackSegment.
	callback proto.Ptr
	// state is the next state pointer, only for the event loop.
	state proto.Ptr
	// async is the last pointer of an asynchronous call, which may be started from any goroutine.
	async atomic.Int64
	// file is the last id of a file transfer, only for the event loop.
	file int64
}

// init starts each space at a random base.
func (ids *identifiers) init() {
	ids.callback = randomBase(minCallbackPtr, maxCallbackBase)
	ids.state = randomBase(minStatePtr, maxStateBase)
	ids.async.Store(int64(randomBase(1, maxSeqBase)))
	ids.file = int64(randomBase(1, maxSeqBase))
}

// randomBase returns a random number within [min, max).
func randomBase(min, max proto.Ptr) proto.Ptr {
	var buf [8]byte
	_, _ = rand.Read(buf[:]) // never returns an error, see crypto/rand
	return min + proto.Ptr(binary.LittleEndian.Uint64(buf[:])%uint64(max-min))
}

// nextAsync returns the pointer of the next asynchronous call.
func (ids *identifiers) nextAsync() proto.Ptr {
	return proto.Ptr(ids.async.Add(1))
}

// nextFile returns the id of the next file transfer.
func (ids *identifiers) nextFile() int64 {
	ids.file++
	return ids.file
}

// callbackSegment holds the callbacks of one rendered tree. Their pointers form the interval
// [base, base+len(fns)), so a pointer resolves by index, and the whole segment is released at once.
type callbackSegment struct {
	base proto.Ptr
	fns  []func()
	// keys is parallel to fns, once any callback has a key, see [MountKeyedCallback]. Otherwise it is empty,
	// so a tree without keys costs nothing.
	keys    []CallbackKey
	hasKeys bool
	// byKey finds the pointer of a key; 0 marks a key which is used more than once.
	byKey map[CallbackKey]proto.Ptr
	// calledAt is when a callback of this tree was called last.
	calledAt time.Time
}

// restart empties the segment and continues at the next free pointer of the scope.
func (c *callbackSegment) restart(next proto.Ptr) {
	clear(c.fns) // release the closures, the backing array is reused
	c.fns = c.fns[:0]
	clear(c.keys)
	c.keys = c.keys[:0]
	c.hasKeys = false
	clear(c.byKey)
	c.calledAt = time.Time{}
	c.base = next
}

// dropKeys forgets the keys of the segment, so that none of its callbacks is redirected once it is superseded.
func (c *callbackSegment) dropKeys() {
	clear(c.keys)
	c.keys = c.keys[:0]
	c.hasKeys = false
	clear(c.byKey)
}

// mount appends f and returns its pointer.
func (c *callbackSegment) mount(f func()) proto.Ptr {
	ptr := c.base + proto.Ptr(len(c.fns))
	c.fns = append(c.fns, f)
	if c.hasKeys {
		c.keys = append(c.keys, CallbackKey{})
	}
	return ptr
}

// mountKeyed appends f under key and returns its pointer.
func (c *callbackSegment) mountKeyed(key CallbackKey, f func()) proto.Ptr {
	if !c.hasKeys {
		// the callbacks so far have no key
		c.hasKeys = true
		c.keys = slices.Grow(c.keys[:0], len(c.fns)+1)[:len(c.fns)]
	}

	ptr := c.mount(f)
	c.keys[len(c.keys)-1] = key

	if c.byKey == nil {
		c.byKey = map[CallbackKey]proto.Ptr{}
	}

	if _, used := c.byKey[key]; used {
		c.byKey[key] = 0 // ambiguous
	} else {
		c.byKey[key] = ptr
	}

	return ptr
}

// next returns the first pointer after the segment.
func (c *callbackSegment) next() proto.Ptr {
	return c.base + proto.Ptr(len(c.fns))
}

// lookup returns the callback of ptr or nil, if the pointer does not belong to this segment.
func (c *callbackSegment) lookup(ptr proto.Ptr) func() {
	if ptr < c.base || ptr >= c.next() {
		return nil
	}

	return c.fns[ptr-c.base]
}

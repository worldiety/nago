// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package sbox

import "sync"

// CapBuffer is a size-capped in-memory [io.Writer] for capturing the output of
// an untrusted process.
//
// A sandboxed process may misbehave and emit an unbounded amount of data on
// stdout/stderr. Wiring such output directly into an unbounded [bytes.Buffer]
// would let the child exhaust the host's memory. CapBuffer accepts at most Limit
// bytes and silently discards the rest (while still consuming the writes, so the
// child never blocks and no broken-pipe error is raised). Use it as Stdout
// and/or Stderr of a [Cmd].
//
// The zero value is not usable; construct one with [NewCapBuffer]. CapBuffer is
// safe for concurrent use, so the same buffer may back both Stdout and Stderr.
type CapBuffer struct {
	mu       sync.Mutex
	data     []byte
	limit    int
	dropped  int
	overflow bool
}

// NewCapBuffer returns a CapBuffer that retains at most limit bytes. A limit of
// zero or less defaults to 1 MiB.
func NewCapBuffer(limit int) *CapBuffer {
	if limit <= 0 {
		limit = 1 << 20 // 1 MiB
	}
	return &CapBuffer{limit: limit}
}

// Write implements [io.Writer]. It appends up to the remaining capacity and
// discards any excess. It always reports len(p) written and never returns an
// error, so the sandboxed process is never blocked or signalled on overflow.
func (b *CapBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	remaining := b.limit - len(b.data)
	if remaining <= 0 {
		b.overflow = true
		b.dropped += len(p)
		return len(p), nil
	}
	if len(p) > remaining {
		b.data = append(b.data, p[:remaining]...)
		b.overflow = true
		b.dropped += len(p) - remaining
		return len(p), nil
	}
	b.data = append(b.data, p...)
	return len(p), nil
}

// Bytes returns the captured bytes. The returned slice is a copy and safe to
// retain and mutate.
func (b *CapBuffer) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]byte, len(b.data))
	copy(out, b.data)
	return out
}

// String returns the captured output as a string.
func (b *CapBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.data)
}

// Len returns the number of bytes currently retained (never more than the
// limit).
func (b *CapBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.data)
}

// Overflowed reports whether the output exceeded the limit and some bytes were
// discarded.
func (b *CapBuffer) Overflowed() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.overflow
}

// Dropped returns the number of bytes that were discarded because the limit was
// reached.
func (b *CapBuffer) Dropped() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.dropped
}

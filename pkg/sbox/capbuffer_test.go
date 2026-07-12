// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package sbox

import (
	"strings"
	"testing"
)

func TestCapBufferWithinLimit(t *testing.T) {
	b := NewCapBuffer(100)
	n, err := b.Write([]byte("hello"))
	if err != nil || n != 5 {
		t.Fatalf("Write = %d, %v", n, err)
	}
	if b.String() != "hello" {
		t.Fatalf("String = %q", b.String())
	}
	if b.Overflowed() || b.Dropped() != 0 {
		t.Fatalf("unexpected overflow: overflow=%v dropped=%d", b.Overflowed(), b.Dropped())
	}
	if b.Len() != 5 {
		t.Fatalf("Len = %d", b.Len())
	}
}

func TestCapBufferOverflow(t *testing.T) {
	b := NewCapBuffer(4)
	// A single write exceeding the limit is truncated, not rejected.
	n, err := b.Write([]byte("abcdef"))
	if err != nil || n != 6 {
		t.Fatalf("Write = %d, %v (must consume all bytes, never error)", n, err)
	}
	if b.String() != "abcd" {
		t.Fatalf("String = %q, want %q", b.String(), "abcd")
	}
	if !b.Overflowed() {
		t.Fatalf("expected Overflowed")
	}
	if b.Dropped() != 2 {
		t.Fatalf("Dropped = %d, want 2", b.Dropped())
	}

	// Further writes are fully dropped but still consumed.
	n, err = b.Write([]byte("xyz"))
	if err != nil || n != 3 {
		t.Fatalf("Write = %d, %v", n, err)
	}
	if b.String() != "abcd" {
		t.Fatalf("String = %q after overflow", b.String())
	}
	if b.Dropped() != 5 {
		t.Fatalf("Dropped = %d, want 5", b.Dropped())
	}
}

func TestCapBufferDefaultLimit(t *testing.T) {
	b := NewCapBuffer(0)
	big := strings.Repeat("x", 2<<20) // 2 MiB into a defaulted 1 MiB buffer
	if _, err := b.Write([]byte(big)); err != nil {
		t.Fatalf("Write err: %v", err)
	}
	if b.Len() != 1<<20 {
		t.Fatalf("Len = %d, want %d", b.Len(), 1<<20)
	}
	if !b.Overflowed() {
		t.Fatalf("expected Overflowed")
	}
}

func TestCapBufferBytesIsCopy(t *testing.T) {
	b := NewCapBuffer(10)
	b.Write([]byte("abc"))
	out := b.Bytes()
	out[0] = 'X'
	if b.String() != "abc" {
		t.Fatalf("Bytes must return a copy; buffer mutated to %q", b.String())
	}
}

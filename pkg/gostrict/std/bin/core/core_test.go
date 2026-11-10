// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"bytes"
	"testing"
)

// TestBufferLimit checks, that a buffer without a limit of the host grows beyond the former default of 64 MiB, and that
// a limit of the host keeps the bytes within it and fails at Freeze instead of cutting them silently.
func TestBufferLimit(t *testing.T) {
	const former = 64 << 20
	cases := []struct {
		name  string
		limit int
		n     int
		// want is the length after appending n bytes, fail tells, whether Freeze fails
		want int64
		fail bool
	}{
		{"no limit beyond the former default", 0, former + 1, former + 1, false},
		{"no limit for a negative limit", -1, 10, 10, false},
		{"at the limit of the host", 8, 8, 8, false},
		{"one byte beyond the limit of the host", 8, 9, 8, true},
	}
	block := bytes.Repeat([]byte{0x5a}, 1<<20)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b := NewBuffer(0, c.limit)
			// a small buffer gets single bytes, because an append beyond the limit appends none of its bytes
			step := min(c.n, len(block))
			if c.n < 100 {
				step = 1
			}
			for left := c.n; left > 0; left -= step {
				b.AppendBytes(block[:min(left, step)])
			}
			if b.Len() != c.want {
				t.Fatalf("Len is %d, want %d", b.Len(), c.want)
			}
			if got := int64(len(b.Snapshot())); got != c.want {
				t.Fatalf("Snapshot has %d bytes, want %d", got, c.want)
			}
			data, reason := b.Freeze()
			if c.fail {
				if reason != "bin: the buffer exceeds its limit of 8 bytes" || data != nil {
					t.Fatalf("Freeze returned %d bytes and %q, want the failure", len(data), reason)
				}
				return
			}
			if reason != "" || int64(len(data)) != c.want || data[len(data)-1] != 0x5a {
				t.Fatalf("Freeze returned %d bytes and %q", len(data), reason)
			}
		})
	}
}

// TestNewBufferReserve checks, that a capacity is only a hint: a huge one neither fails nor takes its memory at once.
func TestNewBufferReserve(t *testing.T) {
	cases := []struct {
		capacity int64
		limit    int
		want     int
	}{
		{-1, 0, 0},
		{100, 0, 100},
		{1 << 40, 0, MaxReserve},
		{100, 8, 8},
	}
	for _, c := range cases {
		if got := cap(NewBuffer(c.capacity, c.limit).data); got != c.want {
			t.Errorf("NewBuffer(%d, %d) reserves %d, want %d", c.capacity, c.limit, got, c.want)
		}
	}
}

// TestOffsets checks the offsets of int64 beyond the range of int32 and below zero.
func TestOffsets(t *testing.T) {
	data := []byte("0123456789")
	cases := []struct {
		name string
		read func(r *Reader) string
		want string
	}{
		{"seek within", func(r *Reader) string { r.SeekTo(4); return r.Text(6) }, ""},
		{"seek beyond int32", func(r *Reader) string { r.SeekTo(1 << 32); return r.Text(1) }, "bin: seeking to 4294967296 exceeds the length 10 at offset 0"},
		{"skip beyond int32", func(r *Reader) string { r.Skip(1<<31 + 3); return "" }, "bin: reading 2147483651 bytes exceeds the length 10 at offset 0"},
		{"negative bytes", func(r *Reader) string { return string(r.Bytes(-1)) }, "bin: reading -1 bytes exceeds the length 10 at offset 0"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := NewReader(data)
			c.read(r)
			if got := r.Close(); got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
	if got := string(Slice(data, -1<<40, 1<<40)); got != "0123456789" {
		t.Fatalf("Slice clamps to %q", got)
	}
	if got := Slice(data, 1<<40, 1<<41); got != nil {
		t.Fatalf("Slice beyond the end is %q", got)
	}
}

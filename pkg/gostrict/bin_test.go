// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/binary"
)

const binHeader = `// Package hdr calls functions on another goroutine.
//
//gostrict:header
package hdr

// Later calls done after it returned on another goroutine.
//
//gostrict:host
//gostrict:calls concurrent
func Later(done func()) {}
`

// TestBinRules checks, that the compiler separates what can change from what can not.
func TestBinRules(t *testing.T) {
	cases := []struct {
		name, src string
		// code is the expected finding or empty
		code gostrict.Code
	}{
		{"bytes in a concurrent function", `func Run(b bin.Bytes) { hdr.Later(func() { _ = b.Len() }) }`, ""},
		{"a buffer in a concurrent function", `func Run(b bin.Buffer) { hdr.Later(func() { _ = b.Len() }) }`, gostrict.CodeConcurrent},
		{"a reader in a concurrent function", `func Run(d bin.Bytes) int32 {
	v, err := bin.Decode(d, func(r bin.Reader) int32 {
		hdr.Later(func() { _ = r.Pos() })
		return 0
	})
	if err != nil {
		return 0
	}
	return v
}`, gostrict.CodeConcurrent},
		{"bytes as a package variable", `var magic = bin.Of(1, 2)

func Run() int32 { return magic.Len() }`, ""},
		{"a buffer as a package variable", `var shared = bin.NewBuffer(0)

func Run() int32 { return shared.Len() }`, gostrict.CodeGlobal},
		{"== of bytes", `func Run(a, b bin.Bytes) bool { return a == b }`, gostrict.CodeBigEqual},
		{"a byte without a check", `func Run(b bin.Bytes) uint32 {
	v, _ := b.Get(3)
	return v
}`, gostrict.CodeGuard},
		{"a byte after a proof of the index", `func Run(b bin.Bytes) uint32 {
	var sum uint32
	for i := range b.Len() {
		v, _ := b.Get(i)
		sum += v
	}
	return sum
}`, ""},
		{"a result of Decode without a check", `func Run(b bin.Bytes) uint32 {
	v, _ := bin.Decode(b, func(r bin.Reader) uint32 { return r.U8() })
	return v
}`, gostrict.CodeGuard},
		{"a bit field of a variable width", `func Run(x uint32, n uint32) uint32 { return x & (1<<n - 1) }`, ""},
		{"a signed shift count", `func Run(x uint32, n int32) uint32 { return x << n }`, gostrict.CodeBits},
		{"a constant count beyond 32 bits", `func Run(x uint32) uint32 { return x << 32 }`, gostrict.CodeBits},
		{"a constant count within 64 bits", `func Run(x uint64) uint64 { return x << 63 }`, ""},
		{"a constant count beyond 64 bits", `func Run(x int64) int64 { return x >> 64 }`, gostrict.CodeBits},
		{"int stays forbidden", `func Run(x int) int { return x }`, gostrict.CodeNumber},
		{"byte stays forbidden", `func Run(x byte) byte { return x }`, gostrict.CodeNumber},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := "// Package p is a test.\n//\n//gostrict:terminating\npackage p\n\nimport (\n\t\"go.wdy.de/nago/pkg/gostrict/std/bin\"\n\t\"llm.local/p/hdr\"\n)\n\nvar _ = bin.Empty\n\nfunc use() { hdr.Later(func() {}) }\n\n" + c.src + "\n"
			prog, err := gostrict.LoadModules("llm.local/p", gostrict.Module{Path: "llm.local/p", Trusted: true, FS: fstest.MapFS{
				"p.go":     {Data: []byte(src)},
				"hdr/h.go": {Data: []byte(binHeader)},
			}})
			if err != nil {
				t.Fatal(err)
			}
			diags := gostrict.Check(prog)
			if c.code == "" {
				if len(diags) > 0 {
					t.Fatalf("want no finding, got:\n%s", diags.Text())
				}
				return
			}
			for _, d := range diags {
				if d.Code == c.code {
					return
				}
			}
			t.Fatalf("want %s, got:\n%s", c.code, diags.Text())
		})
	}
}

// TestBinaryRoundTrip checks the content of a recording, which TestDifferential only compares between the engines.
func TestBinaryRoundTrip(t *testing.T) {
	hex, err := binary.Encode(7, -42, 1.5, "Grüße", 5)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(hex, "574b44410008a1ffffffffffffffd63fc00000") {
		t.Fatalf("the head is wrong: %s", hex)
	}
	h, total, err := binary.Decode(hex, 8, 5)
	if err != nil {
		t.Fatal(err)
	}
	want := binary.Header{Magic: 0x04030201, Version: 8, Flags: 5<<8 | 1, Serial: -42, Scale: 3, Name: "Grüße"}
	h.Offset = 0
	if h != want {
		t.Fatalf("got %+v, want %+v", h, want)
	}
	// the samples are i*37-100 of 12 bits
	if total != -500+37*10 {
		t.Fatalf("the sum of the samples is %d", total)
	}

	if _, _, err := binary.Decode(hex, 8, 1000); err == nil || !strings.Contains(err.Error(), "bits exceeds") {
		t.Fatalf("want an error about the end of the data, got %v", err)
	}
	if _, _, err := binary.Decode("00112233", 0, 0); err == nil || !strings.Contains(err.Error(), "wrong magic") {
		t.Fatalf("want the reason of Fail, got %v", err)
	}
}

// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"math"
	"testing"

	"go.wdy.de/nago/pkg/gostrict/std/text"
)

// The native side of the tests in textnum_test.go: the functions of std/text in Go against the same rows.

// TestParseFloatNative compares text.ParseFloat in Go with the contract, and the value beside an error with goja and
// the vm.
func TestParseFloatNative(t *testing.T) {
	engines := loadTextNum(t, textNumSrc)
	for _, r := range parseFloatRows() {
		f, err := text.ParseFloat(r.in)
		msg := ""
		if err != nil {
			msg = err.Error()
		}
		checkParseFloat(t, "native", r, f, math.Signbit(f), msg, err != nil)
		if err != nil {
			for _, e := range engines {
				if g, _, _, _ := e.parseFloat(t, r.in); math.Float64bits(g) != math.Float64bits(f) {
					t.Errorf("ParseFloat(%s): the value beside the error is %#016x natively, %#016x in %s", abbrev(r.in),
						math.Float64bits(f), math.Float64bits(g), e.name)
				}
			}
		}
	}
}

// TestParseFloatRandomNative compares text.ParseFloat in Go with strconv.ParseFloat.
func TestParseFloatRandomNative(t *testing.T) {
	bad := 0
	ins := randomInputs()
	for _, s := range ins {
		bits, msg := strconvExpect(s)
		f, err := text.ParseFloat(s)
		var ok bool
		if msg != "" {
			ok = err != nil && err.Error() == msg
		} else {
			ok = err == nil && math.Float64bits(f) == bits
		}
		if !ok {
			bad++
			if bad <= 20 {
				t.Errorf("ParseFloat(%s) = %#016x, %v, strconv %#016x, %q", abbrev(s), math.Float64bits(f), err, bits, msg)
			}
		}
	}
	if bad > 0 {
		t.Errorf("%d of %d differ from strconv.ParseFloat", bad, len(ins))
	}
}

// TestSliceNative compares text.Slice in Go with the contract and with text.Chars.
func TestSliceNative(t *testing.T) {
	for _, r := range sliceRows {
		if got := text.Slice(r.s, r.from, r.to); got != r.want {
			t.Errorf("Slice(%q, %d, %d) = %q, want %q", r.s, r.from, r.to, got, r.want)
		}
	}
	for _, c := range charCounts {
		cs := text.Chars(c.s)
		if cs.Len() != c.n || text.RuneCount(c.s) != c.n {
			t.Errorf("Chars(%q) has %d characters, RuneCount %d, want %d", c.s, cs.Len(), text.RuneCount(c.s), c.n)
		}
		all := ""
		for i := int32(-1); i <= c.n; i++ {
			at, _ := cs.Get(i)
			if got := text.Slice(c.s, i, i+1); got != at {
				t.Errorf("Slice(%q, %d, %d) = %q, Chars has %q", c.s, i, i+1, got, at)
			}
			all += at
		}
		if got := text.Slice(c.s, 0, c.n); got != all {
			t.Errorf("Slice(%q, 0, %d) = %q, the characters are %q", c.s, c.n, got, all)
		}
	}
}

// TestGroupDigitsNative compares text.GroupDigits in Go with the contract.
func TestGroupDigitsNative(t *testing.T) {
	for _, r := range groupRows() {
		if got := text.GroupDigits(r.s, r.decimalMark, r.groupMark); got != r.want {
			t.Errorf("GroupDigits(%s, %q, %q) = %s, want %s", abbrev(r.s), r.decimalMark, r.groupMark, abbrev(got), abbrev(r.want))
		}
	}
	for _, r := range groupFloatRows {
		if got := text.GroupDigits(text.FromFloat(r.f, r.decimals), ",", "."); got != r.want {
			t.Errorf("GroupDigits(FromFloat(%v, %d), \",\", \".\") = %q, want %q", r.f, r.decimals, got, r.want)
		}
	}
}

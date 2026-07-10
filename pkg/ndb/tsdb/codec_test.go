// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"math"
	"testing"

	"go.wdy.de/nago/pkg/xbytes"
)

func TestZigzagRoundTrip(t *testing.T) {
	for _, v := range []int64{0, 1, -1, 2, -2, 1 << 40, -(1 << 40), math.MaxInt64, math.MinInt64 + 1} {
		if got := unzigzag(zigzag(v)); got != v {
			t.Fatalf("zigzag round trip %d -> %d", v, got)
		}
	}
}

func TestScaleUnscale(t *testing.T) {
	cases := []struct {
		v    float64
		d    uint8
		want int64
	}{
		{1.2345, 2, 123},
		{1.235, 2, 124}, // round half away
		{-1.235, 2, -124},
		{0, 4, 0},
		{100.0, 0, 100},
	}
	for _, c := range cases {
		if got := scale(c.v, c.d); got != c.want {
			t.Fatalf("scale(%v,%d)=%d want %d", c.v, c.d, got, c.want)
		}
	}
	if !math.IsNaN(unscale(NA, 2)) {
		t.Fatal("unscale(NA) should be NaN")
	}
	if got := unscale(123, 2); got != 1.23 {
		t.Fatalf("unscale=%v", got)
	}
}

func TestTimestampCodecEquidistant(t *testing.T) {
	// equidistant series → delta-of-delta all zero → compact
	ts := make([]int64, 1000)
	base := int64(1_700_000_000_000)
	for i := range ts {
		ts[i] = base + int64(i)*20 // 50 Hz
	}
	var b xbytes.Buffer
	encodeTimestamps(&b, ts)
	encoded := b.Pos
	// first ts (~5 bytes) + first delta (~1 byte) + 998 zero dods (1 byte each)
	if encoded > 1100 {
		t.Fatalf("equidistant timestamps did not compress: %d bytes for %d points", encoded, len(ts))
	}
	rb := xbytes.Buffer{Buf: b.Buf[:b.Pos]}
	got, err := decodeTimestamps(&rb, len(ts), nil)
	if err != nil {
		t.Fatal(err)
	}
	for i := range ts {
		if got[i] != ts[i] {
			t.Fatalf("ts[%d]=%d want %d", i, got[i], ts[i])
		}
	}
}

func TestValueCodecRoundTrip(t *testing.T) {
	vals := []int64{100, 101, 100, 99, 100, NA, 100, 5000, 4999}
	var b xbytes.Buffer
	encodeValues(&b, vals)
	rb := xbytes.Buffer{Buf: b.Buf[:b.Pos]}
	got, err := decodeValues(&rb, len(vals), nil)
	if err != nil {
		t.Fatal(err)
	}
	for i := range vals {
		if got[i] != vals[i] {
			t.Fatalf("val[%d]=%d want %d", i, got[i], vals[i])
		}
	}
}

func TestEquidistantDetectionAndRoundTrip(t *testing.T) {
	ts := make([]int64, 1000)
	base := int64(1_700_000_000_000)
	for i := range ts {
		ts[i] = base + int64(i)*20
	}
	equi, step := isEquidistant(ts)
	if !equi || step != 20 {
		t.Fatalf("expected equidistant step=20, got equi=%v step=%d", equi, step)
	}

	bd := &blockData{ts: ts, vals: make([]int64, len(ts))}
	for i := range bd.vals {
		bd.vals[i] = int64(100 + i%7)
	}
	frame, err := encodeBlock(nil, bd, SchemeDecimal, false)
	if err != nil {
		t.Fatal(err)
	}
	h, err := parseBlockHeader(frame)
	if err != nil {
		t.Fatal(err)
	}
	if h.flags&flagEquidistantTS == 0 {
		t.Fatal("expected equidistant flag to be set")
	}
	var out blockData
	if err := decodeBlockBody(frame, h, SchemeDecimal, &out); err != nil {
		t.Fatal(err)
	}
	for i := range ts {
		if out.ts[i] != ts[i] || out.vals[i] != bd.vals[i] {
			t.Fatalf("mismatch at %d: ts %d/%d val %d/%d", i, out.ts[i], ts[i], out.vals[i], bd.vals[i])
		}
	}

	// a single hole must disable the equidistant encoding (fallback to dod)
	ts2 := append([]int64(nil), ts...)
	ts2[500] += 7
	if e, _ := isEquidistant(ts2); e {
		t.Fatal("series with a hole must not be equidistant")
	}
	bd2 := &blockData{ts: ts2, vals: bd.vals}
	frame2, _ := encodeBlock(nil, bd2, SchemeDecimal, false)
	h2, _ := parseBlockHeader(frame2)
	if h2.flags&flagEquidistantTS != 0 {
		t.Fatal("hole block must not set equidistant flag")
	}
	var out2 blockData
	if err := decodeBlockBody(frame2, h2, SchemeDecimal, &out2); err != nil {
		t.Fatal(err)
	}
	for i := range ts2 {
		if out2.ts[i] != ts2[i] {
			t.Fatalf("dod fallback mismatch at %d: %d/%d", i, out2.ts[i], ts2[i])
		}
	}
}

func TestEquidistantIsSmaller(t *testing.T) {
	ts := make([]int64, 4096)
	base := int64(1_700_000_000_000)
	for i := range ts {
		ts[i] = base + int64(i)*20
	}
	vals := make([]int64, len(ts))
	equiFrame, _ := encodeBlock(nil, &blockData{ts: ts, vals: vals}, SchemeDecimal, false)

	ts2 := append([]int64(nil), ts...)
	ts2[2000] += 20 // one gap → non-equidistant
	dodFrame, _ := encodeBlock(nil, &blockData{ts: ts2, vals: vals}, SchemeDecimal, false)

	if len(equiFrame) >= len(dodFrame) {
		t.Fatalf("equidistant frame (%d B) should be smaller than dod frame (%d B)", len(equiFrame), len(dodFrame))
	}
	t.Logf("equidistant block: %d B, delta-of-delta block: %d B (%d points)", len(equiFrame), len(dodFrame), len(ts))
}

func TestConstValueRoundTrip(t *testing.T) {
	ts := make([]int64, 4096)
	vals := make([]int64, 4096)
	base := int64(1_700_000_000_000)
	for i := range ts {
		ts[i] = base + int64(i)*20
		vals[i] = 2000 // constant
	}
	if !allEqualI64(vals) {
		t.Fatal("expected constant values")
	}
	frame, err := encodeBlock(nil, &blockData{ts: ts, vals: vals}, SchemeDecimal, false)
	if err != nil {
		t.Fatal(err)
	}
	h, _ := parseBlockHeader(frame)
	if h.flags&flagConstVal == 0 {
		t.Fatal("expected constant-value flag")
	}
	if h.flags&flagEquidistantTS == 0 {
		t.Fatal("expected equidistant flag too (flat + equidistant compose)")
	}
	var out blockData
	if err := decodeBlockBody(frame, h, SchemeDecimal, &out); err != nil {
		t.Fatal(err)
	}
	if len(out.vals) != 4096 {
		t.Fatalf("count=%d", len(out.vals))
	}
	for i := range out.vals {
		if out.vals[i] != 2000 || out.ts[i] != ts[i] {
			t.Fatalf("mismatch at %d: ts=%d val=%d", i, out.ts[i], out.vals[i])
		}
	}
	// a flat + equidistant 4096-point block must be tiny, independent of count.
	if len(frame) > 128 {
		t.Fatalf("flat equidistant block unexpectedly large: %d bytes", len(frame))
	}
	t.Logf("flat equidistant 4096-point block: %d bytes total", len(frame))
}

func TestConstValueBreaksOnChange(t *testing.T) {
	vals := make([]int64, 100)
	for i := range vals {
		vals[i] = 5
	}
	vals[50] = 6 // one change
	if allEqualI64(vals) {
		t.Fatal("non-constant values must not be detected as constant")
	}
	ts := make([]int64, 100)
	for i := range ts {
		ts[i] = int64(i)
	}
	frame, _ := encodeBlock(nil, &blockData{ts: ts, vals: vals}, SchemeDecimal, false)
	h, _ := parseBlockHeader(frame)
	if h.flags&flagConstVal != 0 {
		t.Fatal("changed block must not set constant flag")
	}
	var out blockData
	if err := decodeBlockBody(frame, h, SchemeDecimal, &out); err != nil {
		t.Fatal(err)
	}
	for i := range vals {
		if out.vals[i] != vals[i] {
			t.Fatalf("mismatch at %d: %d/%d", i, out.vals[i], vals[i])
		}
	}
}

func TestBlockRoundTrip(t *testing.T) {
	bd := &blockData{
		ts:   []int64{10, 20, 30, 40},
		vals: []int64{1, 2, 3, 4},
	}
	frame, err := encodeBlock(nil, bd, SchemeDecimal, false)
	if err != nil {
		t.Fatal(err)
	}
	h, err := parseBlockHeader(frame)
	if err != nil {
		t.Fatal(err)
	}
	if h.minMillis != 10 || h.maxMillis != 40 || h.count != 4 {
		t.Fatalf("header wrong: %+v", h)
	}
	var out blockData
	if err := decodeBlockBody(frame, h, SchemeDecimal, &out); err != nil {
		t.Fatal(err)
	}
	if len(out.ts) != 4 || out.vals[3] != 4 || out.ts[0] != 10 {
		t.Fatalf("decoded wrong: %+v", out)
	}
}

func TestBlockCRCDetectsCorruption(t *testing.T) {
	bd := &blockData{ts: []int64{1, 2, 3}, vals: []int64{9, 8, 7}}
	frame, _ := encodeBlock(nil, bd, SchemeDecimal, false)
	h, _ := parseBlockHeader(frame)
	frame[h.bodyOff] ^= 0xFF // flip a body byte
	var out blockData
	if err := decodeBlockBody(frame, h, SchemeDecimal, &out); err == nil {
		t.Fatal("expected CRC failure")
	}
}

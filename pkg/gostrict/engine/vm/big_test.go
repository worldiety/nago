// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vm_test

import (
	"context"
	"math/rand/v2"
	"strings"
	"testing"
	"testing/fstest"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/vm"
	"go.wdy.de/nago/pkg/gostrict/std/big"
)

// TestBig compares the functions of std/big in the vm, which implements them with math/big, with std/big itself.
func TestBig(t *testing.T) {
	src := `package s

import "go.wdy.de/nago/pkg/gostrict/std/big"

func Two(a, b big.Int) (big.Int, big.Int, big.Int, big.Int, bool, big.Int, bool, big.Int, big.Int, bool, int32) {
	q, r, q2, r2 := big.Zero(), big.Zero(), big.Zero(), big.Zero()
	qv, qok := big.Quo(a, b)
	if qok {
		q = qv
	}
	rv, rok := big.Rem(a, b)
	if rok {
		r = rv
	}
	qv2, rv2, ok := big.QuoRem(a, b)
	if ok {
		q2, r2 = qv2, rv2
	}
	return big.Add(a, b), big.Sub(a, b), big.Mul(a, b), q, qok, r, rok, q2, r2, ok, big.Cmp(a, b)
}

func One(a big.Int, n int32, u uint32, s string) (big.Int, big.Int, int32, int32, bool, float64, big.Int, big.Int, big.Int, string, big.Int, bool, big.Int) {
	i, ok := big.Int32(a)
	p := big.Zero()
	pv, pok := big.Parse(s)
	if pok {
		p = pv
	}
	return big.Neg(a), big.Abs(a), big.Sign(a), i, ok, big.Float64(a), big.FromInt32(n), big.FromUint32(u), big.Pow10(n), big.String(a), p, pok, big.Zero()
}
`
	b, diags, err := gostrict.NewCache(1).Compile("llm.local/s", gostrict.Module{Path: "llm.local/s", FS: fstest.MapFS{"s.go": {Data: []byte(src)}}})
	if err != nil || len(diags) > 0 {
		t.Fatal(err, diags.Text())
	}
	in, err := new(vm.Engine).Load(context.Background(), b, nil)
	if err != nil {
		t.Fatal(err)
	}
	rnd := rand.New(rand.NewPCG(3, 4))
	ints := []string{"0", "1", "-1", "7", "-7", "2147483647", "-2147483648", "2147483648", "9223372036854775808", "-" + strings.Repeat("9", 40)}
	randInt := func() string {
		if rnd.IntN(3) == 0 {
			return ints[rnd.IntN(len(ints))]
		}
		var sb strings.Builder
		if rnd.IntN(2) == 0 {
			sb.WriteString("-")
		}
		sb.WriteByte(byte('1' + rnd.IntN(9)))
		for range rnd.IntN(30) {
			sb.WriteByte(byte('0' + rnd.IntN(10)))
		}
		return sb.String()
	}
	parse := func(s string) big.Int {
		v, _ := big.Parse(s)
		return v
	}
	for range 500 {
		x, y := randInt(), randInt()
		a, c := parse(x), parse(y)
		q, qok := big.Quo(a, c)
		r, rok := big.Rem(a, c)
		q2, r2, ok := big.QuoRem(a, c)
		want := []any{big.String(big.Add(a, c)), big.String(big.Sub(a, c)), big.String(big.Mul(a, c)), big.String(q), qok, big.String(r), rok, big.String(q2), big.String(r2), ok, big.Cmp(a, c)}
		got, err := in.Call(context.Background(), "Two", x, y)
		if err != nil || !gostrict.SameValues(got, want) {
			t.Fatalf("Two(%s, %s): vm %v, %v, std/big %v", x, y, got, err, want)
		}

		n := int32(rnd.IntN(60) - 10)
		if rnd.IntN(20) == 0 {
			n = 5000
		}
		u := rnd.Uint32()
		s := []string{x, "+" + y, " 1", "1e3", "", "--1", "12x"}[rnd.IntN(7)]
		i, iok := big.Int32(a)
		p, pok := big.Parse(s)
		want = []any{big.String(big.Neg(a)), big.String(big.Abs(a)), big.Sign(a), i, iok, big.Float64(a), big.String(big.FromInt32(n)), big.String(big.FromUint32(u)), big.String(big.Pow10(n)), big.String(a), big.String(p), pok, big.String(big.Zero())}
		got, err = in.Call(context.Background(), "One", x, n, u, s)
		if err != nil || !gostrict.SameValues(got, want) {
			t.Fatalf("One(%s, %d, %d, %q): vm %v, %v, std/big %v", x, n, u, s, got, err, want)
		}
	}
}

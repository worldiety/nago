// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gojaengine

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"

	"go.wdy.de/nago/pkg/gostrict"
)

// Mismatch is a call, for which the JavaScript binding of a function of a header package and its native
// implementation return different values.
type Mismatch struct {
	Func   string
	Args   []any
	Native []any
	JS     []any
	Err    error
}

func (m Mismatch) String() string {
	if m.Err != nil {
		return fmt.Sprintf("%s%v: %v", m.Func, m.Args, m.Err)
	}
	return fmt.Sprintf("%s%v: native %v, javascript %v", m.Func, m.Args, m.Native, m.JS)
}

// CheckBindings calls each function of a header package, which the bundle uses and the host implements, with
// edge cases and n random arguments, natively and through its JavaScript binding, and returns the calls, whose
// results differ, floats beyond the [Tolerance]. The bodies of header packages are trusted, so this is the test, that the two implementations
// of a function agree.
func CheckBindings(ctx context.Context, b *gostrict.Bundle, host gostrict.Funcs, n int) ([]Mismatch, error) {
	inst, err := new(Engine).Load(ctx, b, nil)
	if err != nil {
		return nil, err
	}
	defer inst.Close()
	in := inst.(*instance)

	var res []Mismatch
	rnd := rand.New(rand.NewPCG(1, 2))
	for _, x := range b.Externs {
		native, ok := host[x.Name]
		if !ok || hasHandle(x.Func) {
			// a handle refers to an object of one implementation, which the other one can not compare
			continue
		}
		for i := 0; i < n+len(edges); i++ {
			args := make([]any, len(x.Params))
			for j, t := range x.Params {
				args[j] = sample(rnd, t.Kind, i)
			}

			want := native(args)
			got, err := in.CallExtern(ctx, x.Func, args...)
			switch {
			case err != nil:
				res = append(res, Mismatch{Func: x.Name, Args: args, Native: want, Err: err})
			case !gostrict.CloseValues(want, got, Tolerance):
				res = append(res, Mismatch{Func: x.Name, Args: args, Native: want, JS: got})
			}
		}
	}
	return res, nil
}

// Tolerance is the relative tolerance, up to which the floats of a native function and its binding may differ.
const Tolerance = 1e-9

// edges are the indices of the samples, which are edge cases rather than random values.
var edges = make([]struct{}, 8)

func sample(rnd *rand.Rand, k gostrict.ValueKind, i int) any {
	edge := i < len(edges)
	switch k {
	case gostrict.ValueBool:
		return i%2 == 0
	case gostrict.ValueInt32:
		if edge {
			return []int32{0, 1, -1, 2, -2, math.MaxInt32, math.MinInt32, 1000}[i]
		}
		return int32(rnd.Uint32())
	case gostrict.ValueUint32:
		if edge {
			return []uint32{0, 1, 2, 255, 1 << 31, math.MaxUint32, 1000, 7}[i]
		}
		return rnd.Uint32()
	case gostrict.ValueFloat64:
		if edge {
			return []float64{0, math.Copysign(0, -1), 0.5, -2.5, math.NaN(), math.Inf(1), math.Inf(-1), math.MaxFloat64}[i]
		}
		return (rnd.Float64() - 0.5) * math.Pow(10, float64(rnd.IntN(20)-10))
	case gostrict.ValueString:
		if edge {
			return []string{"", "a", "Grüße", "😀", " x ", "\u0085", "�", "a,b"}[i]
		}
		runes := []rune("abcäöü😀 ,.-\t")
		var s []rune
		for range rnd.IntN(8) {
			s = append(s, runes[rnd.IntN(len(runes))])
		}
		return string(s)
	}
	return nil
}

func hasHandle(f gostrict.Func) bool {
	for _, t := range append(append([]gostrict.Type{}, f.Params...), f.Results...) {
		if t.Kind == gostrict.ValueHandle {
			return true
		}
	}
	return false
}

// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package decimal provides exact decimal numbers of any size, e.g. for amounts of money, prices with more decimal
// places and the totals of an invoice. A [Dec] never changes: each function returns a new one. Addition,
// subtraction and multiplication are exact, a division and [Round] round to a given scale with an explicit
// [Rounding], e.g. [HalfUp] for the commercial rounding of an invoice.
//
// The package is written in the subset itself on top of package big, so it computes the same in Go and in
// JavaScript. No function panics: a division by zero reports it by its result.
package decimal

import (
	"strings"

	"go.wdy.de/nago/pkg/gostrict/std/big"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

// Dec is the integer M scaled by 10^-S, e.g. 12.30 is {M: 1230, S: 2}. Create a Dec with [FromInt32], [New] or
// [Parse]. The scale is from 0 to [MaxScale]; the functions treat a scale out of this range, which only a literal can
// create, as clamped into it.
type Dec struct {
	M big.Int
	S int32
}

// MaxScale is the greatest scale of a [Dec], far more than any amount needs. [Mul] rounds a product with a greater
// scale by [HalfEven].
const MaxScale = 1000

// Rounding tells, how a result is rounded to its scale.
type Rounding int32

const (
	// HalfUp rounds to the nearest and a half away from zero, the commercial rounding, e.g. 2.5 to 3 and -2.5 to
	// -3.
	HalfUp Rounding = iota
	// HalfEven rounds to the nearest and a half to the even neighbor, e.g. 2.5 to 2 and 3.5 to 4.
	HalfEven
	// Down rounds towards zero.
	Down
	// Up rounds away from zero.
	Up
	// Floor rounds towards negative infinity.
	Floor
	// Ceiling rounds towards positive infinity.
	Ceiling
)

// Zero returns 0.
func Zero() Dec {
	return Dec{M: big.Zero(), S: 0}
}

// FromInt32 returns n.
func FromInt32(n int32) Dec {
	return Dec{M: big.FromInt32(n), S: 0}
}

// FromBig returns n.
func FromBig(n big.Int) Dec {
	return Dec{M: n, S: 0}
}

// New returns m scaled by 10^-s, e.g. New(big.FromInt32(1230), 2) is 12.30. A negative s multiplies m by 10^-s.
// New clamps s into -MaxScale..MaxScale.
func New(m big.Int, s int32) Dec {
	s = min(max(s, 0-MaxScale), MaxScale)
	if s < 0 {
		return Dec{M: big.Mul(m, big.Pow10(0-s)), S: 0}
	}
	return Dec{M: m, S: s}
}

// Parse parses a decimal with an optional sign and an optional fraction, e.g. "-12.30". The scale is the number of
// digits of the fraction. It returns false for anything else, e.g. "1e3", ".5" or "1,5", and for more than
// [MaxScale] digits of the fraction.
func Parse(s string) (Dec, bool) {
	sign := ""
	rest := s
	if strings.HasPrefix(rest, "-") {
		sign = "-"
		rest = strings.TrimPrefix(rest, "-")
	} else if strings.HasPrefix(rest, "+") {
		rest = strings.TrimPrefix(rest, "+")
	}
	if strings.HasPrefix(rest, "-") || strings.HasPrefix(rest, "+") {
		return Zero(), false
	}

	parts := text.Split(rest, ".")
	whole, ok := parts.Get(0)
	if !ok || whole == "" {
		return Zero(), false
	}
	fraction := ""
	switch parts.Len() {
	case 1:
	case 2:
		fraction, _ = parts.Get(1)
		if fraction == "" || text.RuneCount(fraction) > MaxScale {
			return Zero(), false
		}
	default:
		return Zero(), false
	}

	m, valid := big.Parse(sign + whole + fraction)
	if !valid {
		return Zero(), false
	}
	return Dec{M: m, S: text.RuneCount(fraction)}, true
}

// String formats d with its scale, e.g. "-12.30".
func String(d Dec) string {
	d = norm(d)
	digits := big.String(big.Abs(d.M))
	sign := ""
	if big.Sign(d.M) < 0 {
		sign = "-"
	}
	if d.S == 0 {
		return sign + digits
	}

	// pad, so that the whole part has at least one digit
	if missing := d.S + 1 - text.RuneCount(digits); missing > 0 {
		digits = text.Repeat("0", missing) + digits
	}
	chars := text.Chars(digits)
	split := chars.Len() - d.S
	whole := list.New[string]()
	fraction := list.New[string]()
	for i, c := range chars.All() {
		if i < split {
			whole.Append(c)
		} else {
			fraction.Append(c)
		}
	}
	return sign + text.Join(whole, "") + "." + text.Join(fraction, "")
}

// norm returns d with its scale clamped into 0..MaxScale.
func norm(d Dec) Dec {
	return Dec{M: d.M, S: min(max(d.S, 0), MaxScale)}
}

// rescale returns d with the greater scale s, which is exact.
func rescale(d Dec, s int32) Dec {
	if s <= d.S {
		return d
	}
	return Dec{M: big.Mul(d.M, big.Pow10(s-d.S)), S: s}
}

// Add returns a + b with the greater scale of both.
func Add(a, b Dec) Dec {
	a, b = norm(a), norm(b)
	s := max(a.S, b.S)
	return Dec{M: big.Add(rescale(a, s).M, rescale(b, s).M), S: s}
}

// Sub returns a - b with the greater scale of both.
func Sub(a, b Dec) Dec {
	a, b = norm(a), norm(b)
	s := max(a.S, b.S)
	return Dec{M: big.Sub(rescale(a, s).M, rescale(b, s).M), S: s}
}

// Mul returns a * b with the sum of both scales, which is exact up to [MaxScale].
func Mul(a, b Dec) Dec {
	a, b = norm(a), norm(b)
	p := Dec{M: big.Mul(a.M, b.M), S: a.S + b.S}
	if p.S > MaxScale {
		return round(p, MaxScale, HalfEven)
	}
	return p
}

// Neg returns -d.
func Neg(d Dec) Dec {
	d = norm(d)
	return Dec{M: big.Neg(d.M), S: d.S}
}

// Abs returns the absolute value of d.
func Abs(d Dec) Dec {
	d = norm(d)
	return Dec{M: big.Abs(d.M), S: d.S}
}

// Cmp returns -1, if a < b, 0, if a and b are equal, also with different scales like 1.0 and 1.00, and +1, if
// a > b.
func Cmp(a, b Dec) int32 {
	a, b = norm(a), norm(b)
	s := max(a.S, b.S)
	return big.Cmp(rescale(a, s).M, rescale(b, s).M)
}

// Sign returns -1, 0 or +1.
func Sign(d Dec) int32 {
	return big.Sign(d.M)
}

// Round returns d rounded to the scale s, e.g. Round(12.345, 2, HalfUp) is 12.35. A greater scale adds zeros. Round
// clamps s into 0..MaxScale.
func Round(d Dec, s int32, mode Rounding) Dec {
	return round(norm(d), min(max(s, 0), MaxScale), mode)
}

// round rounds d with a scale up to 2*MaxScale to s from 0 to MaxScale.
func round(d Dec, s int32, mode Rounding) Dec {
	if s >= d.S {
		return rescale(d, s)
	}
	m, _ := divRound(d.M, big.Pow10(d.S-s), mode)
	return Dec{M: m, S: s}
}

// Quo returns a / b rounded to the scale s and false, if b is zero. Quo clamps s into 0..MaxScale.
func Quo(a, b Dec, s int32, mode Rounding) (Dec, bool) {
	a, b = norm(a), norm(b)
	s = min(max(s, 0), MaxScale)
	n := a.M
	d := b.M
	// a.M/10^a.S / (b.M/10^b.S) * 10^s = a.M * 10^(s+b.S-a.S) / b.M
	if e := s + b.S - a.S; e >= 0 {
		n = big.Mul(n, big.Pow10(e))
	} else {
		d = big.Mul(d, big.Pow10(0-e))
	}
	m, ok := divRound(n, d, mode)
	if !ok {
		return Zero(), false
	}
	return Dec{M: m, S: s}, true
}

// divRound returns n / d rounded by the mode and false, if d is zero.
func divRound(n, d big.Int, mode Rounding) (big.Int, bool) {
	q, r, ok := big.QuoRem(n, d)
	if !ok {
		return big.Zero(), false
	}
	if big.Sign(r) == 0 {
		return q, true
	}

	// the quotient was truncated towards zero, away from zero means one more step in its direction
	negative := big.Sign(n)*big.Sign(d) < 0
	away := false
	switch mode {
	case Down:
		away = false
	case Up:
		away = true
	case Floor:
		away = negative
	case Ceiling:
		away = !negative
	case HalfUp, HalfEven:
		c := big.Cmp(big.Mul(big.Abs(r), big.FromInt32(2)), big.Abs(d))
		switch {
		case c > 0:
			away = true
		case c < 0:
			away = false
		case mode == HalfUp:
			away = true
		default:
			odd, ok := big.Rem(q, big.FromInt32(2))
			away = ok && big.Sign(odd) != 0
		}
	}

	if !away {
		return q, true
	}
	if negative {
		return big.Sub(q, big.FromInt32(1)), true
	}
	return big.Add(q, big.FromInt32(1)), true
}

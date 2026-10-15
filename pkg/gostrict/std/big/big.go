// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package big provides integers of any size, e.g. for amounts of money in cents, which exceed int32, see also
// package decimal. An [Int] never changes: each function returns a new one. In JavaScript, an Int is a BigInt, in Go
// it is a math/big.Int. Compare two of them with [Cmp], because == is not allowed for them.
//
// No function panics: a division by zero and a conversion out of range report it by their result.
//
//gostrict:header
package big

import (
	"math"
	"math/big"
	"regexp"

	"go.wdy.de/nago/pkg/gostrict/std/js"
)

// Int is an integer of any size. Create one with [Zero], [FromInt32] or [Parse].
type Int struct {
	ref js.Ref
}

func wrap(v *big.Int) Int {
	return Int{ref: js.NewRef(v)}
}

func val(i Int) *big.Int {
	if v, ok := i.ref.Value().(*big.Int); ok {
		return v
	}
	return new(big.Int)
}

// Zero returns 0.
//
//gostrict:js () => 0n
func Zero() Int {
	return wrap(new(big.Int))
}

// FromInt32 returns n.
//
//gostrict:js (n) => BigInt(n)
func FromInt32(n int32) Int {
	return wrap(big.NewInt(int64(n)))
}

// FromUint32 returns n.
//
//gostrict:js (n) => BigInt(n)
func FromUint32(n uint32) Int {
	return wrap(new(big.Int).SetUint64(uint64(n)))
}

var digits = regexp.MustCompile(`^[+-]?[0-9]+$`)

// Parse parses a decimal integer with an optional sign, e.g. "-1234". It returns false for anything else.
//
//gostrict:js (s) => /^[+-]?[0-9]+$/.test(s) ? [BigInt(s), true] : [0n, false]
func Parse(s string) (Int, bool) {
	if !digits.MatchString(s) {
		return Zero(), false
	}
	v, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return Zero(), false
	}
	return wrap(v), true
}

// String formats a in decimal, e.g. "-1234".
//
//gostrict:js (a) => a.toString()
func String(a Int) string {
	return val(a).String()
}

// Add returns a + b.
//
//gostrict:js (a, b) => a + b
func Add(a, b Int) Int {
	return wrap(new(big.Int).Add(val(a), val(b)))
}

// Sub returns a - b.
//
//gostrict:js (a, b) => a - b
func Sub(a, b Int) Int {
	return wrap(new(big.Int).Sub(val(a), val(b)))
}

// Mul returns a * b.
//
//gostrict:js (a, b) => a * b
func Mul(a, b Int) Int {
	return wrap(new(big.Int).Mul(val(a), val(b)))
}

// Quo returns a / b truncated towards zero, like the division of Go, and false, if b is zero.
//
//gostrict:js (a, b) => b === 0n ? [0n, false] : [a / b, true]
func Quo(a, b Int) (Int, bool) {
	if val(b).Sign() == 0 {
		return Zero(), false
	}
	return wrap(new(big.Int).Quo(val(a), val(b))), true
}

// Rem returns the remainder of a / b, which has the sign of a, like % of Go, and false, if b is zero.
//
//gostrict:js (a, b) => b === 0n ? [0n, false] : [a % b, true]
func Rem(a, b Int) (Int, bool) {
	if val(b).Sign() == 0 {
		return Zero(), false
	}
	return wrap(new(big.Int).Rem(val(a), val(b))), true
}

// QuoRem returns the quotient and the remainder of a / b like [Quo] and [Rem] and false, if b is zero.
//
//gostrict:js (a, b) => b === 0n ? [0n, 0n, false] : [a / b, a % b, true]
func QuoRem(a, b Int) (Int, Int, bool) {
	if val(b).Sign() == 0 {
		return Zero(), Zero(), false
	}
	q, r := new(big.Int).QuoRem(val(a), val(b), new(big.Int))
	return wrap(q), wrap(r), true
}

// Neg returns -a.
//
//gostrict:js (a) => -a
func Neg(a Int) Int {
	return wrap(new(big.Int).Neg(val(a)))
}

// Abs returns the absolute value of a.
//
//gostrict:js (a) => (a < 0n ? -a : a)
func Abs(a Int) Int {
	return wrap(new(big.Int).Abs(val(a)))
}

// Cmp returns -1, if a < b, 0, if a == b, and +1, if a > b.
//
//gostrict:js (a, b) => (a < b ? -1 : a > b ? 1 : 0)
func Cmp(a, b Int) int32 {
	return int32(val(a).Cmp(val(b)))
}

// Sign returns -1, 0 or +1.
//
//gostrict:js (a) => (a < 0n ? -1 : a > 0n ? 1 : 0)
func Sign(a Int) int32 {
	return int32(val(a).Sign())
}

// MaxPow10 is the greatest exponent of [Pow10].
const MaxPow10 = 4096

// Pow10 returns 10^n for n from 0 to [MaxPow10] and clamps any other n into this range.
//
//gostrict:js (n) => 10n ** BigInt(Math.min(Math.max(n, 0), 4096))
func Pow10(n int32) Int {
	return wrap(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(min(max(n, 0), MaxPow10))), nil))
}

// Int32 returns a as int32 and false, if it is out of the range of int32.
//
//gostrict:js (a) => (a >= -2147483648n && a <= 2147483647n ? [Number(a), true] : [0, false])
func Int32(a Int) (int32, bool) {
	v := val(a)
	if !v.IsInt64() || v.Int64() < math.MinInt32 || v.Int64() > math.MaxInt32 {
		return 0, false
	}
	return int32(v.Int64()), true
}

// Float64 returns the nearest float64 of a, ±Inf beyond its range.
//
// The binding parses the digits beyond the range of int64, because goja converts a BigInt through int64.
//
//gostrict:js (a) => (a >= -9223372036854775808n && a < 9223372036854775808n ? Number(a) : parseFloat(a.toString()))
func Float64(a Int) float64 {
	f, _ := new(big.Float).SetInt(val(a)).Float64()
	return f
}

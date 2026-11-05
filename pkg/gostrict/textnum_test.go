// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/rand/v2"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"go.wdy.de/nago/pkg/gostrict"
)

// The tests of text.ParseFloat, text.Slice and text.GroupDigits compare characters and the bits of floats, which the
// differential test does not see, in goja and in the vm (this file) and natively (textnum_native_test.go). The
// expectations come from the contract R-SCRIPT-GS-TEXT-NUM, not from an engine. The functions are called from a
// script, which the engines translate at run time, so that this file compiles also without them.

const textNumSrc = `package s

import (
	"math"

	"go.wdy.de/nago/pkg/gostrict/std/text"
)

// PF returns the value, its sign bit, as the program sees it, the message of the error and whether there is one.
func PF(s string) (float64, bool, string, bool) {
	f, err := text.ParseFloat(s)
	if err != nil {
		return f, math.Signbit(f), err.Error(), true
	}
	return f, math.Signbit(f), "", false
}

func Slice(s string, from, to int32) string { return text.Slice(s, from, to) }

func Group(s, decimalMark, groupMark string) string { return text.GroupDigits(s, decimalMark, groupMark) }

func GroupFloat(f float64, decimals int32, decimalMark, groupMark string) string {
	return text.GroupDigits(text.FromFloat(f, decimals), decimalMark, groupMark)
}

func CharAt(s string, i int32) string {
	c, _ := text.Chars(s).Get(i)
	return c
}

func Count(s string) int32 { return text.RuneCount(s) }
`

type textNumEngine struct {
	name string
	in   gostrict.Instance
}

// loadTextNum translates the script and loads it into goja and the vm.
func loadTextNum(t *testing.T, src string) []textNumEngine {
	t.Helper()
	b, diags, err := gostrict.NewCache(1).Compile("llm.local/s", script(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) > 0 {
		t.Fatalf("the script does not compile:\n%s", diags.Text())
	}
	var res []textNumEngine
	for _, e := range []struct {
		name string
		e    gostrict.Engine
	}{{"goja", goEngine}, {"vm", vmEngine}} {
		in, err := e.e.Load(context.Background(), b, nil)
		if err != nil {
			t.Fatalf("%s: %v", e.name, err)
		}
		t.Cleanup(func() { in.Close() })
		res = append(res, textNumEngine{e.name, in})
	}
	return res
}

func (e textNumEngine) call(t *testing.T, name string, args ...any) []any {
	t.Helper()
	res, err := e.in.Call(context.Background(), name, args...)
	if err != nil {
		t.Fatalf("%s: %s(%s): %v", e.name, name, shortArgs(args), err)
	}
	return res
}

// parseFloat calls PF and returns the value, the sign bit in the program, the message and whether it failed.
func (e textNumEngine) parseFloat(t *testing.T, s string) (float64, bool, string, bool) {
	t.Helper()
	res := e.call(t, "PF", s)
	return res[0].(float64), res[1].(bool), res[2].(string), res[3].(bool)
}

// chars returns s with each byte of invalid UTF-8 replaced by "�", like text.Chars counts it, so that two strings
// are equal, if they have the same characters.
func chars(s string) string {
	return string([]rune(s))
}

// abbrev shortens a long input for a message or the name of a subtest.
func abbrev(s string) string {
	if len(s) <= 48 {
		return strconv.Quote(s)
	}
	return fmt.Sprintf("%q…(%d bytes)", s[:40], len(s))
}

func shortArgs(args []any) string {
	var parts []string
	for _, a := range args {
		if s, ok := a.(string); ok {
			parts = append(parts, abbrev(s))
		} else {
			parts = append(parts, fmt.Sprint(a))
		}
	}
	return strings.Join(parts, ", ")
}

// ParseFloat

const (
	pfSyntax = "invalid syntax"
	pfRange  = "value out of range"
	negZero  = 0x8000000000000000
	// the float64 0.1 and its successor
	tenthBits     = 0x3FB999999999999A
	tenthNextBits = 0x3FB999999999999B
)

// contractHalf is the literal of the contract table for the middle between 0 and the least subnormal number. It
// has 75 significant digits and is rounded up in the last one, so it lies slightly above the middle, see
// TestParseFloatPremises.
const contractHalf = "2.4703282292062327208828439643411068618252990130716238221279284125033775364e-324"

// tenth40Up and tenth40Down have 40 decimals near 0.1, whose last one decides: the middle between 0.1 and its
// successor lies between them, and the first 20 digits of both lie below it, see TestParseFloatPremises.
const (
	tenth40Up   = "0.1000000000000000124900090270330110797659"
	tenth40Down = "0.1000000000000000124900090270330110797658"
)

// onePlusHalfUlp is exactly 1 + 2^-53, the middle between 1 and its successor.
const onePlusHalfUlp = "1.00000000000000011102230246251565404236316680908203125"

// pfForm is the form of ParseFloat in the contract.
var pfForm = regexp.MustCompile(`^[+-]?(\d+(\.\d*)?|\.\d+)([eE][+-]?\d+)?$`)

type pfRow struct {
	in   string
	bits uint64 // if err is ""
	err  string // "", pfSyntax or pfRange
}

func (r pfRow) message() string {
	if r.err == "" {
		return ""
	}
	return "text.ParseFloat: " + r.err + ": " + r.in
}

func pfOK(s string, f float64) pfRow   { return pfRow{in: s, bits: math.Float64bits(f)} }
func pfBits(s string, b uint64) pfRow  { return pfRow{in: s, bits: b} }
func pfErr(s string, err string) pfRow { return pfRow{in: s, err: err} }
func pfSyn(s ...string) (rows []pfRow) { return pfErrs(pfSyntax, s) }

func pfErrs(err string, s []string) (rows []pfRow) {
	for _, in := range s {
		rows = append(rows, pfErr(in, err))
	}
	return rows
}

// halfMinSubnormal is the exact decimal of 2^-1075, the middle between 0 and the least subnormal number 2^-1074,
// with all its 752 significant digits: 0.000…000247032822920623272088284396434110686182529901307162382212792841250337753635104…125.
func halfMinSubnormal() string {
	half := new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Lsh(big.NewInt(1), 1075))
	return half.FloatString(1075)
}

// maxHalf is the middle between the largest float64 and 2^1024, which a tie rounds to, because the mantissa of the
// largest float64 is odd.
func maxHalf() *big.Int {
	return new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 1024), new(big.Int).Lsh(big.NewInt(1), 970))
}

// parseFloatRows are the rows of the contract and the limits around them.
func parseFloatRows() []pfRow {
	half := halfMinSubnormal()
	sig := strings.TrimLeft(half[2:], "0")
	halfE := sig[:1] + "." + sig[1:] + "e-324"
	mh := maxHalf().String()
	mhE := mh[:1] + "." + mh[1:] + "e308"
	below := new(big.Int).Sub(maxHalf(), big.NewInt(1)).String()
	nines := strings.Repeat("9", 400)

	rows := []pfRow{
		// the rows of the contract
		pfOK("1.5", 1.5), pfOK("-.5", -0.5), pfOK("5.", 5), pfOK("1e3", 1000), pfOK("+2E-2", 0.02), pfOK("007", 7),
		pfBits("-0", negZero), pfBits("-1e-400", negZero),
	}
	rows = append(rows, pfSyn("1,5", " 1", "1_0", "Inf", "NaN", "0x1", "", ".", "1e", "+")...)
	rows = append(rows,
		pfErr("1e309", pfRange), pfErr("1e99999999999999999999", pfRange),
		pfBits("0e99999999999", 0), pfBits("1e-400", 0),

		// the exact middle between 0 and the least subnormal number is a tie, which goes to the even 0, and one more
		// digit 1 of the mantissa lies above it
		pfBits(half, 0), pfBits(half+"1", 1), pfBits(halfE, 0), pfBits(sig[:1]+"."+sig[1:]+"1e-324", 1),
		pfBits("-"+half, negZero), pfBits("-"+half+"1", negZero|1),
		// the literal of the contract table lies above the middle, so the nearest float64 is the least subnormal
		pfBits(contractHalf, 1),
		pfBits(strings.Replace(contractHalf, "75364e", "753635e", 1), 0),

		// 40 decimals of 0.1, whose last one decides; the first 20 digits of both give 0.1
		pfBits(tenth40Up, tenthNextBits), pfBits(tenth40Down, tenthBits),
		pfBits("0.1", tenthBits), pfBits("-"+tenth40Up, negZero|tenthNextBits),
		pfBits(tenth40Up+"e0", tenthNextBits), pfBits(tenth40Up[2:]+"e-40", tenthNextBits),

		// ties to even beyond 15 digits
		pfBits(onePlusHalfUlp, 0x3FF0000000000000), pfBits(onePlusHalfUlp+"000000000001", 0x3FF0000000000001),
		pfBits(strings.TrimSuffix(onePlusHalfUlp, "5")+"4999999999999", 0x3FF0000000000000),
		pfOK("9007199254740993", 9007199254740992), pfOK("9007199254740995", 9007199254740996),
		pfOK("9007199254740993.00000000000000000001", 9007199254740994),
		pfOK("123456789012345678901234567890", 123456789012345678901234567890),

		// the largest float64 and the middle to 2^1024, which rounds to infinity
		pfBits("1.7976931348623157e308", 0x7FEFFFFFFFFFFFFF), pfBits(below, 0x7FEFFFFFFFFFFFFF),
		pfErr(mh, pfRange), pfErr(mhE, pfRange), pfErr("-"+mh, pfRange), pfErr("-1e309", pfRange),
		pfBits("-1.7976931348623157e308", 0xFFEFFFFFFFFFFFFF), pfOK("1e308", 1e308),

		// the least normal and subnormal numbers
		pfBits("2.2250738585072014e-308", 0x0010000000000000), pfBits("2.2250738585072012e-308", 0x0010000000000000),
		pfBits("2.2250738585072011e-308", 0x000FFFFFFFFFFFFF), pfBits("4.9406564584124654e-324", 1),
		pfBits("5e-324", 1), pfBits("1e-323", 2), pfBits("3e-324", 1), pfBits("2e-324", 0),

		// huge exponents, also beyond the range of int32 and of the exact integers of a double
		pfBits("1e-99999999999999999999", 0), pfBits("-1e-99999999999999999999", negZero),
		pfBits("0e-99999999999999999999", 0), pfErr("1e"+nines, pfRange), pfBits("1e-"+nines, 0),
		pfBits("0e"+nines, 0), pfBits("-0e"+nines, negZero), pfErr("-0.5e"+nines, pfRange),
		pfOK("0."+strings.Repeat("0", 400)+"1e401", 1), pfOK("1"+strings.Repeat("0", 400)+"e-400", 1),
		pfOK("1e00000000000000000000000000000000000000308", 1e308),
		pfErr("1e4294967296", pfRange), pfBits("1e-4294967296", 0), pfErr("1e4294967297", pfRange),
		pfErr("1e2147483648", pfRange), pfBits("1e-2147483649", 0), pfErr("1e9007199254740993", pfRange),
		pfBits("1e-9007199254740993", 0),

		// the other forms of the grammar
		pfOK("1.e5", 100000), pfOK("+.5e+1", 5), pfOK("-5.E-1", -0.5), pfOK("00000000000000000000000000000000000001", 1),
		pfBits("0.0", 0), pfBits("+0", 0), pfBits("-0.000", negZero), pfBits("-0e5", negZero), pfBits("-.0", negZero),
		pfOK("100e-2", 1), pfOK("0.000001e6", 1),
	)
	rows = append(rows, pfSyn("1e+", "1e-", "e5", ".e5", "+-1", "--1", "1.5.", "1..2", "1e5.5", "1 ", "\t1", "1\n",
		"１", "١", "infinity", "+Inf", "-Inf", "nan", "0b1", "1d", "1f", "1E+-5", "1\xff", "1e5e5", "5-",
		"1+1", "0x1p-2", "1_000.5", "-", "e", "E5", "1e 5", "1.5e+", "+.e1", strings.Repeat("1", 100)+"x")...)
	return rows
}

// checkParseFloat compares a result of ParseFloat with the row: the bits, the sign bit, which the program sees, and
// for an error its message by characters.
func checkParseFloat(t *testing.T, where string, r pfRow, f float64, sign bool, msg string, failed bool) {
	t.Helper()
	if r.err != "" {
		if !failed || chars(msg) != chars(r.message()) {
			t.Errorf("%s: ParseFloat(%s) = %v, %q, want the error %q", where, abbrev(r.in), f, msg, abbrev(r.message()))
		}
		return
	}
	if failed || math.Float64bits(f) != r.bits || sign != (r.bits>>63 == 1) {
		t.Errorf("%s: ParseFloat(%s) = %#016x (%v, sign bit %v), error %q, want %#016x (%v)", where, abbrev(r.in),
			math.Float64bits(f), f, sign, msg, r.bits, math.Float64frombits(r.bits))
	}
}

// TestParseFloatBits compares the bits and the messages of text.ParseFloat in goja and in the vm with the contract
// (L-N1-2).
func TestParseFloatBits(t *testing.T) {
	engines := loadTextNum(t, textNumSrc)
	for i, r := range parseFloatRows() {
		t.Run(fmt.Sprintf("%d %s", i, abbrev(r.in)), func(t *testing.T) {
			var values []uint64
			for _, e := range engines {
				f, sign, msg, failed := e.parseFloat(t, r.in)
				checkParseFloat(t, e.name, r, f, sign, msg, failed)
				values = append(values, math.Float64bits(f))
			}
			// the value beside an error is not in the contract, but the same in each engine
			if r.err != "" && values[0] != values[1] {
				t.Errorf("the value beside the error differs: goja %#016x, vm %#016x", values[0], values[1])
			}
		})
	}
}

// decimalParts splits a text of the form of ParseFloat into its sign, its significant digits without leading and
// trailing zeros and the exponent, so that its value is 0.digits * 10^exp. It clamps the exponent to ±10^6, beyond
// which any value is 0 or out of range.
func decimalParts(s string) (neg bool, digits string, exp int64) {
	if s != "" && (s[0] == '+' || s[0] == '-') {
		neg = s[0] == '-'
		s = s[1:]
	}
	mant, e, _ := strings.Cut(strings.ToLower(s), "e")
	if e != "" {
		n, ok := new(big.Int).SetString(e, 10)
		if !ok {
			panic("no exponent: " + s)
		}
		switch {
		case n.Cmp(big.NewInt(1e6)) > 0:
			exp = 1e6
		case n.Cmp(big.NewInt(-1e6)) < 0:
			exp = -1e6
		default:
			exp = n.Int64()
		}
	}
	ip, fp, _ := strings.Cut(mant, ".")
	all := ip + fp
	exp += int64(len(ip))
	trimmed := strings.TrimLeft(all, "0")
	exp -= int64(len(all) - len(trimmed))
	digits = strings.TrimRight(trimmed, "0")
	return neg, digits, exp
}

// sigDigits returns the number of significant digits of a text of the form of ParseFloat.
func sigDigits(s string) int {
	_, d, _ := decimalParts(s)
	return len(d)
}

// exactValue returns the absolute value of a text of the form of ParseFloat, or false, if its exponent is beyond
// ±10^5.
func exactValue(s string) (*big.Rat, bool) {
	_, d, e := decimalParts(s)
	if d == "" {
		return new(big.Rat), true
	}
	if e > 1e5 || e < -1e5 {
		return nil, false
	}
	m, _ := new(big.Int).SetString(d, 10)
	r := new(big.Rat).SetInt(m)
	p := new(big.Int).Exp(big.NewInt(10), big.NewInt(abs64(e-int64(len(d)))), nil)
	if e-int64(len(d)) >= 0 {
		return r.Mul(r, new(big.Rat).SetInt(p)), true
	}
	return r.Quo(r, new(big.Rat).SetInt(p)), true
}

func abs64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

// nearestPremise proves with exact arithmetic, that bits is the float64 nearest to s, ties to even, with the sign of
// s also for 0.
func nearestPremise(s string, bits uint64) error {
	neg, d, e := decimalParts(s)
	f := math.Float64frombits(bits)
	if math.Signbit(f) != neg {
		return fmt.Errorf("the sign of %#016x is not the one of the text", bits)
	}
	a := math.Abs(f)
	if math.IsInf(a, 0) || math.IsNaN(a) {
		return fmt.Errorf("%#016x is not finite", bits)
	}
	x, ok := exactValue(s)
	if !ok {
		// beyond any float64, so it can only be 0
		if e < 0 && a == 0 {
			return nil
		}
		return fmt.Errorf("the exponent %d of %d digits gives %v", e, len(d), a)
	}
	dist := func(g *big.Rat) *big.Rat {
		r := new(big.Rat).Sub(x, g)
		return r.Abs(r)
	}
	da := dist(new(big.Rat).SetFloat64(a))
	even := math.Float64bits(a)&1 == 0
	var neighbors []*big.Rat
	if a > 0 {
		neighbors = append(neighbors, new(big.Rat).SetFloat64(math.Nextafter(a, 0)))
	}
	if a == math.MaxFloat64 {
		neighbors = append(neighbors, new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), 1024)))
	} else {
		neighbors = append(neighbors, new(big.Rat).SetFloat64(math.Nextafter(a, math.Inf(1))))
	}
	for _, n := range neighbors {
		switch c := da.Cmp(dist(n)); {
		case c > 0:
			return fmt.Errorf("%v is nearer than %#016x", n.FloatString(30), bits)
		case c == 0 && !even:
			return fmt.Errorf("a tie between %#016x and %v goes to the even one", bits, n.FloatString(30))
		}
	}
	return nil
}

// rangePremise proves, that s rounds to infinity: its absolute value is at least the middle between the largest
// float64 and 2^1024.
func rangePremise(s string) error {
	_, d, e := decimalParts(s)
	x, ok := exactValue(s)
	if !ok {
		if e > 0 && d != "" {
			return nil
		}
		return fmt.Errorf("the exponent %d of %d digits is not out of range", e, len(d))
	}
	if x.Cmp(new(big.Rat).SetInt(maxHalf())) < 0 {
		return fmt.Errorf("%s is below the middle to 2^1024", abbrev(s))
	}
	return nil
}

// TestParseFloatPremises proves the expectations of the rows with exact arithmetic instead of an engine, and the
// premises of the constructed rows.
func TestParseFloatPremises(t *testing.T) {
	for _, r := range parseFloatRows() {
		inForm := pfForm.MatchString(r.in)
		var err error
		switch r.err {
		case "":
			err = nearestPremise(r.in, r.bits)
		case pfRange:
			err = rangePremise(r.in)
		}
		if err != nil {
			t.Errorf("ParseFloat(%s): %v", abbrev(r.in), err)
		}
		if inForm != (r.err != pfSyntax) {
			t.Errorf("ParseFloat(%s): the row expects %q, but the text is in the form %v", abbrev(r.in), r.err, inForm)
		}
	}

	// the exact middle between 0 and 2^-1074 has 752 significant digits, the literal of the contract table lies above it
	half, _ := exactValue(halfMinSubnormal())
	if want := new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Lsh(big.NewInt(1), 1075)); half.Cmp(want) != 0 {
		t.Fatalf("halfMinSubnormal is not 2^-1075")
	}
	if n := sigDigits(halfMinSubnormal()); n != 752 {
		t.Errorf("2^-1075 has %d significant digits", n)
	}
	lit, _ := exactValue(contractHalf)
	if lit.Cmp(half) <= 0 {
		t.Errorf("the literal of the contract does not lie above the middle")
	}

	// the middle between 0.1 and its successor lies between the two texts with 40 decimals, and the first 20
	// significant digits of the upper one lie below it, so they round to 0.1 and not to the successor
	mid := new(big.Rat).Add(new(big.Rat).SetFloat64(math.Float64frombits(tenthBits)), new(big.Rat).SetFloat64(math.Float64frombits(tenthNextBits)))
	mid.Quo(mid, big.NewRat(2, 1))
	up, _ := exactValue(tenth40Up)
	down, _ := exactValue(tenth40Down)
	if len(tenth40Up)-2 != 40 || len(tenth40Down)-2 != 40 || up.Cmp(mid) <= 0 || down.Cmp(mid) >= 0 {
		t.Errorf("the middle %s does not lie between %s and %s", mid.FloatString(60), tenth40Up, tenth40Down)
	}
	first20 := "0." + tenth40Up[2:2+20]
	if sigDigits(tenth40Up) != 40 || nearestPremise(first20, tenthBits) != nil || nearestPremise(tenth40Up, tenthNextBits) != nil {
		t.Errorf("the first 20 digits %s do not round differently than %s", first20, tenth40Up)
	}
}

// randomDecimals returns texts of the form of ParseFloat with up to 40 significant digits and random exponents, a
// half of them near the middle between two floats, where the digits after the 20th decide, also exact ties.
func randomDecimals(n int) []string {
	rnd := rand.New(rand.NewPCG(20261105, 9))
	signs := []string{"", "", "+", "-"}
	digits := func(k int) string {
		b := make([]byte, k)
		for i := range b {
			b[i] = byte('0' + rnd.IntN(10))
		}
		return string(b)
	}
	var res []string
	for len(res) < n {
		sign := signs[rnd.IntN(len(signs))]
		e := "e"
		if rnd.IntN(2) == 0 {
			e = "E"
		}
		if rnd.IntN(2) == 0 {
			// any digits with the point anywhere and an exponent, which moves them over the whole range
			k := 1 + rnd.IntN(40)
			d := digits(k)
			if rnd.IntN(3) == 0 {
				d = strings.Repeat("0", rnd.IntN(5)) + d
			}
			s := d
			switch p := rnd.IntN(len(d) + 2); {
			case p == 0:
				s = "." + d
			case p <= len(d):
				s = d[:p] + "." + d[p:]
			}
			if rnd.IntN(4) > 0 {
				x := rnd.IntN(700) - 360
				es := strconv.Itoa(x)
				if x >= 0 && rnd.IntN(3) == 0 {
					es = "+" + es
				}
				s += e + es
			}
			res = append(res, sign+s)
			continue
		}

		// near the middle between a float64 and its successor: its exact digits, cut after 16 to 40 digits and
		// either kept, which lies below it, or increased by one, which lies above it, or the middle itself, if it
		// has few digits
		var f float64
		switch rnd.IntN(3) {
		case 0:
			f = math.Float64frombits(rnd.Uint64() &^ (1 << 63))
		case 1:
			f = math.Ldexp(1+rnd.Float64(), rnd.IntN(60)+20)
		default:
			f = math.Float64frombits(rnd.Uint64() & 0x000FFFFFFFFFFFFF)
		}
		next := math.Nextafter(f, math.Inf(1))
		if math.IsInf(f, 0) || math.IsNaN(f) || math.IsInf(next, 0) {
			continue
		}
		mid := new(big.Rat).Add(new(big.Rat).SetFloat64(f), new(big.Rat).SetFloat64(next))
		mid.Quo(mid, big.NewRat(2, 1))
		txt := new(big.Float).SetPrec(4096).SetRat(mid).Text('e', 900)
		mant, ex, _ := strings.Cut(txt, "e")
		md := strings.TrimRight(strings.Replace(mant, ".", "", 1), "0")
		x, _ := strconv.Atoi(ex)
		k := 16 + rnd.IntN(25)
		if len(md) > k {
			md = md[:k]
			if rnd.IntN(2) == 0 {
				m, _ := new(big.Int).SetString(md, 10)
				md = m.Add(m, big.NewInt(1)).String()
				if len(md) > k {
					// 99…9 + 1 has one more digit
					x++
				}
			}
		}
		res = append(res, sign+md[:1]+"."+md[1:]+e+strconv.Itoa(x))
	}
	return res
}

// randomForms returns short random texts of the characters of the form, most of them not in it.
func randomForms(n int) []string {
	rnd := rand.New(rand.NewPCG(20261105, 10))
	const alphabet = "01234567890123456789..eeEE++--"
	res := make([]string, n)
	for i := range res {
		b := make([]byte, rnd.IntN(9))
		for j := range b {
			b[j] = alphabet[rnd.IntN(len(alphabet))]
		}
		res[i] = string(b)
	}
	return res
}

// strconvExpect returns the bits, which strconv.ParseFloat of Go gives as the reference, or the message of
// text.ParseFloat for its error.
func strconvExpect(s string) (uint64, string) {
	f, err := strconv.ParseFloat(s, 64)
	var ne *strconv.NumError
	switch {
	case err == nil:
		return math.Float64bits(f), ""
	case errors.As(err, &ne) && errors.Is(ne.Err, strconv.ErrRange):
		return 0, "text.ParseFloat: value out of range: " + s
	default:
		return 0, "text.ParseFloat: invalid syntax: " + s
	}
}

// randomInputs are the random inputs of the tests, which compare with strconv.ParseFloat.
func randomInputs() []string {
	return append(randomDecimals(20000), randomForms(5000)...)
}

// TestParseFloatRandomPremises checks, that the random inputs contain the cases, which they are for.
func TestParseFloatRandomPremises(t *testing.T) {
	var long, first20, ties, subnormal, overflow, syntax, form int
	for _, s := range randomInputs() {
		if !pfForm.MatchString(s) {
			syntax++
			continue
		}
		form++
		bits, msg := strconvExpect(s)
		if msg != "" {
			overflow++
			continue
		}
		if f := math.Abs(math.Float64frombits(bits)); f != 0 && f < 0x1p-1022 {
			subnormal++
		}
		_, d, e := decimalParts(s)
		if len(d) > 15 {
			long++
		}
		if len(d) > 20 {
			if b20, _ := strconvExpect(signOf(s) + "0." + d[:20] + "e" + strconv.FormatInt(e, 10)); b20 != bits {
				first20++
			}
		}
		if len(d) <= 40 {
			x, ok := exactValue(s)
			f := math.Abs(math.Float64frombits(bits))
			if ok && f < math.MaxFloat64 {
				m := new(big.Rat).Add(new(big.Rat).SetFloat64(f), new(big.Rat).SetFloat64(math.Nextafter(f, math.Inf(1))))
				m.Quo(m, big.NewRat(2, 1))
				if x.Cmp(m) == 0 {
					ties++
				}
				if f > 0 {
					m.Add(new(big.Rat).SetFloat64(f), new(big.Rat).SetFloat64(math.Nextafter(f, 0)))
					m.Quo(m, big.NewRat(2, 1))
					if x.Cmp(m) == 0 {
						ties++
					}
				}
			}
		}
	}
	t.Logf("in the form %d, not %d; more than 15 digits %d, the first 20 digits round differently %d, ties %d, subnormal %d, out of range %d",
		form, syntax, long, first20, ties, subnormal, overflow)
	if long < 8000 || first20 < 1000 || ties < 500 || subnormal < 1000 || overflow < 100 || syntax < 2000 || form < 1000 {
		t.Fatal("the random inputs do not contain enough of the cases, which they are for")
	}
}

func signOf(s string) string {
	if strings.HasPrefix(s, "-") {
		return "-"
	}
	return ""
}

// TestParseFloatRandom compares text.ParseFloat in goja and in the vm with strconv.ParseFloat of Go as the
// reference.
func TestParseFloatRandom(t *testing.T) {
	engines := loadTextNum(t, textNumSrc)
	ins := randomInputs()
	bad := 0
	for _, s := range ins {
		bits, msg := strconvExpect(s)
		for _, e := range engines {
			f, sign, got, failed := e.parseFloat(t, s)
			var ok bool
			if msg != "" {
				ok = failed && got == msg
			} else {
				ok = !failed && math.Float64bits(f) == bits && sign == (bits>>63 == 1)
			}
			if !ok {
				bad++
				if bad <= 20 {
					t.Errorf("%s: ParseFloat(%s) = %#016x, %q, strconv %#016x, %q", e.name, abbrev(s), math.Float64bits(f), got, bits, msg)
				}
			}
		}
	}
	if bad > 0 {
		t.Errorf("%d of %d calls differ from strconv.ParseFloat", bad, len(ins)*len(engines))
	}
}

// Slice

type sliceRow struct {
	s        string
	from, to int32
	want     string
}

var sliceRows = []sliceRow{
	// the rows of the contract
	{"Grüße", 1, 3, "rü"}, {"😀ab", 0, 1, "😀"}, {"abc", -5, 99, "abc"}, {"abc", 2, 1, ""}, {"M\xfcller", 0, 2, "M�"},
	// from below 0 counts as 0, not from the end like slice of JavaScript, and to at the length and one more
	{"abc", -1, 3, "abc"}, {"abc", 0, -1, ""}, {"abc", -2, -1, ""}, {"abc", -1, 1, "a"}, {"abc", 0, 3, "abc"},
	{"abc", 0, 4, "abc"}, {"abc", 0, 2, "ab"}, {"abc", 1, 1, ""}, {"abc", 1, 2, "b"}, {"abc", 2, 3, "c"},
	{"abc", 3, 4, ""}, {"abc", 3, 3, ""}, {"abc", 4, 5, ""}, {"", 0, 1, ""}, {"", -1, 0, ""},
	{"abc", math.MinInt32, math.MaxInt32, "abc"}, {"abc", math.MaxInt32, math.MinInt32, ""},
	{"abc", math.MaxInt32, math.MaxInt32, ""}, {"abc", math.MinInt32, math.MinInt32 + 1, ""},
	// characters, not units of UTF-16 or bytes
	{"a😀b", 1, 2, "😀"}, {"a😀b", 2, 3, "b"}, {"a😀b", 0, 2, "a😀"}, {"😀😀", 1, 2, "😀"}, {"Grüße", 3, 5, "ße"},
	{"x\U0010FFFFy", 1, 3, "\U0010FFFFy"}, {"é", 0, 1, "e"}, {"é", 1, 2, "́"},
	// each byte of invalid UTF-8 is a "�" of its own, like in text.Chars
	{"M\xfcller", 1, 2, "�"}, {"M\xfcller", 2, 6, "ller"}, {"M\xfcller", 0, 6, "M�ller"},
	{"a\xe2\x82b", 1, 3, "��"}, {"a\xe2\x82b", 3, 4, "b"}, {"\xed\xa0\x80", 2, 3, "�"},
	{"\xf0\x9f\x98", 0, 9, "���"}, {"\xc0\x80x", 2, 3, "x"}, {"�\xff", 0, 2, "��"},
}

// charCounts are strings with the number of their characters by the contract, also of invalid UTF-8.
var charCounts = []struct {
	s string
	n int32
}{
	{"Grüße", 5}, {"😀ab", 3}, {"M\xfcller", 6}, {"a\xe2\x82b", 4}, {"\xed\xa0\x80", 3}, {"\xf0\x9f\x98", 3},
	{"é", 2}, {"", 0}, {"x\U0010FFFFy", 3}, {"�\xff", 2}, {"\xc0\x80x", 3},
}

// TestSliceChars compares text.Slice in goja and in the vm with the contract (L-N1-3).
func TestSliceChars(t *testing.T) {
	engines := loadTextNum(t, textNumSrc)
	for _, r := range sliceRows {
		for _, e := range engines {
			if got := e.call(t, "Slice", r.s, r.from, r.to)[0].(string); got != r.want {
				t.Errorf("%s: Slice(%q, %d, %d) = %q, want %q", e.name, r.s, r.from, r.to, got, r.want)
			}
		}
	}

	// Slice counts like Chars: the character at i is Slice(s, i, i+1), and all of them are Slice(s, 0, n)
	for _, c := range charCounts {
		for _, e := range engines {
			n := e.call(t, "Count", c.s)[0].(int32)
			if n != c.n {
				t.Errorf("%s: RuneCount(%q) = %d, want %d", e.name, c.s, n, c.n)
			}
			all := ""
			for i := int32(-1); i <= c.n; i++ {
				at := e.call(t, "CharAt", c.s, i)[0].(string)
				if got := e.call(t, "Slice", c.s, i, i+1)[0].(string); got != at {
					t.Errorf("%s: Slice(%q, %d, %d) = %q, Chars has %q", e.name, c.s, i, i+1, got, at)
				}
				all += at
			}
			if got := e.call(t, "Slice", c.s, int32(0), c.n)[0].(string); got != all {
				t.Errorf("%s: Slice(%q, 0, %d) = %q, the characters are %q", e.name, c.s, c.n, got, all)
			}
		}
	}
}

// GroupDigits

type groupRow struct {
	s, decimalMark, groupMark string
	want                      string
}

func groupRows() []groupRow {
	rows := []groupRow{
		// the rows of the contract
		{"1234567.89", ",", ".", "1.234.567,89"}, {"-1234", ",", ".", "-1.234"}, {"999.00", ",", ".", "999,00"},
		{"1234.5", ",", "", "1234,5"}, {"NaN", ",", ".", "NaN"}, {"1e5", ",", ".", "1e5"}, {"12a", ",", ".", "12a"},
		// groups of three from the right, the sign in front, the decimals not grouped
		{"0", ",", ".", "0"}, {"123", ",", ".", "123"}, {"1234", ",", ".", "1.234"}, {"12345", ",", ".", "12.345"},
		{"123456", ",", ".", "123.456"}, {"1234567", ",", ".", "1.234.567"}, {"-123", ",", ".", "-123"},
		{"-123456", ",", ".", "-123.456"}, {"+1234", ",", ".", "+1.234"}, {"+123456", ",", ".", "+123.456"},
		{"1234.5678", ",", ".", "1.234,5678"}, {"0001234", ",", ".", "0.001.234"}, {"-0.00", ",", ".", "-0,00"},
		{"1234567.5", ".", ",", "1,234,567.5"}, {"1234.5", "", "", "12345"}, {"1234.5", "", ".", "1.2345"},
		// marks of more characters and with the patterns of replace of JavaScript are taken as they are
		{"1234567.5", "<>", "'", "1'234'567<>5"}, {"1234567", ",", "$&", "1$&234$&567"}, {"1.5", "$&", ".", "1$&5"},
		{"1234.5", "$1", "$2", "1$2234$15"}, {"1234567", ",", "$$", "1$$234$$567"}, {"1234.5", "$'", "$`", "1$`234$'5"},
		{"1234", ",", "😀", "1😀234"}, {"1234.5", " ", " ", "1 234 5"},
		// a long number, like decimal.String gives it
		{"1" + strings.Repeat("000", 133), ",", ".", "1" + strings.Repeat(".000", 133)},
		{"-" + strings.Repeat("9", 400) + ".5", ",", ".", "-9" + strings.Repeat(".999", 133) + ",5"},
	}
	// anything else comes back as it is
	for _, s := range []string{"", "-", "+", ".5", "5.", "1.", "1.2.3", "1,234", " 1234", "1234 ", "+Inf", "-Inf",
		"Inf", "--1", "+-1", "١٢٣٤", "１２３４", "1_000", "0x10", "-1e5",
		"1234\n", "12\xff34", "1.5e3", "١٢٣٤.5", "1234.", "-.5"} {
		rows = append(rows, groupRow{s, ",", ".", s})
	}
	return rows
}

// groupFloatRows format with text.FromFloat and group the result.
var groupFloatRows = []struct {
	f        float64
	decimals int32
	want     string
}{
	{-0.001, 2, "-0,00"}, {1234567.891, 2, "1.234.567,89"}, {math.NaN(), 2, "NaN"}, {math.Inf(1), 2, "+Inf"},
	{math.Inf(-1), 0, "-Inf"}, {1e21, 0, "1.000.000.000.000.000.000.000"}, {-1234.5, 0, "-1.234"}, {0.125, 2, "0,12"},
}

// TestGroupDigitsChars compares text.GroupDigits in goja and in the vm with the contract (L-N1-4).
func TestGroupDigitsChars(t *testing.T) {
	engines := loadTextNum(t, textNumSrc)
	for _, r := range groupRows() {
		for _, e := range engines {
			if got := e.call(t, "Group", r.s, r.decimalMark, r.groupMark)[0].(string); chars(got) != chars(r.want) {
				t.Errorf("%s: GroupDigits(%s, %q, %q) = %s, want %s", e.name, abbrev(r.s), r.decimalMark, r.groupMark, abbrev(got), abbrev(r.want))
			}
		}
	}
	for _, r := range groupFloatRows {
		for _, e := range engines {
			if got := e.call(t, "GroupFloat", r.f, r.decimals, ",", ".")[0].(string); got != r.want {
				t.Errorf("%s: GroupDigits(FromFloat(%v, %d), \",\", \".\") = %q, want %q", e.name, r.f, r.decimals, got, r.want)
			}
		}
	}
}

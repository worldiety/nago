// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb_test

import (
	"cmp"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"slices"
	"strconv"
	"testing"

	"go.wdy.de/nago/pkg/ndb/tsdb"
)

// The data of the acceptance tests of the faster format (wokoda N22, L-N22-22 and L-N22-24) and of the benchmarks of
// L-N22-25: the kinds of data of the measurement M22d, from its generator, and the columns of the fixture
// testdata/f4aca80c5, which nago f4aca80c5, before the faster format, wrote (TestN22WriteFixture).

// n22Base is the first time of every series: 2025-01-01T00:00:00Z, the start of a quarter in Berlin.
const n22Base int64 = 1_735_689_600_000

// n22Kind is a kind of data of M22d.
type n22Kind struct {
	name     string
	decimals uint8
	enum     bool // the values are the ids of the states, written as their names into an enum column
}

// The kinds of data of M22d: power as a random walk with a daily pattern (1 decimal), with 0.1 % and 1 % NA, a counter,
// a status, which holds one value 98.7 % of the time, noise of 64 bits, a temperature with 2 and a voltage with 3
// decimals.
var n22Kinds = []n22Kind{
	{name: "power", decimals: 1},
	{name: "power_na01", decimals: 1},
	{name: "power_na1", decimals: 1},
	{name: "counter", decimals: 0},
	{name: "status", decimals: 0, enum: true},
	{name: "noise", decimals: 0},
	{name: "temp", decimals: 2},
	{name: "voltage", decimals: 3},
}

func (k n22Kind) schema() tsdb.Schema {
	if k.enum {
		return tsdb.Schema{Scheme: tsdb.SchemeEnum}
	}
	return tsdb.Schema{Scheme: tsdb.SchemeDecimal, Decimals: k.decimals}
}

// ---- the generator of M22 and M22d (scratchpad m22d/prog/main.go, seeds as there) ----

type n22Spec struct {
	base, sd, theta, dayAmp, lo, hi float64
}

var n22Day [144]float64

func init() {
	for i := range n22Day {
		n22Day[i] = math.Sin(2*math.Pi*float64(i)/144 - math.Pi/2)
	}
}

type n22Gen struct {
	r     *rand.Rand
	sp    n22Spec
	x     float64
	k     int
	state int
	left  int
}

func newN22Gen(t, j int, sp n22Spec) *n22Gen {
	return &n22Gen{r: rand.New(rand.NewPCG(uint64(t)*1000+uint64(j), 0x9e3779b97f4a7c15)), sp: sp, x: sp.base}
}

func (g *n22Gen) nextF() float64 {
	sp := &g.sp
	g.x += sp.theta*(sp.base-g.x) + sp.sd*g.r.NormFloat64()
	v := g.x + sp.dayAmp*n22Day[g.k%144]
	g.k++
	return min(max(v, sp.lo), sp.hi)
}

// nextEnum is a status of 8 states, which leaves state 0 at about 0.2 % of the points for 1 to 12 points.
func (g *n22Gen) nextEnum() int64 {
	if g.state == 0 {
		if g.r.IntN(500) == 0 {
			g.state = 1 + g.r.IntN(7)
			g.left = 1 + g.r.IntN(12)
		}
	} else {
		g.left--
		if g.left <= 0 {
			g.state = 0
		}
	}
	return int64(g.state)
}

var (
	n22SpecPower   = n22Spec{1500, 40, 0.01, 300, 0, 3600}
	n22SpecGearbox = n22Spec{55, 0.4, 0.02, 6, -10, 95}
	n22SpecVoltage = n22Spec{690, 1.5, 0.05, 2, 600, 760}
)

// n22Values returns n values of the kind as scaled int64: tsdb.NA for NA, the id of a state for the status.
func n22Values(kind string, n int) []int64 {
	v := make([]int64, n)
	switch kind {
	case "power", "power_na01", "power_na1":
		g := newN22Gen(0, 2, n22SpecPower)
		r := rand.New(rand.NewPCG(11, 12))
		q := map[string]int{"power_na01": 1000, "power_na1": 100}[kind]
		for i := range v {
			v[i] = int64(math.Round(g.nextF() * 10))
			if q > 0 && r.IntN(q) == 0 {
				v[i] = tsdb.NA
			}
		}
	case "counter":
		r := rand.New(rand.NewPCG(7, 8))
		c := int64(1_000_000)
		for i := range v {
			c += int64(r.IntN(16))
			v[i] = c
		}
	case "status":
		g := newN22Gen(0, 19, n22Spec{})
		for i := range v {
			v[i] = g.nextEnum()
		}
	case "noise":
		r := rand.New(rand.NewPCG(9, 10))
		for i := range v {
			x := int64(r.Uint64())
			if x == tsdb.NA {
				x++
			}
			v[i] = x
		}
	case "temp":
		g := newN22Gen(0, 10, n22SpecGearbox)
		for i := range v {
			v[i] = int64(math.Round(g.nextF() * 100))
		}
	case "voltage":
		g := newN22Gen(0, 14, n22SpecVoltage)
		for i := range v {
			v[i] = int64(math.Round(g.nextF() * 1000))
		}
	default:
		panic("unknown kind " + kind)
	}
	return v
}

// n22StateName is the name of a state of the status in the enum column.
func n22StateName(id int64) string {
	if id == 0 {
		return "ok"
	}
	return "state-" + strconv.FormatInt(id, 10)
}

// ---- points and the columns of the fixture ----

// n22Point is a point of a column: the time and the scaled value of a decimal column, or the text of an enum or a
// string column.
type n22Point struct {
	ts  int64
	v   int64
	str string
}

// n22Life is what one life of the database writes into a column, in this order: the appended points, then the
// corrections (an out-of-order put each) and the deletions; it ends with Close, or with a crash after a Sync, if crash
// is set (the crash copy). syncAfter are indexes of points, after which the life calls Sync.
type n22Life struct {
	points    []n22Point
	syncAfter []int
	fix       []n22Point
	del       []int64
	flush     bool // Flush before the end of the life
	crash     bool
}

// n22FixCol is a column of the fixture.
type n22FixCol struct {
	name     string
	schema   tsdb.Schema
	compress bool
	lives    []n22Life
}

// n22Grid returns the points of vals in a grid of step milliseconds from start.
func n22Grid(start, step int64, vals []int64) []n22Point {
	p := make([]n22Point, len(vals))
	for i, v := range vals {
		p[i] = n22Point{ts: start + int64(i)*step, v: v}
	}
	return p
}

// n22FixtureCols are the columns of the fixture testdata/f4aca80c5, all in bucket n22. Their points are those of the
// generator, so that a test knows each value besides the checksums of the fixture.
func n22FixtureCols() []n22FixCol {
	const sec = 1000
	const n = 4096 + 1000 // a full block and the open rest, which Close seals as a block of its own
	var cols []n22FixCol
	for _, k := range n22Kinds {
		vals := n22Values(k.name, n)
		pts := n22Grid(n22Base, sec, vals)
		if k.enum {
			for i := range pts {
				pts[i] = n22Point{ts: pts[i].ts, str: n22StateName(vals[i])}
			}
		}
		cols = append(cols, n22FixCol{name: k.name, schema: k.schema(), compress: true, lives: []n22Life{{points: pts}}})
	}
	// power without compression
	cols = append(cols, n22FixCol{name: "power_raw", schema: tsdb.Schema{Scheme: tsdb.SchemeDecimal, Decimals: 1},
		lives: []n22Life{{points: n22Grid(n22Base, sec, n22Values("power", n))}}})

	// gaps: an equidistant block, a block with irregular holes, a block with a single hole, the rest with holes of one step
	{
		vals := n22Values("temp", 3*4096+17)
		r := rand.New(rand.NewPCG(21, 22))
		pts := make([]n22Point, len(vals))
		ts := n22Base
		for i, v := range vals {
			pts[i] = n22Point{ts: ts, v: v}
			step := int64(sec)
			switch {
			case i >= 4096 && i < 2*4096:
				step = sec * int64(1+r.IntN(300)) // irregular
			case i == 2*4096+2000:
				step = 3600 * sec // one hole of an hour
			case i >= 3*4096:
				step = sec * int64(1+i%2)
			}
			ts += step
		}
		cols = append(cols, n22FixCol{name: "gaps", schema: tsdb.Schema{Scheme: tsdb.SchemeDecimal, Decimals: 2}, compress: true,
			lives: []n22Life{{points: pts}}})
	}
	// constant blocks: 4242, NA, 0, a varying rest of a block and -7
	{
		var vals []int64
		for _, c := range []int64{4242, tsdb.NA, 0} {
			for range 4096 {
				vals = append(vals, c)
			}
		}
		vals = append(vals, n22Values("temp", 100)...)
		for range 4096 {
			vals = append(vals, -7)
		}
		cols = append(cols, n22FixCol{name: "const", schema: tsdb.Schema{Scheme: tsdb.SchemeDecimal, Decimals: 2}, compress: true,
			lives: []n22Life{{points: n22Grid(n22Base, sec, vals)}}})
	}
	// the bounds of int64: runs of extremes, alternating extremes, whose deltas overflow, and NA between them
	{
		ext := []int64{math.MaxInt64, math.MinInt64 + 1, 0, -1, 1, math.MaxInt64 - 1, tsdb.NA}
		var vals []int64
		for i := range 4096 {
			vals = append(vals, ext[i%len(ext)])
		}
		for i := range 4096 {
			if i%2 == 0 {
				vals = append(vals, math.MaxInt64)
			} else {
				vals = append(vals, math.MinInt64+1)
			}
		}
		for i := range 5 {
			vals = append(vals, ext[(i*3)%len(ext)])
		}
		cols = append(cols, n22FixCol{name: "limits", schema: tsdb.Schema{Scheme: tsdb.SchemeDecimal}, compress: true,
			lives: []n22Life{{points: n22Grid(n22Base, sec, vals)}}})
	}
	// NA: the first and the last point, runs of 1 to 300 and single NA between power values, a third of all points
	{
		vals := n22Values("power", 2*4096+33)
		r := rand.New(rand.NewPCG(31, 32))
		vals[0] = tsdb.NA
		for i := 1; i < len(vals)-1; {
			if r.IntN(10) < 3 {
				run := 1 + r.IntN(300)
				for j := i; j < min(i+run, len(vals)); j++ {
					vals[j] = tsdb.NA
				}
				i += run
			}
			i += 1 + r.IntN(200)
		}
		vals[len(vals)-1] = tsdb.NA
		cols = append(cols, n22FixCol{name: "na", schema: tsdb.Schema{Scheme: tsdb.SchemeDecimal, Decimals: 1}, compress: true,
			lives: []n22Life{{points: n22Grid(n22Base, sec, vals)}}})
	}
	// block sizes: each life ends with Close, which seals the open block as a block of its own: blocks of 127, 128, 129,
	// 4095, 4096 and 4096 and 1 points
	{
		vals := n22Values("power", 127+128+129+4095+4096+4097)
		pts := n22Grid(n22Base, sec, vals)
		var lives []n22Life
		at := 0
		for _, size := range []int{127, 128, 129, 4095, 4096, 4097} {
			lives = append(lives, n22Life{points: pts[at : at+size]})
			at += size
		}
		cols = append(cols, n22FixCol{name: "sizes", schema: tsdb.Schema{Scheme: tsdb.SchemeDecimal, Decimals: 1}, compress: true,
			lives: lives})
	}
	// a finalized and an open chunk: Flush finalizes the first life
	{
		pts := n22Grid(n22Base, sec, n22Values("voltage", 10_000))
		cols = append(cols, n22FixCol{name: "multi", schema: tsdb.Schema{Scheme: tsdb.SchemeDecimal, Decimals: 3}, compress: true,
			lives: []n22Life{{points: pts[:5000], flush: true}, {points: pts[5000:]}}})
	}
	// an open chunk left by a crash: sealed blocks and two continuation records of the open block behind them
	{
		pts := n22Grid(n22Base, sec, n22Values("power", 4096+500+4096+300+200))
		cols = append(cols, n22FixCol{name: "resume", schema: tsdb.Schema{Scheme: tsdb.SchemeDecimal, Decimals: 1}, compress: true,
			lives: []n22Life{{points: pts[:4596]}, {points: pts[4596:], syncAfter: []int{4096 + 300 - 1, 4096 + 300 + 200 - 1}, crash: true}}})
	}
	// a head left by a crash: a correction and a deletion in head.wal over sealed blocks
	{
		pts := n22Grid(n22Base, sec, n22Values("counter", 4096+100))
		cols = append(cols, n22FixCol{name: "wal", schema: tsdb.Schema{Scheme: tsdb.SchemeDecimal}, compress: true,
			lives: []n22Life{{points: pts[:4096]}, {points: pts[4096:], syncAfter: []int{99},
				fix: []n22Point{{ts: pts[10].ts, v: -12345}, {ts: pts[4096+50].ts, v: math.MaxInt64}}, del: []int64{pts[20].ts, pts[4096+60].ts}, crash: true}}})
	}
	// an enum of many names with constant runs
	{
		var pts []n22Point
		r := rand.New(rand.NewPCG(41, 42))
		for i := range 2*4096 + 77 {
			name := "v" + strconv.Itoa(r.IntN(600))
			if i >= 4096 && i < 4096+1500 {
				name = "steady"
			}
			pts = append(pts, n22Point{ts: n22Base + int64(i)*sec, str: name})
		}
		cols = append(cols, n22FixCol{name: "enum", schema: tsdb.Schema{Scheme: tsdb.SchemeEnum}, compress: true,
			lives: []n22Life{{points: pts}}})
	}
	// text: empty, short, Unicode and long strings, with holes in time
	{
		var pts []n22Point
		texts := []string{"", "a", "Grüße 😀", fmt.Sprintf("%01000d", 7), "zeile\nzwei\x00null", "ok"}
		ts := n22Base
		for i := range 4096 + 50 {
			pts = append(pts, n22Point{ts: ts, str: texts[i%len(texts)] + strconv.Itoa(i%13)})
			ts += sec * int64(1+i%3)
		}
		cols = append(cols, n22FixCol{name: "text", schema: tsdb.Schema{Scheme: tsdb.SchemeString}, compress: true,
			lives: []n22Life{{points: pts}}})
	}
	return cols
}

// n22Model returns the points, which a column holds after its lives, in ascending time: the appended points with the
// corrections and without the deletions.
func n22Model(c n22FixCol) []n22Point {
	m := map[int64]n22Point{}
	for _, l := range c.lives {
		for _, p := range l.points {
			m[p.ts] = p
		}
		for _, p := range l.fix {
			m[p.ts] = p
		}
		for _, ts := range l.del {
			delete(m, ts)
		}
	}
	out := make([]n22Point, 0, len(m))
	for _, p := range m {
		out = append(out, p)
	}
	slices.SortFunc(out, func(a, b n22Point) int { return cmp.Compare(a.ts, b.ts) })
	return out
}

// ---- reading a column into checksums ----

// n22Read is what a column reads with each read method of tsdb, as a number of points and a SHA-256 of the times and
// the exact values: the scaled int64, the bits of the float64 (NaN of NA included) or the bytes of the text.
type n22Read struct {
	Points     int    `json:"points"`
	ScanI64    string `json:"scanI64,omitempty"`
	ScanF64    string `json:"scanF64,omitempty"`
	IterF64    string `json:"iterF64,omitempty"`
	ScanString string `json:"scanString,omitempty"`
	IterString string `json:"iterString,omitempty"`
}

type n22Hash struct {
	buf []byte
	h   interface {
		Write([]byte) (int, error)
		Sum([]byte) []byte
	}
}

func newN22Hash() *n22Hash { return &n22Hash{h: sha256.New()} }

func (h *n22Hash) u64(ts int64, v uint64) {
	h.buf = binary.LittleEndian.AppendUint64(binary.LittleEndian.AppendUint64(h.buf[:0], uint64(ts)), v)
	_, _ = h.h.Write(h.buf)
}

func (h *n22Hash) str(ts int64, s string) {
	h.buf = binary.LittleEndian.AppendUint64(binary.LittleEndian.AppendUint64(h.buf[:0], uint64(ts)), uint64(len(s)))
	_, _ = h.h.Write(append(h.buf, s...))
}

func (h *n22Hash) hex() string { return hex.EncodeToString(h.h.Sum(nil)) }

// n22ReadColumn reads the whole column c with every read method of its scheme.
func n22ReadColumn(t testing.TB, c *tsdb.Column) n22Read {
	t.Helper()
	var r n22Read
	switch c.Schema().Scheme {
	case tsdb.SchemeDecimal:
		hi, hf, hit := newN22Hash(), newN22Hash(), newN22Hash()
		must(t, c.ScanI64(minTS, maxTS, func(ts, vals []int64) bool {
			for i := range ts {
				hi.u64(ts[i], uint64(vals[i]))
				r.Points++
			}
			return true
		}))
		must(t, c.ScanF64(minTS, maxTS, func(ts []int64, vals []float64) bool {
			for i := range ts {
				hf.u64(ts[i], math.Float64bits(vals[i]))
			}
			return true
		}))
		n := 0
		for p := range c.IterF64(minTS, maxTS) {
			hit.u64(int64(p.X), math.Float64bits(float64(p.Y)))
			n++
		}
		if n != r.Points {
			t.Errorf("IterF64 yields %d points, ScanI64 %d", n, r.Points)
		}
		r.ScanI64, r.ScanF64, r.IterF64 = hi.hex(), hf.hex(), hit.hex()
	default:
		hs, his := newN22Hash(), newN22Hash()
		must(t, c.ScanString(minTS, maxTS, func(ts []int64, vals []string) bool {
			for i := range ts {
				hs.str(ts[i], vals[i])
				r.Points++
			}
			return true
		}))
		n := 0
		for p := range c.IterString(minTS, maxTS) {
			his.str(int64(p.X), p.V)
			n++
		}
		if n != r.Points {
			t.Errorf("IterString yields %d points, ScanString %d", n, r.Points)
		}
		r.ScanString, r.IterString = hs.hex(), his.hex()
	}
	return r
}

// n22Expect returns the checksums of the points of a model, as a column of the schema reads them by the contract:
// ScanF64 and IterF64 yield NaN (math.NaN()) for NA and float64(v)/10^n else (L-N22-22).
func n22Expect(schema tsdb.Schema, pts []n22Point) n22Read {
	r := n22Read{Points: len(pts)}
	if schema.Scheme == tsdb.SchemeDecimal {
		hi, hf := newN22Hash(), newN22Hash()
		p := math.Pow10(int(schema.Decimals))
		for _, x := range pts {
			hi.u64(x.ts, uint64(x.v))
			f := math.NaN()
			if x.v != tsdb.NA {
				f = float64(x.v) / p
			}
			hf.u64(x.ts, math.Float64bits(f))
		}
		r.ScanI64, r.ScanF64 = hi.hex(), hf.hex()
		r.IterF64 = r.ScanF64
		return r
	}
	hs := newN22Hash()
	for _, x := range pts {
		hs.str(x.ts, x.str)
	}
	r.ScanString = hs.hex()
	r.IterString = r.ScanString
	return r
}

// n22Full reports whether the tests run in full size: without -short or with N22_FULL=1.
func n22Full() bool { return !testing.Short() || os.Getenv("N22_FULL") == "1" }

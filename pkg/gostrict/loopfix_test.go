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
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"go.wdy.de/nago/pkg/gostrict"
)

// The tests of this file check the messages and fixes of GS040 for loops after the contract R-SCRIPT-GS-LOOP-FIX of
// wokoda (L-N9-2). Each row is a program with one loop in the terminating profile. The message is the text of the
// diagnostic (Msg, Why and Fix), compared without white space and case, so that the rewrite may name the loop
// itself or the placeholders of the contract (i, n, m, L, B): the programs use these names. A fix is applied with
// [gostrict.ApplyEdits]; the fixed program has no diagnostic, and the program and its fix run in goja and in the vm
// without the profile with the same number of passes, which each program counts. The bodies of the loops with a fix
// hold no call, no conversion and no change of a list, so that the conditions of a fix hold under every reading.
// The two gaps (a counter of type uint32 downwards with >= 0, a bound, which a captured callback changes) are
// observed in the engines: the program does not end within its time, its counter row ends.

// loopFile is the file of the programs in the diagnostics.
const loopFile = "llm.local/s/s.go"

// the message and the remedy of GS040 for a loop before N9 (check_term.go), which the other reasons keep
const (
	loopWhy = "the terminating profile requires, that every loop ends"
	loopFix = "count with a bound, e.g. for i := int32(0); i < n; i++, range over a list or a number, or limit the loop with fuel: f := fuel.New(1000); for f.Burn() && cond { ... }"
)

// loopKind is the rewrite, which the message names.
type loopKind int

const (
	loopOther    loopKind = iota // the message of today, see loopRow.old
	loopInclUp                   // end := n + 1 and i < end
	loopInclDown                 // end := m - 1 and i > end
	loopLen                      // n := L.Len() and i < n
	loopExpr                     // the expression in a variable before the loop
	loopCompound                 // the rest as if !(B) { break } at the start of the body
)

type loopRow struct {
	name string
	// src is the program without the package clause; the line of the loop contains anchor exactly once
	src    string
	anchor string
	kind   loopKind
	// expr is the bound for loopExpr, as the message may name it
	expr string
	// old is the message of today for loopOther
	old string
	fix bool
	// for a row with a fix: the calls of Run and their results, which the program and its fix must give
	calls [][]any
	want  [][]any
	// after checks the source of the fix
	after func(t *testing.T, fixed string)
}

func loopProgram(src string) string {
	return "package s\n\n" + src
}

func loopCheck(t *testing.T, src string) gostrict.Diagnostics {
	t.Helper()
	prog, err := gostrict.LoadModules("llm.local/s", gostrict.Module{Path: "llm.local/s", FS: fstest.MapFS{"s.go": {Data: []byte(src)}}})
	if err != nil {
		t.Fatal(err)
	}
	prog.Terminating = true
	return gostrict.Check(prog)
}

// loopLine returns the line, which contains the anchor exactly once.
func loopLine(t *testing.T, src, anchor string) int {
	t.Helper()
	res := 0
	for i, l := range strings.Split(src, "\n") {
		if strings.Contains(l, anchor) {
			if res != 0 {
				t.Fatalf("anchor %q is ambiguous", anchor)
			}
			res = i + 1
		}
	}
	if res == 0 {
		t.Fatalf("anchor %q not found", anchor)
	}
	return res
}

// loopNorm is s without white space in lower case.
func loopNorm(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), ""))
}

func loopText(d gostrict.Diagnostic) string {
	return d.Msg + "\n" + d.Why + "\n" + d.Fix
}

// loopNames tells, whether the message names the rewrite of its kind.
func loopNames(kind loopKind, expr, msg string) bool {
	m := loopNorm(msg)
	has := func(parts ...string) bool {
		for _, p := range parts {
			if !strings.Contains(m, loopNorm(p)) {
				return false
			}
		}
		return true
	}
	switch kind {
	case loopInclUp:
		return has("end := n + 1", "i < end")
	case loopInclDown:
		return has("end := m - 1", "i > end")
	case loopLen:
		return has("n := L.Len()", "i < n")
	case loopExpr:
		// the expression of the loop in a variable, or the example of the contract
		return has(":= "+expr) || has("n := L.Len()", "i < n")
	case loopCompound:
		return has("if !(", "{ break }")
	}
	return false
}

// loopRun compiles the program without the profile and calls Run in goja and in the vm.
func loopRun(t *testing.T, src string, args []any) map[string][]any {
	t.Helper()
	b, diags, err := gostrict.NewCache(1).Compile("llm.local/s", script(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) > 0 {
		t.Fatalf("the program does not compile without the profile:\n%s\n%s", diags.Text(), src)
	}
	res := map[string][]any{}
	for _, e := range []struct {
		name string
		e    gostrict.Engine
	}{{"goja", goEngine}, {"vm", vmEngine}} {
		in := mustLoad(t, e.e, b)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		got, err := in.Call(ctx, "Run", args...)
		cancel()
		_ = in.Close()
		if err != nil {
			t.Fatalf("%s: Run(%v): %v", e.name, args, err)
		}
		res[e.name] = got
	}
	return res
}

// loopListFn builds a list of k elements in the function, before the loop, so that the body changes no list.
const loopListFn = "\tl := list.New[int32]()\n\tfor j := range k {\n\t\tl.Append(j)\n\t}\n"

const loopListImport = "import \"go.wdy.de/nago/pkg/gostrict/std/list\"\n\n"

// loopBefore checks, that the line directly before the line of the anchor (the loop or its label) declares a
// variable with decl.
func loopBefore(anchor, decl string) func(t *testing.T, fixed string) {
	return func(t *testing.T, fixed string) {
		t.Helper()
		lines := strings.Split(fixed, "\n")
		for i, l := range lines {
			if strings.Contains(l, anchor) {
				if i == 0 || !strings.Contains(lines[i-1], ":=") || !strings.Contains(loopNorm(lines[i-1]), loopNorm(decl)) {
					t.Errorf("the line before %q does not declare a variable with %q:\n%s", anchor, decl, fixed)
				}
				return
			}
		}
		t.Errorf("the fix lost %q:\n%s", anchor, fixed)
	}
}

// loopContains checks, that the fix contains s, compared without white space.
func loopContains(s string) func(t *testing.T, fixed string) {
	return func(t *testing.T, fixed string) {
		t.Helper()
		if !strings.Contains(loopNorm(fixed), loopNorm(s)) {
			t.Errorf("the fix lacks %q:\n%s", s, fixed)
		}
	}
}

func loopAll(fs ...func(t *testing.T, fixed string)) func(t *testing.T, fixed string) {
	return func(t *testing.T, fixed string) {
		t.Helper()
		for _, f := range fs {
			f(t, fixed)
		}
	}
}

var loopRows = []loopRow{
	// the first table: each reason with a met condition gets its rewrite and a fix
	{
		name: "expression n+1, n not changed: rewrite and fix",
		src: `func Run(n int32) int32 {
	c := int32(0)
	for i := int32(0); i < n+1; i++ {
		c++
	}
	return c
}
`,
		anchor: "for i :=", kind: loopExpr, expr: "n + 1", fix: true,
		calls: [][]any{{int32(5)}, {int32(0)}, {int32(-1)}}, want: [][]any{{int32(6)}, {int32(1)}, {int32(0)}},
		after: loopBefore("for i :=", "n + 1"),
	},
	{
		name: "expression L.Len(), the body changes no list: rewrite n := L.Len() and fix",
		src: loopListImport + `func Run(k int32) int32 {
` + loopListFn + `	c := int32(0)
	for i := int32(0); i < l.Len(); i++ {
		c++
	}
	return c
}
`,
		anchor: "for i :=", kind: loopLen, fix: true,
		calls: [][]any{{int32(3)}, {int32(0)}}, want: [][]any{{int32(3)}, {int32(0)}},
		after: loopBefore("for i :=", "l.Len()"),
	},
	{
		name: "expression L.Len()+1 (no call but Len): rewrite and fix",
		src: loopListImport + `func Run(k int32) int32 {
` + loopListFn + `	c := int32(0)
	for i := int32(0); i < l.Len()+1; i++ {
		c++
	}
	return c
}
`,
		anchor: "for i :=", kind: loopExpr, expr: "l.Len() + 1", fix: true,
		calls: [][]any{{int32(3)}, {int32(0)}}, want: [][]any{{int32(4)}, {int32(1)}},
		after: loopBefore("for i :=", "l.Len()"),
	},
	{
		name: "compound i < n && B with B a variable, which the body changes: rewrite and fix",
		src: `func Run(n int32, more bool, stop int32) int32 {
	c := int32(0)
	for i := int32(0); i < n && more; i++ {
		c++
		if i == stop {
			more = false
		}
	}
	return c
}
`,
		anchor: "for i :=", kind: loopCompound, fix: true,
		calls: [][]any{{int32(10), true, int32(-1)}, {int32(10), true, int32(3)}, {int32(10), false, int32(-1)}, {int32(0), true, int32(-1)}},
		want:  [][]any{{int32(10)}, {int32(4)}, {int32(0)}, {int32(0)}},
		after: loopContains("for i := int32(0); i < n; i++ { if !(more) { break }"),
	},
	{
		name: "compound i < n && B with B a comparison: rewrite and fix with !(B)",
		src: `func Run(n int32, x float64) int32 {
	c := int32(0)
	for i := int32(0); i < n && x > 1; i++ {
		x /= 2
		c++
	}
	return c
}
`,
		anchor: "for i :=", kind: loopCompound, fix: true,
		calls: [][]any{{int32(10), float64(20)}, {int32(2), float64(20)}, {int32(10), float64(1)}},
		want:  [][]any{{int32(5)}, {int32(2)}, {int32(0)}},
		after: loopContains("for i := int32(0); i < n; i++ { if !(x > 1) { break }"),
	},

	// the second table: an inclusive bound, which is no constant, has a rewrite without fix
	{
		name: "inclusive i <= n upwards: rewrite end := n + 1, no fix",
		src: `func Run(n int32) int32 {
	c := int32(0)
	for i := int32(0); i <= n; i++ {
		c++
	}
	return c
}
`,
		anchor: "for i :=", kind: loopInclUp,
	},
	{
		name: "inclusive i >= m downwards: rewrite end := m - 1, no fix",
		src: `func Run(m int32) int32 {
	c := int32(0)
	for i := int32(10); i >= m; i-- {
		c++
	}
	return c
}
`,
		anchor: "for i :=", kind: loopInclDown,
	},

	// the third table: the condition of a fix does not hold
	{
		name: "expression n+1 with n = n - 1 in the body: rewrite, no fix",
		src: `func Run(n int32) int32 {
	c := int32(0)
	for i := int32(0); i < n+1; i++ {
		n = n - 1
		c++
	}
	return c
}
`,
		anchor: "for i :=", kind: loopExpr, expr: "n + 1",
	},
	{
		name: "L.Len() with l.Append in the body: rewrite, no fix",
		src: loopListImport + `func Run(k int32) int32 {
` + loopListFn + `	c := int32(0)
	for i := int32(0); i < l.Len(); i++ {
		l.Append(i)
		c++
	}
	return c
}
`,
		anchor: "for i :=", kind: loopLen,
	},
	{
		name: "L.Len() with Append through an alias m := l: rewrite, no fix",
		src: loopListImport + `func Run(k int32) int32 {
` + loopListFn + `	m := l
	c := int32(0)
	for i := int32(0); i < l.Len(); i++ {
		m.Append(i)
		c++
	}
	return c
}
`,
		anchor: "for i :=", kind: loopLen,
	},
	{
		name: "L.Len() with Append through a call f(l): rewrite, no fix",
		src: loopListImport + `func grow(l list.List[int32]) {
	l.Append(1)
}

func Run(k int32) int32 {
` + loopListFn + `	c := int32(0)
	for i := int32(0); i < l.Len(); i++ {
		grow(l)
		c++
	}
	return c
}
`,
		anchor: "for i :=", kind: loopLen,
	},
	{
		name: "L.Len() with Set, Delete and Pop of another list in the body: rewrite, no fix",
		src: loopListImport + `func Run(k int32) int32 {
` + loopListFn + `	other := list.Of[int32](1, 2, 3)
	c := int32(0)
	for i := int32(0); i < l.Len(); i++ {
		other.Set(0, i)
		other.Delete(0)
		other.Pop()
		c++
	}
	return c
}
`,
		anchor: "for i :=", kind: loopLen,
	},
	{
		name: "bound with another call: rewrite, no fix",
		src: `import "go.wdy.de/nago/pkg/gostrict/std/text"

func Run(s string) int32 {
	c := int32(0)
	for i := int32(0); i < text.RuneCount(s); i++ {
		c++
	}
	return c
}
`,
		anchor: "for i :=", kind: loopExpr, expr: "text.RuneCount(s)",
	},
	{
		name: "B && i < 5: rewrite, no fix",
		src: `func Run(b bool) int32 {
	c := int32(0)
	for i := int32(0); b && i < 5; i++ {
		c++
	}
	return c
}
`,
		anchor: "for i :=", kind: loopCompound,
	},
	{
		name: "i < n+1 && b: rewrite, no fix",
		src: `func Run(n int32, b bool) int32 {
	c := int32(0)
	for i := int32(0); i < n+1 && b; i++ {
		c++
	}
	return c
}
`,
		anchor: "for i :=", kind: loopCompound,
	},
	{
		name: "i < n || x: rewrite, no fix",
		src: `func Run(n int32, x bool) int32 {
	c := int32(0)
	for i := int32(0); i < n || x; i++ {
		c++
	}
	return c
}
`,
		anchor: "for i :=", kind: loopCompound,
	},
	{
		name: "i < n && B with a call in B: rewrite, no fix",
		src: `func far(x float64) bool {
	return x > 1
}

func Run(n int32, x float64) int32 {
	c := int32(0)
	for i := int32(0); i < n && far(x); i++ {
		x /= 2
		c++
	}
	return c
}
`,
		anchor: "for i :=", kind: loopCompound,
	},

	// the fourth table: a free name, the declaration before the label
	{
		name: "fix with the name n taken by an outer variable, which the body reads",
		src: loopListImport + `func Run(k, n int32) int32 {
` + loopListFn + `	c := int32(0)
	if n > 0 {
		for i := int32(0); i < l.Len(); i++ {
			c += n
		}
		c += n
	}
	return c
}
`,
		anchor: "for i :=", kind: loopLen, fix: true,
		calls: [][]any{{int32(3), int32(10)}, {int32(0), int32(10)}}, want: [][]any{{int32(40)}, {int32(10)}},
	},
	{
		name: "fix with the name n declared after the loop in the same block",
		src: loopListImport + `func Run(k int32) int32 {
` + loopListFn + `	c := int32(0)
	for i := int32(0); i < l.Len(); i++ {
		c++
	}
	n := c * 2
	return n
}
`,
		anchor: "for i :=", kind: loopLen, fix: true,
		calls: [][]any{{int32(3)}}, want: [][]any{{int32(6)}},
	},
	{
		name: "fix with the name n taken by a constant of the package, which the body reads",
		src: loopListImport + `const n = 7

func Run(k int32) int32 {
` + loopListFn + `	c := int32(0)
	for i := int32(0); i < l.Len(); i++ {
		c += n
	}
	return c
}
`,
		anchor: "for i :=", kind: loopLen, fix: true,
		calls: [][]any{{int32(3)}}, want: [][]any{{int32(21)}},
	},
	{
		name: "fix with the name end taken by an outer variable, which the body reads",
		src: `func Run(n, end int32) int32 {
	c := int32(0)
	if n >= 0 {
		for i := int32(0); i < n+1; i++ {
			c += end
		}
		c += end
	}
	return c
}
`,
		anchor: "for i :=", kind: loopExpr, expr: "n + 1", fix: true,
		calls: [][]any{{int32(4), int32(7)}, {int32(0), int32(7)}}, want: [][]any{{int32(42)}, {int32(14)}},
	},
	{
		name: "fix at a loop with a label: the declaration before the label",
		src: loopListImport + `func Run(k int32) int32 {
` + loopListFn + `	c := int32(0)
	zero := int32(0)
outer:
	for i := int32(0); i < l.Len(); i++ {
		for j := zero; j < 3; j++ {
			if j > i {
				continue outer
			}
			c++
		}
	}
	return c
}
`,
		anchor: "for i :=", kind: loopLen, fix: true,
		calls: [][]any{{int32(3)}, {int32(1)}}, want: [][]any{{int32(6)}, {int32(1)}},
		after: loopAll(loopBefore("outer:", "l.Len()"), loopContains("outer: for i :=")),
	},

	// the other reasons keep the message and the remedy of today
	{
		name: "other reason: no condition", old: "the loop has no condition",
		src: `func Run() int32 {
	c := int32(0)
	for {
		c++
		if c > 3 {
			break
		}
	}
	return c
}
`,
		anchor: "for {",
	},
	{
		name: "other reason: no counter", old: "the loop has no counter, which its post statement increments or decrements by a constant",
		src: `func Run(n, step int32) int32 {
	c := int32(0)
	for i := int32(0); i < n; i += step {
		c++
	}
	return c
}
`,
		anchor: "for i :=",
	},
	{
		name: "other reason: the body assigns the counter", old: "the body assigns the counter i",
		src: `func Run(n int32) int32 {
	c := int32(0)
	for i := int32(0); i < n; i++ {
		i += 1
		c++
	}
	return c
}
`,
		anchor: "for i :=",
	},
	{
		name: "other reason: no comparison of the counter (a variable)", old: "the condition does not compare the counter i with a bound",
		src: `func Run(more bool) int32 {
	c := int32(0)
	for i := int32(0); more; i++ {
		c++
		more = i < 3
	}
	return c
}
`,
		anchor: "for i :=",
	},
	{
		name: "other reason: no comparison of the counter (another variable)", old: "the condition does not compare the counter i with a bound",
		src: `func Run(x float64) int32 {
	c := int32(0)
	for i := int32(0); x > 1; i++ {
		x /= 2
		c++
	}
	return c
}
`,
		anchor: "for i :=",
	},
	{
		name: "other reason: wrong direction", old: "the condition does not stop the counter i in its direction",
		src: `func Run() int32 {
	c := int32(0)
	for i := int32(0); i > -5; i++ {
		c++
	}
	return c
}
`,
		anchor: "for i :=",
	},
	{
		name: "other reason: a constant bound lets the counter wrap around", old: "the counter wraps around at the bound",
		src: `func Run() int32 {
	c := int32(0)
	for i := int32(0); i <= 2147483647; i++ {
		c++
	}
	return c
}
`,
		anchor: "for i :=",
	},
	{
		name: "other reason: the body assigns the bound", old: "the body assigns the bound n",
		src: `func Run(n int32) int32 {
	c := int32(0)
	for i := int32(0); i < n; i++ {
		n++
		c++
	}
	return c
}
`,
		anchor: "for i :=",
	},
}

// TestLoopFix checks each row of R-SCRIPT-GS-LOOP-FIX: the message with the rewrite, a fix only where its condition
// holds, after the fix no finding and the same passes in both engines; the other reasons keep their message.
func TestLoopFix(t *testing.T) {
	for _, r := range loopRows {
		t.Run(r.name, func(t *testing.T) {
			src := loopProgram(r.src)
			line := loopLine(t, src, r.anchor)
			diags := loopCheck(t, src)
			var d gostrict.Diagnostic
			n := 0
			for _, x := range diags {
				if x.Code == gostrict.CodeTermination && x.File == loopFile && x.Line == line {
					d = x
					n++
				}
			}
			if n != 1 || len(diags) != 1 {
				t.Fatalf("want exactly one diagnostic, GS040 at line %d, got:\n%s\n%s", line, diags.Text(), src)
			}

			if r.kind == loopOther {
				if d.Msg != r.old || d.Why != loopWhy || d.Fix != loopFix {
					t.Errorf("the message changed:\n got  %q / %q / %q\n want %q / %q / %q", d.Msg, d.Why, d.Fix, r.old, loopWhy, loopFix)
				}
			} else if !loopNames(r.kind, r.expr, loopText(d)) {
				t.Errorf("the message does not name the rewrite:\n%s", loopText(d))
			}

			edits := diags.Edits(loopFile)
			if !r.fix {
				if len(d.Edits) > 0 || len(edits) > 0 {
					t.Errorf("a fix, where its condition does not hold: %+v", d.Edits)
				}
				return
			}
			if len(d.Edits) == 0 {
				t.Fatalf("no fix:\n%s", loopText(d))
			}
			fixed, err := gostrict.ApplyEdits([]byte(src), edits)
			if err != nil {
				t.Fatal(err)
			}
			if ds := loopCheck(t, string(fixed)); len(ds) > 0 {
				t.Fatalf("the fixed program has diagnostics:\n%s\n%s", ds.Text(), fixed)
			}
			if r.after != nil {
				r.after(t, string(fixed))
			}
			for i, args := range r.calls {
				before := loopRun(t, src, args)
				after := loopRun(t, string(fixed), args)
				for _, e := range []string{"goja", "vm"} {
					if !gostrict.SameValues(before[e], r.want[i]) {
						t.Errorf("premise, %s: Run(%v) of the program = %v, want %v", e, args, before[e], r.want[i])
					}
					if !gostrict.SameValues(after[e], before[e]) {
						t.Errorf("%s: Run(%v) after the fix = %v, before %v\n%s", e, args, after[e], before[e], fixed)
					}
				}
			}
		})
	}
}

// loopGapRow is a loop, which does not end, and its counter row, which ends without a finding.
type loopGapRow struct {
	name          string
	endless, ends string
	args          []any
	want          []any // the result of the counter row
}

var loopGapRows = []loopGapRow{
	{
		name: "counter of type uint32 downwards with >= 0",
		endless: `func Run() int32 {
	c := int32(0)
	for i := uint32(3); i >= 0; i-- {
		c++
	}
	return c
}
`,
		ends: `func Run() int32 {
	c := int32(0)
	for i := int32(3); i >= 0; i-- {
		c++
	}
	return c
}
`,
		want: []any{int32(4)},
	},
	{
		name: "bound changed by a captured callback, which the body calls",
		endless: `func Run(n int32) int32 {
	c := int32(0)
	grow := func() {
		n++
	}
	for i := int32(0); i < n; i++ {
		grow()
		c++
	}
	return c
}
`,
		ends: `func Run(n int32) int32 {
	c := int32(0)
	grow := func() {
		n++
	}
	for i := int32(0); i < n; i++ {
		c++
	}
	grow()
	return c + n
}
`,
		args: []any{int32(3)},
		want: []any{int32(7)},
	},
}

// TestLoopGaps checks the two loops, which GS040 must report in addition (R-SCRIPT-GS-LOOP-FIX): they do not end in
// goja and in the vm, and their counter rows end and have no finding.
func TestLoopGaps(t *testing.T) {
	for _, r := range loopGapRows {
		t.Run(r.name, func(t *testing.T) {
			endless, ends := loopProgram(r.endless), loopProgram(r.ends)

			// the premise: the loop does not end, its counter row does
			b, diags, err := gostrict.NewCache(1).Compile("llm.local/s", script(endless))
			if err != nil || len(diags) > 0 {
				t.Fatalf("the loop does not compile without the profile: %v\n%s", err, diags.Text())
			}
			for _, e := range []gostrict.Engine{goEngine, vmEngine} {
				in := mustLoad(t, e, b)
				ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
				res, err := in.Call(ctx, "Run", r.args...)
				cancel()
				_ = in.Close()
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Errorf("premise, %T: the loop ended with %v, %v", e, res, err)
				}
			}
			for e, got := range loopRun(t, ends, r.args) {
				if !gostrict.SameValues(got, r.want) {
					t.Errorf("premise, %s: the counter row gives %v, want %v", e, got, r.want)
				}
			}

			line := loopLine(t, endless, "for i :=")
			found := false
			for _, d := range loopCheck(t, endless) {
				found = found || d.Code == gostrict.CodeTermination && d.File == loopFile && d.Line == line
			}
			if !found {
				t.Errorf("no GS040 at the loop, line %d:\n%s", line, endless)
			}
			if ds := loopCheck(t, ends); len(ds) > 0 {
				t.Errorf("the counter row has diagnostics:\n%s", ds.Text())
			}
		})
	}
}

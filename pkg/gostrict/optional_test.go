// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"go.wdy.de/nago/pkg/gostrict"
)

// The tests of this file check the rule GS033 for the values of optional results after the contract
// R-SCRIPT-GS-OPTIONAL of wokoda (L-N11-1 to L-N11-4 and L-N11-7): every result before a last bool or error of a call
// with two or more results, whatever its type and whoever returns it, may only be read, where a guard after its
// binding proves ok resp. err == nil, or where a fact secures the element of Get at list.List, list.Frozen or
// dict.Map. Each row of the table of the detail is a program next to its counter rows. A row without a marked line is
// free: the checker reports nothing at all, so that the program also compiles. A row with lines marked "// at" has
// GS033 at one of them and no other finding; the contract does not fix, whether the finding stands at the binding or
// at the reading, so a row marks both, where both are possible. The finding names the value (where the row names
// it), the way out (GetOr, opt.OrZero, opt.OrZeroErr, the guard or passing the error on) and the line of the killing
// of a fact (the line marked "// kill"). The fixes and the values after them are checked in optional_fix_test.go.
//
// The programs take their lists and maps as parameters, so that no fact secures an element by accident, and they use
// int32 elements, whose zero value contains no nil: before N11 GS033 did not check them, so these rows show the new
// rule and not the rule of R-SCRIPT-GS-GETOR.

// optFile is the file of the programs in the diagnostics.
const optFile = "llm.local/s/s.go"

// optLib is a library of the module of the programs (Bibliothek of the detail).
const optLib = `package lib

import "go.wdy.de/nago/pkg/gostrict/std/list"

// Find returns the element i of [5, 6] and whether there is one.
func Find(i int32) (int32, bool) {
	if i == 0 {
		return 5, true
	}
	if i == 1 {
		return 6, true
	}
	return 0, false
}

// Bag holds items, which only this package sets.
type Bag struct{ items list.List[int32] }

// NewBag returns a bag with one item.
func NewBag() Bag { return Bag{items: list.Of[int32](1)} }

// Len is the number of the items.
func (b Bag) Len() int32 { return b.items.Len() }
`

// optProgram is the program of the declarations with the package clause and the imports, which they use.
func optProgram(decls string) string {
	var imp strings.Builder
	for _, p := range []string{"errors", "go.wdy.de/nago/pkg/gostrict/std/big", "go.wdy.de/nago/pkg/gostrict/std/decimal",
		"go.wdy.de/nago/pkg/gostrict/std/dict", "go.wdy.de/nago/pkg/gostrict/std/list", "go.wdy.de/nago/pkg/gostrict/std/msg",
		"go.wdy.de/nago/pkg/gostrict/std/opt", "go.wdy.de/nago/pkg/gostrict/std/text", "llm.local/s/lib"} {
		name := p[strings.LastIndex(p, "/")+1:]
		if regexp.MustCompile(`\b` + name + `\.`).MatchString(decls) {
			imp.WriteString("\t\"" + p + "\"\n")
		}
	}
	if imp.Len() == 0 {
		return "package s\n" + decls + "\n"
	}
	return "package s\n\nimport (\n" + imp.String() + ")\n" + decls + "\n"
}

// optModule is the module of a program with its library.
func optModule(src string) gostrict.Module {
	return gostrict.Module{Path: "llm.local/s", FS: fstest.MapFS{"s.go": {Data: []byte(src)}, "lib/lib.go": {Data: []byte(optLib)}}}
}

func optCheck(t *testing.T, src string) gostrict.Diagnostics {
	t.Helper()
	prog, err := gostrict.LoadModules("llm.local/s", optModule(src))
	if err != nil {
		t.Fatal(err)
	}
	return gostrict.Check(prog)
}

// optMarked returns the lines, which contain the marker as a comment.
func optMarked(src, marker string) []int {
	var res []int
	re := regexp.MustCompile(`//\s*` + marker + `\b`)
	for i, l := range strings.Split(src, "\n") {
		if re.MatchString(l) {
			res = append(res, i+1)
		}
	}
	return res
}

// optNumbered is the program with line numbers for a message.
func optNumbered(src string) string {
	var b strings.Builder
	for i, l := range strings.Split(src, "\n") {
		fmt.Fprintf(&b, "%3d %s\n", i+1, l)
	}
	return b.String()
}

func optText(d gostrict.Diagnostic) string {
	return d.Msg + "\n" + d.Why + "\n" + d.Fix
}

// optWay tells, whether the text of a finding names a way out: GetOr, opt.OrZero, opt.OrZeroErr, the guard or passing
// the error on.
var optWay = regexp.MustCompile(`GetOr|OrZero|if \w+, \w+ :=|\bif !?ok\b|err != nil|return [^\n]*err`)

// optRow is a program of the table of R-SCRIPT-GS-OPTIONAL.
type optRow struct {
	name string
	src  string
	// val is the name of the value, which the finding names, or "" (the row does not check it)
	val string
	// discard marks a row, which throws an error away: the finding has no fix and names passing the error on and
	// opt.OrZeroErr
	discard bool
	// code is the code of the finding, GS033 if empty
	code gostrict.Code
}

// optEval checks a row.
func optEval(t *testing.T, r optRow) {
	t.Helper()
	src := optProgram(r.src)
	at := optMarked(src, "at")
	kills := optMarked(src, "kill")
	diags := optCheck(t, src)
	if len(at) == 0 {
		if len(diags) > 0 {
			t.Errorf("want no finding, got:\n%s\n%s", diags.Text(), optNumbered(src))
		}
		return
	}
	code := r.code
	if code == "" {
		code = gostrict.CodeGuard
	}
	var hits []gostrict.Diagnostic
	for _, d := range diags {
		switch {
		case d.Code != code || d.File != optFile:
			t.Errorf("a finding besides %s: %s:%d %s %s", code, d.File, d.Line, d.Code, d.Msg)
		case !slices.Contains(at, d.Line):
			t.Errorf("%s at line %d, which the row does not mark (marked: %v)", code, d.Line, at)
		default:
			hits = append(hits, d)
		}
	}
	if len(hits) == 0 {
		t.Errorf("no %s at the lines %v:\n%s\n%s", code, at, diags.Text(), optNumbered(src))
		return
	}
	if code != gostrict.CodeGuard {
		return
	}
	var all strings.Builder
	for _, d := range hits {
		all.WriteString(optText(d) + "\n")
	}
	text := all.String()
	if r.val != "" && !regexp.MustCompile(`\b`+regexp.QuoteMeta(r.val)+`\b`).MatchString(text) {
		t.Errorf("the finding does not name the value %s:\n%s", r.val, text)
	}
	if !optWay.MatchString(text) {
		t.Errorf("the finding names no way out (GetOr, opt.OrZero, opt.OrZeroErr, the guard, passing the error on):\n%s", text)
	}
	for _, k := range kills {
		if !regexp.MustCompile(fmt.Sprintf(`\b%d\b`, k)).MatchString(text) {
			t.Errorf("the finding does not name the line %d of the killing of the fact:\n%s\n%s", k, text, optNumbered(src))
		}
	}
	if r.discard {
		if !strings.Contains(text, "OrZeroErr") || !regexp.MustCompile(`err != nil|return`).MatchString(text) {
			t.Errorf("the finding of a thrown away error names neither opt.OrZeroErr nor passing the error on:\n%s", text)
		}
		if e := diags.Edits(optFile); len(e) > 0 {
			t.Errorf("a fix for a thrown away error: %+v", e)
		}
	}
}

func optRun(t *testing.T, rows []optRow) {
	t.Helper()
	names := map[string]bool{}
	for _, r := range rows {
		if names[r.name] {
			t.Fatalf("the row %q is twice in the table", r.name)
		}
		names[r.name] = true
		t.Run(r.name, func(t *testing.T) { optEval(t, r) })
	}
}

// optTypes holds the types of the rows of the first line of the table.
const optTypes = `
type P struct {
	X int32
	S string
}

type R struct {
	N int32
	L list.List[int32]
}
`

// TestOptionalTypes is the first line of the table: v, _ := l.Get(i) with v of type int32, string, opt.Option, a
// struct without lists and a struct with lists is GS033; counter row: the guard if v, ok := l.Get(i); ok. The fixes
// are in TestOptionalFix.
func TestOptionalTypes(t *testing.T) {
	var rows []optRow
	for _, c := range []struct{ typ, read string }{
		{"int32", "elem"},
		{"string", "text.RuneCount(elem)"},
		{"opt.Option[int32]", "elem.Or(-1)"},
		{"P", "elem.X + text.RuneCount(elem.S)"},
		{"R", "elem.N + elem.L.Len()"},
	} {
		rows = append(rows,
			optRow{name: c.typ + ": v, _ := l.Get(i)", val: "elem", src: optTypes + `
func Run(l list.List[` + c.typ + `], i int32) int32 {
	elem, _ := l.Get(i) // at
	return ` + c.read + ` // at
}`},
			optRow{name: c.typ + ": counter row: if v, ok := l.Get(i); ok", src: optTypes + `
func Run(l list.List[` + c.typ + `], i int32) int32 {
	if elem, ok := l.Get(i); ok {
		return ` + c.read + `
	}
	return 0
}`})
	}
	optRun(t, rows)
}

// TestOptionalGuards holds the lines of the table about guards: the example of the owner of the rights, each form of
// a guard, the guard before the binding, ok reassigned before and after the guard, in a loop and in a function
// literal, and the bindings with = and by redeclaration.
func TestOptionalGuards(t *testing.T) {
	optRun(t, []optRow{
		// v, ok := l.Get(3); if ok { … }; v.X (the example of the owner of the rights)
		{name: "v.X after if ok { … } (example of the owner of the rights)", val: "elem", src: optTypes + `
func Run(l list.List[P]) int32 {
	n := int32(0)
	elem, ok := l.Get(3) // at
	if ok {
		n = 1
	}
	return n + elem.X // at
}`},
		{name: "counter row: v.X in the body of if ok", src: optTypes + `
func Run(l list.List[P]) int32 {
	n := int32(0)
	elem, ok := l.Get(3)
	if ok {
		n = 1 + elem.X
	}
	return n
}`},

		// each form of a guard
		{name: "guard: the body of if ok", src: `
func Run(l list.List[int32], i int32) int32 {
	v, ok := l.Get(i)
	if ok {
		return v
	}
	return 0
}`},
		{name: "guard: the body of if ok && x", src: `
func Run(l list.List[int32], i int32, x bool) int32 {
	v, ok := l.Get(i)
	if ok && x {
		return v
	}
	return 0
}`},
		{name: "guard: the else of if !ok", src: `
func Run(l list.List[int32], i int32) int32 {
	v, ok := l.Get(i)
	if !ok {
		return 0
	} else {
		return v
	}
}`},
		{name: "guard: the right part of ok && …", src: `
func Run(l list.List[int32], i int32) bool {
	v, ok := l.Get(i)
	return ok && v > 3
}`},
		{name: "guard: the right part of !ok || …", src: `
func Run(l list.List[int32], i int32) bool {
	v, ok := l.Get(i)
	return !ok || v > 3
}`},
		{name: "guard: if !ok || x { continue } before", src: `
func Run(l list.List[int32], x bool) int32 {
	s := int32(0)
	for i := range int32(4) {
		v, ok := l.Get(i)
		if !ok || x {
			continue
		}
		s += v
	}
	return s
}`},
		{name: "guard: if !ok { break } before", src: `
func Run(l list.List[int32]) int32 {
	s := int32(0)
	for i := range int32(4) {
		v, ok := l.Get(i)
		if !ok {
			break
		}
		s += v
	}
	return s
}`},
		{name: "guard: an early return in a case", src: `
func Run(l list.List[int32], i, k int32) int32 {
	v, ok := l.Get(i)
	switch k {
	case 1:
		if !ok {
			return 0
		}
		return v
	}
	return -1
}`},
		{name: "counter row: if ok || x { … v … }", val: "v", src: `
func Run(l list.List[int32], i int32, x bool) int32 {
	v, ok := l.Get(i) // at
	if ok || x {
		return v // at
	}
	return 0
}`},
		{name: "counter row: if !ok { x = 1 } without return, then v", val: "v", src: `
func Run(l list.List[int32], i int32) int32 {
	x := int32(0)
	v, ok := l.Get(i) // at
	if !ok {
		x = 1
	}
	return v + x // at
}`},

		// if !ok { return } after the binding; counter row: the guard before the binding
		{name: "if !ok { return } after the binding, then v.X", src: optTypes + `
func Run(l list.List[P]) int32 {
	v, ok := l.Get(5)
	if !ok {
		return 0
	}
	return v.X
}`},
		{name: "counter row: the guard before the binding", val: "v", src: optTypes + `
func Run(l list.List[P]) int32 {
	_, ok := l.Get(0)
	if !ok {
		return 0
	}
	v, ok := l.Get(5) // at
	return v.X // at
}`},

		// ok reassigned between the binding and the guard; counter row: after the guard
		{name: "ok reassigned between the binding and the guard", val: "v", src: `
func Run(l list.List[int32], i int32) int32 {
	v, ok := l.Get(i) // at
	ok = i > 0
	if ok {
		return v // at
	}
	return 0
}`},
		{name: "counter row: ok reassigned after the guard", src: `
func Run(l list.List[int32], i int32) int32 {
	v, ok := l.Get(i)
	if !ok {
		return 0
	}
	ok = i > 0
	if ok {
		return v + 1
	}
	return v
}`},

		// the guard in a loop, ok reassigned in its body afterwards; the guard in a function literal, ok reassigned
		// after the binding; a function literal, which assigns ok, called in the loop with the guard
		{name: "guard in a loop, ok reassigned afterwards in its body", val: "v", src: `
func Run(l list.List[int32]) int32 {
	s := int32(0)
	v, ok := l.Get(0) // at
	for i := range int32(3) {
		if ok {
			s += v // at
		}
		ok = i > 5
	}
	return s
}`},
		{name: "guard in a function literal, ok reassigned after the binding", val: "v", src: `
func Run(l list.List[int32]) int32 {
	v, ok := l.Get(0) // at
	f := func() int32 {
		if ok {
			return v // at
		}
		return 0
	}
	ok = true
	return f()
}`},
		{name: "set := func() { ok = true } and set() in the loop with the guard", val: "v", src: `
func Run(l list.List[int32]) int32 {
	s := int32(0)
	v, ok := l.Get(0) // at
	set := func() { ok = true }
	for range int32(3) {
		if ok {
			s += v // at
		}
		set()
	}
	return s
}`},
		{name: "counter row: ok reassigned after a guard outside of any loop", src: `
func Run(l list.List[int32]) int32 {
	v, ok := l.Get(0)
	if ok {
		ok = false
		return v
	}
	return 0
}`},

		// bindings with = and by redeclaration; counter rows: the form if t, ok := …; ok { v = t } and var v, ok = …
		{name: "v, ok = l.Get(i)", src: `
func Run(l list.List[int32], i int32) int32 {
	var v int32
	var ok bool
	v, ok = l.Get(i) // at
	if ok {
		return v // at
	}
	return 0
}`},
		{name: "s.F, ok = l.Get(i)", src: `
type S struct{ F int32 }

func Run(l list.List[int32], i int32) int32 {
	var s S
	var ok bool
	s.F, ok = l.Get(i) // at
	if ok {
		return s.F // at
	}
	return 0
}`},
		{name: "v, ok2 := b.Get(9) with an existing v", src: `
func Run(b list.List[int32]) int32 {
	v := int32(1)
	v, ok2 := b.Get(9) // at
	if ok2 {
		return v // at
	}
	return 0
}`},
		{name: "counter row: if t, ok := l.Get(i); ok { v = t }", src: `
func Run(l list.List[int32], i int32) int32 {
	var v int32
	if t, ok := l.Get(i); ok {
		v = t
	}
	return v
}`},
		{name: "counter row: var v, ok = l.Get(i) with a guard", src: `
func Run(l list.List[int32], i int32) int32 {
	var v, ok = l.Get(i)
	if ok {
		return v
	}
	return 0
}`},

		// the guard outside, the access in a function literal in its scope
		{name: "guard outside, access in a function literal in its scope", src: `
func Run(l list.List[int32], i int32) int32 {
	v, ok := l.Get(i)
	if !ok {
		return 0
	}
	f := func() int32 { return v * 2 }
	return f()
}`},
		{name: "guard outside, access in a function literal in the body of if ok", src: `
func Run(l list.List[int32], i int32) int32 {
	v, ok := l.Get(i)
	if ok {
		f := func() int32 { return v * 2 }
		return f()
	}
	return 0
}`},
	})
}

// optCallees are the declarations, which return (int32, bool) in each form of a callee.
const optCallees = `
func find(i int32) (int32, bool) {
	if i == 0 {
		return 5, true
	}
	return 0, false
}

type Box struct{ l list.List[int32] }

func (b Box) Find(i int32) (int32, bool) {
	return b.l.Get(i)
}

type Finder interface {
	Find(i int32) (int32, bool)
}
`

// TestOptionalCallees is the line of the table about the callees and types: an own function with (T, bool), a
// method, a method value, a function value, a method of an interface and of a type parameter, a library, a header of
// std (text.ParseInt), decimal.Parse, big.QuoRem with three results, list.Pop and opt.Option.Get: each without a guard
// GS033, each counter row with a guard free. r.Text is a header of wokoda and checked there.
func TestOptionalCallees(t *testing.T) {
	type callee struct {
		name string
		// pre are the statements before the call, call is the call, read the expression of the value v
		pre, call, read string
		ret             string // the type of Run
		err             bool   // the last result is error
	}
	callees := []callee{
		{name: "own function with (T, bool)", call: "find(i)", read: "v", ret: "int32"},
		{name: "method", pre: "\tb := Box{l: list.New[int32]()}\n", call: "b.Find(i)", read: "v", ret: "int32"},
		{name: "method value", pre: "\tf := Box{l: list.New[int32]()}.Find\n", call: "f(i)", read: "v", ret: "int32"},
		{name: "function value", pre: "\tf := find\n", call: "f(i)", read: "v", ret: "int32"},
		{name: "method of an interface", pre: "\tvar f Finder = Box{l: list.New[int32]()}\n", call: "f.Find(i)", read: "v", ret: "int32"},
		{name: "library", call: "lib.Find(i)", read: "v", ret: "int32"},
		{name: "header of std: text.ParseInt", pre: "\ts := text.FromInt(i)\n", call: "text.ParseInt(s)", read: "v", ret: "int32", err: true},
		{name: "decimal.Parse", pre: "\ts := text.FromInt(i)\n", call: "decimal.Parse(s)", read: "decimal.String(v)", ret: "string"},
		{name: "list.Pop", pre: "\tl := list.New[int32]()\n\tl.Append(i)\n", call: "l.Pop()", read: "v", ret: "int32"},
		{name: "opt.Option.Get", pre: "\to := opt.Some(i)\n", call: "o.Get()", read: "v", ret: "int32"},
	}
	zero := map[string]string{"int32": "0", "string": `""`}
	var rows []optRow
	for _, c := range callees {
		bad := "!ok"
		if c.err {
			bad = "ok != nil"
		}
		rows = append(rows,
			optRow{name: c.name + " without a guard", val: "v", src: optCallees + `
func Run(i int32) ` + c.ret + ` {
` + c.pre + `	v, _ := ` + c.call + ` // at
	return ` + c.read + ` // at
}`},
			optRow{name: c.name + ": counter row with a guard", src: optCallees + `
func Run(i int32) ` + c.ret + ` {
` + c.pre + `	v, ok := ` + c.call + `
	if ` + bad + ` {
		return ` + zero[c.ret] + `
	}
	return ` + c.read + `
}`})
	}
	rows = append(rows,
		// the method of a type parameter
		optRow{name: "method of a type parameter without a guard", val: "v", src: optCallees + `
func first[F Finder](f F, i int32) int32 {
	v, _ := f.Find(i) // at
	return v // at
}

func Run(i int32) int32 {
	return first(Box{l: list.New[int32]()}, i)
}`},
		optRow{name: "method of a type parameter: counter row with a guard", src: optCallees + `
func first[F Finder](f F, i int32) int32 {
	if v, ok := f.Find(i); ok {
		return v
	}
	return 0
}

func Run(i int32) int32 {
	return first(Box{l: list.New[int32]()}, i)
}`},
		// big.QuoRem has three results, the last bool
		optRow{name: "big.QuoRem with three results without a guard", val: "q", src: `
func Run(a, b int32) string {
	q, r, _ := big.QuoRem(big.FromInt32(a), big.FromInt32(b)) // at
	return big.String(q) + big.String(r) // at
}`},
		optRow{name: "big.QuoRem: counter row with a guard", src: `
func Run(a, b int32) string {
	q, r, ok := big.QuoRem(big.FromInt32(a), big.FromInt32(b))
	if !ok {
		return ""
	}
	return big.String(q) + big.String(r)
}`},
	)
	optRun(t, rows)
}

// TestOptionalErrors holds the lines of the table about errors: (T, error) with _ and read before or in the body of
// if err != nil, (T, bool, error), throwing an error away (L-N11-7) with its counter rows and opt.OrZeroErr with a
// list (GS032), and an error, which is reassigned before each reading.
func TestOptionalErrors(t *testing.T) {
	const funcs = `
type P struct{ X int32 }

func load(s string) (int32, error) {
	if s == "" {
		return 0, errors.New("empty")
	}
	return 1, nil
}

func probe(i int32) (P, bool, error) {
	if i < 0 {
		return P{}, false, errors.New("negative")
	}
	if i == 0 {
		return P{}, false, nil
	}
	return P{X: i}, true, nil
}

func save(n int32) (int32, error) {
	if n < 0 {
		return 0, errors.New("negative")
	}
	return n, nil
}

func check(n int32) error {
	if n < 0 {
		return errors.New("negative")
	}
	return nil
}

func items(n int32) (list.List[int32], error) {
	if n < 0 {
		return list.New[int32](), errors.New("negative")
	}
	return list.Of(n), nil
}
`
	optRun(t, []optRow{
		// (T, error)
		{name: "(T, error): x, _ := f()", val: "x", src: funcs + `
func Run(s string) int32 {
	x, _ := load(s) // at
	return x // at
}`},
		{name: "(T, error): x read before if err != nil { return }", val: "x", src: funcs + `
func Run(s string) int32 {
	x, err := load(s) // at
	y := x + 1 // at
	if err != nil {
		return 0
	}
	return y
}`},
		{name: "(T, error): x read in the body of if err != nil", val: "x", src: funcs + `
func Run(s string) int32 {
	x, err := load(s) // at
	if err != nil {
		return x // at
	}
	return 0
}`},
		{name: "(T, error): counter row: x after the guard", src: funcs + `
func Run(s string) int32 {
	x, err := load(s)
	if err != nil {
		return 0
	}
	return x
}`},
		{name: "(T, error): counter row: x in the body of if err == nil", src: funcs + `
func Run(s string) int32 {
	x, err := load(s)
	if err == nil {
		return x
	}
	return 0
}`},

		// (T, bool, error)
		{name: "(T, bool, error): b, _, err := f(); if err != nil { return }; b.X", val: "b", src: funcs + `
func Run(i int32) int32 {
	b, _, err := probe(i) // at
	if err != nil {
		return -1
	}
	return b.X // at
}`},
		{name: "(T, bool, error): counter row: a guard on both", src: funcs + `
func Run(i int32) int32 {
	b, ok, err := probe(i)
	if err != nil || !ok {
		return -1
	}
	return b.X
}`},

		// throwing an error away (Answer „Verbieten, opt.OrZeroErr bleibt“)
		{name: "_, _ = f() with error", discard: true, src: funcs + `
func Run(n int32) int32 {
	_, _ = save(n) // at
	return n
}`},
		{name: "_ = f() with only error", discard: true, src: funcs + `
func Run(n int32) int32 {
	_ = check(n) // at
	return n
}`},
		{name: "a call as a statement, whose last result is error", discard: true, src: funcs + `
func Run(n int32) int32 {
	save(n) // at
	return n
}`},
		{name: "a call as a statement, whose only result is error", discard: true, src: funcs + `
func Run(n int32) int32 {
	check(n) // at
	return n
}`},
		{name: "a call of a header of std as a statement, whose last result is error", discard: true, src: `
func Run(s string) string {
	text.ParseInt(s) // at
	return s
}`},
		{name: "counter row: if _, err := f(); err != nil { return …, err }", src: funcs + `
func Run(n int32) (int32, error) {
	if _, err := save(n); err != nil {
		return 0, err
	}
	if err := check(n); err != nil {
		return 0, err
	}
	return n, nil
}`},
		{name: "counter row: y := opt.OrZeroErr(f()) with a result without nil", src: funcs + `
func Run(n int32) int32 {
	y := opt.OrZeroErr(save(n))
	return y
}`},
		{name: "counter row: opt.OrZeroErr(f()) with a list is GS032", code: gostrict.CodeNil, src: funcs + `
func Run(n int32) int32 {
	y := opt.OrZeroErr(items(n)) // at
	return y.Len()
}`},

		// an error, which is reassigned before each reading
		{name: "_, err = a(); _, err = b(); if err != nil { … }: at the first assignment", src: funcs + `
func Run(n int32) (int32, error) {
	var err error
	_, err = save(n) // at
	_, err = save(n + 1)
	if err != nil {
		return 0, err
	}
	return n, nil
}`},
		{name: "counter row: the assignments to err in separate cases, each checked", src: funcs + `
func Run(n, k int32) (int32, error) {
	var err error
	switch k {
	case 1:
		_, err = save(n)
		if err != nil {
			return 0, err
		}
	case 2:
		_, err = save(n + 1)
		if err != nil {
			return 0, err
		}
	}
	return n, nil
}`},
	})
}

// TestOptionalPassOn holds the lines of the table about passing a result on and about a call as an argument.
func TestOptionalPassOn(t *testing.T) {
	const decls = `
type P struct{ X int32 }

func check(t string) error {
	if t == "" {
		return errors.New("empty")
	}
	return nil
}

func show(v int32, ok bool) int32 {
	if ok {
		return v
	}
	return -1
}
`
	optRun(t, []optRow{
		{name: "return v, ok", src: decls + `
func Find(l list.List[int32], i int32) (int32, bool) {
	v, ok := l.Get(i)
	return v, ok
}`},
		{name: "return v, false", src: decls + `
func Find(l list.List[int32], i int32) (int32, bool) {
	v, ok := l.Get(i)
	if !ok {
		return v, false
	}
	return v + 1, true
}`},
		{name: "return v, err", src: decls + `
func Load(s string) (int32, error) {
	v, err := text.ParseInt(s)
	return v, err
}`},
		{name: `return v, errors.New("…")`, src: decls + `
func Load(s string) (int32, error) {
	v, err := text.ParseInt(s)
	if err != nil {
		return v, errors.New("no number")
	}
	return v, nil
}`},
		{name: "return v, e with an error e in the body of the guard if e != nil", src: decls + `
func Load(s, t string) (int32, error) {
	v, err := text.ParseInt(s)
	if e := check(t); e != nil {
		return v, e
	}
	return v, err
}`},
		{name: "return g()", src: decls + `
func Find(l list.List[int32], i int32) (int32, bool) {
	return l.Get(i)
}

func Load(s string) (int32, error) {
	return text.ParseInt(s)
}`},
		{name: "counter row: return v.X, ok", val: "v", src: decls + `
func Find(l list.List[P], i int32) (int32, bool) {
	v, ok := l.Get(i) // at
	return v.X, ok // at
}`},
		{name: "counter row: return v, true", val: "v", src: decls + `
func Find(l list.List[int32], i int32, strict bool) (int32, bool) {
	v, ok := l.Get(i) // at
	if strict {
		return v, ok
	}
	return v, true // at
}`},
		{name: "counter row: return v, nil", val: "v", src: decls + `
func Load(s string, strict bool) (int32, error) {
	v, err := text.ParseInt(s) // at
	if strict {
		return v, err
	}
	return v, nil // at
}`},
		{name: "counter row: return a, ok2 with the ok of another call", val: "a", src: decls + `
func Find(l list.List[int32], i int32) (int32, bool) {
	a, ok := l.Get(i) // at
	_, ok2 := l.Get(i + 1)
	if !ok2 {
		return 0, ok
	}
	return a, ok2 // at
}`},
		{name: "counter row: return v, ok after _, ok = m.Get(k)", val: "v", src: decls + `
func Find(m dict.Map[string, int32], k string) (int32, bool) {
	v, ok := m.Get(k) // at
	_, ok = m.Get(k + "x")
	return v, ok // at
}`},

		// a call as an argument
		{name: "show(l.Get(5))", src: decls + `
func Run(l list.List[int32]) int32 {
	return show(l.Get(5)) // at
}`},
		{name: "counter row: opt.OrZero(l.Get(5))", src: decls + `
func Run(l list.List[int32]) int32 {
	return opt.OrZero(l.Get(5))
}`},
		{name: "counter row: if v, ok := l.Get(5); ok { show(v, ok) }", src: decls + `
func Run(l list.List[int32]) int32 {
	if v, ok := l.Get(5); ok {
		return show(v, ok)
	}
	return 0
}`},
	})
}

// TestOptionalIndexFacts holds the lines of the table about the sources of a fact of an index and their counter rows.
func TestOptionalIndexFacts(t *testing.T) {
	optRun(t, []optRow{
		// for i := a; i < n; i++ with a constant a >= 0 and n := L.Len() before
		{name: "for i := 0; i < n; i++ with n := L.Len()", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	n := L.Len()
	for i := int32(0); i < n; i++ {
		v, _ := L.Get(i)
		s += v
	}
	return s
}`},
		{name: "for i := 2; i < n; i++ with n := L.Len()", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	n := L.Len()
	for i := int32(2); i < n; i++ {
		v, _ := L.Get(i)
		s += v
	}
	return s
}`},
		{name: "counter row: a = -1", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	n := L.Len()
	for i := int32(-1); i < n; i++ {
		v, _ := L.Get(i) // at
		s += v // at
	}
	return s
}`},
		{name: "counter row: a is not constant", src: `
func Run(L list.List[int32], a int32) int32 {
	s := int32(0)
	n := L.Len()
	for i := a; i < n; i++ {
		v, _ := L.Get(i) // at
		s += v // at
	}
	return s
}`},
		{name: "counter row: n := M.Len() with an access to L in the counting loop", src: `
func Run(L, M list.List[int32]) int32 {
	s := int32(0)
	n := M.Len()
	for i := int32(0); i < n; i++ {
		v, _ := L.Get(i) // at
		s += v // at
	}
	return s
}`},
		{name: "counter row: i <= n", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	n := L.Len()
	for i := int32(0); i <= n; i++ {
		v, _ := L.Get(i) // at
		s += v // at
	}
	return s
}`},

		// for i := range n with n := L.Len() before, for i := range L.Len()
		{name: "for i := range n with n := L.Len()", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	n := L.Len()
	for i := range n {
		v, _ := L.Get(i)
		s += v
	}
	return s
}`},
		{name: "for i := range L.Len()", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	for i := range L.Len() {
		v, _ := L.Get(i)
		s += v
	}
	return s
}`},
		{name: "counter row: n := M.Len() with an access to L in range n", src: `
func Run(L, M list.List[int32]) int32 {
	s := int32(0)
	n := M.Len()
	for i := range n {
		v, _ := L.Get(i) // at
		s += v // at
	}
	return s
}`},

		// for i := L.Len() - 1; i >= 0; i--
		{name: "for i := L.Len() - 1; i >= 0; i--", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	for i := L.Len() - 1; i >= 0; i-- {
		v, _ := L.Get(i)
		s += v
	}
	return s
}`},
		{name: "counter row: for i := L.Len(); i >= 0; i--", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	for i := L.Len(); i >= 0; i-- {
		v, _ := L.Get(i) // at
		s += v // at
	}
	return s
}`},
		{name: "counter row: for i := M.Len() - 1; i >= 0; i-- with an access to L", src: `
func Run(L, M list.List[int32]) int32 {
	s := int32(0)
	for i := M.Len() - 1; i >= 0; i-- {
		v, _ := L.Get(i) // at
		s += v // at
	}
	return s
}`},

		// the body of if i >= 0 && i < L.Len() and after if i < 0 || i >= L.Len() { return … }
		{name: "the body of if i >= 0 && i < L.Len()", src: `
func Run(L list.List[int32], i int32) int32 {
	if i >= 0 && i < L.Len() {
		v, _ := L.Get(i)
		return v
	}
	return 0
}`},
		{name: "after if i < 0 || i >= L.Len() { return … }", src: `
func Run(L list.List[int32], i int32) int32 {
	if i < 0 || i >= L.Len() {
		return 0
	}
	v, _ := L.Get(i)
	return v
}`},
		{name: "counter row: i < L.Len() without i >= 0", src: `
func Run(L list.List[int32], i int32) int32 {
	if i < L.Len() {
		v, _ := L.Get(i) // at
		return v // at
	}
	return 0
}`},

		// for i := range L.All()
		{name: "for i := range L.All()", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	for i := range L.All() {
		v, _ := L.Get(i)
		s += v
	}
	return s
}`},

		// after L := list.Of(…) with k elements, Get(c) with a constant 0 <= c < k
		{name: "after L := list.Of(…) with 3 elements: Get(0) and Get(2)", src: `
func Run(a int32) int32 {
	L := list.Of(a, a+1, a+2)
	v, _ := L.Get(0)
	w, _ := L.Get(2)
	return v + w
}`},
		{name: "counter row: Get(3) after list.Of with 3 elements", src: `
func Run(a int32) int32 {
	L := list.Of(a, a+1, a+2)
	v, _ := L.Get(3) // at
	return v // at
}`},

		// the checks of the length against a constant k
		{name: "after if L.Len() != 3 { return … }: Get(2)", src: `
func Run(L list.List[int32]) int32 {
	if L.Len() != 3 {
		return 0
	}
	v, _ := L.Get(2)
	return v
}`},
		{name: "counter row: Get(3) after if L.Len() != 3 { return … }", src: `
func Run(L list.List[int32]) int32 {
	if L.Len() != 3 {
		return 0
	}
	v, _ := L.Get(3) // at
	return v // at
}`},
		{name: "after if L.Len() < 3 { return … }: Get(2)", src: `
func Run(L list.List[int32]) int32 {
	if L.Len() < 3 {
		return 0
	}
	v, _ := L.Get(2)
	return v
}`},
		{name: "counter row: Get(3) after if L.Len() < 3 { return … }", src: `
func Run(L list.List[int32]) int32 {
	if L.Len() < 3 {
		return 0
	}
	v, _ := L.Get(3) // at
	return v // at
}`},
		{name: "the body of if L.Len() == 3: Get(2)", src: `
func Run(L list.List[int32]) int32 {
	if L.Len() == 3 {
		v, _ := L.Get(2)
		return v
	}
	return 0
}`},
		{name: "counter row: Get(3) in the body of if L.Len() == 3", src: `
func Run(L list.List[int32]) int32 {
	if L.Len() == 3 {
		v, _ := L.Get(3) // at
		return v // at
	}
	return 0
}`},
		{name: "the body of if L.Len() >= 3: Get(2)", src: `
func Run(L list.List[int32]) int32 {
	if L.Len() >= 3 {
		v, _ := L.Get(2)
		return v
	}
	return 0
}`},
		{name: "counter row: Get(3) in the body of if L.Len() >= 3", src: `
func Run(L list.List[int32]) int32 {
	if L.Len() >= 3 {
		v, _ := L.Get(3) // at
		return v // at
	}
	return 0
}`},
		{name: "counter row: k is not constant", src: `
func Run(L list.List[int32], k int32) int32 {
	if L.Len() != k {
		return 0
	}
	v, _ := L.Get(2) // at
	return v // at
}`},

		// after if L.Len() == 0 { return … } for Get(0) and Get(L.Len()-1)
		{name: "after if L.Len() == 0 { return … }: Get(0) and Get(L.Len()-1)", src: `
func Run(L list.List[int32]) int32 {
	if L.Len() == 0 {
		return 0
	}
	v, _ := L.Get(0)
	w, _ := L.Get(L.Len() - 1)
	return v + w
}`},
		{name: "counter row: Get(1) after if L.Len() == 0 { return … }", src: `
func Run(L list.List[int32]) int32 {
	if L.Len() == 0 {
		return 0
	}
	v, _ := L.Get(1) // at
	return v // at
}`},

		// the path of L: a field of a local variable without a pointer; counter row: through a pointer
		{name: "a field path of a local variable without a pointer", src: `
type Box struct{ L list.List[int32] }

func Run(b Box) int32 {
	s := int32(0)
	n := b.L.Len()
	for i := range n {
		v, _ := b.L.Get(i)
		s += v
	}
	return s
}`},
		{name: "counter row: a field path through a pointer", src: `
type Box struct{ L list.List[int32] }

func Run(b *Box) int32 {
	s := int32(0)
	n := b.L.Len()
	for i := range n {
		v, _ := b.L.Get(i) // at
		s += v // at
	}
	return s
}`},
	})
}

// TestOptionalIndexKills holds the line of the table about the killing of a fact of an index: each killing is GS033
// with the line of the killing in the finding; counter rows: Append and Set in the body, Pop on a list of another
// concrete element type and a function literal as the argument of list.SortFunc.
func TestOptionalIndexKills(t *testing.T) {
	optRun(t, []optRow{
		{name: "n reassigned in the body", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	n := L.Len()
	for i := int32(0); i < n; i++ {
		v, _ := L.Get(i) // at
		s += v // at
		n = n + 1 // kill
	}
	return s
}`},
		{name: "L.Delete between n := L.Len() and the loop", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	n := L.Len()
	L.Delete(0) // kill
	for i := range n {
		v, _ := L.Get(i) // at
		s += v // at
	}
	return s
}`},
		{name: "L.Delete in the body", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	n := L.Len()
	for i := int32(0); i < n; i++ {
		v, _ := L.Get(i) // at
		s += v // at
		L.Delete(0) // kill
	}
	return s
}`},
		{name: "a call f(L) in the body", src: `
func drop(l list.List[int32]) {
	l.Delete(0)
}

func Run(L list.List[int32]) int32 {
	s := int32(0)
	n := L.Len()
	for i := range n {
		drop(L) // kill
		v, _ := L.Get(i) // at
		s += v // at
	}
	return s
}`},
		{name: "counter row: a call of an own function, whose arguments reach no list", src: `
func twice(n int32) int32 {
	return 2 * n
}

func Run(L list.List[int32]) int32 {
	s := int32(0)
	n := L.Len()
	for i := range n {
		s += twice(i)
		v, _ := L.Get(i)
		s += v
	}
	return s
}`},
		{name: "Delete through an alias m := L", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	m := L
	n := L.Len()
	for i := range n {
		v, _ := L.Get(i) // at
		s += v // at
		m.Delete(0) // kill
	}
	return s
}`},
		{name: "Pop through another type parameter", src: `
func pick[U any](L list.List[int32], M list.List[U]) int32 {
	s := int32(0)
	n := L.Len()
	for i := range n {
		v, _ := L.Get(i) // at
		s += v // at
		M.Pop() // kill
	}
	return s
}

func Run(L list.List[int32], M list.List[string]) int32 {
	return pick(L, M)
}`},
		{name: "Pop through another type parameter, L of a type parameter", src: `
func pick[T any, U any](L list.List[T], M list.List[U]) list.List[T] {
	out := list.New[T]()
	n := L.Len()
	for i := range n {
		v, _ := L.Get(i) // at
		out.Append(v) // at
		M.Pop() // kill
	}
	return out
}

func Run(L list.List[int32], M list.List[string]) int32 {
	return pick(L, M).Len()
}`},
		{name: "counter row: Pop on a list of another concrete element type", src: `
func Run(L list.List[int32], M list.List[string]) int32 {
	s := int32(0)
	n := L.Len()
	for i := range n {
		v, _ := L.Get(i)
		s += v
		M.Pop()
	}
	return s
}`},
		{name: "an assignment to the field on the path", src: `
type Box struct{ L list.List[int32] }

func Run(b Box) int32 {
	s := int32(0)
	n := b.L.Len()
	for i := range n {
		v, _ := b.L.Get(i) // at
		s += v // at
		b.L = list.New[int32]() // kill
	}
	return s
}`},
		{name: "an access in a function literal, which runs later", src: `
func Run(L list.List[int32]) int32 {
	later := func() int32 { return 0 }
	n := L.Len()
	for i := range n {
		later = func() int32 {
			v, _ := L.Get(i) // at
			return v // at
		}
	}
	L.Delete(0)
	return later()
}`},
		{name: "counter row: Append in the body", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	n := L.Len()
	for i := range n {
		v, _ := L.Get(i)
		s += v
		L.Append(v)
	}
	return s
}`},
		{name: "counter row: Set in the body", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	n := L.Len()
	for i := range n {
		v, _ := L.Get(i)
		s += v
		L.Set(i, v+1)
	}
	return s
}`},
		{name: "counter row: a function literal as the argument of list.SortFunc", src: `
func Run(L list.List[int32], W list.List[string]) int32 {
	if L.Len() < 2 {
		return 0
	}
	s := int32(0)
	list.SortFunc(W, func(a, b string) int32 {
		v, _ := L.Get(1)
		s += v
		return 0
	})
	return s
}`},
	})
}

// TestOptionalKeyFacts holds the lines of the table about the facts of a key, about range L.All() with Delete and
// about the list.Frozen, which never shrinks.
func TestOptionalKeyFacts(t *testing.T) {
	optRun(t, []optRow{
		// the sources of a fact of a key
		{name: "for k := range D.Keys()", src: `
func Run(D dict.Map[string, int32]) int32 {
	s := int32(0)
	for k := range D.Keys() {
		v, _ := D.Get(k)
		s += v
	}
	return s
}`},
		{name: "for k := range D.All()", src: `
func Run(D dict.Map[string, int32]) int32 {
	s := int32(0)
	for k := range D.All() {
		v, _ := D.Get(k)
		s += v
	}
	return s
}`},
		{name: "the body of if D.Has(k)", src: `
func Run(D dict.Map[string, int32], k string) int32 {
	if D.Has(k) {
		v, _ := D.Get(k)
		return v
	}
	return 0
}`},
		{name: "counter row: Delete on a map of the same value type before the access", src: `
func Run(D, E dict.Map[string, int32], k string) int32 {
	if D.Has(k) {
		E.Delete(k) // kill
		v, _ := D.Get(k) // at
		return v // at
	}
	return 0
}`},
		{name: "counter row: if D.Has(k) with the access D.Get(k2)", src: `
func Run(D dict.Map[string, int32], k, k2 string) int32 {
	if D.Has(k) {
		v, _ := D.Get(k2) // at
		return v // at
	}
	return 0
}`},

		// range L.All() with Delete after and before the access and L = M after the access
		{name: "range L.All(), Delete after the access", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	for i := range L.All() {
		v, _ := L.Get(i)
		s += v
		L.Delete(0)
	}
	return s
}`},
		{name: "range L.All(), Delete before the access", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	for i := range L.All() {
		L.Delete(0) // kill
		v, _ := L.Get(i) // at
		s += v // at
	}
	return s
}`},
		{name: "range L.All(), L = M after the access", src: `
func Run(L, M list.List[int32]) int32 {
	s := int32(0)
	for i := range L.All() {
		v, _ := L.Get(i) // at
		s += v // at
		L = M // kill
	}
	return s
}`},

		// for i := range L.Len() { v, _ := L.Get(i); L.Delete(0) }; counter row with range L.All()
		{name: "for i := range L.Len() { v, _ := L.Get(i); L.Delete(0) }", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	for i := range L.Len() {
		v, _ := L.Get(i) // at
		s += v // at
		L.Delete(0) // kill
	}
	return s
}`},
		{name: "counter row: the same with range L.All()", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	for i := range L.All() {
		v, _ := L.Get(i)
		s += v
		L.Delete(0)
	}
	return s + 0
}`},

		// for k := range D.Keys() { v, _ := D.Get(k); D.Delete(k) }; the same with Delete before the access
		{name: "for k := range D.Keys() { v, _ := D.Get(k); D.Delete(k) }", src: `
func Run(D dict.Map[string, int32]) int32 {
	s := int32(0)
	for k := range D.Keys() {
		v, _ := D.Get(k)
		s += v
		D.Delete(k)
	}
	return s
}`},
		{name: "for k := range D.Keys() with Delete before the access", src: `
func Run(D dict.Map[string, int32]) int32 {
	s := int32(0)
	for k := range D.Keys() {
		D.Delete(k) // kill
		v, _ := D.Get(k) // at
		s += v // at
	}
	return s
}`},

		// an access in a loop in the body of if D.Has(k), Delete after the access in the same loop
		{name: "an access in a loop in the body of if D.Has(k), Delete after the access", src: `
func Run(D dict.Map[string, int32], k string) int32 {
	s := int32(0)
	if D.Has(k) {
		for range int32(2) {
			v, _ := D.Get(k) // at
			s += v // at
			D.Delete(k) // kill
		}
	}
	return s
}`},
		{name: "counter row: the same without Delete", src: `
func Run(D dict.Map[string, int32], k string) int32 {
	s := int32(0)
	if D.Has(k) {
		for range int32(2) {
			v, _ := D.Get(k)
			s += v
		}
	}
	return s
}`},

		// range F.All() over a list.Frozen, Delete on a list.List of the same element type in the body
		{name: "range F.All() over a list.Frozen, Delete on a list.List of the same element type before the access", src: `
func Run(F list.Frozen[int32], L list.List[int32]) int32 {
	s := int32(0)
	for i := range F.All() {
		L.Delete(0)
		v, _ := F.Get(i)
		s += v
	}
	return s
}`},
		{name: "counter row: the same over a list.List", src: `
func Run(F list.List[int32], L list.List[int32]) int32 {
	s := int32(0)
	for i := range F.All() {
		L.Delete(0) // kill
		v, _ := F.Get(i) // at
		s += v // at
	}
	return s
}`},
		{name: "counter row: an assignment to F in the body", src: `
func Run(F, G list.Frozen[int32]) int32 {
	s := int32(0)
	for i := range F.All() {
		v, _ := F.Get(i) // at
		s += v // at
		F = G // kill
	}
	return s
}`},
	})
}

// TestOptionalExamplesAndStd is the last line of the table for nago: each example of nago (internal/examples) and
// each part of std, which is written in gostrict (a package without //gostrict:header, checked as a program of a user)
// has no finding at all, so neither GS033 nor a conversion of R-SCRIPT-GS-FROZEN. The guides, seeds and tests of
// wokoda and the time tracking are checked in wokoda.
func TestOptionalExamplesAndStd(t *testing.T) {
	dirs, err := filepath.Glob("internal/examples/*")
	if err != nil || len(dirs) == 0 {
		t.Fatalf("no examples: %v", err)
	}
	for _, d := range dirs {
		t.Run(d, func(t *testing.T) {
			prog, err := gostrict.Load(d)
			if err != nil {
				t.Fatal(err)
			}
			if diags := gostrict.Check(prog); len(diags) > 0 {
				t.Errorf("the example has findings:\n%s", diags.Text())
			}
		})
	}
	stds, err := filepath.Glob("std/*")
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, d := range stds {
		files, _ := filepath.Glob(filepath.Join(d, "*.go"))
		fs := fstest.MapFS{}
		header := false
		for _, f := range files {
			if strings.HasSuffix(f, "_test.go") {
				continue
			}
			b, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			header = header || regexp.MustCompile(`(?m)^//gostrict:header\s*$`).Match(b)
			fs[filepath.Base(f)] = &fstest.MapFile{Data: b}
		}
		if header || len(fs) == 0 {
			continue
		}
		checked++
		t.Run(d, func(t *testing.T) {
			prog, err := gostrict.LoadModules("llm.local/std", gostrict.Module{Path: "llm.local/std", FS: fs})
			if err != nil {
				t.Fatal(err)
			}
			if diags := gostrict.Check(prog); len(diags) > 0 {
				t.Errorf("the part of std has findings as a program of a user:\n%s", diags.Text())
			}
		})
	}
	// premise: std has parts in gostrict, at least decimal
	if checked == 0 {
		t.Errorf("no part of std without //gostrict:header")
	}
}

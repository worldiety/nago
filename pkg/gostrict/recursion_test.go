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

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/gojaengine"
	"go.wdy.de/nago/pkg/gostrict/engine/vm"
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
)

// The tests of this file check the recursion of the terminating profile after the contract R-SCRIPT-GS-RECURSION of
// wokoda (L-N2-2 to L-N2-4): each row of its table of cases is a program of the subset with //gostrict:terminating,
// next to a counter row, which differs in the decisive property. The expectation is observed twice: the checker
// reports GS040 or nothing, and the program runs in goja and in the vm. A program with a finding runs without the
// directive and exhausts the depth of calls, a program without one ends with its result. So a finding is not only
// claimed, but the recursion exists, and a missing finding is not only claimed, but the program ends.

const recDirective = "//gostrict:terminating"

// recDepth limits the depth of calls in both engines. The programs without a finding call at most a few levels
// deep, so the limit separates an endless recursion from a deep one.
const recDepth = 400

// recCase is a program of the subset.
type recCase struct {
	name string
	// files maps the path of a file in the module llm.local/r to its source.
	files map[string]string
	// main is the package below llm.local/r, whose exported function entry runs, "" for the root.
	main  string
	entry string
	args  []any
	// want is the result of entry, if the program ends.
	want []any
	// finding lists the lines, where GS040 must be reported: each element is a set of alternatives, a line of a file
	// as "file|substring", one of which must have a finding. may lists further lines, which may have one, e.g. the
	// other functions of the cycle. A row without finding has no diagnostic at all.
	finding [][]string
	may     []string
	// loops tells, that the program calls itself endlessly. A row with a finding and without loops is the price
	// of L-N2-3: two functions of the same signature count as possible, while the program ends.
	loops bool
	// ui tells, that the program uses the header package lib/ui, whose host functions both engines need.
	ui bool
	// untranslatable is the part of the message of GS099, with which the translation rejects a program, which the
	// checker accepts, so that it can not run.
	untranslatable string
}

func (c recCase) module(directive bool) gostrict.Module {
	fsys := fstest.MapFS{}
	for name, src := range c.files {
		if !directive {
			src = strings.ReplaceAll(src, recDirective, "")
		}
		fsys[name] = &fstest.MapFile{Data: []byte(src)}
	}
	return gostrict.Module{Path: "llm.local/r", FS: fsys}
}

func (c recCase) mainPath() string {
	if c.main == "" {
		return "llm.local/r"
	}
	return "llm.local/r/" + c.main
}

// file returns the name of the file in diagnostics.
func recFile(name string) string {
	return "llm.local/r/" + name
}

// line returns the line of the file, which contains the substring, which must occur exactly once.
func (c recCase) line(t *testing.T, anchor string) (string, int) {
	t.Helper()
	file, sub, ok := strings.Cut(anchor, "|")
	if !ok {
		t.Fatalf("anchor %q has no file", anchor)
	}
	src, ok := c.files[file]
	if !ok {
		t.Fatalf("anchor %q names an unknown file", anchor)
	}
	res := 0
	for i, l := range strings.Split(src, "\n") {
		if strings.Contains(l, sub) {
			if res != 0 {
				t.Fatalf("anchor %q is ambiguous", anchor)
			}
			res = i + 1
		}
	}
	if res == 0 {
		t.Fatalf("anchor %q not found", anchor)
	}
	return recFile(file), res
}

func recCheck(t *testing.T, c recCase) gostrict.Diagnostics {
	t.Helper()
	prog, err := gostrict.LoadModules(c.mainPath(), c.module(true))
	if err != nil {
		t.Fatal(err)
	}
	return gostrict.Check(prog)
}

func recEngines() []struct {
	name string
	e    gostrict.Engine
} {
	return []struct {
		name string
		e    gostrict.Engine
	}{
		{"goja", &gojaengine.Engine{MaxCallStackSize: recDepth}},
		{"vm", &vm.Engine{MaxCallDepth: recDepth}},
	}
}

// recRun compiles the program without the directive and calls its entry in both engines. So the run observes the
// program and not the checker of the recursion. The program must compile without any diagnostic, so that GS040 is
// the only objection against a program with a finding.
func recRun(t *testing.T, c recCase) {
	t.Helper()
	b, diags, err := gostrict.NewCache(1).Compile(c.mainPath(), c.module(false))
	if err != nil {
		t.Fatal(err)
	}
	if c.untranslatable != "" {
		if len(diags) != 1 || diags[0].Code != gostrict.CodeInternal || !strings.Contains(diags[0].Msg, c.untranslatable) {
			t.Fatalf("want only GS099 %q, got:\n%s", c.untranslatable, diags.Text())
		}
		t.Logf("the program can not run: %s", diags[0].Msg)
		return
	}
	if len(diags) > 0 {
		t.Fatalf("the program does not compile without the directive:\n%s", diags.Text())
	}
	var host gostrict.Funcs
	if c.ui {
		host = ui.Host()
	}
	for _, e := range recEngines() {
		in, err := e.e.Load(context.Background(), b, host)
		if err != nil {
			t.Fatalf("%s: %v", e.name, err)
		}
		res, err := in.Call(context.Background(), c.entry, c.args...)
		_ = in.Close()
		switch {
		case c.loops && !errors.Is(err, gostrict.ErrStackExhausted):
			t.Errorf("%s: %s(%v) = %v, %v, want an exhausted stack", e.name, c.entry, c.args, res, err)
		case !c.loops && err != nil:
			t.Errorf("%s: %s(%v) failed: %v", e.name, c.entry, c.args, err)
		case !c.loops && c.want != nil && !gostrict.SameValues(res, c.want):
			t.Errorf("%s: %s(%v) = %v, want %v", e.name, c.entry, c.args, res, c.want)
		}
	}
}

// recExpect compares the diagnostics with the expected findings of the row.
func recExpect(t *testing.T, c recCase, diags gostrict.Diagnostics) {
	t.Helper()
	if len(c.finding) == 0 {
		if len(diags) > 0 {
			t.Errorf("want no diagnostic, got:\n%s", diags.Text())
		}
		return
	}
	type pos struct {
		file string
		line int
	}
	allowed := map[pos]bool{}
	for _, alts := range c.finding {
		found := false
		for _, a := range alts {
			f, l := c.line(t, a)
			allowed[pos{f, l}] = true
			for _, d := range diags {
				if d.Code == gostrict.CodeTermination && d.File == f && d.Line == l {
					found = true
				}
			}
		}
		if !found {
			t.Errorf("want GS040 at one of %v, got:\n%s", alts, diags.Text())
		}
	}
	for _, a := range c.may {
		f, l := c.line(t, a)
		allowed[pos{f, l}] = true
	}
	for _, d := range diags {
		switch {
		case d.Code != gostrict.CodeTermination:
			t.Errorf("unexpected diagnostic %s:%d %s %s", d.File, d.Line, d.Code, d.Msg)
		case !allowed[pos{d.File, d.Line}]:
			t.Errorf("GS040 outside the cycle at %s:%d: %s", d.File, d.Line, d.Msg)
		}
	}
}

// The use case x of CR-02 with a local function value and the views, which call it from the callback of a button.
const (
	cr02CaseValue = `// Package x is the use case x of CR-02 with a local function value.
//
//gostrict:terminating
package x

// Invoke runs the use case.
func Invoke() float64 {
	return invoke()
}

func invoke() float64 {
	return wert(2)
}

func wert(n float64) float64 {
	z := func(s string) float64 {
		return n
	}
	return z("n")
}
`
	cr02CaseNoValue = `// Package x is the use case x of CR-02 without a function value.
//
//gostrict:terminating
package x

// Invoke runs the use case.
func Invoke() float64 {
	return invoke()
}

func invoke() float64 {
	return wert(2)
}

func wert(n float64) float64 {
	return n
}
`
	// the literal of the function value calls the use case back, a recursion, which the signature finds
	cr02CaseRecursive = `// Package x is the use case x of CR-02, whose function value calls the use case.
//
//gostrict:terminating
package x

// Invoke runs the use case.
func Invoke() float64 {
	return invoke()
}

func invoke() float64 {
	return wert(2)
}

func wert(n float64) float64 {
	z := func(s string) float64 {
		return Invoke()
	}
	return z("n")
}
`
	// the function value has the signature func() of the callback of the button
	cr02CaseCallbackSignature = `// Package x is the use case x of CR-02, whose function value has the signature of a callback.
//
//gostrict:terminating
package x

// Invoke runs the use case.
func Invoke() float64 {
	return invoke()
}

func invoke() float64 {
	return wert(2)
}

func wert(n float64) float64 {
	z := func() {}
	z()
	return n
}
`
	cr02ViewY = `// Package y is a view, which calls the use case x from the callback of a button.
//
//gostrict:terminating
package y

import (
	"go.wdy.de/nago/pkg/gostrict/lib/ui"

	"llm.local/r/usecase/x"
)

func View() ui.View {
	return ui.PrimaryButton(func() {
		_ = x.Invoke()
	})
}
`
	cr02ViewX = `// Package x is a view with the name of the use case x, which calls it from the callback of a button.
//
//gostrict:terminating
package x

import (
	"go.wdy.de/nago/pkg/gostrict/lib/ui"

	uc "llm.local/r/usecase/x"
)

func View() ui.View {
	return ui.PrimaryButton(func() {
		_ = uc.Invoke()
	})
}
`
	// the view calls the use case also directly, so that the run passes through wert
	cr02ViewYDirect = `// Package y is a view, which calls the use case x directly and from the callback of a button.
//
//gostrict:terminating
package y

import (
	"go.wdy.de/nago/pkg/gostrict/lib/ui"

	"llm.local/r/usecase/x"
)

func View() ui.View {
	_ = x.Invoke()
	return ui.PrimaryButton(func() {
		_ = x.Invoke()
	})
}
`
)

// recSingle returns the files of a program of one package r in the file r.go.
func recSingle(src string) map[string]string {
	return map[string]string{"r.go": "// Package r is a program of the terminating profile.\n//\n//gostrict:terminating\npackage r\n\n" + src}
}

// recCases is the table of cases of R-SCRIPT-GS-RECURSION, each row next to its counter row.
var recCases = []recCase{
	// CR-02, four variants
	{
		name:  "CR-02 function value in the case, no view",
		files: map[string]string{"usecase/x/x.go": cr02CaseValue},
		main:  "usecase/x", entry: "Invoke", want: []any{2.0},
	},
	{
		name:  "CR-02 counter: the function value calls the case back",
		files: map[string]string{"usecase/x/x.go": cr02CaseRecursive},
		main:  "usecase/x", entry: "Invoke", loops: true,
		finding: [][]string{{"usecase/x/x.go|func wert("}},
		may:     []string{"usecase/x/x.go|func Invoke(", "usecase/x/x.go|func invoke(", "usecase/x/x.go|z := func("},
	},
	{
		name:  "CR-02 function value in the case, view y calls Invoke in the callback of a button",
		files: map[string]string{"usecase/x/x.go": cr02CaseValue, "view/y/y.go": cr02ViewY},
		main:  "view/y", entry: "View", ui: true,
	},
	{
		name:  "CR-02 function value in the case, view x calls Invoke in the callback of a button",
		files: map[string]string{"usecase/x/x.go": cr02CaseValue, "view/x/x.go": cr02ViewX},
		main:  "view/x", entry: "View", ui: true,
	},
	{
		name:  "CR-02 no function value in the case, view x calls Invoke in the callback of a button",
		files: map[string]string{"usecase/x/x.go": cr02CaseNoValue, "view/x/x.go": cr02ViewX},
		main:  "view/x", entry: "View", ui: true,
	},
	{
		name:  "CR-02 counter: the function value has the signature func() of the callback (price of L-N2-3)",
		files: map[string]string{"usecase/x/x.go": cr02CaseCallbackSignature, "view/y/y.go": cr02ViewYDirect},
		main:  "view/y", entry: "View", ui: true,
		finding: [][]string{{"usecase/x/x.go|func wert("}},
		may:     []string{"usecase/x/x.go|func Invoke(", "usecase/x/x.go|func invoke(", "usecase/x/x.go|z := func(", "view/y/y.go|func View(", "view/y/y.go|return ui.PrimaryButton("},
	},

	// TabsPage of CR-02
	{
		name: "TabsPage next to a local function value of another signature",
		files: recSingle(`import "go.wdy.de/nago/pkg/gostrict/lib/ui"

func View() ui.View {
	return ui.VStack(tabelle(), ui.Tabs(ui.TabsPage("A", func() ui.View {
		return tabelle()
	})))
}

func tabelle() ui.View {
	k := func(t string, rechts bool) ui.TTableColumn {
		return ui.TableColumn(ui.Text(t))
	}
	return ui.Table(k("a", false), k("b", true))
}
`),
		entry: "View", ui: true,
	},
	{
		name: "TabsPage counter: the function value has the signature of the body (price of L-N2-3)",
		files: recSingle(`import "go.wdy.de/nago/pkg/gostrict/lib/ui"

func View() ui.View {
	return ui.VStack(tabelle(), ui.Tabs(ui.TabsPage("A", func() ui.View {
		return tabelle()
	})))
}

func tabelle() ui.View {
	k := func() ui.View {
		return ui.Text("a")
	}
	return k()
}
`),
		entry: "View", ui: true,
		finding: [][]string{{"r.go|func tabelle("}},
		may:     []string{"r.go|func View(", "r.go|ui.TabsPage(", "r.go|k := func("},
	},

	// a cycle of literals alone
	{
		name: "literals: a literal calls itself through the variable",
		files: recSingle(`func Run() int32 {
	f := func(n int32) int32 { return n }
	f = func(n int32) int32 { return f(n) }
	return f(1)
}
`),
		entry: "Run", loops: true,
		finding: [][]string{{"r.go|f = func(", "r.go|func Run("}},
		may:     []string{"r.go|f := func("},
	},
	{
		name: "literals counter: literals of different signatures",
		files: recSingle(`func Run() int32 {
	f := func(n int32) int32 { return n }
	g := func(n int32, m int32) int32 { return f(n) + m }
	return g(1, 2)
}
`),
		entry: "Run", want: []any{int32(3)},
	},

	// range over a function
	{
		name: "range: the iterator calls the function, which ranges over it",
		files: recSingle(`func nums(yield func(int32) bool) {
	_ = Total()
}

func Total() int32 {
	s := int32(0)
	for x := range nums {
		s += x
	}
	return s
}
`),
		entry: "Total", loops: true,
		finding: [][]string{{"r.go|func Total("}},
		may:     []string{"r.go|func nums("},
	},
	{
		name: "range counter: an iterator, whose body does not call back",
		files: recSingle(`func nums(yield func(int32) bool) {
	for i := range int32(3) {
		if !yield(i) {
			return
		}
	}
}

func Total() int32 {
	s := int32(0)
	for x := range nums {
		s += x
	}
	return s
}
`),
		entry: "Total", want: []any{int32(3)},
	},
	{
		name: "range body: the iterator calls yield, the body calls the function",
		files: recSingle(`func nums(yield func(int32) bool) {
	_ = yield(1)
}

func Total() int32 {
	s := int32(0)
	for x := range nums {
		s += x + Total()
	}
	return s
}
`),
		entry: "Total", loops: true,
		finding: [][]string{{"r.go|func Total("}},
		may:     []string{"r.go|func nums(", "r.go|for x := range nums"},
	},
	{
		name: "range body counter: the iterator never calls yield, the body is a node of its own",
		files: recSingle(`func none(yield func(int32) bool) {
}

func Total() int32 {
	s := int32(0)
	for x := range none {
		s += x + Total()
	}
	return s
}
`),
		entry: "Total", want: []any{int32(0)},
	},

	// the initializer of a package variable
	{
		name: "package variable: the method calls the function of the variable, which calls the method",
		files: recSingle(`type shape interface{ area() int32 }

type sq struct{ n int32 }

func (q sq) area() int32 { return apply(q) }

var apply = func(s shape) int32 { return s.area() }

func Run() int32 { return sq{n: 2}.area() }
`),
		entry: "Run", loops: true,
		finding: [][]string{{"r.go|func (q sq) area("}},
		may:     []string{"r.go|var apply ="},
	},
	{
		name: "package variable counter: the method does not call the variable",
		files: recSingle(`type shape interface{ area() int32 }

type sq struct{ n int32 }

func (q sq) area() int32 { return q.n }

var apply = func(s shape) int32 { return s.area() }

func Run() int32 { return apply(sq{n: 2}) }
`),
		entry: "Run", want: []any{int32(2)},
	},
	{
		name: "package variable twice: the function passes itself to the function of the variable (found today)",
		files: recSingle(`var twice = func(f func(int32) int32, n int32) int32 { return f(f(n)) }

func Run(n int32) int32 { return twice(Run, n) }
`),
		entry: "Run", args: []any{int32(1)}, loops: true,
		finding: [][]string{{"r.go|func Run("}},
		may:     []string{"r.go|var twice ="},
	},
	{
		name: "package variable twice counter: another function is passed",
		files: recSingle(`var twice = func(f func(int32) int32, n int32) int32 { return f(f(n)) }

func double(n int32) int32 { return 2 * n }

func Run(n int32) int32 { return twice(double, n) }
`),
		entry: "Run", args: []any{int32(3)}, want: []any{int32(12)},
	},

	// a type parameter, whose core type is a function
	{
		name: "type parameter: the function passes itself as a type parameter, which is called",
		files: recSingle(`func apply[F ~func(int32) int32](f F, n int32) int32 { return f(n) }

func Run(n int32) int32 { return apply(Run, n) }
`),
		entry: "Run", args: []any{int32(1)}, loops: true,
		finding: [][]string{{"r.go|func Run("}},
		may:     []string{"r.go|func apply["},
	},
	{
		name: "type parameter counter: another function is passed",
		files: recSingle(`func apply[F ~func(int32) int32](f F, n int32) int32 { return f(n) }

func double(n int32) int32 { return 2 * n }

func Run(n int32) int32 { return apply(double, n) }
`),
		entry: "Run", args: []any{int32(3)}, want: []any{int32(6)},
	},

	// a method value of an interface
	{
		name: "method value: run calls the method value of the interface, the method calls run",
		files: recSingle(`type shape interface{ area() int32 }

type sq struct{ n int32 }

func (q sq) area() int32 { return run(q) }

func run(s shape) int32 {
	f := s.area
	return f()
}

func Run() int32 { return run(sq{n: 2}) }
`),
		entry: "Run", loops: true,
		finding: [][]string{{"r.go|func run("}, {"r.go|func (q sq) area("}},
	},
	{
		name: "method value counter: the method does not call run",
		files: recSingle(`type shape interface{ area() int32 }

type sq struct{ n int32 }

func (q sq) area() int32 { return q.n }

func run(s shape) int32 {
	f := s.area
	return f()
}

func Run() int32 { return run(sq{n: 2}) }
`),
		entry: "Run", want: []any{int32(2)},
	},

	// a function argument of a header, which calls it at once, and one, which calls it later
	{
		name: "callback argument: TBox.WithFrame calls the literal, which calls card",
		files: recSingle(`import "go.wdy.de/nago/pkg/gostrict/lib/ui"

func card(a int32) ui.View {
	return ui.BoxAlign(ui.Center, ui.Text("x")).WithFrame(func(fr ui.Frame) ui.Frame {
		_ = card(a)
		return fr
	})
}

func View() ui.View { return card(1) }
`),
		entry: "View", ui: true, loops: true,
		finding: [][]string{{"r.go|func card("}},
		may:     []string{"r.go|WithFrame(func("},
	},
	{
		name: "callback argument counter: the callback of a button runs later",
		files: recSingle(`import "go.wdy.de/nago/pkg/gostrict/lib/ui"

func render() ui.View {
	return ui.PrimaryButton(func() {
		_ = render()
	})
}

func View() ui.View { return render() }
`),
		entry: "View", ui: true,
	},

	// the exception for a decreasing argument
	{
		name: "decreasing: the function calls itself also through a function value with the same argument",
		files: recSingle(`func F(n int32) int32 {
	if n <= 0 {
		return 0
	}
	g := F
	return F(n-1) + g(n)
}
`),
		entry: "F", args: []any{int32(1)}, loops: true,
		finding: [][]string{{"r.go|func F("}},
	},
	{
		name: "decreasing counter: the function calls itself only statically with a decreasing argument",
		files: recSingle(`func F(n int32) int32 {
	if n <= 0 {
		return 0
	}
	return F(n-1) + n
}
`),
		entry: "F", args: []any{int32(3)}, want: []any{int32(6)},
	},

	// a function of the standard packages used as a value (found today)
	{
		name: "list.SortFunc as a value: cmp calls Sort",
		files: recSingle(`import "go.wdy.de/nago/pkg/gostrict/std/list"

func cmp(a, b int32) int32 {
	_ = Sort(list.Of(a, b))
	return a - b
}

func Sort(l list.List[int32]) int32 {
	s := list.SortFunc[int32]
	s(l, cmp)
	return l.Len()
}

func Run() int32 { return Sort(list.Of(int32(2), 1)) }
`),
		entry: "Run", loops: true, untranslatable: "index expressions are not supported",
		finding: [][]string{{"r.go|func Sort("}},
		may:     []string{"r.go|func cmp("},
	},
	{
		name: "list.SortFunc as a value counter: cmp does not call Sort",
		files: recSingle(`import "go.wdy.de/nago/pkg/gostrict/std/list"

func cmp(a, b int32) int32 {
	return a - b
}

func Sort(l list.List[int32]) int32 {
	s := list.SortFunc[int32]
	s(l, cmp)
	return l.Len()
}

func Run() int32 { return Sort(list.Of(int32(2), 1)) }
`),
		entry: "Run", want: []any{int32(2)}, untranslatable: "index expressions are not supported",
	},
	{
		name: "a function, which passes its function parameter to list.SortFunc, as a value: cmp calls Sort",
		files: recSingle(`import "go.wdy.de/nago/pkg/gostrict/std/list"

func cmp(a, b int32) int32 {
	_ = Sort(list.Of(a, b))
	return a - b
}

func sorter(l list.List[int32], c func(a, b int32) int32) {
	list.SortFunc(l, c)
}

func Sort(l list.List[int32]) int32 {
	s := sorter
	s(l, cmp)
	return l.Len()
}

func Run() int32 { return Sort(list.Of(int32(2), 1)) }
`),
		entry: "Run", loops: true,
		finding: [][]string{{"r.go|func Sort("}},
		may:     []string{"r.go|func cmp(", "r.go|func sorter("},
	},
	{
		name: "a function, which passes its function parameter to list.SortFunc, as a value counter: cmp does not call Sort",
		files: recSingle(`import "go.wdy.de/nago/pkg/gostrict/std/list"

func cmp(a, b int32) int32 {
	return a - b
}

func sorter(l list.List[int32], c func(a, b int32) int32) {
	list.SortFunc(l, c)
}

func Sort(l list.List[int32]) int32 {
	s := sorter
	s(l, cmp)
	return l.Len()
}

func Run() int32 { return Sort(list.Of(int32(2), 1)) }
`),
		entry: "Run", want: []any{int32(2)},
	},

	// a generic library function with a function parameter
	{
		name: "generic Map: the case passes a literal, which does not call back",
		files: map[string]string{
			"lib/lib.go": recMapLib,
			"fall/fall.go": `// Package fall uses the generic library.
//
//gostrict:terminating
package fall

import (
	"go.wdy.de/nago/pkg/gostrict/std/list"

	"llm.local/r/lib"
)

func Total(l list.List[string]) float64 {
	m := lib.Map(l, func(s string) float64 { return 1 })
	sum := float64(0)
	for _, v := range m.All() {
		sum += v
	}
	return sum
}

func Run() float64 { return Total(list.Of("a", "b")) }
`,
		},
		main: "fall", entry: "Run", want: []any{2.0},
	},
	{
		name: "generic Map counter: the literal calls the case back",
		files: map[string]string{
			"lib/lib.go": recMapLib,
			"fall/fall.go": `// Package fall uses the generic library.
//
//gostrict:terminating
package fall

import (
	"go.wdy.de/nago/pkg/gostrict/std/list"

	"llm.local/r/lib"
)

func Total(l list.List[string]) float64 {
	m := lib.Map(l, func(s string) float64 { return Total(list.Of(s)) })
	sum := float64(0)
	for _, v := range m.All() {
		sum += v
	}
	return sum
}

func Run() float64 { return Total(list.Of("a", "b")) }
`,
		},
		main: "fall", entry: "Run", loops: true,
		finding: [][]string{{"fall/fall.go|func Total("}},
		may:     []string{"lib/lib.go|func Map[", "fall/fall.go|m := lib.Map("},
	},
}

const recMapLib = `// Package lib is a generic library.
//
//gostrict:terminating
package lib

import "go.wdy.de/nago/pkg/gostrict/std/list"

// Map returns the results of f for the elements of l.
func Map[T, U any](l list.List[T], f func(T) U) list.List[U] {
	res := list.New[U]()
	for _, x := range l.All() {
		res.Append(f(x))
	}
	return res
}
`

// TestRecursionCases checks each row of the table of cases of R-SCRIPT-GS-RECURSION: the finding or its absence and
// the run, which shows, that the program calls itself endlessly or ends.
func TestRecursionCases(t *testing.T) {
	names := map[string]bool{}
	for _, c := range recCases {
		if names[c.name] {
			t.Fatalf("the name %q is not unique", c.name)
		}
		names[c.name] = true
		t.Run(c.name, func(t *testing.T) {
			t.Run("check", func(t *testing.T) {
				recExpect(t, c, recCheck(t, c))
			})
			t.Run("run", func(t *testing.T) {
				recRun(t, c)
			})
		})
	}
}

// TestRecursionCasesPaired makes sure, that the table has the counter rows: each row with a finding, which really
// loops, has a row without finding and each row without finding a row with one.
func TestRecursionCasesPaired(t *testing.T) {
	var with, without, price int
	for _, c := range recCases {
		switch {
		case len(c.finding) > 0 && c.loops:
			with++
		case len(c.finding) > 0:
			price++
		case c.loops:
			t.Errorf("%s: a program, which loops, without a finding", c.name)
		default:
			without++
		}
	}
	if with == 0 || without == 0 || price == 0 {
		t.Fatalf("rows with finding %d, without %d, price %d", with, without, price)
	}
	t.Logf("rows with finding %d, without %d, price of L-N2-3 %d", with, without, price)
}

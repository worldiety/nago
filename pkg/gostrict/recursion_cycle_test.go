// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"go.wdy.de/nago/pkg/gostrict"
)

// The tests of this file check the cycle, which GS040 reports, after L-N2-4 of the contract R-SCRIPT-GS-RECURSION:
// the message "<paket>.<f> may call itself: <glied> → … → <paket>.<f>", a member "<paket>.<f> (<datei>:<zeile>)" with
// the call site, which leads to the next member, a non static edge with its kind and expression, the fix for a
// function value and a shortest cycle with the first sequence of call sites on a tie. The contract leaves open,
// whether <paket> is the path or the name of the package and <datei> the whole name of the file or its base, so both
// are accepted, but consistently within a message.

// recDiag returns the finding GS040 at the line of the file, which contains the anchor.
func recDiag(t *testing.T, c recCase, diags gostrict.Diagnostics, anchor string) gostrict.Diagnostic {
	t.Helper()
	f, l := c.line(t, anchor)
	for _, d := range diags {
		if d.Code == gostrict.CodeTermination && d.File == f && d.Line == l {
			return d
		}
	}
	t.Fatalf("no GS040 at %s:%d (%s):\n%s", f, l, anchor, diags.Text())
	return gostrict.Diagnostic{}
}

// recForms returns the forms of the package and of the file, which a message may use: path and name, whole name and
// base.
func recForms(pkgPath, file string) (pkgs, files []string) {
	pkgs = []string{pkgPath, pkgPath[strings.LastIndex(pkgPath, "/")+1:]}
	files = []string{recFile(file), file[strings.LastIndex(file, "/")+1:]}
	return pkgs, files
}

// recMatch tells, whether the message matches the pattern for one of the forms of the package and the file. The
// pattern contains {P} for the package and {F} for the file.
func recMatch(msg, pattern, pkgPath, file string) bool {
	pkgs, files := recForms(pkgPath, file)
	for _, p := range pkgs {
		for _, f := range files {
			re := strings.ReplaceAll(regexp.QuoteMeta(pattern), regexp.QuoteMeta("{P}"), regexp.QuoteMeta(p))
			re = strings.ReplaceAll(re, regexp.QuoteMeta("{F}"), regexp.QuoteMeta(f))
			re = strings.ReplaceAll(re, regexp.QuoteMeta("{ANY}"), ".*")
			if regexp.MustCompile("^" + re + "$").MatchString(msg) {
				return true
			}
		}
	}
	return false
}

const fixFunctionValue = "call a named function instead of the function value "

// TestRecursionMessageStatic checks the whole message of a cycle of static calls and that its fix does not name a
// function value.
func TestRecursionMessageStatic(t *testing.T) {
	c := recCase{files: recSingle(`func ping(n int32) int32 {
	return pong(n)
}

func pong(n int32) int32 {
	return ping(n)
}
`)}
	diags := recCheck(t, c)
	_, lp := c.line(t, "r.go|return pong(n)")
	_, lq := c.line(t, "r.go|return ping(n)")
	for _, w := range []struct{ anchor, pattern string }{
		{"r.go|func ping(", fmt.Sprintf("{P}.ping may call itself: {P}.ping ({F}:%d) → {P}.pong ({F}:%d) → {P}.ping", lp, lq)},
		{"r.go|func pong(", fmt.Sprintf("{P}.pong may call itself: {P}.pong ({F}:%d) → {P}.ping ({F}:%d) → {P}.pong", lq, lp)},
	} {
		d := recDiag(t, c, diags, w.anchor)
		if !recMatch(d.Msg, w.pattern, "llm.local/r", "r.go") {
			t.Errorf("message\n  %s\nwant\n  %s", d.Msg, w.pattern)
		}
		if strings.HasPrefix(d.Fix, fixFunctionValue) {
			t.Errorf("the fix of a static cycle names a function value: %s", d.Fix)
		}
	}
}

// TestRecursionMessageFunctionValue checks the message of CR-03: the cycle of the use case through its local function
// value with the edge, which the signature resolves, and the fix, which names the function value. Its counter part
// is the static cycle of TestRecursionMessageStatic, whose fix does not.
func TestRecursionMessageFunctionValue(t *testing.T) {
	c := recCase{files: map[string]string{"usecase/x/x.go": cr02CaseRecursive}, main: "usecase/x"}
	diags := recCheck(t, c)
	_, li := c.line(t, "usecase/x/x.go|return invoke()")
	_, lw := c.line(t, "usecase/x/x.go|return wert(2)")
	_, lz := c.line(t, "usecase/x/x.go|return z(\"n\")")
	_, ldecl := c.line(t, "usecase/x/x.go|z := func(")
	_, llit := c.line(t, "usecase/x/x.go|return Invoke()")
	d := recDiag(t, c, diags, "usecase/x/x.go|func wert(")

	ok := false
	// the position of the function value is its call or its declaration, and the literal may carry the package
	for _, zpos := range []int{lz, ldecl} {
		for _, lit := range []string{"function literal", "{P}.function literal"} {
			pattern := fmt.Sprintf("{P}.wert may call itself: {P}.wert ({F}:%d) → via the function value z ({F}:%d), resolved by its signature to %s ({F}:%d) → {P}.Invoke ({F}:%d) → {P}.invoke ({F}:%d) → {P}.wert", lz, zpos, lit, llit, li, lw)
			if recMatch(d.Msg, pattern, "llm.local/r/usecase/x", "usecase/x/x.go") {
				ok = true
			}
		}
	}
	if !ok {
		t.Errorf("message\n  %s\nwant the cycle x.wert (x.go:%d) → via the function value z (x.go:%d), resolved by its signature to function literal (x.go:%d) → x.Invoke (x.go:%d) → x.invoke (x.go:%d) → x.wert", d.Msg, lz, lz, llit, li, lw)
	}
	if !strings.HasPrefix(d.Fix, fixFunctionValue+"z") {
		t.Errorf("fix %q, want the prefix %q", d.Fix, fixFunctionValue+"z")
	}
}

// TestRecursionMessageKinds checks, that each kind of non static edge stands in the message with its kind and its
// expression, and the names of the members, which are no named functions.
func TestRecursionMessageKinds(t *testing.T) {
	byName := map[string]recCase{}
	for _, c := range recCases {
		byName[c.name] = c
	}
	for _, w := range []struct {
		row, anchor string
		// contains are the parts of the message in their order
		contains []string
		// fix are the function values, one of which the fix names, none for a cycle without function value
		fix []string
	}{
		{"range: the iterator calls the function, which ranges over it", "r.go|func Total(", []string{"Total may call itself: ", "via the range", "nums"}, nil},
		{"range body: the iterator calls yield, the body calls the function", "r.go|func Total(", []string{"Total may call itself: ", "via the range", "nums", "via the function value yield", "range body"}, []string{"yield"}},
		{"package variable: the method calls the function of the variable, which calls the method", "r.go|func (q sq) area(", []string{"area may call itself: ", "via the function value apply", "via the interface method", "s.area"}, []string{"apply"}},
		{"method value: run calls the method value of the interface, the method calls run", "r.go|func run(", []string{"run may call itself: ", "via the function value f", "area"}, []string{"f"}},
		{"callback argument: TBox.WithFrame calls the literal, which calls card", "r.go|func card(", []string{"card may call itself: ", "via the callback argument", "function literal"}, nil},
		{"literals: a literal calls itself through the variable", "", []string{"function literal", "via the function value f"}, []string{"f"}},
		{"type parameter: the function passes itself as a type parameter, which is called", "r.go|func Run(", []string{"Run may call itself: ", "apply", "via the function value f", "Run"}, []string{"f"}},
		{"package variable twice: the function passes itself to the function of the variable (found today)", "r.go|func Run(", []string{"Run may call itself: ", "via the function value twice", "via the function value f", "Run"}, []string{"twice", "f"}},
	} {
		t.Run(w.row, func(t *testing.T) {
			c, ok := byName[w.row]
			if !ok {
				t.Fatalf("no row %q", w.row)
			}
			diags := recCheck(t, c)
			var d gostrict.Diagnostic
			if w.anchor == "" {
				if len(diags) == 0 {
					t.Fatal("no finding")
				}
				d = diags[0]
			} else {
				d = recDiag(t, c, diags, w.anchor)
			}
			rest := d.Msg
			for _, part := range w.contains {
				i := strings.Index(rest, part)
				if i < 0 {
					t.Fatalf("message\n  %s\nwant %q after the parts before it, in the order %q", d.Msg, part, w.contains)
				}
				rest = rest[i+len(part):]
			}
			named := ""
			for _, f := range w.fix {
				if strings.HasPrefix(d.Fix, fixFunctionValue+f) {
					named = f
				}
			}
			switch {
			case len(w.fix) > 0 && named == "":
				t.Errorf("fix %q, want the prefix %q and one of %q", d.Fix, fixFunctionValue, w.fix)
			case len(w.fix) == 0 && strings.HasPrefix(d.Fix, fixFunctionValue):
				t.Errorf("the fix names a function value, which is not in the cycle: %s", d.Fix)
			}
		})
	}
}

// TestRecursionShortestCycle checks, that the message shows a shortest cycle through the reported function, and on a
// tie the one, whose sequence of call sites comes first by file, line and column, not by name or declaration.
func TestRecursionShortestCycle(t *testing.T) {
	c := recCase{files: map[string]string{
		"a.go": `// Package r has several cycles through each function.
//
//gostrict:terminating
package r

// S has a cycle of three functions, whose call comes first, and one of two.
func S() int32 { return s1() + s2() }

func s1() int32 { return s3() }

func s3() int32 { return S() }

func s2() int32 { return S() }

// G calls both functions on the same line, the column decides.
func G() int32 { return bb() + ab() }

func ab() int32 { return G() }

func bb() int32 { return G() }

// H calls both functions on different lines, the line decides.
func H() int32 {
	x := hb()
	return x + ha()
}

func ha() int32 { return H() }

func hb() int32 { return H() }

// F calls one of two functions through a function value at the same site, the next call site decides by its file.
func F(c bool) int32 {
	z := zlate
	if c {
		z = aearly
	}
	return z()
}

func zlate() int32 { return F(false) }
`,
		"b.go": `package r

func aearly() int32 { return F(false) }
`,
	}}
	diags := recCheck(t, c)
	line := func(anchor string) int {
		_, l := c.line(t, anchor)
		return l
	}
	for _, w := range []struct {
		anchor, pattern string
	}{
		{"a.go|func S(", fmt.Sprintf("{P}.S may call itself: {P}.S ({F}:%d) → {P}.s2 ({F}:%d) → {P}.S", line("a.go|func S("), line("a.go|func s2("))},
		{"a.go|func G(", fmt.Sprintf("{P}.G may call itself: {P}.G ({F}:%d) → {P}.bb ({F}:%d) → {P}.G", line("a.go|func G("), line("a.go|func bb("))},
		{"a.go|func H(", fmt.Sprintf("{P}.H may call itself: {P}.H ({F}:%d) → {P}.hb ({F}:%d) → {P}.H", line("a.go|x := hb()"), line("a.go|func hb("))},
		// zlate in a.go comes before aearly in b.go, although its line is greater and its name comes later
		{"a.go|func F(", fmt.Sprintf("{P}.F may call itself: {P}.F ({F}:%d) → via the function value z ({ANY}), resolved by its signature to {P}.zlate ({F}:%d) → {P}.F", line("a.go|return z()"), line("a.go|func zlate("))},
	} {
		d := recDiag(t, c, diags, w.anchor)
		if !recMatch(d.Msg, w.pattern, "llm.local/r", "a.go") {
			t.Errorf("message\n  %s\nwant\n  %s", d.Msg, w.pattern)
		}
	}
	if line("b.go|func aearly(") >= line("a.go|func zlate(") {
		t.Fatal("the fixture needs aearly on an earlier line than zlate")
	}
}

// recLink is an expected member of a cycle: the function, the line of its call, which leads to the next member, the
// kind of the call and its expression.
type recLink struct {
	pkg, fn string
	// at is the anchor "file|substring" of the line of the call.
	at   string
	call gostrict.CallKind
	// expr is the expression of the call or, if it starts with "…", the end of it.
	expr string
}

// TestRecursionCycleStructure checks Diagnostic.Cycle: each member with the path of its package, its function,
// "function literal" or "range body", the file and line of the call, which leads to the next member, the kind of the
// call and its expression, starting at the reported function. A method may be named with its type.
func TestRecursionCycleStructure(t *testing.T) {
	byName := map[string]recCase{}
	for _, c := range recCases {
		byName[c.name] = c
	}
	const r, x = "llm.local/r", "llm.local/r/usecase/x"
	for _, w := range []struct {
		row, anchor string
		links       []recLink
	}{
		{"CR-02 counter: the function value calls the case back", "usecase/x/x.go|func wert(", []recLink{
			{x, "wert", `usecase/x/x.go|return z("n")`, gostrict.CallValue, "z"},
			{x, gostrict.LinkLiteral, "usecase/x/x.go|return Invoke()", gostrict.CallStatic, "Invoke"},
			{x, "Invoke", "usecase/x/x.go|return invoke()", gostrict.CallStatic, "invoke"},
			{x, "invoke", "usecase/x/x.go|return wert(2)", gostrict.CallStatic, "wert"},
		}},
		{"range: the iterator calls the function, which ranges over it", "r.go|func Total(", []recLink{
			{r, "Total", "r.go|for x := range nums", gostrict.CallRange, "nums"},
			{r, "nums", "r.go|_ = Total()", gostrict.CallStatic, "Total"},
		}},
		{"range body: the iterator calls yield, the body calls the function", "r.go|func Total(", []recLink{
			{r, "Total", "r.go|for x := range nums", gostrict.CallRange, "nums"},
			{r, "nums", "r.go|_ = yield(1)", gostrict.CallValue, "yield"},
			{r, gostrict.LinkRangeBody, "r.go|s += x + Total()", gostrict.CallStatic, "Total"},
		}},
		{"package variable: the method calls the function of the variable, which calls the method", "r.go|func (q sq) area(", []recLink{
			{r, "area", "r.go|func (q sq) area(", gostrict.CallValue, "apply"},
			{r, gostrict.LinkLiteral, "r.go|var apply =", gostrict.CallInterface, "s.area"},
		}},
		{"callback argument: TBox.WithFrame calls the literal, which calls card", "r.go|func card(", []recLink{
			{r, "card", "r.go|return ui.BoxAlign(", gostrict.CallCallback, "…WithFrame"},
			{r, gostrict.LinkLiteral, "r.go|_ = card(a)", gostrict.CallStatic, "card"},
		}},
		{"literals: a literal calls itself through the variable", "r.go|f = func(", []recLink{
			{r, gostrict.LinkLiteral, "r.go|f = func(", gostrict.CallValue, "f"},
		}},
		{"generic Map counter: the literal calls the case back", "fall/fall.go|func Total(", []recLink{
			{r + "/fall", "Total", "fall/fall.go|m := lib.Map(", gostrict.CallStatic, "lib.Map"},
			{r + "/lib", "Map", "lib/lib.go|for _, x := range l.All()", gostrict.CallRange, "l.All()"},
			{r + "/lib", gostrict.LinkRangeBody, "lib/lib.go|res.Append(f(x))", gostrict.CallValue, "f"},
			{r + "/fall", gostrict.LinkLiteral, "fall/fall.go|m := lib.Map(", gostrict.CallStatic, "Total"},
		}},
	} {
		t.Run(w.row, func(t *testing.T) {
			c, ok := byName[w.row]
			if !ok {
				t.Fatalf("no row %q", w.row)
			}
			d := recDiag(t, c, recCheck(t, c), w.anchor)
			if len(d.Cycle) != len(w.links) {
				t.Fatalf("cycle %+v, want %d links", d.Cycle, len(w.links))
			}
			for i, want := range w.links {
				got := d.Cycle[i]
				file, line := c.line(t, want.at)
				exprOK := got.Expr == want.expr
				if end, ok := strings.CutPrefix(want.expr, "…"); ok {
					exprOK = strings.HasSuffix(got.Expr, end)
				}
				fnOK := got.Func == want.fn || strings.HasSuffix(got.Func, "."+want.fn)
				if got.Package != want.pkg || !fnOK || got.File != file || got.Line != line || got.Call != want.call || !exprOK {
					t.Errorf("link %d is %+v, want %s %s at %s:%d, %s %q", i, got, want.pkg, want.fn, file, line, want.call, want.expr)
				}
				// the column points into the call on its line
				src := strings.Split(c.files[strings.TrimPrefix(file, "llm.local/r/")], "\n")[line-1]
				ident := strings.TrimSuffix(strings.TrimPrefix(want.expr, "…"), "()")
				ident = ident[strings.LastIndex(ident, ".")+1:]
				if got.Col < 1 || got.Col > len(src) || !strings.Contains(src[got.Col-1:], ident) {
					t.Errorf("link %d: column %d is not before %q in %q", i, got.Col, ident, src)
				}
				if got.Arity {
					t.Errorf("link %d: the signatures can be compared, but the link reaches by arity", i)
				}
			}
		})
	}
}

// TestRecursionCycleInMessage checks for each finding of a cycle in the table of cases, that Diagnostic.Cycle starts
// at the reported function and that the message shows its links in their order, so that a host, which composes its
// own message from Cycle, says the same.
func TestRecursionCycleInMessage(t *testing.T) {
	n := 0
	for _, c := range recCases {
		if len(c.finding) == 0 {
			continue
		}
		for _, d := range recCheck(t, c) {
			if d.Code != gostrict.CodeTermination {
				continue
			}
			n++
			if len(d.Cycle) == 0 {
				t.Errorf("%s: %s:%d has no cycle: %s", c.name, d.File, d.Line, d.Msg)
				continue
			}
			first := d.Cycle[0].Package + "." + d.Cycle[0].Func
			rest, ok := strings.CutPrefix(d.Msg, first+" may call itself: ")
			// the last call leads back to the first link, after "→" or after "to" of a call, which is not static
			if !ok || !(strings.HasSuffix(d.Msg, " → "+first) || strings.HasSuffix(d.Msg, " to "+first)) {
				t.Errorf("%s: the message does not start and end with %s: %s", c.name, first, d.Msg)
				continue
			}
			for _, l := range d.Cycle {
				part := fmt.Sprintf("%s.%s (%s:%d)", l.Package, l.Func, l.File, l.Line)
				i := strings.Index(rest, part)
				if i < 0 {
					t.Errorf("%s: the message misses %q in its order: %s", c.name, part, d.Msg)
					break
				}
				rest = rest[i+len(part):]
			}
		}
	}
	if n == 0 {
		t.Fatal("no finding")
	}
}

// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"regexp"
	"slices"
	"testing"

	"go.wdy.de/nago/pkg/gostrict"
)

// The way out of a finding of GS033 at a binding with = (wokoda R-SCRIPT-GS-OPTIONAL, "Meldung und Fix": the message
// names the way): if the ok resp. err is bound with = to a variable or a field, the way out names both targets, e.g.
// out.HasRevenue next to out.Revenue for out.Revenue, out.HasRevenue = r.Number("revenue") (found by the check of the
// surface in wokoda); for v, _ = … it names only v. The wording is the matter of the implementation: a row checks only,
// that each target stands in the way out (the Fix of the finding) as a name, and the counter rows, that the way out
// assigns to no other target than v.

// wayTargets are the targets of the assignments with = in a way out, e.g. v in "if t, ok := l.Get(i); ok { v = t }".
var wayTargets = regexp.MustCompile(`([A-Za-z_][\w.]*)\s*=[^=]`)

// wayRow is a program with the binding and the reading after it marked "// at" (the contract does not fix, at which
// of them the finding stands) and the targets, which the way out names.
type wayRow struct {
	name string
	src  string
	// targets must each stand in the way out as a name
	targets []string
	// only marks a counter row: the way out assigns to no target besides targets
	only bool
}

func wayEval(t *testing.T, r wayRow) {
	t.Helper()
	src := optProgram(r.src)
	at := optMarked(src, "at")
	var hits []gostrict.Diagnostic
	for _, d := range optCheck(t, src) {
		switch {
		case d.Code != gostrict.CodeGuard || d.File != optFile:
			t.Errorf("a finding besides GS033: %s:%d %s %s", d.File, d.Line, d.Code, d.Msg)
		case slices.Contains(at, d.Line):
			hits = append(hits, d)
		}
	}
	if len(hits) == 0 {
		t.Fatalf("no GS033 at the lines %v:\n%s", at, optNumbered(src))
	}
	// one finding names every target in its way out
	named := false
	for _, d := range hits {
		all := true
		for _, n := range r.targets {
			all = all && regexp.MustCompile(`(^|[^\w.])`+regexp.QuoteMeta(n)+`\b`).MatchString(d.Fix)
		}
		named = named || all
	}
	if !named {
		for _, d := range hits {
			t.Errorf("the way out does not name all of %v:\n%s", r.targets, optText(d))
		}
	}
	if r.only {
		for _, d := range hits {
			for _, m := range wayTargets.FindAllStringSubmatch(d.Fix, -1) {
				if !slices.Contains(r.targets, m[1]) {
					t.Errorf("the way out assigns to %s, want only %v:\n%s", m[1], r.targets, optText(d))
				}
			}
		}
	}
}

// TestOptionalWayOutTargets: out.Revenue, out.HasRevenue = l.Get(i), v, found = l.Get(i), the same with a function with
// (T, error) and with text.ParseInt, which is marked, name both targets in the way out; counter rows v, _ = l.Get(i),
// out.N, _ = f() and n, _ = g() with (T, error) name only the first target.
func TestOptionalWayOutTargets(t *testing.T) {
	for _, r := range []wayRow{
		{name: "out.Revenue, out.HasRevenue = l.Get(i): both fields", targets: []string{"out.Revenue", "out.HasRevenue"}, src: `
type O struct {
	Revenue    int32
	HasRevenue bool
}

func Run(l list.List[int32], i int32) O {
	var out O
	out.Revenue, out.HasRevenue = l.Get(i) // at
	return out // at
}`},
		{name: "v, found = l.Get(i): both variables", targets: []string{"v", "found"}, src: `
func Run(l list.List[int32], i int32) int32 {
	var v int32
	var found bool
	v, found = l.Get(i) // at
	if found {
		return v // at
	}
	return 0
}`},
		{name: "out.N, out.Err = f() with (T, error): both fields", targets: []string{"out.N", "out.Err"}, src: `
type O struct {
	N   int32
	Err error
}

func f() (int32, error) { return 1, nil }

func Run() O {
	var out O
	out.N, out.Err = f() // at
	return out // at
}`},
		{name: "n, failure = text.ParseInt(s), which is marked: both variables", targets: []string{"n", "failure"}, src: `
func Run(s string) int32 {
	var n int32
	var failure error
	n, failure = text.ParseInt(s) // at
	if failure != nil {
		return 0
	}
	return n // at
}`},
		{name: "counter row: v, _ = l.Get(i) names only v", targets: []string{"v"}, only: true, src: `
func Run(l list.List[int32], i int32) int32 {
	var v int32
	v, _ = l.Get(i) // at
	return v // at
}`},
		{name: "counter row: out.N, _ = f() with (T, bool) names only out.N", targets: []string{"out.N"}, only: true, src: `
type O struct {
	N int32
}

func f() (int32, bool) { return 1, true }

func Run() O {
	var out O
	out.N, _ = f() // at
	return out // at
}`},
		{name: "counter row: n, _ = g() with (T, error) names only n", targets: []string{"n"}, only: true, src: `
func g() (int32, error) { return 1, nil }

func Run() int32 {
	var n int32
	n, _ = g() // at
	return n // at
}`},
	} {
		t.Run(r.name, func(t *testing.T) { wayEval(t, r) })
	}
}

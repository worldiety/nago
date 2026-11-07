// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"testing"

	"go.wdy.de/nago/pkg/gostrict"
)

// The tests of this file check the contract R-SCRIPT-GS-FROZEN of wokoda (L-N11-5): no conversion, whose operand or
// target has the underlying type of list.List resp. list.Frozen or runs over a type parameter, whose type set holds
// such a type, except between identical types, and list.Freeze as a copy, which does not change with the list. The
// code and the message of the finding are the choice of the implementation, so a row only requires a finding at the
// line marked "// at" (neither GS000, which would mean, that the row is not valid Go, nor GS099); a counter row has no
// finding at all.

// frozenRows are the rows of the table of the detail.
var frozenRows = []optRow{
	{name: "list.Frozen[int32](l)", src: `
func Run(l list.List[int32]) int32 {
	f := list.Frozen[int32](l) // at
	return f.Len()
}`},
	{name: "list.List[int32](f)", src: `
func Run(f list.Frozen[int32]) int32 {
	l := list.List[int32](f) // at
	return l.Len()
}`},
	{name: "list.List[int32](f).Delete(0)", src: `
func Run(f list.Frozen[int32]) int32 {
	list.List[int32](f).Delete(0) // at
	return f.Len()
}`},
	{name: "type X list.Frozen[int32]: X(f)", src: `
type X list.Frozen[int32]

func Run(f list.Frozen[int32]) int32 {
	x := X(f) // at
	return frozenLen(x)
}

func frozenLen(x X) int32 {
	return 0
}`},
	{name: "type X list.Frozen[int32]: list.List[int32](X(f))", src: `
type X list.Frozen[int32]

func Run(f list.Frozen[int32]) int32 {
	l := list.List[int32](X(f)) // at
	return l.Len()
}`},
	{name: "type Y list.List[int32]: Y(f)", src: `
type Y list.List[int32]

func Run(f list.Frozen[int32]) int32 {
	y := Y(f) // at
	return yLen(y)
}

func yLen(y Y) int32 {
	return 0
}`},
	{name: "type Y list.List[int32]: list.List[int32](Y(f))", src: `
type Y list.List[int32]

func Run(f list.Frozen[int32]) int32 {
	l := list.List[int32](Y(f)) // at
	return l.Len()
}`},
	{name: "type Y list.List[int32]: Y(l)", src: `
type Y list.List[int32]

func Run(l list.List[int32]) int32 {
	y := Y(l) // at
	return yLen(y)
}

func yLen(y Y) int32 {
	return 0
}`},
	{name: "type Y list.List[int32]: list.Frozen[int32](Y(l))", src: `
type Y list.List[int32]

func Run(l list.List[int32]) int32 {
	f := list.Frozen[int32](Y(l)) // at
	return f.Len()
}`},
	{name: "func g[T list.List[int32] | list.Frozen[int32]](v T) with list.List[int32](v)", src: `
func g[T list.List[int32] | list.Frozen[int32]](v T) int32 {
	l := list.List[int32](v) // at
	return l.Len()
}

func Run(f list.Frozen[int32]) int32 {
	return g(f)
}`},
	{name: "counter row: list.Freeze(l)", src: `
func Run(l list.List[int32]) int32 {
	f := list.Freeze(l)
	return f.Len()
}`},
	{name: "counter row: list.List[int32](l) and list.Frozen[int32](f), identical types", src: `
func Run(l list.List[int32], f list.Frozen[int32]) int32 {
	a := list.List[int32](l)
	b := list.Frozen[int32](f)
	return a.Len() + b.Len()
}`},
}

// TestFrozenConversion checks the conversions of the table of R-SCRIPT-GS-FROZEN.
func TestFrozenConversion(t *testing.T) {
	for _, r := range frozenRows {
		t.Run(r.name, func(t *testing.T) {
			src := optProgram(r.src)
			at := optMarked(src, "at")
			diags := optCheck(t, src)
			if len(at) == 0 {
				if len(diags) > 0 {
					t.Errorf("want no finding, got:\n%s\n%s", diags.Text(), optNumbered(src))
				}
				return
			}
			found := false
			for _, d := range diags {
				if d.Code == gostrict.CodeTypeError || d.Code == gostrict.CodeInternal {
					t.Errorf("the row is not checked: %s:%d %s %s", d.File, d.Line, d.Code, d.Msg)
					continue
				}
				found = found || d.File == optFile && d.Line == at[0]
			}
			if !found {
				t.Errorf("no finding at line %d:\n%s\n%s", at[0], diags.Text(), optNumbered(src))
			}
		})
	}
}

// TestFrozenCopy checks, that list.Freeze copies: l.Delete(0) after f := list.Freeze(l) does not change f.Len(), in
// goja and in the vm; the counter value, the length of l after the Delete, shows, that the Delete happened.
func TestFrozenCopy(t *testing.T) {
	src := optProgram(`
func Run() (int32, int32, int32) {
	l := list.Of[int32](1, 2, 3)
	f := list.Freeze(l)
	l.Delete(0)
	l.Append(9)
	l.Delete(0)
	return f.Len(), l.Len(), f.GetOr(0, -1)
}`)
	got := optRunChecked(t, src, nil)
	want := []any{int32(3), int32(2), int32(1)}
	for _, e := range optEngines {
		if !gostrict.SameValues(got[e.name], want) {
			t.Errorf("%s: (f.Len(), l.Len(), f.GetOr(0, -1)) = %v, want %v", e.name, got[e.name], want)
		}
	}
}

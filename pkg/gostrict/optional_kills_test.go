// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import "testing"

// TestOptionalHeaderKills holds the lines of the table of R-SCRIPT-GS-OPTIONAL about the killing of a fact by a
// header (L-N11-2 after the supplement of wokoda 7add6b40, R-SCRIPT-GS-HEADER-LISTS): a header kills only through the
// functions, which it calls; a function literal as the argument of a header with //gostrict:calls now runs like the
// body of a loop, any other function value kills always, a header without a directive with a function parameter
// counts like now, one with //gostrict:calls later never kills, ui.RadioStateGroup.Notify (//gostrict:calls earlier)
// always; a range over a function value is its call; big.Int is a number; a call, in whose argument the access
// lies, runs after it. The rows use the invented header h of optHeader (Plain without a directive, Now, Later, Two),
// lib/ui and list.SortFunc. A row with a killing names its line "// kill", where the killing has one line.
func TestOptionalHeaderKills(t *testing.T) {
	const shrink = `
func shrink(l list.List[int32], v ui.TText) int32 {
	l.Delete(0)
	return 0
}
`
	optRun(t, []optRow{
		// L.Get(i) in the literal of ui.Text("x").WithFrame(…), the argument of shrink(L, …) on its own line
		{name: "WithFrame in the argument of shrink(L, …) in the body of a counting loop", src: shrink + `
func Run(L list.List[int32]) int32 {
	s, got := int32(0), int32(0)
	n := L.Len()
	for i := range n {
		s += shrink(L, // kill
			ui.Text("x").WithFrame(func(fr ui.Frame) ui.Frame {
				v, _ := L.Get(i) // at
				got = v // at
				return fr
			}))
	}
	return s + got
}`},
		{name: "counter row: the same call with L.Get(0) after if L.Len() == 0 { return … } without a loop", src: shrink + `
func Run(L list.List[int32]) int32 {
	if L.Len() == 0 {
		return 0
	}
	got := int32(0)
	s := shrink(L,
		ui.Text("x").WithFrame(func(fr ui.Frame) ui.Frame {
			v, _ := L.Get(0)
			got = v
			return fr
		}))
	return s + got
}`},
		{name: "counter row: the same call with L.Get(i) in for i := range L.All()", src: shrink + `
func Run(L list.List[int32]) int32 {
	s, got := int32(0), int32(0)
	for i := range L.All() {
		s += shrink(L,
			ui.Text("x").WithFrame(func(fr ui.Frame) ui.Frame {
				v, _ := L.Get(i)
				got = v
				return fr
			}))
	}
	return s + got
}`},

		// list.SortFunc(L, …) and WithFrame in the body of a counting loop over L, L.Get(i) before the call
		{name: "list.SortFunc(L, literal, which does not kill) after L.Get(i) in a counting loop", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	n := L.Len()
	for i := range n {
		v, _ := L.Get(i)
		s += v
		list.SortFunc(L, func(a, b int32) int32 { return a - b })
	}
	return s
}`},
		{name: "ui.Text(x).WithFrame(literal, which does not kill) after L.Get(i) in a counting loop", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	n := L.Len()
	for i := range n {
		v, _ := L.Get(i)
		s += v
		_ = ui.Text("x").WithFrame(func(fr ui.Frame) ui.Frame { return fr })
	}
	return s
}`},
		{name: "counter row: list.SortFunc with a function value instead of the literal", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	cmp := func(a, b int32) int32 { return a - b }
	n := L.Len()
	for i := range n {
		v, _ := L.Get(i) // at
		s += v // at
		list.SortFunc(L, cmp) // kill
	}
	return s
}`},
		{name: "counter row: WithFrame with a function value instead of the literal", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	frame := func(fr ui.Frame) ui.Frame { return fr }
	n := L.Len()
	for i := range n {
		v, _ := L.Get(i) // at
		s += v // at
		_ = ui.Text("x").WithFrame(frame) // kill
	}
	return s
}`},
		{name: "counter row: list.SortFunc with L.Delete(0) in the literal", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	n := L.Len()
	for i := range n {
		v, _ := L.Get(i) // at
		s += v // at
		list.SortFunc(L, func(a, b int32) int32 {
			L.Delete(0) // kill
			return a - b
		})
	}
	return s
}`},
		{name: "counter row: WithFrame with L.Delete(0) in the literal", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	n := L.Len()
	for i := range n {
		v, _ := L.Get(i) // at
		s += v // at
		_ = ui.Text("x").WithFrame(func(fr ui.Frame) ui.Frame {
			L.Delete(0) // kill
			return fr
		})
	}
	return s
}`},

		// in the body of for i := range L.All(): a literal of list.SortFunc and an inner loop, which do not hold the source
		{name: "range L.All(): L.Get(i) in the literal of list.SortFunc(W, …) with L.Delete(0) after the access", src: `
func Run(L list.List[int32], W list.List[string]) int32 {
	s := int32(0)
	for i := range L.All() {
		list.SortFunc(W, func(a, b string) int32 {
			v, _ := L.Get(i) // at
			s += v // at
			L.Delete(0) // kill
			return 0
		})
	}
	return s
}`},
		{name: "range L.All(): L.Get(i) in an inner for range 2 with L.Delete(0) after the access", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	for i := range L.All() {
		for range int32(2) {
			v, _ := L.Get(i) // at
			s += v // at
			L.Delete(0) // kill
		}
	}
	return s
}`},
		{name: "counter row: the literal of list.SortFunc(W, …) without Delete", src: `
func Run(L list.List[int32], W list.List[string]) int32 {
	s := int32(0)
	for i := range L.All() {
		list.SortFunc(W, func(a, b string) int32 {
			v, _ := L.Get(i)
			s += v
			return 0
		})
	}
	return s
}`},
		{name: "counter row: the inner for range 2 without Delete", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	for i := range L.All() {
		for range int32(2) {
			v, _ := L.Get(i)
			s += v
		}
	}
	return s
}`},
		{name: "counter row: L.Delete(0) after the call of list.SortFunc in the body of the range", src: `
func Run(L list.List[int32], W list.List[string]) int32 {
	s := int32(0)
	for i := range L.All() {
		list.SortFunc(W, func(a, b string) int32 {
			v, _ := L.Get(i)
			s += v
			return 0
		})
		L.Delete(0)
	}
	return s
}`},

		// a header without a directive, with now and with later (the invented header h)
		{name: "the literal with the access as the argument of a header without a directive: the fact ends", src: `
func Run(L list.List[int32]) int32 {
	if L.Len() == 0 {
		return 0
	}
	return h.Plain(func(x int32) int32 {
		v, _ := L.Get(0) // at
		return v + x // at
	})
}`},
		{name: "counter row: the same with a header with //gostrict:calls now and a literal, which does not kill", src: `
func Run(L list.List[int32]) int32 {
	if L.Len() == 0 {
		return 0
	}
	return h.Now(func(x int32) int32 {
		v, _ := L.Get(0)
		return v + x
	})
}`},
		{name: "a function value, which shortens L, as the argument of a header without a directive in a counting loop", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	drop := func(x int32) int32 {
		L.Delete(0)
		return x
	}
	n := L.Len()
	for i := range n {
		v, _ := L.Get(i) // at
		s += v // at
		s += h.Plain(drop) // kill
	}
	return s
}`},
		{name: "counter row: the same function value as the argument of a header with //gostrict:calls later", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	drop := func(x int32) int32 {
		L.Delete(0)
		return x
	}
	n := L.Len()
	for i := range n {
		v, _ := L.Get(i)
		s += v
		s += h.Later(drop)
	}
	return s
}`},
		{name: "a header with now and two function parameters: the access in the first literal, the second shortens L", src: `
func Run(L list.List[int32], i int32) int32 {
	drop := func(x int32) int32 {
		L.Delete(0)
		return x
	}
	if i >= 0 && i < L.Len() {
		return h.Two(func(x int32) int32 {
			v, _ := L.Get(i) // at
			return v + x // at
		}, drop)
	}
	return 0
}`},
		{name: "counter row: both literals do not kill", src: `
func Run(L list.List[int32], i int32) int32 {
	if i >= 0 && i < L.Len() {
		return h.Two(func(x int32) int32 {
			v, _ := L.Get(i)
			return v + x
		}, func(x int32) int32 {
			return x
		})
	}
	return 0
}`},

		// ui.RadioStateGroup.Notify calls the observers from Observe at once (//gostrict:calls earlier)
		{name: "g.Notify() of ui.RadioStateGroup in a counting loop, g.Observe with an observer, which shortens L", src: `
func Run(wnd ui.Window, L list.List[int32]) int32 {
	g := ui.AutoRadioStateGroup(wnd, "g", 3)
	g.Observe(func(idx int32) {
		L.Delete(0)
	})
	s := int32(0)
	n := L.Len()
	for i := range n {
		v, _ := L.Get(i) // at
		s += v // at
		g.Notify() // kill
	}
	return s
}`},
		{name: "counter row: g.SelectedIndex() at the same handle in the body", src: `
func Run(wnd ui.Window, L list.List[int32]) int32 {
	g := ui.AutoRadioStateGroup(wnd, "g", 3)
	g.Observe(func(idx int32) {
		L.Delete(0)
	})
	s := int32(0)
	n := L.Len()
	for i := range n {
		v, _ := L.Get(i)
		s += v
		s += g.SelectedIndex()
	}
	return s
}`},

		// a range over a function value is its call
		{name: "for x := range seq over a function value, which shortens L, in a counting loop", src: `
func Run(L list.List[int32]) int32 {
	seq := func(yield func(int32) bool) {
		L.Delete(0)
		yield(1)
	}
	s := int32(0)
	n := L.Len()
	for i := range n {
		v, _ := L.Get(i) // at
		s += v // at
		for x := range seq { // kill
			s += x
		}
	}
	return s
}`},
		{name: "counter row: for x := range L.All() in the body", src: `
func Run(L list.List[int32]) int32 {
	s := int32(0)
	n := L.Len()
	for i := range n {
		v, _ := L.Get(i)
		s += v
		for x := range L.All() {
			s += x
		}
	}
	return s
}`},

		// big.Int is a number
		{name: "a call of an own function with arguments of the types big.Int and decimal.Dec only", src: `
func calc(a big.Int, d decimal.Dec) int32 {
	return big.Sign(a) + decimal.Sign(d)
}

func Run(L list.List[int32]) int32 {
	s := int32(0)
	n := L.Len()
	for i := range n {
		v, _ := L.Get(i)
		s += v
		s += calc(big.FromInt32(i), decimal.FromInt32(i))
	}
	return s
}`},
		{name: "counter row: the same call with L as a further argument", src: `
func calc(a big.Int, d decimal.Dec, l list.List[int32]) int32 {
	return big.Sign(a) + decimal.Sign(d) + l.Len()
}

func Run(L list.List[int32]) int32 {
	s := int32(0)
	n := L.Len()
	for i := range n {
		v, _ := L.Get(i) // at
		s += v // at
		s += calc(big.FromInt32(i), decimal.FromInt32(i), L) // kill
	}
	return s
}`},
		{name: "counter row: drop(d) through a function value with one decimal argument only", src: `
func Run(L list.List[int32]) int32 {
	drop := func(d decimal.Dec) int32 {
		return decimal.Sign(d)
	}
	s := int32(0)
	n := L.Len()
	for i := range n {
		v, _ := L.Get(i) // at
		s += v // at
		s += drop(decimal.FromInt32(i)) // kill
	}
	return s
}`},
	})
}

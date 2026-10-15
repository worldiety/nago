// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package alias checks the value semantics of structs, whose values the translated code shares, if their type is
// never changed, and copies otherwise.
package alias

import (
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

// Line is never changed, so its values are shared.
type Line struct {
	Width float64
	Color string
}

// Style is changed through a local variable.
type Style struct {
	Bold bool
	Top  Line
}

// Cell is never changed, but holds a Style, which is.
type Cell struct {
	Text  string
	Style Style
}

// Outer is changed through a chain of fields, so Inner is as well.
type Outer struct {
	In Inner
	N  int32
}

type Inner struct {
	Deep Line2
}

type Line2 struct {
	W int32
}

// Node is changed through a pointer.
type Node struct {
	V    int32
	Line Line
}

// Pair is replaced as a whole, which copies into the existing value, but its Line is shared.
type Pair struct {
	A Line
	B int32
}

type Box[T any] struct {
	V T
}

type Shape interface{ isShape() }

type Circle struct{ R float64 }

func (Circle) isShape() {}

var thin = Line{Width: 1, Color: "black"}

var base = Style{Bold: false, Top: thin}

func Run() list.List[string] {
	out := list.New[string]()
	add := func(s string) { out.Append(s) }

	// a shared immutable value
	cells := list.New[Cell]()
	for i := range int32(3) {
		cells.Append(Cell{Text: text.FromInt(i), Style: base})
	}
	s := base
	s.Bold = true
	s.Top = Line{Width: 2, Color: "red"}
	c0, _ := cells.Get(0)
	add("style " + yes(c0.Style.Bold) + " " + c0.Style.Top.Color + " " + yes(s.Bold) + " " + s.Top.Color + " " + yes(base.Bold))

	// a chain of fields changes Outer and Inner
	o := Outer{In: Inner{Deep: Line2{W: 1}}}
	p := o
	p.In.Deep.W = 7
	q := o.In
	o.In.Deep.W = 3
	add("chain " + text.FromInt(o.In.Deep.W) + " " + text.FromInt(p.In.Deep.W) + " " + text.FromInt(q.Deep.W))

	// a pointer and a copy of its value
	n := &Node{V: 1, Line: thin}
	v := *n
	n.V = 2
	*n = Node{V: 5, Line: Line{Width: 3, Color: "blue"}}
	add("pointer " + text.FromInt(v.V) + " " + v.Line.Color + " " + text.FromInt(n.V) + " " + n.Line.Color + " " + thin.Color)

	// a whole assignment keeps the shared line of the former value
	pair := Pair{A: thin, B: 1}
	a := pair.A
	pair = Pair{A: Line{Width: 9, Color: "green"}, B: 2}
	add("pair " + a.Color + " " + pair.A.Color + " " + thin.Color)

	// the elements of a list are copies
	nodes := list.Of(Node{V: 1, Line: thin}, Node{V: 2, Line: thin})
	first, _ := nodes.Get(0)
	first.V = 10
	again, _ := nodes.Get(0)
	for _, x := range nodes.All() {
		x.V = 100
		_ = x
	}
	sum := int32(0)
	for x := range nodes.Values() {
		sum += x.V
	}
	add("list " + text.FromInt(first.V) + " " + text.FromInt(again.V) + " " + text.FromInt(sum))

	// a generic type, changed through a field
	b1 := Box[Line]{V: thin}
	b2 := b1
	b2.V = Line{Width: 4, Color: "white"}
	add("box " + b1.V.Color + " " + b2.V.Color)

	// := assigns an existing variable
	keep := thin
	keep, ok := pick(true)
	add("define " + keep.Color + " " + yes(ok) + " " + thin.Color)

	// range assigns existing variables
	var last Node
	for _, last = range nodes.All() {
	}
	last.V = 77
	again2, _ := nodes.Get(1)
	add("range " + text.FromInt(last.V) + " " + text.FromInt(again2.V))

	// a variant of a sealed interface
	var sh Shape = Circle{R: 1}
	switch c := sh.(type) {
	case Circle:
		c.R = 5
		add("variant " + text.FromFloat(c.R, 0))
	}
	switch c := sh.(type) {
	case Circle:
		add("variant " + text.FromFloat(c.R, 0))
	}
	return out
}

func pick(b bool) (Line, bool) {
	if b {
		return Line{Width: 8, Color: "pink"}, true
	}
	return thin, false
}

func yes(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// Order evaluates the fields of a literal in the order of the source, like Go, not in the order of the struct.
func Order() string {
	log := list.New[string]()
	next := func(s string) int32 {
		log.Append(s)
		return log.Len()
	}
	p := Pair2{B: next("b"), A: next("a")}
	return text.Join(log, ",") + " " + text.FromInt(p.A) + " " + text.FromInt(p.B)
}

type Pair2 struct {
	A, B int32
}

// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package core uses the statements and expressions of the subset without function literals, iterators and
// interfaces, e.g. for the core of an engine: comparisons of structs, named results, several results as
// arguments, the operations of lists, maps and options, labels and the forms of switch.
package core

import (
	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/opt"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

type Point struct {
	X, Y float64
}

type Label struct {
	Name string
	At   Point
}

// Style has more fields than fit into the registers of an engine and never changes, so engines may share it.
type Style struct {
	Font, Color, Background string
	Size, Weight, Spacing   float64
	Bold, Italic, Underline bool
	Border                  Point
}

var (
	plain  = Style{Font: "sans", Size: 12}
	strong = Style{Font: "sans", Size: 12, Bold: true}
	primes = list.Freeze(list.Of[int32](2, 3, 5, 7, 11, 13))
)

func yes(b bool) string {
	if b {
		return "y"
	}
	return "n"
}

// Equal compares structs, also with strings, nested structs, shared structs and NaN.
func Equal(x, y float64) string {
	a := Label{Name: "a", At: Point{x, y}}
	b := Label{Name: "a", At: Point{x, y}}
	c := Label{Name: "b", At: Point{x, y}}
	s := strong
	s.Bold = false
	res := yes(a == b) + yes(a != c) + yes(a.At == c.At) + yes(Point{x, y} == Point{y, x})
	res += yes(plain == strong) + yes(plain == Style{Font: "sans", Size: 12}) + yes(s == plain)
	return res
}

// split returns several results, which a call passes on.
func split(n int32) (int32, int32) {
	return n / 10, n % 10
}

func join(tens, ones int32) string {
	return text.FromInt(tens) + "|" + text.FromInt(ones)
}

// stats has named results and returns them with a bare return.
func stats(l list.List[int32]) (sum, min, max int32, ok bool) {
	if l.Len() == 0 {
		return
	}
	min, max = 1<<31-1, -1<<31
	for _, v := range l.All() {
		sum += v
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	ok = true
	return
}

// Results passes several results on and returns named results.
func Results(n int32) (string, int32, int32, int32, bool, bool) {
	sum, lo, hi, ok := stats(list.Of(n, -n, 2*n, 7))
	_, _, _, empty := stats(list.New[int32]())
	return join(split(n)), sum, lo, hi, ok, empty
}

func byAge(a, b Label) int32 {
	if a.At.X < b.At.X {
		return -1
	}
	if a.At.X > b.At.X {
		return 1
	}
	return 0
}

// Lists uses the operations of lists of structs.
func Lists(n int32) string {
	l := list.Make[Point](n)
	for i := range n {
		l.Set(i, Point{float64(i), float64(n - i)})
	}
	last, ok := l.Pop()
	out := text.FromFloat(last.X, 0) + yes(ok)
	l.Delete(0)
	c := l.Clone()
	c.Set(0, Point{-1, -1})
	first, _ := l.Get(0)
	cfirst, _ := c.Get(0)
	_, missing := l.Get(99)
	out += "," + text.FromFloat(first.X, 0) + "," + text.FromFloat(cfirst.X, 0) + yes(missing) + "," + text.FromInt(l.Len())

	labels := list.Of(Label{"c", Point{3, 0}}, Label{"a", Point{1, 0}}, Label{"b", Point{1, 1}}, Label{"d", Point{0, 0}})
	list.SortFunc(labels, byAge)
	for _, lb := range labels.All() {
		out += lb.Name
	}
	for p := range primes.Values() {
		if p > n {
			break
		}
		out += "." + text.FromInt(p)
	}
	empty := list.New[Point]()
	_, gotEmpty := empty.Pop()
	return out + yes(gotEmpty)
}

// Maps uses maps with keys of each kind and changes them during an iteration.
func Maps(n int32) string {
	m := dict.New[uint32, string]()
	for i := range uint32(n) {
		m.Set(i*2654435761, text.FromInt(int32(i)))
	}
	m.Set(0, "zero")
	out := ""
	for k, v := range m.All() {
		if k%3 == 0 {
			m.Delete(k + 2654435761)
		}
		m.Set(k+1, "new")
		out += v + ";"
	}
	key := uint32(2654435761)
	v, ok := m.Get(key * 2)
	out += v + yes(ok) + yes(m.Has(1)) + text.FromInt(m.Len())

	words := dict.New[string, Point]()
	words.Set("b", Point{1, 2})
	words.Set("a", Point{3, 4})
	words.Set("b", Point{5, 6})
	for k := range words.Keys() {
		p, _ := words.Get(k)
		out += k + text.FromFloat(p.X+p.Y, 0)
	}
	words.Delete("b")
	_, has := words.Get("b")
	flags := dict.New[bool, int32]()
	flags.Set(true, 1)
	flags.Set(false, 2)
	flags.Set(true, 3)
	t, _ := flags.Get(true)
	return out + yes(has) + text.FromInt(words.Len()) + text.FromInt(t)
}

type Node struct {
	Value int32
	Next  opt.Option[*Node]
}

// Options holds pointers and structs in options.
func Options(n int32) string {
	head := opt.None[*Node]()
	for i := range n {
		head = opt.Some(&Node{Value: i, Next: head})
	}
	out := ""
	for cur := head; cur.IsSome(); {
		node, ok := cur.Get()
		if !ok {
			break
		}
		out += text.FromInt(node.Value)
		cur = node.Next
	}
	p := opt.Some(Point{1, 2})
	q := opt.None[Point]()
	pv, pok := p.Get()
	qv, qok := q.Get()
	d := q.Or(Point{7, 8})
	e := p.Or(Point{7, 8})
	s := opt.Some("x")
	return out + text.FromFloat(pv.Y+qv.Y+d.X+e.X, 0) + yes(pok) + yes(qok) + s.Or("y") + opt.None[string]().Or("z")
}

type Kind int32

const (
	Small Kind = iota
	Medium
	Large
)

func kindOf(n int32) Kind {
	switch {
	case n < 10:
		return Small
	case n < 100:
		return Medium
	}
	return Large
}

// Control uses labels, switch with and without tag, an initializer and break out of a switch.
func Control(n int32) string {
	out := ""
outer:
	for i := int32(0); i < n; i++ {
		switch k := kindOf(i * i); k {
		case Small:
			if i%2 == 0 {
				continue outer
			}
			out += "s"
		case Medium, Large:
			for j := range i {
				if j == 3 {
					continue outer
				}
				if i*j > 40 {
					break outer
				}
				out += "m"
			}
		default:
			out += "?"
		}
		switch i % 3 {
		case 0:
			out += "0"
			break
		case 1:
			out += "1"
		}
		out += "."
	}
	x := uint32(7)
	x -= 9
	y := int32(-7)
	y >>= 1
	z := ^x &^ 0xf0
	return out + text.FromInt(int32(x>>28)) + text.FromInt(y) + text.FromInt(int32(z))
}

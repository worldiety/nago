// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package funcs uses function literals, iterators, interfaces and generics in the corners, where an engine may
// differ: the variables of loops, which closures capture, branches out of the bodies of ranges over iterators,
// method values, pointer variants and errors of the program.
package funcs

import (
	"iter"

	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

// Captures captures the variables of loops, which each iteration has of its own, and a counter, which the closures
// change.
func Captures(n int32) string {
	fs := list.New[func() int32]()
	for i := int32(0); i < n; i++ {
		fs.Append(func() int32 { return i * i })
	}
	for j := range n {
		fs.Append(func() int32 { return -j })
	}
	count := int32(0)
	inc := func(d int32) int32 {
		count += d
		return count
	}
	out := ""
	for _, f := range fs.All() {
		out += text.FromInt(f()) + ","
	}
	inc(2)
	inc(3)
	// a closure sees the later assignment of the variable
	late := int32(1)
	seen := func() int32 { return late }
	late = 12
	adder := func(base int32) func(int32) int32 {
		return func(x int32) int32 {
			base += x
			return base
		}
	}
	a := adder(10)
	a(1)
	return out + text.FromInt(count) + "," + text.FromInt(seen()) + "," + text.FromInt(a(5))
}

// Grid yields the cells of a grid of n x n.
func Grid(n int32) iter.Seq2[int32, int32] {
	return func(yield func(int32, int32) bool) {
		for r := range n {
			for c := range n {
				if !yield(r, c) {
					return
				}
			}
		}
	}
}

func upTo(n int32) iter.Seq[int32] {
	return func(yield func(int32) bool) {
		for i := range n {
			if !yield(i) {
				return
			}
		}
	}
}

// Branches leaves the bodies of ranges over iterators with labeled branches and returns.
func Branches(n int32) (string, int32) {
	out := ""
rows:
	for i := range upTo(n) {
		for r, c := range Grid(n) {
			switch {
			case r == i && c == i:
				out += "d"
				continue rows
			case r+c > n+i:
				break rows
			case c == 0:
				continue
			}
			out += "."
		}
		out += "|"
	}
	total := int32(0)
	for v := range upTo(n) {
		for w := range upTo(n) {
			total += v * w
			if total > 50 {
				return out, total
			}
		}
	}
	return out, -total
}

// first returns the first even value with a named result and a bare return out of the body of a range.
func first(seq iter.Seq[int32]) (v int32, ok bool) {
	for x := range seq {
		if x%2 == 0 && x > 0 {
			v, ok = x, true
			return
		}
	}
	return
}

// Shape is a sum of a value and a pointer variant.
type Shape interface {
	Area() float64
	isShape()
}

type Square struct{ Side float64 }

func (s Square) Area() float64 { return s.Side * s.Side }
func (Square) isShape()        {}

type Blob struct{ Size float64 }

func (b *Blob) Area() float64  { return b.Size }
func (b *Blob) Grow(d float64) { b.Size += d }
func (*Blob) isShape()         {}

type Counter struct{ N int32 }

func (c *Counter) Inc() int32 {
	c.N++
	return c.N
}

func (c Counter) Get() int32 { return c.N }

// Problem is an error of the program.
type Problem struct{ Code int32 }

func (p Problem) Error() string { return "problem " + text.FromInt(p.Code) }

func check(n int32) error {
	if n > 3 {
		return Problem{Code: n}
	}
	return nil
}

// Pair is a generic type with a method.
type Pair[A, B any] struct {
	First  A
	Second B
}

func (p Pair[A, B]) Swap() Pair[B, A] {
	return Pair[B, A]{First: p.Second, Second: p.First}
}

func Map[T, R any](l list.List[T], f func(T) R) list.List[R] {
	res := list.New[R]()
	for v := range l.Values() {
		res.Append(f(v))
	}
	return res
}

// Methods uses method values, interfaces with pointer variants and errors of the program.
func Methods(n int32) string {
	b := &Blob{Size: 1}
	shapes := list.Of[Shape](Square{Side: 2}, b)
	grow := b.Grow
	grow(float64(n))
	area := 0.0
	for s := range shapes.Values() {
		area += s.Area()
	}
	var c Counter
	inc := c.Inc
	get := c.Get
	inc()
	inc()
	same := shapes.Len() == 2 && Shape(Square{Side: 2}) == Square{Side: 2}
	kinds := ""
	for _, s := range shapes.All() {
		switch s := s.(type) {
		case Square:
			kinds += "square" + text.FromFloat(s.Side, 0)
		case *Blob:
			s.Grow(1)
			kinds += "blob"
		}
	}
	errs := ""
	for i := range n {
		if err := check(i); err != nil {
			errs += err.Error() + ";"
		}
	}
	p := Pair[string, int32]{First: "a", Second: n}.Swap()
	lens := Map(list.Of("x", "yy", "zzz"), func(s string) int32 { return text.RuneCount(s) * n })
	ev, ok := first(upTo(n))
	return text.FromFloat(area, 1) + " " + text.FromInt(c.N) + " " + text.FromInt(get()) + " " + kinds + " " +
		text.FromFloat(b.Size, 0) + " " + errs + " " + text.FromInt(p.First) + p.Second + " " +
		text.FromInt(lens.Len()) + " " + text.FromInt(ev) + boolStr(ok) + boolStr(same)
}

func boolStr(b bool) string {
	if b {
		return "y"
	}
	return "n"
}

func pick(a, b int32) int32             { return a*10 + b }
func ignore(a int32, b int32) int32     { return b }
func twice(x int32) int32               { return x + x }
func corner(r Rect2) float64            { return r.Min.X + r.Max.Y }
func (p Point2) Plus(q Point2) Point2   { return Point2{p.X + q.X, p.Y + q.Y} }
func (p Point2) Scale(f float64) Point2 { return Point2{p.X * f, p.Y * f} }
func depth(n int32) int32 {
	if n <= 0 {
		return 0
	}
	return depth(n-1) + 1
}

type Point2 struct{ X, Y float64 }

type Rect2 struct{ Min, Max Point2 }

// Inlined calls small functions, which an engine may inline, with arguments, whose order of evaluation and side
// effects must stay.
func Inlined(n int32) string {
	count := int32(0)
	next := func() int32 {
		count++
		return count
	}
	a := pick(next(), next())
	b := ignore(next(), 5)
	c := twice(next())
	p := Point2{1, 2}
	p = p.Plus(p.Scale(float64(n))).Plus(Point2{X: 0.5})
	r := Rect2{Min: Point2{1, 2}, Max: p}
	return text.FromInt(a) + "," + text.FromInt(b) + "," + text.FromInt(c) + "," + text.FromInt(count) + "," +
		text.FromFloat(p.X, 1) + "," + text.FromFloat(p.Y, 1) + "," + text.FromFloat(corner(r), 1) + "," + text.FromInt(depth(n))
}

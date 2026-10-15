// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package generics sorts and maps with generic functions and generic methods.
package generics

import (
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/num"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

type Number interface {
	~int32 | ~uint32 | ~float64
}

func Sum[T Number](l list.List[T]) T {
	var s T
	for _, v := range l.All() {
		s += v
	}
	return s
}

func Max[T Number](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Seq wraps a list with generic methods.
type Seq[T any] struct {
	items list.List[T]
}

func From[T any](l list.List[T]) Seq[T] {
	return Seq[T]{items: l}
}

func (s Seq[T]) Map[U any](f func(T) U) Seq[U] {
	r := list.New[U]()
	for _, v := range s.items.All() {
		r.Append(f(v))
	}
	return Seq[U]{items: r}
}

func (s Seq[T]) Filter(keep func(T) bool) Seq[T] {
	r := list.New[T]()
	for v := range s.items.Values() {
		if keep(v) {
			r.Append(v)
		}
	}
	return Seq[T]{items: r}
}

func (s Seq[T]) List() list.List[T] {
	return s.items
}

type Pair[A, B any] struct {
	First  A
	Second B
}

func Zip[A, B any](a list.List[A], b list.List[B]) list.List[Pair[A, B]] {
	r := list.New[Pair[A, B]]()
	for i, x := range a.All() {
		if y, ok := b.Get(i); ok {
			r.Append(Pair[A, B]{First: x, Second: y})
		}
	}
	return r
}

type Person struct {
	Name string
	Age  int32
}

func Run() string {
	people := list.Of(Person{"Ana", 31}, Person{"Bo", 25}, Person{"Cy", 31}, Person{"Dee", 19})
	list.SortFunc(people, func(a, b Person) int32 { return b.Age - a.Age })

	adults := From(people).Filter(func(p Person) bool { return p.Age > 20 })
	names := adults.Map(func(p Person) string { return p.Name + ":" + text.FromInt(p.Age) })
	ages := adults.Map[float64](func(p Person) float64 { return float64(p.Age) / 2 })

	out := text.Join(names.List(), ",")
	out += "|" + text.FromInt(Sum(list.Of[int32](1, 2, 3)))
	out += "|" + text.FromInt(num.Round32(Sum(ages.List())))
	out += "|" + text.FromInt(Max[int32](4, -9))

	for _, p := range Zip(list.Of("a", "b", "c"), list.Of[int32](1, 2)).All() {
		out += "|" + p.First + text.FromInt(p.Second)
	}
	return out
}

func SumFloats(a, b, c float64) float64 {
	return Sum(list.Of(a, b, c))
}

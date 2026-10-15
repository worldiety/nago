// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package values relies on the value semantics of structs, pointers and interfaces.
package values

import (
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

type Counter struct {
	N int32
}

func (c *Counter) Inc() {
	c.N++
}

func (c Counter) Get() int32 {
	return c.N
}

// Bump changes its copy only.
func (c Counter) Bump() int32 {
	c.N += 100
	return c.N
}

type Getter interface {
	Get() int32
}

type Pair struct {
	A, B Counter
}

type Registry struct {
	last *Counter
}

func (r *Registry) Keep(c *Counter) {
	r.last = c
}

func show(n int32) string {
	return text.FromInt(n) + " "
}

func Run() string {
	out := ""
	a := Counter{N: 1}
	b := a
	b.Inc()
	out += show(a.N) + show(b.N)

	p := &Counter{N: 5}
	p.Inc()
	out += show(p.N)

	var g Getter = a
	a.Inc()
	out += show(g.Get()) + show(a.Get())

	f := a.Get
	a.Inc()
	out += show(f()) + show(a.Bump()) + show(a.N)

	pr := Pair{A: a, B: b}
	pr2 := pr
	pr2.A.Inc()
	same := pr == pr2
	pr2.A = pr.A
	out += text.FromInt(boolInt(same)) + text.FromInt(boolInt(pr == pr2)) + " "

	*p = Counter{N: 10}
	l := list.Of(*p)
	p.Inc()
	if c, ok := l.Get(0); ok {
		out += show(c.N)
	}
	c0, _ := l.Get(0)
	c0.Inc()
	out += show(c0.N)

	r := &Registry{last: p}
	q := &Counter{N: 20}
	r.Keep(q)
	*q = Counter{N: 30}
	out += show(r.last.N)

	var getters = list.Of[Getter](a, *p, g)
	sum := int32(0)
	for _, x := range getters.All() {
		sum += x.Get()
	}
	out += show(sum)
	return out
}

func boolInt(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package trees builds recursive structures with opt.Option: a linked list and a binary search tree.
package trees

import (
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/opt"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

// Node is a node of a linked list.
type Node struct {
	Value int32
	Next  opt.Option[*Node]
}

// Tree is a node of a binary search tree.
type Tree struct {
	Key         int32
	Left, Right opt.Option[*Tree]
}

// Point is a value, which an option copies.
type Point struct {
	X, Y float64
}

func insert(t opt.Option[*Tree], key int32) opt.Option[*Tree] {
	n, ok := t.Get()
	if !ok {
		return opt.Some(&Tree{Key: key})
	}
	if key < n.Key {
		n.Left = insert(n.Left, key)
	} else if key > n.Key {
		n.Right = insert(n.Right, key)
	}
	return t
}

func walk(t opt.Option[*Tree], out list.List[string]) {
	n, ok := t.Get()
	if !ok {
		return
	}
	walk(n.Left, out)
	out.Append(text.FromInt(n.Key))
	walk(n.Right, out)
}

func depth(t opt.Option[*Tree]) int32 {
	n, ok := t.Get()
	if !ok {
		return 0
	}
	return 1 + max(depth(n.Left), depth(n.Right))
}

// Run builds a list and a tree and returns the sum of the list, the sorted keys and the depth of the tree.
func Run(n int32) (int32, string, int32) {
	head := opt.None[*Node]()
	for i := range n {
		head = opt.Some(&Node{Value: i * i, Next: head})
	}
	sum := int32(0)
	for cur := head; cur.IsSome(); {
		node, ok := cur.Get()
		if !ok {
			break
		}
		sum += node.Value
		cur = node.Next
	}

	tree := opt.None[*Tree]()
	x := uint32(12345)
	for range n {
		x ^= x << 13
		x ^= x >> 17
		x ^= x << 5
		tree = insert(tree, int32(x%1000))
	}
	keys := list.New[string]()
	walk(tree, keys)
	return sum, text.Join(keys, ","), depth(tree)
}

// Values checks, that an option copies a struct in and out.
func Values() (float64, float64, bool) {
	p := Point{X: 1, Y: 2}
	o := opt.Some(p)
	p.X = 10
	q, _ := o.Get()
	q.Y = 20
	r := o.Or(Point{})
	empty := opt.None[Point]()
	d := empty.Or(Point{X: -1})
	return r.X + d.X, r.Y, empty.IsSome()
}

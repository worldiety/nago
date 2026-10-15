// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package widgets is a header package like lib/ui: views are handles with builder methods, which a sealed interface
// unites, and their values are records, enums and variables of the package. Each function has a binding and a
// native body, so that the test compares both.
//
//gostrict:header
package widgets

import (
	"strconv"
	"strings"

	"go.wdy.de/nago/pkg/gostrict/std/js"
)

// View is any view of the package.
type View interface{ isView() }

// TLabel is a text.
type TLabel struct{ ref js.Ref }

// TBox lays out its children.
type TBox struct{ ref js.Ref }

func (TLabel) isView() {}
func (TBox) isView()   {}

// Align aligns the children of a box.
type Align int32

const (
	Start Align = iota
	Center
	End
)

// Insets are the distances to the border of a view.
type Insets struct {
	Top, Left float64
}

// All returns the insets with v on each side.
//
//gostrict:js (i, v) => ({ Top: v, Left: v })
func (i Insets) All(v float64) Insets {
	return Insets{Top: v, Left: v}
}

// Wide are wide insets.
//
//gostrict:js ({ Top: 0, Left: 100 })
var Wide Insets = Insets{Top: 0, Left: 100}

type node struct {
	kind     string
	text     string
	pad      Insets
	align    Align
	children []*node
	clicked  int32
	onClick  func() int32
}

func get(ref js.Ref) *node {
	n, _ := ref.Value().(*node)
	if n == nil {
		return &node{}
	}
	c := *n
	return &c
}

// Label returns a text.
//
//gostrict:js (t) => ({ k: "label", t: t, p: null, a: 0, cs: [], c: null })
func Label(text string) TLabel {
	return TLabel{ref: js.NewRef(&node{kind: "label", text: text})}
}

// Padding sets the insets.
//
//gostrict:js (l, p) => ({ ...l, p: p })
func (l TLabel) Padding(p Insets) TLabel {
	n := get(l.ref)
	n.pad = p
	return TLabel{ref: js.NewRef(n)}
}

// OnClick sets the function, which a click calls, and which returns a number for the test.
//
//gostrict:js (l, f) => ({ ...l, c: f })
func (l TLabel) OnClick(f func() int32) TLabel {
	n := get(l.ref)
	n.onClick = f
	return TLabel{ref: js.NewRef(n)}
}

// Box returns a box of the children.
//
//gostrict:js (children) => ({ k: "box", t: "", p: null, a: 0, cs: children.map((c) => c.v), c: null })
func Box(children ...View) TBox {
	n := &node{kind: "box"}
	for _, c := range children {
		switch c := c.(type) {
		case TLabel:
			n.children = append(n.children, get(c.ref))
		case TBox:
			n.children = append(n.children, get(c.ref))
		}
	}
	return TBox{ref: js.NewRef(n)}
}

// Align aligns the children.
//
//gostrict:js (b, a) => ({ ...b, a: a })
func (b TBox) Align(a Align) TBox {
	n := get(b.ref)
	n.align = a
	return TBox{ref: js.NewRef(n)}
}

// Describe returns the tree of the view as text and clicks each label once.
//
//gostrict:js (v) => { const d = (n) => n.k + "(" + n.t + (n.p === null ? "" : " " + n.p.Top + "/" + n.p.Left) + " " + n.a + (n.c === null ? "" : " click " + n.c()) + n.cs.map((c) => " " + d(c)).join("") + ")"; return d(v.v); }
func Describe(v View) string {
	var n *node
	switch v := v.(type) {
	case TLabel:
		n = get(v.ref)
	case TBox:
		n = get(v.ref)
	}
	return describe(n)
}

func describe(n *node) string {
	var sb strings.Builder
	sb.WriteString(n.kind + "(" + n.text)
	if n.pad != (Insets{}) {
		sb.WriteString(" " + num(n.pad.Top) + "/" + num(n.pad.Left))
	}
	sb.WriteString(" " + strconv.Itoa(int(n.align)))
	if n.onClick != nil {
		sb.WriteString(" click " + strconv.Itoa(int(n.onClick())))
	}
	for _, c := range n.children {
		sb.WriteString(" " + describe(c))
	}
	sb.WriteString(")")
	return sb.String()
}

func num(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

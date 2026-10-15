// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package three

import (
	"fmt"
	"strings"
	"sync"

	"go.wdy.de/nago/pkg/gostrict/std/js"
)

// node is an object of the native scene graph.
type node struct {
	kind     string
	children []*node
	pos      [3]float64
	rot      [3]float64
	target   [3]float64
	scale    float64
	color    uint32
	aspect   float64
	geometry *node
	material *node
	name     string
	hidden   bool
	opacity  float64
	vertices [][3]float64
}

var (
	mu    sync.Mutex
	shown *node
)

func newNode(kind string) js.Ref {
	return js.NewRef(&node{kind: kind, scale: 1, opacity: 1})
}

func newMaterial(color uint32) Material {
	n := newNode("Material")
	nodeOf(n).color = color
	return Material{n}
}

func nodeOf(r js.Ref) *node {
	return r.Value().(*node)
}

func refOf[T Mesh | Group | Light | Camera | Scene | Line](h T) js.Ref {
	switch h := any(h).(type) {
	case Line:
		return h.ref
	case Mesh:
		return h.ref
	case Group:
		return h.ref
	case Light:
		return h.ref
	case Camera:
		return h.ref
	case Scene:
		return h.ref
	}
	panic("unreachable")
}

// Dump describes the scene, which the program has shown last, e.g. for a test.
func Dump() string {
	mu.Lock()
	defer mu.Unlock()
	if shown == nil {
		return ""
	}
	var sb strings.Builder
	dump(&sb, shown, 0)
	return sb.String()
}

func dump(sb *strings.Builder, n *node, depth int) {
	fmt.Fprintf(sb, "%s%s pos=%.2f rot=%.2f scale=%.2f", strings.Repeat("  ", depth), n.kind, n.pos, n.rot, n.scale)
	if n.material != nil {
		fmt.Fprintf(sb, " color=%06x", n.material.color)
	}
	if n.name != "" {
		fmt.Fprintf(sb, " name=%q", n.name)
	}
	if n.hidden {
		sb.WriteString(" hidden")
	}
	sb.WriteByte('\n')
	for _, c := range n.children {
		dump(sb, c, depth+1)
	}
}

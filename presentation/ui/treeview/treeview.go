// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package treeview

import (
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
)

type Node[T any, ID comparable] struct {
	ID                 ID
	Icon               core.SVG
	Expandable         bool
	Label              string
	AccessibilityLabel string
	Children           []*Node[T, ID]
	Data               T
}

func (n *Node[T, ID]) Walk(visit func(node *Node[T, ID]) bool) bool {
	if n == nil {
		return false
	}

	if !visit(n) {
		return false
	}

	for _, child := range n.Children {
		if !child.Walk(visit) {
			return false
		}
	}

	return true
}

// The TreeStateModel separates the expanded and selected state of the entire tree from the actual nodes.
// This makes it easy to refresh the tree in each rendering but keeps the expanded and selected states alive.
type TreeStateModel[ID comparable] struct {
	Expanded map[ID]bool
	Selected map[ID]bool
}

func (t *TreeStateModel[ID]) init() {
	if t.Expanded == nil {
		t.Expanded = map[ID]bool{}
	}

	if t.Selected == nil {
		t.Selected = map[ID]bool{}
	}
}

func (t TreeStateModel[ID]) FirstSelected() (ID, bool) {
	for k, v := range t.Selected {
		if v {
			return k, true
		}
	}

	var zero ID
	return zero, false
}

type TTreeView[T any, ID comparable] struct {
	root        *Node[T, ID]
	indentDP    float64
	action      func(*Node[T, ID])
	frame       ui.Frame
	state       *core.State[TreeStateModel[ID]]
	multiselect bool
}

func TreeView[T any, ID comparable](root *Node[T, ID], state *core.State[TreeStateModel[ID]]) TTreeView[T, ID] {
	return TTreeView[T, ID]{root: root, indentDP: 16, frame: ui.Frame{}.FullWidth(), state: state}
}

func (c TTreeView[T, ID]) Frame(frame ui.Frame) TTreeView[T, ID] {
	c.frame = frame
	return c
}

func (c TTreeView[T, ID]) Multiselect(multiselect bool) TTreeView[T, ID] {
	c.multiselect = multiselect
	return c
}

// Expand expands or collapses the node and all its children recursively.
func (c TTreeView[T, ID]) Expand(id ID, expanded bool) {
	var root *Node[T, ID]
	c.root.Walk(func(n *Node[T, ID]) bool {
		if n.ID == id {
			root = n
			return false
		}

		return true
	})

	state := c.state.Get()
	state.init()

	if root != nil {
		c.root.Walk(func(node *Node[T, ID]) bool {
			state.Expanded[node.ID] = expanded
			return true
		})
	}

	c.state.Set(state)
	c.state.Notify()
	c.state.Invalidate()
}

// Select selects the node and all its children recursively.
func (c TTreeView[T, ID]) Select(id ID, selected bool) {
	var root *Node[T, ID]
	c.root.Walk(func(n *Node[T, ID]) bool {
		if n.ID == id {
			root = n
			return false
		}

		return true
	})

	state := c.state.Get()
	state.init()

	if root != nil {
		c.root.Walk(func(node *Node[T, ID]) bool {
			state.Selected[node.ID] = selected
			return true
		})
	}

	c.state.Set(state)
	c.state.Notify()
	c.state.Invalidate()
}

func (c TTreeView[T, ID]) Render(ctx core.RenderContext) core.RenderNode {
	if c.root == nil {
		return ui.VStack().Render(ctx)
	}

	var tmp []core.View
	for _, child := range c.root.Children {
		tmp = append(tmp, c.renderNode(child, 0, nil)...)
	}

	return ui.VStack(
		tmp...,
	).Render(ctx)
}

func (c TTreeView[T, ID]) Action(fn func(*Node[T, ID])) TTreeView[T, ID] {
	c.action = fn
	return c
}

func (c TTreeView[T, ID]) renderNode(n *Node[T, ID], indent int, dst []core.View) []core.View {
	var ico core.SVG

	n.Expandable = n.Expandable || len(n.Children) > 0

	nExpanded := c.state.Get().Expanded[n.ID]
	nSelected := c.state.Get().Selected[n.ID]

	if n.Expandable && !nExpanded {
		ico = icons.ChevronRight
	} else if n.Expandable {
		ico = icons.ChevronDown
	}

	var bgColor ui.Color
	if nSelected {
		bgColor = ui.ColorInteractive
	}

	v := ui.HStack(
		ui.If(len(ico) == 0, ui.Space(ui.L24)),
		ui.If(len(ico) != 0, ui.ImageIcon(ico)),
		ui.If(len(n.Icon) != 0, ui.ImageIcon(n.Icon)),
		ui.Text(n.Label),
	).
		Alignment(ui.Leading).
		Action(func() {
			state := c.state.Get()
			state.init()

			if !c.multiselect {
				for k := range state.Selected {
					state.Selected[k] = false
				}
			}

			state.Selected[n.ID] = true
			state.Expanded[n.ID] = !state.Expanded[n.ID]
			c.state.Set(state)

			if c.action != nil {
				c.action(n)
			}

			c.state.Notify()
			c.state.Invalidate()
		}).
		BackgroundColor(bgColor).
		Padding(ui.Padding{Left: ui.L(float64(indent) * c.indentDP), Top: ui.L4, Bottom: ui.L4, Right: ui.L4}).
		Border(ui.Border{}.Radius(ui.L4)).
		Frame(c.frame).
		AccessibilityLabel(n.AccessibilityLabel)

	dst = append(dst, v)
	if nExpanded {
		for _, child := range n.Children {
			dst = c.renderNode(child, indent+1, dst)
		}
	}

	return dst
}

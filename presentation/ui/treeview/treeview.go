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

type Node[T any] struct {
	ID                 string
	Icon               core.SVG
	Expandable         bool
	Expanded           bool
	Label              string
	AccessibilityLabel string
	Children           []*Node[T]
	Data               T
	Selected           bool
}

// Select recursively selects or unselects the node and all its children.
func (n *Node[T]) Select(selected bool) {
	n.Selected = selected
	for _, child := range n.Children {
		child.Select(selected)
	}
}

// Expand recursively expands or collapses the node and all its children.
func (n *Node[T]) Expand(expanded bool) *Node[T] {
	n.Expanded = expanded
	for _, child := range n.Children {
		child.Expand(expanded)
	}

	return n
}

type TTreeView[T any] struct {
	root     *Node[T]
	indentDP float64
	action   func(*Node[T])
	frame    ui.Frame
}

func TreeView[T any](root *Node[T]) TTreeView[T] {
	return TTreeView[T]{root: root, indentDP: 16, frame: ui.Frame{}.FullWidth()}
}

func (c TTreeView[T]) Frame(frame ui.Frame) TTreeView[T] {
	c.frame = frame
	return c
}

func (c TTreeView[T]) Render(ctx core.RenderContext) core.RenderNode {
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

func (c TTreeView[T]) Action(fn func(*Node[T])) TTreeView[T] {
	c.action = fn
	return c
}

func (c TTreeView[T]) renderNode(n *Node[T], indent int, dst []core.View) []core.View {
	var ico core.SVG

	n.Expandable = n.Expandable || len(n.Children) > 0

	if n.Expandable && !n.Expanded {
		ico = icons.ChevronRight
	} else if n.Expandable {
		ico = icons.ChevronDown
	}

	var bgColor ui.Color
	if n.Selected {
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
			if c.action != nil {
				c.action(n)
			}
		}).
		BackgroundColor(bgColor).
		Padding(ui.Padding{Left: ui.L(float64(indent) * c.indentDP), Top: ui.L4, Bottom: ui.L4, Right: ui.L4}).
		Border(ui.Border{}.Radius(ui.L4)).
		Frame(c.frame).
		AccessibilityLabel(n.AccessibilityLabel)

	dst = append(dst, v)
	if n.Expanded {
		for _, child := range n.Children {
			dst = c.renderNode(child, indent+1, dst)
		}
	}

	return dst
}

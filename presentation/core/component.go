// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"go.wdy.de/nago/presentation/proto"
)

const Debug = false // TODO make be a compile time flagged const

type RenderContext interface {
	// Window returns the associated Window instance.
	Window() Window

	// MountCallback returns for non-nil funcs a pointer. This pointer is only unique for the current render state.
	// This means, that subsequent calls which result in the same structural ora tree, will have the same
	// pointers. This allows more efficient model deltas. The largest downside is, that an outdated frontend
	// may invoke the wrong callbacks.
	// All callbacks are removed between render calls.
	MountCallback(func()) proto.Ptr
}

type RenderNode = proto.Component

type View interface {
	Render(RenderContext) RenderNode
}

type ViewPadding struct {
	parent  View
	padding *proto.Padding
}

func NewViewPadding(parent View, padding *proto.Padding) ViewPadding {
	return ViewPadding{parent: parent, padding: padding}
}

func (p ViewPadding) Top(pad proto.Length) View {
	p.padding.Top = pad
	return p.parent
}

func (p ViewPadding) All(pad proto.Length) View {
	p.padding.Left = pad
	p.padding.Right = pad
	p.padding.Bottom = pad
	p.padding.Top = pad
	return p.parent
}

func (p ViewPadding) Vertical(pad proto.Length) View {
	p.padding.Bottom = pad
	p.padding.Top = pad
	return p.parent
}

func (p ViewPadding) Horizontal(pad proto.Length) View {
	p.padding.Left = pad
	p.padding.Right = pad
	return p.parent
}

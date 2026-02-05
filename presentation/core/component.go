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

// RenderView is a lazy delayed view factory, which immediately calls Render on the returned view
type RenderView func(wnd Window) View

func (f RenderView) Render(ctx RenderContext) RenderNode {
	return f(ctx.Window()).Render(ctx)
}

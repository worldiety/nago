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

	// MountCallback returns for non-nil funcs a pointer, which is unique within the scope and never reused. The
	// callback is only valid for the tree which is rendered right now: all callbacks are removed between render
	// calls, and a call with a pointer of a former tree is discarded, e.g. the second click of a double click.
	// It is never redirected to whatever is at the same position now.
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

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ssr

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// ssrRenderContext implements core.RenderContext for server-side rendering.
// MountCallback always returns 0 – no callbacks are needed in a static HTML pass.
type ssrRenderContext struct {
	wnd core.Window
}

// NewRenderContext creates a new SSR render context backed by the given window.
func NewRenderContext(wnd core.Window) core.RenderContext {
	return &ssrRenderContext{wnd: wnd}
}

func (c *ssrRenderContext) Window() core.Window              { return c.wnd }
func (c *ssrRenderContext) MountCallback(_ func()) proto.Ptr { return 0 }

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package webview

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/internal"
)

// TWebView is a basic component (WebView).
// It shows a html page in the context of your view.
// A Web-Renderer must use an IFrame to avoid any problems or sandboxing security flaws.
type TWebView struct {
	frame          ui.Frame
	src            core.URI
	raw            string
	allow          string
	title          string
	referrerpolicy string
}

func WebView() TWebView {
	return TWebView{}
}

func (c TWebView) Frame(frame ui.Frame) TWebView {
	c.frame = frame
	return c
}

func (c TWebView) Allow(allow string) TWebView {
	c.allow = allow
	return c
}

func (c TWebView) Title(title string) TWebView {
	c.title = title
	return c
}

func (c TWebView) ReferrerPolicy(referrerpolicy string) TWebView {
	c.referrerpolicy = referrerpolicy
	return c
}

func (c TWebView) Src(src core.URI) TWebView {
	c.src = src
	return c
}

// Raw html code which is passed into the frame.
func (c TWebView) Raw(raw string) TWebView {
	c.raw = raw
	return c
}

func (c TWebView) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.WebView{
		URI:            proto.URI(c.src),
		Frame:          internal.FrameToOra(c.frame),
		Title:          proto.Str(c.title),
		Allow:          proto.Str(c.allow),
		ReferrerPolicy: proto.Str(c.referrerpolicy),
		Raw:            proto.Str(c.raw),
	}
}

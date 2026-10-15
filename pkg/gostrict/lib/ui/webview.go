// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/webview"
)

// TWebView is a page, which is shown in the context of a view, like webview.TWebView of nago. See WebView.
type TWebView struct{ ref js.Ref }

func (TWebView) isView() {}

// WebView returns a web view, like webview.WebView of nago.
//
//gostrict:host
func WebView() TWebView {
	return TWebView{ref: wrap(webview.WebView())}
}

// Frame sets the frame of the web view.
//
//gostrict:host
func (c TWebView) Frame(frame Frame) TWebView {
	return TWebView{ref: wrap(get[webview.TWebView](c.ref).Frame(frame.nago()))}
}

// Allow sets the allow attribute of the frame.
//
//gostrict:host
func (c TWebView) Allow(allow string) TWebView {
	return TWebView{ref: wrap(get[webview.TWebView](c.ref).Allow(allow))}
}

// Title sets the title of the frame.
//
//gostrict:host
func (c TWebView) Title(title string) TWebView {
	return TWebView{ref: wrap(get[webview.TWebView](c.ref).Title(title))}
}

// ReferrerPolicy sets the referrer policy of the frame.
//
//gostrict:host
func (c TWebView) ReferrerPolicy(referrerpolicy string) TWebView {
	return TWebView{ref: wrap(get[webview.TWebView](c.ref).ReferrerPolicy(referrerpolicy))}
}

// Src sets the URI of the page.
//
//gostrict:host
func (c TWebView) Src(src URI) TWebView {
	return TWebView{ref: wrap(get[webview.TWebView](c.ref).Src(core.URI(src)))}
}

// Raw sets the HTML code, which is passed into the frame.
//
//gostrict:host
func (c TWebView) Raw(raw string) TWebView {
	return TWebView{ref: wrap(get[webview.TWebView](c.ref).Raw(raw))}
}

// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgssr

import (
	"bytes"
	"io/fs"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/dom"
	"go.wdy.de/nago/pkg/sitemap"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/web/ssr"
	"golang.org/x/text/language"
)

// ssrPage holds the result of a single SSR render pass.
type ssrPage struct {
	Body  []byte // the rendered <div id="app">…</div> HTML
	Title string // the window title extracted during rendering (may be empty)
}

// Enable integrates a server side render step into the delivered dynamic javascript stub. The primary
// goal is not essentially a 1:1 hydration stub but at least provide enough content for proper SEO indexing.
func Enable(cfg *application.Configurator, root core.NavigationPath) {
	var idxBuf []byte
	for _, fsys := range cfg.Filesystems() {
		idxBuf, _ = fs.ReadFile(fsys, "index.html")
		if len(idxBuf) > 0 {
			break
		}
	}

	if len(idxBuf) == 0 {
		slog.Error("no index.html found in any fsys, hydration is disabled.", "root", root)
		return
	}

	placeholder := []byte(`<div id="app"></div>`)
	if !bytes.Contains(idxBuf, placeholder) {
		slog.Error("index.html does not contain placeholder for hydration, hydration is disabled.", "root", root)
		return
	}

	path := string(root)
	if path == "." {
		path = ""
	}

	path = "/" + path

	cfg.AddSitemapURL(sitemap.URL{
		Loc:        cfg.ContextPathURI(path, nil),
		LastMod:    sitemap.NewW3CTime(time.Now()),
		ChangeFreq: sitemap.ChangeFreqAlways,
		Priority:   0.5,
	})

	slog.Info("hydration enabled", "root", root, "path", path)

	getAnonUserFn := std.Must(cfg.UserManagement()).UseCases.GetAnonUser

	cfg.HandleMethod(http.MethodGet, path, func(writer http.ResponseWriter, request *http.Request) {
		page := renderHydrationHTML(cfg, root, getAnonUserFn, request)
		buf := bytes.Replace(idxBuf, placeholder, page.Body, 1)
		if page.Title != "" {
			buf = bytes.Replace(buf, []byte(`<title>Nago</title>`), []byte(`<title>`+page.Title+`</title>`), 1)
		}
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		writer.Write(buf)
	})

}

// renderHydrationHTML performs a single SSR render pass for the given root view.
// It parses the Accept-Language header, builds a minimal ssrWindow, calls the
// registered factory, renders the proto tree to DOM nodes and returns the result
// as an ssrPage containing the <div id="app">…</div> HTML and the window title.
func renderHydrationHTML(cfg *application.Configurator, root core.NavigationPath, getAnonUser user.GetAnonUser, r *http.Request) ssrPage {
	emptyPage := ssrPage{Body: []byte(`<div id="app"></div>`)}

	// 1. Parse locale from Accept-Language header.
	locale := language.English
	if al := r.Header.Get("Accept-Language"); al != "" {
		if tags, _, err := language.ParseAcceptLanguage(al); err == nil && len(tags) > 0 {
			locale = tags[0]
		}
	}

	// 2. Find the factory for the requested root view.
	id := proto.RootViewID(root)
	factory, ok := cfg.RootViews()[id]
	if !ok {
		slog.Error("SSR hydration: factory not found", "root", root)
		return emptyPage
	}

	// 3. Build a minimal SSR window + render context.
	wnd := ssr.NewWindow(locale, getAnonUser, root)
	ctx := ssr.NewRenderContext(wnd)

	// 4. Render the view tree to a proto.Component tree.
	var tree proto.Component
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				if cfg.IsDebug() {
					debug.PrintStack()
				}
				slog.Error("SSR hydration: panic during render", "panic", rec)
				tree = nil
			}
		}()
		view := factory(wnd)
		if view != nil {
			tree = view.Render(ctx)
		}
	}()

	if tree == nil {
		return emptyPage
	}

	// 5. Extract the window title set during rendering.
	var title string
	if tw, ok := wnd.(interface{ Title() string }); ok {
		title = tw.Title()
	}

	// 6. Convert proto tree → DOM nodes → HTML string.
	node := ssr.RenderComponent(tree)

	wrapper := dom.NewDiv()
	wrapper.SetAttr("id", "app")
	if node != nil {
		wrapper.AppendChild(node)
	}

	htmlStr, err := dom.RenderToString(wrapper)
	if err != nil {
		slog.Error("SSR hydration: failed to render DOM to string", "err", err)
		return emptyPage
	}

	return ssrPage{
		Body:  []byte(htmlStr),
		Title: title,
	}
}

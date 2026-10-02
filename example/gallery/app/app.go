// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package app is the shared part of the component galleries below example/gallery. Each gallery renders the
// components of one group of nago.dev on their own routes, e.g. /button. They are used by docs/screenshots to
// capture the component screenshots and are not tutorials.
//
// Keep each demo close to the snippet shown on the component page, so that image and code match.
package app

import (
	"slices"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

type demo struct {
	route core.NavigationPath
	view  func(wnd core.Window) core.View
}

var demos []demo

// Register adds a demo for the given route. Call it from an init function of the component file.
func Register(route core.NavigationPath, view func(wnd core.Window) core.View) {
	demos = append(demos, demo{route: route, view: view})
}

// Run starts the gallery with all registered demos. The index page lists all routes.
func Run() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nago.gallery")
		cfg.Serve(vuejs.Dist())

		for _, d := range demos {
			// the padding leaves room for the screenshot crop around the demo
			cfg.RootView(d.route, func(wnd core.Window) core.View {
				return VStack(d.view(wnd)).Alignment(TopLeading).Padding(Padding{}.All(L48))
			})
		}

		cfg.RootView(".", func(wnd core.Window) core.View {
			routes := make([]core.NavigationPath, 0, len(demos))
			for _, d := range demos {
				routes = append(routes, d.route)
			}
			slices.Sort(routes)

			return VStack(
				ForEach(routes, func(r core.NavigationPath) core.View {
					return Link(wnd, string(r), "/"+string(r), "")
				})...,
			).Alignment(Leading).Gap(L8).Padding(Padding{}.All(L16))
		})
	}).Run()
}

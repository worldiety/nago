// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"strings"

	"go.wdy.de/nago/presentation/proto"
)

type Navigation interface {
	ForwardTo(id NavigationPath, values Values)
	ForwardToTarget(id NavigationPath, target string, values Values)
	BackwardTo(id NavigationPath, values Values)
	Back()
	Replace(id NavigationPath, values Values)
	ResetTo(id NavigationPath, values Values)
	Reload()

	// Open either launches the denoted application or opens the resources with the associated resource.
	// This incorporates APIs like
	//  - the HTML5 open function or
	//  - the win32 ShellExecute call or
	//  - the MacOS open command or
	//  - xdg-open or gnome-open on linux
	Open(resource URI)
}

type navigationController struct {
	destroyed bool
	scope     *Scope
}

func newNavigationController(scope *Scope) *navigationController {
	return &navigationController{
		scope: scope,
	}
}

func (n *navigationController) Window() Window {
	if n.scope != nil {
		return n.scope.allocatedRootView.UnwrapOr(nil)
	}

	return nil
}

func (n *navigationController) ForwardTo(id NavigationPath, values Values) {
	n.ForwardToTarget(id, "", values)
}

func (n *navigationController) ForwardToTarget(id NavigationPath, target string, values Values) {
	if n.destroyed {
		return
	}

	n.IgnoreNextInvalidation()

	n.scope.Publish(&proto.NavigationForwardToRequested{
		RootView: proto.RootViewID(id),
		Values:   values.proto(),
		Target:   proto.Str(target),
	})
}

func (n *navigationController) BackwardTo(id NavigationPath, values Values) {
	// TODO implement me in ora protocol
	n.ForwardTo(id, values)
}

func (n *navigationController) Back() {
	if n.destroyed {
		return
	}

	n.IgnoreNextInvalidation()

	n.scope.Publish(&proto.NavigationBackRequested{})
}

func (n *navigationController) Replace(id NavigationPath, values Values) {
	if n.destroyed {
		return
	}

	n.IgnoreNextInvalidation()

	n.scope.Publish(&proto.NavigationReplaceRequested{
		RootView: proto.RootViewID(id),
		Values:   values.proto(),
	})
}

func (n *navigationController) ResetTo(id NavigationPath, values Values) {
	if n.destroyed {
		return
	}

	n.IgnoreNextInvalidation()

	n.scope.Publish(&proto.NavigationResetRequested{
		RootView: proto.RootViewID(id),
		Values:   values.proto(),
	})
}

func (n *navigationController) Reload() {
	if n.destroyed {
		return
	}

	n.IgnoreNextInvalidation()

	n.scope.Publish(&proto.NavigationReloadRequested{})
}

func (n *navigationController) Open(resource URI) {
	n.scope.Publish(&proto.OpenHttpLink{
		Url: proto.URI(resource),
	})
}

// IgnoreNextInvalidation sets a scope flag, so that if this event loop cycle ends, any invalidation is ignored.
// This is useful, if you know, that the current render cycle will be without purpose, e.g. due to a navigation
// request. This also simplifies the client implementation, if the navigation cycle is faster than applying
// the rendered view, which may result in invalid screens.
func (n *navigationController) IgnoreNextInvalidation() {
	n.scope.ignoreNextInvalidation.Store(true)
}

// HTTPFlow issues either a WebView or a native browser redirect flow, starting at the given start-uri and in case of a
// result, expects the given redirectTarget.
// Example flow for a mobile app frontend:
//   - Android App opens a WebView with start uri
//   - If webview detects redirectTarget uri, it closes the webview AND
//   - extracts the query parameters AND
//   - invokes the navigation path with the query parameters
//
// Example for a web frontend:
//   - browser navigates to start
//   - browser redirects to redirectTarget which must be actually the redirectNavigation which we cannot intercept
//   - thus, the redirectNavigation is actually ignored for web frontends, however should be defined for completeness
//     for other frontends
//
// This function just uses [Navigation.Open] with the start as resource and the following defined options:
//   - _type is hardcoded as "http-flow"
//   - redirectTarget as declared
//   - redirectNavigation as declared
//
// A web frontend keeps its session, as long as the identity provider redirects back with a GET, because the session
// cookie is SameSite=Lax. Former versions kept the session id in the local storage of the browser to restore it
// after the redirect, which is not required anymore and has been removed. An identity provider which posts its
// result back needs a server route, which takes the POST and redirects to a page by GET.
func HTTPFlow(nav Navigation, start, redirectTarget URI, redirectNavigation NavigationPath) {
	nav.(*navigationController).IgnoreNextInvalidation()

	nav.(*navigationController).scope.Publish(&proto.OpenHttpFlow{
		Url:                proto.URI(start),
		RedirectNavigation: proto.Str(redirectNavigation),
		RedirectTarget:     proto.Str(redirectTarget),
	})
}

// HTTPify inspects the given string and eventually prefixes it with https://
func HTTPify(s string) URI {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return URI(s)
	}

	if strings.HasPrefix(s, "mailto:") {
		return URI(s)
	}

	return "https://" + URI(s)
}

// HTTPOpen just triggers a regular (p)open call for the given http URL. A webfrontend will
// most likely trigger a window.open(url, target). In Javascript, target may be _blank|_self|_parent|_top|_unfencedTop
func HTTPOpen(nav Navigation, url URI, target string) {
	nav.(*navigationController).scope.Publish(&proto.OpenHttpLink{
		Url:    proto.URI(url),
		Target: proto.Str(target),
	})
}

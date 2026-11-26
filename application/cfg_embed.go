// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// embeddedCookieName carries the session id of a page in a frame of another site over http://localhost, see
// [Configurator.EmbeddablePages].
const embeddedCookieName = "wdy-ora-embedded"

// hostEmbeddedCookieName carries the session id of a page in a frame of another site over https, like
// [hostSessionCookieName].
const hostEmbeddedCookieName = "__Host-" + embeddedCookieName

// EmbeddablePages lets the pages of the given paths get a session in a frame of another site. A path is given like
// the id of [Configurator.RootView], thus a wildcard like "e/*" includes all pages below e/. By default, a page in
// a frame of another site has no session, because a browser drops the session cookie with SameSite=Lax there: it
// cannot download, and each new connection of the wire loads the page again.
//
// The navigation of such a frame to one of these pages gets a cookie of its own, which is SameSite=None, Secure and
// Partitioned, thus it only counts within the embedding site, and each embedding site has a session of its own. That
// session never has a user, see [session.NewEmbeddedID], so no other site can show a page of a logged-in user and
// trick the user into clicks (clickjacking). The navigation of such a frame to any other page issues no cookie, as
// before; once issued, the cookie is sent for every page of the application in a frame of that site. If a request
// carries both cookies, the plain session cookie counts.
//
// A browser keeps a cookie with SameSite=None only, if it is Secure, thus the cookie is issued only over https,
// with the prefix __Host-, and over http://localhost, which a browser treats as secure. Over plain http of another
// host and with NO_SSL=true, an embedded page still has no session. A browser without support of Partitioned
// cookies may drop the cookie, if it blocks third-party cookies.
func (c *Configurator) EmbeddablePages(paths ...core.NavigationPath) *Configurator {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, p := range paths {
		id := proto.RootViewID(p)
		if !id.Valid() {
			panic(fmt.Errorf("invalid embeddable page path: %v", id))
		}

		c.embeddablePages = append(c.embeddablePages, id)
	}

	return c
}

// embeddable reports whether the page of the url path may get a session in a frame of another site, see
// [Configurator.EmbeddablePages].
func embeddable(pages []proto.RootViewID, urlPath string) bool {
	id := proto.RootViewID(strings.TrimPrefix(urlPath, "/"))
	if id == "" {
		id = "."
	}

	return slices.ContainsFunc(pages, func(p proto.RootViewID) bool { return p.Matches(id) })
}

// isForeignFrameNavigation reports whether the request loads a page into a frame of another site. Only this request
// tells that a page is embedded: all later requests of the page, like those of the wire, a download or an image,
// are same-origin.
func isForeignFrameNavigation(r *http.Request) bool {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		return false
	}

	if !strings.EqualFold(r.Header.Get("Sec-Fetch-Site"), "cross-site") {
		return false
	}

	dest := r.Header.Get("Sec-Fetch-Dest")
	return strings.EqualFold(dest, "iframe") || strings.EqualFold(dest, "frame")
}

// embeddedCookieNameOf returns the name of the embedded cookie of a response to the request, or false, if a browser
// would drop it: a cookie with SameSite=None must be Secure, and a browser drops a Secure cookie of plain http,
// except for localhost. NO_SSL=true never marks a cookie as Secure.
func (c *Configurator) embeddedCookieNameOf(r *http.Request) (string, bool) {
	switch {
	case c.cookiePolicy == cookieNeverSecure:
		return "", false
	case c.secureCookie(r):
		return hostEmbeddedCookieName, true
	case isLocalHost(hostname(r.Host)):
		return embeddedCookieName, true
	}

	return "", false
}

// embeddedSessionIDOf returns the session id of the embedded cookie of the request. An id, which nago has not
// created for an embedded page, does not count, because its session could have a user. Unlike the plain session
// cookie, no other cookie is ever taken over, see [sessionIDFromCookies].
func (c *Configurator) embeddedSessionIDOf(r *http.Request) string {
	name, ok := c.embeddedCookieNameOf(r)
	if !ok {
		return ""
	}

	cookie, err := r.Cookie(name)
	if err != nil || !session.ID(cookie.Value).Embedded() {
		return ""
	}

	return cookie.Value
}

// newEmbeddedSessionCookie returns the embedded cookie of a response to the request with the session of its
// embedded cookie, or a new one, see [Configurator.EmbeddablePages]. It returns nil, if a browser would drop it.
func (c *Configurator) newEmbeddedSessionCookie(r *http.Request) *http.Cookie {
	name, ok := c.embeddedCookieNameOf(r)
	if !ok {
		return nil
	}

	id := c.embeddedSessionIDOf(r)
	if id == "" {
		id = string(session.NewEmbeddedID())
	}

	return &http.Cookie{
		Name:        name,
		Value:       id,
		Expires:     time.Now().Add(365 * 24 * time.Hour),
		Secure:      true,
		HttpOnly:    true,
		SameSite:    http.SameSiteNoneMode,
		Partitioned: true,
		Path:        "/",
	}
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// sessionCookieName is the http-only cookie which carries the session id.
const sessionCookieName = "wdy-ora-access"

// cookiePolicy decides whether the session cookie is marked as Secure.
type cookiePolicy int

const (
	// cookieAdaptive marks the cookie as Secure, if the request came through https, see requestIsHTTPS. Thus, a
	// server behind an https reverse proxy can be used at the same time directly by plain http, e.g. as a fallback
	// in a local network without DNS.
	cookieAdaptive cookiePolicy = iota
	cookieAlwaysSecure
	cookieNeverSecure
)

// envNoSSL overrides the adaptive decision for special setups: true never marks the cookie as Secure, e.g. behind
// a plain http proxy which sends forwarding headers, false always does.
const envNoSSL = "NO_SSL"

func determineCookiePolicy() cookiePolicy {
	if v, ok := os.LookupEnv(envNoSSL); ok {
		if noSSL, _ := strconv.ParseBool(v); noSSL {
			slog.Info("session cookie is never secure", "env", envNoSSL)
			return cookieNeverSecure
		}

		slog.Info("session cookie is always secure", "env", envNoSSL)
		return cookieAlwaysSecure
	}

	slog.Info("session cookie is secure for https requests, including those through a reverse proxy")
	return cookieAdaptive
}

// secureCookie decides whether the session cookie of a response to the request is marked as Secure. A browser
// drops a Secure cookie of a plain http response, so the request decides.
func (c *Configurator) secureCookie(r *http.Request) bool {
	switch c.cookiePolicy {
	case cookieAlwaysSecure:
		return true
	case cookieNeverSecure:
		return false
	default:
		return requestIsHTTPS(r)
	}
}

// requestIsHTTPS reports whether the browser sent the request through https, either directly or through a reverse
// proxy. A request through a proxy which does not tell the protocol counts as https, because that is the usual
// setup. A forged header only hurts the client which forged it: a Secure cookie of a plain http response is dropped,
// and any hint for https wins over a hint for http.
func requestIsHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}

	protoHint := false
	for _, v := range r.Header.Values("X-Forwarded-Proto") {
		for _, p := range strings.Split(v, ",") {
			protoHint = true
			if strings.EqualFold(strings.TrimSpace(p), "https") {
				return true
			}
		}
	}

	for _, v := range r.Header.Values("Forwarded") {
		for _, part := range strings.FieldsFunc(v, func(r rune) bool { return r == ',' || r == ';' }) {
			k, val, ok := strings.Cut(strings.TrimSpace(part), "=")
			if !ok || !strings.EqualFold(k, "proto") {
				continue
			}

			protoHint = true
			if strings.EqualFold(strings.Trim(val, `"`), "https") {
				return true
			}
		}
	}

	if protoHint {
		return false
	}

	behindProxy := r.Header.Get("X-Forwarded-For") != "" || r.Header.Get("Forwarded") != "" || r.Header.Get("X-Real-IP") != ""
	return behindProxy
}

// newSessionCookie creates the session cookie with the given id for the response to the request.
func (c *Configurator) newSessionCookie(r *http.Request, id string) *http.Cookie {
	return &http.Cookie{
		Name:     sessionCookieName,
		Value:    id,
		Expires:  time.Now().Add(365 * 24 * time.Hour),
		Secure:   c.secureCookie(r),
		HttpOnly: true,
		// Strict lost the session on the return from an identity provider, see the history of this file. Lax is
		// sent on top-level navigations with a safe method, like the redirect back from a login service.
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}
}

// mayIssueSessionCookie reports whether a response to the request may issue a new session cookie, because the
// request came without one. A cross-site navigation or a non-safe method may lack an existing cookie just because of
// SameSite, e.g. the POST of an identity provider: a new cookie would replace the session of the user. Such a page
// gets its cookie from its assets, which are same-origin, or from the wire.
func mayIssueSessionCookie(r *http.Request) bool {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		return false
	}

	return !strings.EqualFold(r.Header.Get("Sec-Fetch-Site"), "cross-site")
}

// checkWireOrigin accepts the websocket of a page of this application only. A cookie is SameSite=Lax, which a
// sibling subdomain of the same site passes, so without this check, a page of another application on the same
// parent domain could open the wire with the cookie of the user and act as the user. Clients without an Origin, like
// native apps, are accepted, because a browser always sends one.
func (c *Configurator) checkWireOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}

	u, err := url.Parse(origin)
	if err == nil && u.Host != "" {
		for _, host := range ownHosts(r) {
			if strings.EqualFold(u.Host, host) {
				return true
			}
		}

		// the development server of the frontend may run on another port of the local machine
		if c.IsDebug() && isLocalHost(u.Hostname()) && isLocalHost(hostname(r.Host)) {
			return true
		}
	}

	slog.Warn("rejected a websocket of a foreign origin", "origin", origin, "host", r.Host)
	return false
}

// ownHosts returns the hosts under which the browser reaches this application: the host of the request and the host
// a reverse proxy forwarded. The context path is no source, because the wire takes it from the first request, if it
// has not been configured.
func ownHosts(r *http.Request) []string {
	hosts := []string{r.Host}
	for _, v := range r.Header.Values("X-Forwarded-Host") {
		for _, h := range strings.Split(v, ",") {
			hosts = append(hosts, strings.TrimSpace(h))
		}
	}

	for _, v := range r.Header.Values("Forwarded") {
		for _, part := range strings.FieldsFunc(v, func(r rune) bool { return r == ',' || r == ';' }) {
			if k, val, ok := strings.Cut(strings.TrimSpace(part), "="); ok && strings.EqualFold(k, "host") {
				hosts = append(hosts, strings.Trim(val, `"`))
			}
		}
	}

	return hosts
}

func hostname(hostport string) string {
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		return h
	}

	return hostport
}

func isLocalHost(host string) bool {
	host = strings.Trim(host, "[]")
	return host == "localhost" || strings.HasSuffix(host, ".localhost") || host == "127.0.0.1" || host == "::1"
}

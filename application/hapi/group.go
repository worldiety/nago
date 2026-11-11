// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package hapi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync/atomic"

	"go.wdy.de/nago/pkg/oas/v31"
)

// Group is a set of endpoints below one path prefix with an OpenAPI document of its own.
//
// The endpoints of the [API] of an application are registered once at its start and never change. Those of a group
// are defined at runtime, e.g. from scripts, and [Group.Replace] exchanges them as a whole and atomically: a request
// is served either by the old or by the new set, never by a mix of both. The group is a [http.Handler], which
// distributes a request to the endpoint of its method and path: a path, which no endpoint takes, is a 404, a path,
// which an endpoint of another method takes, is a 405 with the header Allow. Both are written by [WriteError] as
// problem+json. HEAD and OPTIONS are methods like any other, a GET endpoint does not take them.
type Group struct {
	prefix string
	state  atomic.Pointer[groupState]
}

type groupState struct {
	doc     *oas.OpenAPI
	docJSON []byte
	routes  []groupRoute
}

// NewGroup creates an empty group for the paths below prefix, e.g. /api/areas/sales. The prefix starts with a slash
// and has none at its end; the empty prefix is the root.
func NewGroup(prefix string) *Group {
	return &Group{prefix: strings.TrimSuffix(prefix, "/")}
}

// Prefix returns the path prefix of the group.
func (g *Group) Prefix() string {
	return g.prefix
}

// Replace builds a new set of endpoints with build and replaces the current set and document with it. build
// registers the endpoints at the given api like at the API of an application, with the DSL (e.g. [Get]) or with
// [Handle]; their paths are relative to the prefix. doc is the new document, nil for an empty one; without servers it
// gets the prefix as its server. Replace fails without any change, if build fails or two endpoints of the same method
// take the same paths or a common path, which none of them takes more specifically. A nil build is a group without
// endpoints.
func (g *Group) Replace(doc *oas.OpenAPI, build func(api *API) error) error {
	if doc == nil {
		doc = &oas.OpenAPI{}
	}

	if doc.Openapi == "" {
		doc.Openapi = oas.Version
	}

	if doc.Paths == nil {
		doc.Paths = oas.Paths{}
	}

	if len(doc.Servers) == 0 && g.prefix != "" {
		doc.Servers = []oas.Server{{URL: g.prefix}}
	}

	var rt groupRouter
	api := NewAPI(doc, Options{RegisterHandler: rt.add})
	if build != nil {
		if err := build(api); err != nil {
			return err
		}
	}

	if rt.err != nil {
		return rt.err
	}

	buf, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("hapi: cannot encode the document of the group %s: %w", g.prefix, err)
	}

	g.state.Store(&groupState{doc: doc, docJSON: buf, routes: rt.routes})
	return nil
}

// Document returns the current document of the group, nil before the first [Group.Replace]. It must not be changed.
func (g *Group) Document() *oas.OpenAPI {
	st := g.state.Load()
	if st == nil {
		return nil
	}

	return st.doc
}

// ServeDocument writes the current document of the group as JSON, a 404 before the first [Group.Replace].
func (g *Group) ServeDocument(writer http.ResponseWriter, request *http.Request) {
	st := g.state.Load()
	if st == nil {
		WriteError(writer, request, NewStatusError(http.StatusNotFound, "There is no document."))
		return
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := writer.Write(st.docJSON); err != nil {
		slog.Error("hapi: cannot write the document of a group", "err", err.Error())
	}
}

// StrUnknownPath and StrMethodNotAllowed are the details of the answers of a [Group] without an endpoint.
const (
	StrUnknownPath      = "There is no endpoint at this path."
	StrMethodNotAllowed = "The endpoint at this path does not take this method."
)

// ServeHTTP distributes the request to the endpoint of its method and path. The path is compared segment by segment
// with the decoded segments of the escaped path, so that an encoded slash is a part of the value of a placeholder.
// The handler of the endpoint reads the values of the placeholders with [http.Request.PathValue].
func (g *Group) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	st := g.state.Load()
	segs, ok := g.segments(request)
	if st == nil || !ok {
		WriteError(writer, request, NewStatusError(http.StatusNotFound, StrUnknownPath))
		return
	}

	var best *groupRoute
	var allow []string
	for i := range st.routes {
		r := &st.routes[i]
		if !r.pattern.takes(segs) {
			continue
		}

		if !slices.Contains(allow, r.method) {
			allow = append(allow, r.method)
		}

		if r.method == request.Method && (best == nil || r.pattern.subset(best.pattern)) {
			best = r
		}
	}

	if best == nil {
		if len(allow) == 0 {
			WriteError(writer, request, NewStatusError(http.StatusNotFound, StrUnknownPath))
			return
		}

		slices.Sort(allow)
		writer.Header().Set("Allow", strings.Join(allow, ", "))
		WriteError(writer, request, NewStatusError(http.StatusMethodNotAllowed, StrMethodNotAllowed))
		return
	}

	for name, value := range best.pattern.values(segs) {
		request.SetPathValue(name, value)
	}

	best.handler(writer, request)
}

// segments are the decoded segments of the path of the request below the prefix.
func (g *Group) segments(request *http.Request) ([]string, bool) {
	if request.URL == nil {
		return nil, false
	}

	rest, ok := strings.CutPrefix(request.URL.EscapedPath(), g.prefix)
	if !ok {
		return nil, false
	}

	if rest == "" || rest == "/" {
		return nil, true
	}

	if rest[0] != '/' {
		return nil, false
	}

	raw := strings.Split(rest[1:], "/")
	res := make([]string, len(raw))
	for i, s := range raw {
		v, err := url.PathUnescape(s)
		if err != nil {
			return nil, false
		}

		res[i] = v
	}

	return res, true
}

// groupRoute is an endpoint of a group.
type groupRoute struct {
	method  string
	pattern groupPattern
	handler http.HandlerFunc
}

// groupRouter collects the endpoints of a [Group.Replace] and the first error.
type groupRouter struct {
	routes []groupRoute
	err    error
}

func (rt *groupRouter) add(method, path string, handler http.HandlerFunc) {
	if rt.err != nil {
		return
	}

	p, err := parseGroupPattern(path)
	if err != nil {
		rt.err = err
		return
	}

	if handler == nil {
		rt.err = fmt.Errorf("hapi: the endpoint %s %s has no handler", method, path)
		return
	}

	for _, other := range rt.routes {
		if other.method != method {
			continue
		}

		ab, ba := p.subset(other.pattern), other.pattern.subset(p)
		switch {
		case ab && ba:
			rt.err = fmt.Errorf("hapi: the endpoints %s %s and %s %s take the same paths", method, path, method, other.pattern.raw)
			return
		case !ab && !ba && p.meets(other.pattern):
			rt.err = fmt.Errorf("hapi: the endpoints %s %s and %s %s take a common path and none is more specific", method, path, method, other.pattern.raw)
			return
		}
	}

	rt.routes = append(rt.routes, groupRoute{method: method, pattern: p, handler: handler})
}

// groupPattern is a parsed path of an endpoint of a group: whole segments of text, {name} for one segment and, at
// its end, {name...} for one or more.
type groupPattern struct {
	raw  string
	segs []groupSegment
}

type groupSegment struct {
	text  string
	name  string
	multi bool
}

func (s groupSegment) literal() bool {
	return s.name == ""
}

func parseGroupPattern(path string) (groupPattern, error) {
	res := groupPattern{raw: path}
	if !strings.HasPrefix(path, "/") {
		return res, fmt.Errorf("hapi: the path %q of an endpoint does not start with a slash", path)
	}

	if path == "/" {
		return res, nil
	}

	names := map[string]bool{}
	parts := strings.Split(path[1:], "/")
	for i, p := range parts {
		if p == "" {
			return res, fmt.Errorf("hapi: the path %q of an endpoint has an empty segment", path)
		}

		if !strings.HasPrefix(p, "{") || !strings.HasSuffix(p, "}") {
			if strings.ContainsAny(p, "{}") {
				return res, fmt.Errorf("hapi: the segment %q of the path %q mixes text and a placeholder", p, path)
			}

			res.segs = append(res.segs, groupSegment{text: p})
			continue
		}

		seg := groupSegment{name: p[1 : len(p)-1]}
		if name, ok := strings.CutSuffix(seg.name, "..."); ok {
			seg = groupSegment{name: name, multi: true}
			if i != len(parts)-1 {
				return res, fmt.Errorf("hapi: the placeholder %q of the path %q is not its last segment", p, path)
			}
		}

		if seg.name == "" || strings.ContainsAny(seg.name, "{}") || names[seg.name] {
			return res, fmt.Errorf("hapi: the placeholder %q of the path %q has no valid and unique name", p, path)
		}

		names[seg.name] = true
		res.segs = append(res.segs, seg)
	}

	return res, nil
}

// takes reports whether the pattern takes the decoded segments of a path.
func (p groupPattern) takes(path []string) bool {
	for i := 0; ; i++ {
		switch {
		case i == len(p.segs):
			return i == len(path)
		case i == len(path):
			return false
		case p.segs[i].multi:
			return !slices.Contains(path[i:], "")
		case path[i] == "":
			return false
		case p.segs[i].literal() && p.segs[i].text != path[i]:
			return false
		}
	}
}

// values are the values of the placeholders for a path, which the pattern takes.
func (p groupPattern) values(path []string) map[string]string {
	res := map[string]string{}
	for i, s := range p.segs {
		switch {
		case s.multi:
			res[s.name] = strings.Join(path[i:], "/")
		case !s.literal():
			res[s.name] = path[i]
		}
	}

	return res
}

// subset reports whether every path of p is also a path of q.
func (p groupPattern) subset(q groupPattern) bool {
	a, b := p.segs, q.segs
	for i := 0; ; i++ {
		switch {
		case i == len(b):
			return i == len(a)
		case i == len(a):
			return false
		case b[i].multi:
			return true
		case a[i].multi:
			return false
		case b[i].literal() && (!a[i].literal() || a[i].text != b[i].text):
			return false
		}
	}
}

// meets reports whether a path exists, which both patterns take.
func (p groupPattern) meets(q groupPattern) bool {
	a, b := p.segs, q.segs
	for i := 0; ; i++ {
		switch {
		case i == len(a) || i == len(b):
			return i == len(a) && i == len(b)
		case a[i].multi || b[i].multi:
			return true
		case a[i].literal() && b[i].literal() && a[i].text != b[i].text:
			return false
		}
	}
}

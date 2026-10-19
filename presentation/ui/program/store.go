// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package program

import (
	"container/list"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/presentation/core"
)

// BundlePath is the path, under which [Handler] serves the bundles, which a component has registered.
const BundlePath = "/api/nago/v1/gostrict/"

// maxBundles limits the bundles, which the process keeps for the frontend. A component registers its bundle in
// each render, so a bundle in use is never dropped.
const maxBundles = 1024

// bundles is the store of the process. A bundle is identified by the hash of its script, so the applications of a
// process may share it.
var bundles = newStore(maxBundles)

// bindingKey signs the URLs, which bind a bundle to the session of a window, see [registerFor]. The bundles live in
// the memory of the process, so the key does not need to survive it.
var bindingKey = func() []byte {
	key := make([]byte, 32)
	_, _ = rand.Read(key) // never fails, see the documentation of crypto/rand
	return key
}()

type store struct {
	mu    sync.Mutex
	max   int
	order *list.List
	items map[string]*list.Element
}

type storeEntry struct {
	hash string
	body []byte
	// public is set by [Register], whose URL is valid for any session
	public bool
}

func newStore(max int) *store {
	return &store{max: max, order: list.New(), items: map[string]*list.Element{}}
}

// put keeps the manifest of the bundle. A public bundle stays public, even if a window registers it again.
func (s *store) put(b *gostrict.Bundle, public bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if el, ok := s.items[b.Hash]; ok {
		s.order.MoveToFront(el)
		if public {
			e := el.Value.(storeEntry)
			e.public = true
			el.Value = e
		}
		return
	}

	body, err := json.Marshal(newManifest(b))
	if err != nil {
		// the manifest consists of strings and numbers only
		panic(err)
	}
	s.items[b.Hash] = s.order.PushFront(storeEntry{hash: b.Hash, body: body, public: public})
	for s.order.Len() > s.max {
		last := s.order.Back()
		s.order.Remove(last)
		delete(s.items, last.Value.(storeEntry).hash)
	}
}

func (s *store) get(hash string) (storeEntry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	el, ok := s.items[hash]
	if !ok {
		return storeEntry{}, false
	}
	s.order.MoveToFront(el)
	return el.Value.(storeEntry), true
}

// Register keeps the bundle for the frontend and returns its immutable URL, which any session may load. The
// components of this package use [registerFor] instead, which binds the URL to the session of the window.
func Register(b *gostrict.Bundle) core.URI {
	bundles.put(b, true)
	return core.URI(BundlePath + b.Hash + ".json")
}

// registerFor keeps the bundle for the frontend and returns its immutable URL, which only the session of the window
// may load, so that a hash, which leaks, e.g. from a log, does not grant access to the bundle. A component calls it
// in each render, and the URL stays the same as long as the session does.
func registerFor(wnd core.Window, b *gostrict.Bundle) core.URI {
	bundles.put(b, false)
	return core.URI(BundlePath + b.Hash + ".json?s=" + binding(b.Hash, sessionOf(wnd)))
}

func sessionOf(wnd core.Window) string {
	if wnd == nil {
		return ""
	}
	sess := wnd.Session()
	if sess == nil {
		return ""
	}
	return string(sess.ID())
}

// binding signs the hash of a bundle together with a session id.
func binding(hash, sessionID string) string {
	mac := hmac.New(sha256.New, bindingKey)
	mac.Write([]byte(hash))
	mac.Write([]byte{0})
	mac.Write([]byte(sessionID))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)[:16])
}

// Handler serves the bundles under [BundlePath] like [SessionHandler], but it knows no session, so it only serves
// the bundles of [Register] and those, which a window without session has registered.
func Handler() http.Handler {
	return SessionHandler(nil)
}

// SessionHandler serves the bundles under [BundlePath]. A bundle never changes under its URL, so a browser may cache
// it forever, but no shared cache may, because a component binds the URL to the session of its window. The given
// function returns the session id of a request and may be nil.
func SessionHandler(sessionOf func(r *http.Request) string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, BundlePath)
		hash, ok := strings.CutSuffix(name, ".json")
		if !ok {
			http.NotFound(w, r)
			return
		}
		e, ok := bundles.get(hash)
		if !ok {
			http.NotFound(w, r)
			return
		}
		if !e.public {
			var sessionID string
			if sessionOf != nil {
				sessionID = sessionOf(r)
			}
			// a foreign session looks like an unknown bundle
			if !hmac.Equal([]byte(r.URL.Query().Get("s")), []byte(binding(hash, sessionID))) {
				http.NotFound(w, r)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
		_, _ = w.Write(e.body)
	})
}

// manifest is the form of a bundle, which the frontend loads: the script and the schema of its boundary, which
// the frontend needs to convert the values of the protocol.
type manifest struct {
	Hash      string              `json:"hash"`
	JS        string              `json:"js"`
	Exports   map[string]fnJSON   `json:"exports"`
	Records   map[string]recJSON  `json:"records"`
	Unions    map[string][]string `json:"unions"`
	Handlers  map[string]string   `json:"handlers"`
	States    map[string]string   `json:"states"`
	SourceMap json.RawMessage     `json:"sourceMap,omitempty"`
}

type fnJSON struct {
	Params  []typeJSON `json:"params"`
	Results []typeJSON `json:"results"`
}

type recJSON struct {
	Fields []fieldJSON `json:"fields"`
}

type fieldJSON struct {
	Name string   `json:"name"`
	JS   string   `json:"js"`
	Type typeJSON `json:"type"`
}

type typeJSON struct {
	Kind string    `json:"kind"`
	Name string    `json:"name,omitempty"`
	Elem *typeJSON `json:"elem,omitempty"`
	Key  *typeJSON `json:"key,omitempty"`
}

func newManifest(b *gostrict.Bundle) manifest {
	m := manifest{
		Hash:     b.Hash,
		JS:       b.JS,
		Exports:  map[string]fnJSON{},
		Records:  map[string]recJSON{},
		Unions:   map[string][]string{},
		Handlers: b.Handlers,
		States:   b.States,
	}
	if len(b.SourceMap) > 0 {
		m.SourceMap = json.RawMessage(b.SourceMap)
	}
	for name, f := range b.Exports {
		var fn fnJSON
		for _, t := range f.Params {
			fn.Params = append(fn.Params, newTypeJSON(t))
		}
		for _, t := range f.Results {
			fn.Results = append(fn.Results, newTypeJSON(t))
		}
		m.Exports[name] = fn
	}
	for name, r := range b.Records {
		var rec recJSON
		for _, f := range r.Fields {
			rec.Fields = append(rec.Fields, fieldJSON{Name: f.Name, JS: f.JS, Type: newTypeJSON(f.Type)})
		}
		m.Records[name] = rec
	}
	for name, u := range b.Unions {
		m.Unions[name] = u.Variants
	}
	if m.Handlers == nil {
		m.Handlers = map[string]string{}
	}
	if m.States == nil {
		m.States = map[string]string{}
	}
	return m
}

func newTypeJSON(t gostrict.Type) typeJSON {
	res := typeJSON{Kind: t.Kind.String(), Name: t.Name}
	if t.Elem != nil {
		e := newTypeJSON(*t.Elem)
		res.Elem = &e
	}
	if t.Key != nil {
		k := newTypeJSON(*t.Key)
		res.Key = &k
	}
	return res
}

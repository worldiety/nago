// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package program

import (
	"container/list"
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

type store struct {
	mu    sync.Mutex
	max   int
	order *list.List
	items map[string]*list.Element
}

type storeEntry struct {
	hash string
	body []byte
}

func newStore(max int) *store {
	return &store{max: max, order: list.New(), items: map[string]*list.Element{}}
}

// put keeps the manifest of the bundle and returns its URL.
func (s *store) put(b *gostrict.Bundle) core.URI {
	uri := core.URI(BundlePath + b.Hash + ".json")

	s.mu.Lock()
	defer s.mu.Unlock()
	if el, ok := s.items[b.Hash]; ok {
		s.order.MoveToFront(el)
		return uri
	}

	body, err := json.Marshal(newManifest(b))
	if err != nil {
		// the manifest consists of strings and numbers only
		panic(err)
	}
	s.items[b.Hash] = s.order.PushFront(storeEntry{hash: b.Hash, body: body})
	for s.order.Len() > s.max {
		last := s.order.Back()
		s.order.Remove(last)
		delete(s.items, last.Value.(storeEntry).hash)
	}
	return uri
}

func (s *store) get(hash string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	el, ok := s.items[hash]
	if !ok {
		return nil, false
	}
	s.order.MoveToFront(el)
	return el.Value.(storeEntry).body, true
}

// Register keeps the bundle for the frontend and returns its immutable URL. A component calls it in each render.
func Register(b *gostrict.Bundle) core.URI {
	return bundles.put(b)
}

// Handler serves the bundles under [BundlePath]. A bundle never changes under its URL, so a browser may cache it
// forever.
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, BundlePath)
		hash, ok := strings.CutSuffix(name, ".json")
		if !ok {
			http.NotFound(w, r)
			return
		}
		body, ok := bundles.get(hash)
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		_, _ = w.Write(body)
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

// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"maps"
	"path"
	"slices"
	"strings"
	"sync"

	"go.wdy.de/nago/pkg/gostrict/internal/lru"
)

// Version changes, whenever the translation changes, so that a cache does not return an outdated bundle.
const Version = "gostrict/2"

// Cache compiles programs and keeps the most recently used bundles and diagnostics by the hash of their sources,
// so that a program, e.g. a script of an application, is checked and translated only once per version.
type Cache struct {
	mu      sync.Mutex
	entries *lru.Cache[string, *cacheEntry]
	// inflight deduplicates concurrent compilations of the same sources
	inflight map[string]*compilation
}

type cacheEntry struct {
	bundle *Bundle
	diags  Diagnostics
}

type compilation struct {
	done  chan struct{}
	entry *cacheEntry
	err   error
}

// NewCache returns a cache, which keeps up to max bundles.
func NewCache(max int) *Cache {
	return &Cache{entries: lru.New[string, *cacheEntry](max), inflight: map[string]*compilation{}}
}

// Compile returns the bundle of the package main of the modules, see [LoadModules], or its diagnostics. It hashes
// all Go files of the modules, so a module should contain a program and not a whole repository.
func (c *Cache) Compile(main string, modules ...Module) (*Bundle, Diagnostics, error) {
	key, err := SourceHash(main, modules...)
	if err != nil {
		return nil, nil, err
	}

	c.mu.Lock()
	if e, ok := c.entries.Get(key); ok {
		c.mu.Unlock()
		return e.bundle, e.diags, nil
	}
	if run, ok := c.inflight[key]; ok {
		c.mu.Unlock()
		<-run.done
		if run.err != nil {
			return nil, nil, run.err
		}
		return run.entry.bundle, run.entry.diags, nil
	}
	run := &compilation{done: make(chan struct{})}
	c.inflight[key] = run
	c.mu.Unlock()

	run.entry, run.err = compileEntry(main, modules)
	close(run.done)

	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.inflight, key)
	if run.err != nil {
		return nil, nil, run.err
	}
	c.entries.Put(key, run.entry)
	return run.entry.bundle, run.entry.diags, nil
}

func compileEntry(main string, modules []Module) (*cacheEntry, error) {
	prog, err := LoadModules(main, modules...)
	if err != nil {
		return nil, err
	}
	e := &cacheEntry{}
	if e.diags = Check(prog); len(e.diags) > 0 {
		return e, nil
	}
	e.bundle, e.diags = Emit(prog)
	return e, nil
}

// SourceHash returns the hash of the Go files of the modules, of the main package and of the version of the
// compiler.
func SourceHash(main string, modules ...Module) (string, error) {
	h := sha256.New()
	h.Write([]byte(Version + "\x00" + main + "\x00"))
	for _, m := range modules {
		h.Write([]byte(m.Path + "\x00"))
		err := fs.WalkDir(m.FS, ".", func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || path.Ext(p) != ".go" || strings.HasSuffix(p, "_test.go") {
				return nil
			}
			buf, err := fs.ReadFile(m.FS, p)
			if err != nil {
				return err
			}
			h.Write([]byte(p + "\x00"))
			h.Write(buf)
			h.Write([]byte{0})
			return nil
		})
		if err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(h.Sum(nil)[:16]), nil
}

// Whole is a checked whole program, see [LoadRoots], which translates a bundle once for each set of entry points,
// e.g. for each use case, which a host runs. A changed program is a new Whole, while running calls keep the bundles
// of the former one.
type Whole struct {
	prog *Program
	// mu serializes the translations, which share the program
	mu      sync.Mutex
	bundles map[string]*Bundle
}

// CompileWhole loads and checks a whole program, see [LoadRoots]. If it is not in the subset, it returns the
// diagnostics and no program.
func CompileWhole(roots []string, modules ...Module) (*Whole, Diagnostics, error) {
	prog, err := LoadRoots(roots, modules...)
	if err != nil {
		return nil, nil, err
	}
	if diags := Check(prog); len(diags) > 0 {
		return nil, diags, nil
	}
	return &Whole{prog: prog, bundles: map[string]*Bundle{}}, nil, nil
}

// Entries returns the full names of the entry points, e.g. "example.com/scripts/crm/read.Main".
func (w *Whole) Entries() []string {
	return slices.Sorted(maps.Values(entryPoints(w.prog)))
}

// Bundle returns the bundle of the entry points or of all, if there are none, see [EmitEntries].
func (w *Whole) Bundle(entries ...string) (*Bundle, error) {
	entries = slices.Sorted(slices.Values(entries))
	key := strings.Join(entries, "\x00")
	w.mu.Lock()
	defer w.mu.Unlock()
	if b, ok := w.bundles[key]; ok {
		return b, nil
	}
	b, diags := EmitEntries(w.prog, entries...)
	if len(diags) > 0 {
		return nil, diags
	}
	w.bundles[key] = b
	return b, nil
}

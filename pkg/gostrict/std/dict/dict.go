// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package dict provides the map of the gostrict subset, which replaces the maps of Go. Unlike a map of Go, a
// dict iterates in the order, in which the keys were added first, so a program behaves the same in every run
// and in Go as in JavaScript. A dict has reference semantics like a map of Go.
//
//gostrict:header
package dict

import (
	"iter"
	"slices"
)

// Key is a type, which compares by value in Go and in JavaScript.
type Key interface {
	~string | ~int32 | ~uint32 | ~bool
}

type store[K Key, V any] struct {
	idx  map[K]int
	keys []K
	vals []V
}

// Map maps keys to values. The subset requires a map to be created with [New], so the zero value never occurs in
// a checked program.
type Map[K Key, V any] struct {
	p *store[K, V]
}

// New returns an empty map.
func New[K Key, V any]() Map[K, V] {
	return Map[K, V]{p: &store[K, V]{idx: map[K]int{}}}
}

// Len returns the number of entries.
func (m Map[K, V]) Len() int32 {
	if m.p == nil {
		return 0
	}
	return int32(len(m.p.keys))
}

// Get returns the value of the key and whether the key exists.
func (m Map[K, V]) Get(k K) (V, bool) {
	if m.p != nil {
		if i, ok := m.p.idx[k]; ok {
			return m.p.vals[i], true
		}
	}
	var zero V
	return zero, false
}

// Has tells, whether the key exists.
func (m Map[K, V]) Has(k K) bool {
	_, ok := m.Get(k)
	return ok
}

// Set sets the value of the key. A new key is added at the end of the order, an existing one keeps its place.
func (m Map[K, V]) Set(k K, v V) {
	if m.p == nil {
		return // unreachable in a checked program
	}
	if i, ok := m.p.idx[k]; ok {
		m.p.vals[i] = v
		return
	}
	m.p.idx[k] = len(m.p.keys)
	m.p.keys = append(m.p.keys, k)
	m.p.vals = append(m.p.vals, v)
}

// Delete removes the key, if it exists.
func (m Map[K, V]) Delete(k K) {
	if m.p == nil {
		return
	}
	i, ok := m.p.idx[k]
	if !ok {
		return
	}
	delete(m.p.idx, k)
	m.p.keys = slices.Delete(m.p.keys, i, i+1)
	m.p.vals = slices.Delete(m.p.vals, i, i+1)
	for j := i; j < len(m.p.keys); j++ {
		m.p.idx[m.p.keys[j]] = j
	}
}

// All iterates over the keys, which exist when the iteration starts, in their order. A key deleted during the
// iteration is skipped, a key added is not visited.
func (m Map[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		if m.p == nil {
			return
		}
		for _, k := range slices.Clone(m.p.keys) {
			v, ok := m.Get(k)
			if !ok {
				continue
			}
			if !yield(k, v) {
				return
			}
		}
	}
}

// Keys iterates over the keys like [Map.All].
func (m Map[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range m.All() {
			if !yield(k) {
				return
			}
		}
	}
}

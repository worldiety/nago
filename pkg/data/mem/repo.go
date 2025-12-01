// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mem

import (
	"iter"
	"maps"
	"slices"
	"strings"
	"sync"

	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xslices"
)

type assertT struct{}

func (assertT) Identity() string {
	return ""
}

var _ data.Repository[assertT, string] = (*Repository[assertT, string])(nil)

// Repository is a standalone in-memory implementation which works without any serialization, in contrast to
// [mem.Store] and [json.Repository] combinations.
type Repository[E data.Aggregate[ID], ID ~string] struct {
	mutex    sync.RWMutex
	licenses map[ID]E
	name     string
}

func (r *Repository[E, ID]) FindAllByPrefix(prefix ID) iter.Seq2[E, error] {
	//TODO implement me
	panic("implement me")
}

func (r *Repository[E, ID]) Identifiers() iter.Seq2[ID, error] {
	return xslices.ValuesWithError(slices.Sorted(maps.Keys(r.licenses)), nil)
}

func (r *Repository[E, ID]) IdentifiersByPrefix(prefix ID) iter.Seq2[ID, error] {
	return func(yield func(ID, error) bool) {
		for id, err := range r.Identifiers() {
			if err != nil {
				yield(id, err)
				return
			}

			if strings.HasPrefix(string(id), string(prefix)) {
				if !yield(id, nil) {
					return
				}
			}
		}
	}
}

func (r *Repository[E, ID]) SetName(name string) {
	r.name = name
}

func (r *Repository[E, ID]) Name() string {
	return r.name
}

func (r *Repository[E, ID]) Load(id ID) (E, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	v, ok := r.licenses[id]
	return v, ok
}

func (r *Repository[E, ID]) FindByID(id ID) (std.Option[E], error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	lic, ok := r.licenses[id]
	if !ok {
		return std.None[E](), nil
	}

	return std.Some(lic), nil
}

func (r *Repository[E, ID]) FindAllByID(ids iter.Seq[ID]) iter.Seq2[E, error] {
	return func(yield func(E, error) bool) {
		r.mutex.RLock()
		defer r.mutex.RUnlock()

		for id := range ids {
			l, ok := r.licenses[id]
			if ok {
				if !yield(l, nil) {
					return
				}
			}
		}
	}
}

func (r *Repository[E, ID]) All() iter.Seq2[E, error] {
	tmp := slices.SortedFunc(maps.Values(r.licenses), func(a E, b E) int {
		return strings.Compare(string(a.Identity()), string(b.Identity()))
	})

	return xslices.ValuesWithError(tmp, nil)
}

func (r *Repository[E, ID]) Values() iter.Seq[E] {
	tmp := slices.SortedFunc(maps.Values(r.licenses), func(a E, b E) int {
		return strings.Compare(string(a.Identity()), string(b.Identity()))
	})

	return slices.Values(tmp)
}

func (r *Repository[E, ID]) Count() (int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.licenses), nil
}

func (r *Repository[E, ID]) Len() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.licenses)
}

func (r *Repository[E, ID]) DeleteByID(id ID) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.licenses, id)
	return nil
}

func (r *Repository[E, ID]) Remove(id ID) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.licenses, id)
}

func (r *Repository[E, ID]) DeleteAll() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	clear(r.licenses)
	return nil
}

func (r *Repository[E, ID]) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	clear(r.licenses)
}

func (r *Repository[E, ID]) DeleteAllByID(ids iter.Seq[ID]) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for id := range ids {
		delete(r.licenses, id)
	}

	return nil
}

func (r *Repository[E, ID]) Delete(predicate func(E) (bool, error)) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, val := range r.licenses {
		if ok, err := predicate(val); ok || err != nil {
			if err != nil {
				return err
			}

			delete(r.licenses, val.Identity())
		}
	}

	return nil
}

func (r *Repository[E, ID]) DeleteByEntity(e E) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.licenses, e.Identity())
	return nil
}

func (r *Repository[E, ID]) Save(e E) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.licenses == nil {
		r.licenses = make(map[ID]E)
	}

	r.licenses[e.Identity()] = e
	return nil
}

func (r *Repository[E, ID]) Put(e E) {
	r.Store(e)
}

func (r *Repository[E, ID]) Store(e E) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.licenses == nil {
		r.licenses = make(map[ID]E)
	}

	r.licenses[e.Identity()] = e
}

func (r *Repository[E, ID]) StoreAll(e ...E) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.licenses == nil {
		r.licenses = make(map[ID]E)
	}

	for _, e := range e {
		r.licenses[e.Identity()] = e
	}
}

func (r *Repository[E, ID]) SaveAll(it iter.Seq[E]) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.licenses == nil {
		r.licenses = make(map[ID]E)
	}

	for e := range it {
		r.licenses[e.Identity()] = e
	}
	return nil
}

func (r *Repository[E, ID]) Append(values ...E) {
	r.StoreAll(values...)
}

func (r *Repository[E, ID]) Get(id ID) (E, bool) {
	v, ok := r.licenses[id]
	return v, ok
}

func (r *Repository[E, ID]) DeleteFunc(fn func(E) bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, v := range r.licenses {
		if ok := fn(v); ok {
			delete(r.licenses, v.Identity())
		}
	}

}

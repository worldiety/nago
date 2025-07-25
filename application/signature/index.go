// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package signature

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std/concurrent"
	"iter"
	"sync"
)

type memPtr uint32

// inMemoryIndex has an optimized memory layout which as the following advantages when compared to a naive layout:
//   - byMemPtr has a linear layout without signature pointers reducing fragmentation and pointer chasing
//   - other by* use a shorter 32bit pointer into memPtr table which uses half the space of a pointer and avoids any GC inspections
type inMemoryIndex struct {
	byMemPtr      concurrent.RWMap[memPtr, Signature]
	byID          concurrent.RWMap[ID, memPtr]
	byUser        concurrent.RWMap[user.ID, []memPtr]
	byResource    concurrent.RWMap[user.Resource, []memPtr]
	lastSignature option.Opt[memPtr]
	lastMemPtr    memPtr
	mutex         sync.Mutex
}

// Index inserts the given signature into the index
func (idx *inMemoryIndex) Index(sig Signature) {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	idx.lastMemPtr++
	ptr := idx.lastMemPtr

	idx.byMemPtr.Put(ptr, sig)

	idx.byID.Put(sig.ID, ptr)

	if sig.User != "" {
		sigs, _ := idx.byUser.Get(sig.User)
		sigs = append(sigs, ptr)
		idx.byUser.Put(sig.User, sigs)
	}

	for document := range sig.Documents.All() {
		slice, _ := idx.byResource.Get(document.Resource)
		slice = append(slice, ptr)
		idx.byResource.Put(document.Resource, slice)
	}

	if idx.lastSignature.IsNone() {
		idx.lastSignature = option.Some(ptr)
	} else {
		v, _ := idx.byMemPtr.Get(idx.lastSignature.Unwrap())
		if v.Number < sig.Number {
			idx.lastSignature = option.Some(ptr)
		}
	}
}

func (idx *inMemoryIndex) ByUser(user user.ID) iter.Seq2[Signature, error] {
	return func(yield func(Signature, error) bool) {
		slice, _ := idx.byUser.Get(user)
		for _, ptr := range slice {
			v, ok := idx.byMemPtr.Get(ptr)
			if !ok {
				panic("unreachable")
			}

			if !yield(v, nil) {
				return
			}
		}
	}
}

func (idx *inMemoryIndex) ByResource(res user.Resource) iter.Seq2[Signature, error] {
	return func(yield func(Signature, error) bool) {
		slice, _ := idx.byResource.Get(res)
		for _, ptr := range slice {
			v, ok := idx.byMemPtr.Get(ptr)
			if !ok {
				panic("unreachable")
			}

			if !yield(v, nil) {
				return
			}
		}
	}
}

func (idx *inMemoryIndex) ByID(id ID) (Signature, bool) {
	ptr, ok := idx.byID.Get(id)
	if !ok {
		panic("unreachable")
	}

	return idx.byMemPtr.Get(ptr)
}

func (idx *inMemoryIndex) LastSignature() option.Opt[Signature] {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	if idx.lastSignature.IsNone() {
		return option.Opt[Signature]{}
	}

	v, ok := idx.byMemPtr.Get(idx.lastSignature.Unwrap())
	if !ok {
		panic("unreachable")
	}

	return option.Some(v)
}

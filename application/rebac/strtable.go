// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rebac

import (
	"math"
	"sync"
)

type strTable struct {
	index   map[string]uint32
	reverse map[uint32]string
	mutex   sync.RWMutex
	lastIdx uint32 // 0 is reserved as invalid pointer
}

func newStrTable() *strTable {
	return &strTable{
		index:   map[string]uint32{},
		reverse: map[uint32]string{},
	}
}

// Lookup only returns a pointer if the string is already interned.
func (t *strTable) Lookup(s string) (uint32, bool) {
	if s == "" {
		return 0, true
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	ptr, ok := t.index[s]
	return ptr, ok
}

// Intern returns either an existing pointer or a new one.
func (t *strTable) Intern(s string) uint32 {
	if s == "" {
		return 0
	}

	// the fast path is most likely
	t.mutex.RLock()
	ptr, ok := t.index[s]
	if ok {
		t.mutex.RUnlock()
		return ptr
	}

	// release read lock to upgrade for write lock
	t.mutex.RUnlock()

	// slow path, expected to be rare
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// double check idiom under write lock
	ptr, ok = t.index[s]
	if ok {
		return ptr
	}

	t.lastIdx++
	if math.MaxUint32 == t.lastIdx {
		panic("too many strings in table")
	}

	ptr = t.lastIdx
	t.index[s] = ptr
	t.reverse[ptr] = s

	return ptr
}

func (t *strTable) String(ptr uint32) string {
	if ptr == 0 {
		return ""
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.reverse[ptr]
}

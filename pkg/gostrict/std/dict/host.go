// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package dict

import "reflect"

// The methods of this file convert the maps for a host in Go, see gostrict.ToNeutral. The subset does not see
// them.

// HostEntries returns the keys and values in their order.
func (m Map[K, V]) HostEntries() [][2]any {
	res := make([][2]any, 0, m.Len())
	for k, v := range m.All() {
		res = append(res, [2]any{k, v})
	}
	return res
}

// HostFromEntries returns a map of the entries, whose keys are of type K and whose values are of type V.
func (Map[K, V]) HostFromEntries(entries [][2]any) any {
	m := New[K, V]()
	for _, e := range entries {
		m.Set(e[0].(K), e[1].(V))
	}
	return m
}

// HostKey returns the type of the keys.
func (Map[K, V]) HostKey() reflect.Type {
	return reflect.TypeFor[K]()
}

// HostElem returns the type of the values.
func (Map[K, V]) HostElem() reflect.Type {
	return reflect.TypeFor[V]()
}

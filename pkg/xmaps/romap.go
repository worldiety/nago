// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xmaps

type Map[K comparable, V any] interface {
	Get(K) (V, bool)
}

type readOnlyMap[K comparable, V any] struct {
	m map[K]V
}

func (m readOnlyMap[K, V]) Get(k K) (V, bool) {
	v, ok := m.m[k]
	return v, ok
}

func WrapMap[K comparable, V any](m map[K]V) Map[K, V] {
	return readOnlyMap[K, V]{m: m}
}

// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package lru provides a cache, which drops the least recently used entry. It is not safe for concurrent use.
package lru

import "container/list"

type entry[K comparable, V any] struct {
	key   K
	value V
}

// Cache keeps up to a maximum of entries.
type Cache[K comparable, V any] struct {
	max   int
	order *list.List
	items map[K]*list.Element
}

func New[K comparable, V any](max int) *Cache[K, V] {
	return &Cache[K, V]{max: max, order: list.New(), items: map[K]*list.Element{}}
}

func (c *Cache[K, V]) Get(k K) (V, bool) {
	if el, ok := c.items[k]; ok {
		c.order.MoveToFront(el)
		return el.Value.(entry[K, V]).value, true
	}
	var zero V
	return zero, false
}

func (c *Cache[K, V]) Put(k K, v V) {
	if el, ok := c.items[k]; ok {
		el.Value = entry[K, V]{k, v}
		c.order.MoveToFront(el)
		return
	}
	c.items[k] = c.order.PushFront(entry[K, V]{k, v})
	for c.order.Len() > c.max {
		last := c.order.Back()
		c.order.Remove(last)
		delete(c.items, last.Value.(entry[K, V]).key)
	}
}

func (c *Cache[K, V]) Len() int {
	return c.order.Len()
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package concurrent

import (
	"container/list"
	"sync"
)

type LinkedList[T any] struct {
	mutex sync.RWMutex
	list  *list.List
}

func NewLinkedList[T any]() *LinkedList[T] {
	return &LinkedList[T]{
		list: list.New(),
	}
}

func (ll *LinkedList[T]) Len() int {
	ll.mutex.RLock()
	defer ll.mutex.RUnlock()
	return ll.list.Len()
}

func (ll *LinkedList[T]) Front() T {
	ll.mutex.RLock()
	defer ll.mutex.RUnlock()
	return ll.list.Front().Value.(T)
}

func (ll *LinkedList[T]) Back() T {
	ll.mutex.RLock()
	defer ll.mutex.RUnlock()
	return ll.list.Back().Value.(T)
}

func (ll *LinkedList[T]) PushBack(value T) {
	ll.mutex.Lock()
	defer ll.mutex.Unlock()
	ll.list.PushBack(value)
}

func (ll *LinkedList[T]) PushFront(value T) {
	ll.mutex.Lock()
	defer ll.mutex.Unlock()
	ll.list.PushFront(value)
}

func (ll *LinkedList[T]) PopBack() T {
	ll.mutex.Lock()
	defer ll.mutex.Unlock()
	return ll.list.Remove(ll.list.Back()).(T)
}

func (ll *LinkedList[T]) PopFront() T {
	ll.mutex.Lock()
	defer ll.mutex.Unlock()
	return ll.list.Remove(ll.list.Front()).(T)
}

func (ll *LinkedList[T]) Clear() {
	ll.mutex.Lock()
	defer ll.mutex.Unlock()
	ll.list.Init()
}

// Values copies the entire list into a slice. This has worst case performance.
func (ll *LinkedList[T]) Values() []T {
	ll.mutex.RLock()
	defer ll.mutex.RUnlock()
	values := make([]T, 0, ll.list.Len())
	for e := ll.list.Front(); e != nil; e = e.Next() {
		values = append(values, e.Value.(T))
	}
	return values
}

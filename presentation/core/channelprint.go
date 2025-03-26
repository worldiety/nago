// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"log/slog"
	"sync"
)

type NopChannel struct {
}

func (n NopChannel) Subscribe(f func(msg []byte) error) (destroy func()) {
	return func() {

	}
}

func (n NopChannel) Publish(msg []byte) error {
	return nil
}

// A PrintChannel is for debugging.
type PrintChannel struct {
	mutex            sync.Mutex
	subscribers      map[int]func(msg []byte) error
	debugSubscribers map[int]func(msg []byte) error
	hnd              int
}

func NewPrintChannel() *PrintChannel {
	return &PrintChannel{}
}

func (n *PrintChannel) Subscribe(f func(msg []byte) error) (destroy func()) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if n.subscribers == nil {
		n.subscribers = make(map[int]func(msg []byte) error)
	}

	n.hnd++
	hnd := n.hnd
	n.subscribers[hnd] = f

	return func() {
		n.mutex.Lock()
		defer n.mutex.Unlock()
		delete(n.subscribers, hnd)
	}
}

func (n *PrintChannel) Publish(msg []byte) error {
	slog.Info("publishing message", slog.String("msg", string(msg)))

	n.mutex.Lock()
	// defensive copy to avoid deadlocks
	tmp := make([]func(msg []byte) error, 0, len(n.debugSubscribers))
	for _, f := range n.debugSubscribers {
		tmp = append(tmp, f)
	}
	n.mutex.Unlock()

	for _, f := range tmp {
		if err := f(msg); err != nil {
			return err
		}
	}

	return nil
}

// PublishDebug is dispatched to Subscribe callbacks.
func (n *PrintChannel) PublishDebug(msg []byte) error {
	slog.Info("received subscribed message", slog.String("msg", string(msg)))

	n.mutex.Lock()
	// defensive copy to avoid deadlocks
	tmp := make([]func(msg []byte) error, 0, len(n.subscribers))
	for _, f := range n.subscribers {
		tmp = append(tmp, f)
	}
	n.mutex.Unlock()

	for _, f := range tmp {
		if err := f(msg); err != nil {
			return err
		}
	}

	return nil
}

// SubscribeDebug callback is called for Publish calls.
func (n *PrintChannel) SubscribeDebug(f func(msg []byte) error) (destroy func()) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if n.debugSubscribers == nil {
		n.debugSubscribers = make(map[int]func(msg []byte) error)
	}

	n.hnd++
	hnd := n.hnd
	n.debugSubscribers[hnd] = f

	return func() {
		n.mutex.Lock()
		defer n.mutex.Unlock()
		delete(n.debugSubscribers, hnd)
	}
}

// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package msg provides the outbox of the messages, which a program sends to its backend. A program does not send
// a message itself: an exported function returns an [Out], whose messages the host sends, after the function has
// returned. So a call stays a function of its arguments, which a test or the backend can replay.
//
// A message is a record. An exported function, which takes the state and a record, e.g.
//
//	func OnHit(g *Game, h Hit) msg.Out[Ack]
//
// handles the messages of this type, which the backend sends at any time.
//
//gostrict:header
package msg

import (
	"iter"
	"slices"
)

// Out collects the messages of one type. Create it with [New].
type Out[T any] struct {
	p *[]T
}

// New returns an empty outbox.
func New[T any]() Out[T] {
	var s []T
	return Out[T]{p: &s}
}

// Send adds the message, which the host sends, after the function has returned.
func (o Out[T]) Send(m T) {
	if o.p == nil {
		return // unreachable in a checked program
	}
	*o.p = append(*o.p, m)
}

// Len returns the number of messages.
func (o Out[T]) Len() int32 {
	if o.p == nil {
		return 0
	}
	return int32(len(*o.p))
}

// Values iterates over the messages.
func (o Out[T]) Values() iter.Seq[T] {
	if o.p == nil {
		return slices.Values([]T(nil))
	}
	return slices.Values(*o.p)
}

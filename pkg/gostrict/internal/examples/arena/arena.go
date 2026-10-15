// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package arena is the client of a game, which synchronizes its state with a backend by messages.
package arena

import (
	"math"

	"go.wdy.de/nago/pkg/gostrict/internal/examples/arena/proto"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/msg"
)

// Client is the state, which the host holds.
type Client struct {
	ID     string
	X, Z   float64
	Seq    int32
	HP     int32
	Others list.List[proto.Pos]
}

func NewClient(id string) *Client {
	return &Client{ID: id, HP: 100, Others: list.New[proto.Pos]()}
}

// Frame moves the player along a circle and reports its position to the backend.
func Frame(c *Client, t float64) msg.Out[proto.Move] {
	c.X = 5 * math.Cos(t)
	c.Z = 5 * math.Sin(t)
	c.Seq++
	out := msg.New[proto.Move]()
	out.Send(proto.Move{ID: c.ID, X: c.X, Z: c.Z, Seq: c.Seq})
	return out
}

// OnHit handles a hit, which the backend reports at any time, and confirms it.
func OnHit(c *Client, h proto.Hit) msg.Out[proto.Ack] {
	c.HP = max(c.HP-h.Damage, 0)
	out := msg.New[proto.Ack]()
	out.Send(proto.Ack{ID: c.ID, Seq: h.Seq, HP: c.HP})
	return out
}

// OnWorld takes over the positions of the other players.
func OnWorld(c *Client, w proto.World) {
	c.Others = w.Players.Clone()
}

// Status returns the hit points, the number of moves and the number of other players.
func Status(c *Client) (int32, int32, int32) {
	return c.HP, c.Seq, c.Others.Len()
}

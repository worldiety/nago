// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package orders takes and returns records, like the use case of an application, and keeps a state, like a game.
package orders

import (
	"errors"

	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/num"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

type Item struct {
	Name  string
	Qty   int32
	Price float64
}

type Order struct {
	Customer string
	Items    list.List[Item]
	Discount float64
}

type Summary struct {
	Customer string
	Total    float64
	Lines    list.List[string]
	Biggest  Item
}

// vat is a table, which a package variable may hold, because it can not change.
var vat = list.Freeze(list.Of(0.19, 0.07))

// Summarize totals the order like a use case.
func Summarize(o Order, reduced bool) (Summary, error) {
	if o.Customer == "" {
		return Summary{Lines: list.New[string]()}, errors.New("the order has no customer")
	}
	rate, _ := vat.Get(0)
	if reduced {
		rate, _ = vat.Get(1)
	}

	s := Summary{Customer: o.Customer, Lines: list.New[string]()}
	for _, it := range o.Items.All() {
		if it.Qty <= 0 {
			return Summary{Lines: list.New[string]()}, errors.New("the quantity of " + it.Name + " must be positive")
		}
		sum := float64(it.Qty) * it.Price
		s.Total += sum
		s.Lines.Append(text.FromInt(it.Qty) + "x " + it.Name + " = " + text.FromInt(num.Round32(sum*100)) + " ct")
		if sum > float64(s.Biggest.Qty)*s.Biggest.Price {
			s.Biggest = it
		}
	}
	s.Total = s.Total * (1 - o.Discount) * (1 + rate)
	return s, nil
}

// Game is the state of a game, which the host holds between the frames.
type Game struct {
	Pos    float64
	Vel    float64
	Bounce int32
	Trail  list.List[float64]
}

func NewGame(height float64) *Game {
	return &Game{Pos: height, Trail: list.New[float64]()}
}

// Step lets the ball fall for dt seconds and returns the number of bounces so far.
func Step(g *Game, dt float64) int32 {
	g.Vel -= 9.81 * dt
	g.Pos += g.Vel * dt
	if g.Pos < 0 {
		g.Pos = -g.Pos
		g.Vel = -g.Vel * 0.8
		g.Bounce++
		g.Trail.Append(g.Pos)
	}
	return g.Bounce
}

// Snapshot returns a copy of the state as a record.
func Snapshot(g *Game) Game {
	return Game{Pos: g.Pos, Vel: g.Vel, Bounce: g.Bounce, Trail: g.Trail.Clone()}
}

// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package physics simulates a body, which falls, lands and jumps, like in a platformer.
package physics

import (
	"math"

	"go.wdy.de/nago/pkg/gostrict/std/list"
)

type Vec struct {
	X, Y float64
}

func (v Vec) Add(o Vec) Vec {
	return Vec{v.X + o.X, v.Y + o.Y}
}

func (v Vec) Scale(f float64) Vec {
	return Vec{v.X * f, v.Y * f}
}

func (v Vec) Len() float64 {
	return math.Hypot(v.X, v.Y)
}

type Rect struct {
	X, Y, W, H float64
}

func (r Rect) Contains(p Vec) bool {
	return p.X >= r.X && p.X <= r.X+r.W && p.Y >= r.Y && p.Y <= r.Y+r.H
}

type Body struct {
	Pos, Vel Vec
	OnGround bool
	Jumps    int32
}

type World struct {
	Player  *Body
	Tiles   list.List[Rect]
	Gravity float64
}

func (w *World) Step(dt float64, jump bool) {
	p := w.Player
	if jump && p.OnGround {
		p.Vel.Y = -12
		p.OnGround = false
		p.Jumps++
	}
	p.Vel.Y += w.Gravity * dt
	p.Pos = p.Pos.Add(p.Vel.Scale(dt))
	for _, t := range w.Tiles.All() {
		if t.Contains(p.Pos) && p.Vel.Y > 0 {
			p.Pos.Y = t.Y
			p.Vel.Y = 0
			p.OnGround = true
		}
	}
}

func Simulate(steps int32, jumpEvery int32) (float64, float64, bool, int32) {
	w := &World{
		Player:  &Body{Pos: Vec{0, 0}},
		Tiles:   list.Of(Rect{-100, 5, 200, 2}, Rect{8, 3, 4, 1}),
		Gravity: 30,
	}
	for i := int32(0); i < steps; i++ {
		jump := false
		if jumpEvery > 0 {
			jump = i%jumpEvery == 0
		}
		w.Step(1.0/60, jump)
		w.Player.Pos.X += 0.1
	}
	p := w.Player
	return p.Pos.X, p.Pos.Y, p.OnGround, p.Jumps
}

func Speed(x, y float64) float64 {
	return Vec{x, y}.Scale(math.Sqrt(2)).Len()
}

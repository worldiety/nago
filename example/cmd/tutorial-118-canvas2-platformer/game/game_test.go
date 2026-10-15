// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package game

import (
	"testing"

	"go.wdy.de/nago/pkg/gostrict/lib/input"
)

// TestPlay runs the game natively at 60 frames per second: the player runs to the right, jumps over the gap and
// collects coins.
func TestPlay(t *testing.T) {
	g, hello := Init("ada", "/sprites.png")
	if hello.Len() != 1 {
		t.Fatal("the game must greet its backend")
	}
	Resize(g, 800, 400)
	startX, startY := g.X, g.Y

	frame := 0
	run := func(seconds float64) {
		for range int(seconds * 60) {
			frame++
			Frame(g, float64(frame)/60)
		}
	}
	run(0.5)
	if !g.OnGround || g.Y != startY && g.Y < startY {
		t.Fatalf("the player must stand on the ground: %+v", g)
	}

	OnInput(g, input.Event{Type: input.KeyDown, Code: "ArrowRight"})
	run(0.5)
	if g.X < startX+90 {
		t.Fatalf("the player must run: x %.1f from %.1f", g.X, startX)
	}

	y := g.Y
	OnInput(g, input.Event{Type: input.KeyDown, Code: "Space"})
	run(0.2)
	if g.Y > y-40 {
		t.Fatalf("the player must jump: y %.1f from %.1f", g.Y, y)
	}
	OnInput(g, input.Event{Type: input.KeyUp, Code: "Space"})
	run(1)
	OnInput(g, input.Event{Type: input.KeyUp, Code: "ArrowRight"})
	run(0.5)
	if !g.OnGround {
		t.Fatalf("the player must land: %+v", g)
	}
	t.Logf("after %d frames: x %.1f y %.1f coins %d time %.2f", frame, g.X, g.Y, g.Coins, g.Time)

	// a touch in the middle jumps, its release stops
	OnInput(g, input.Event{Type: input.PointerDown, PointerID: 7, X: 400, Y: 200})
	if !g.Jump {
		t.Fatal("a touch in the middle must jump")
	}
	OnInput(g, input.Event{Type: input.PointerUp, PointerID: 7, X: 400, Y: 200})
	if g.Jump || g.Pointers.Len() != 0 {
		t.Fatal("the release must stop the jump")
	}
}

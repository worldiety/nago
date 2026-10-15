// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package book

import (
	"math"
	"testing"

	"go.wdy.de/nago/example/cmd/tutorial-119-canvas2-photobook/book/proto"
	"go.wdy.de/nago/pkg/gostrict/lib/input"
)

// TestEdit edits the double page natively like a user with the pointer and the keys.
func TestEdit(t *testing.T) {
	b, hello := Init()
	if hello.Len() != 1 {
		t.Fatal("the editor must greet its backend")
	}
	Resize(b, 1000, 600)
	scale, ox, oy := view(b)
	px := func(x, y float64) (float64, float64) { return ox + x*scale, oy + y*scale }
	ev := func(typ input.Type, x, y float64) input.Event {
		sx, sy := px(x, y)
		return input.Event{Type: typ, X: sx, Y: sy}
	}

	if out := OnAdd(b, proto.Add{URL: "/a.jpg", Width: 400, Height: 300}); out.Len() != 1 {
		t.Fatal("an added photo must be saved")
	}
	OnAdd(b, proto.Add{URL: "/b.png", Width: 300, Height: 600})
	a, _ := b.Photos.Get(0)
	if a.W != 120 || a.H != 90 {
		t.Fatalf("unexpected size %v", a)
	}

	// drag the second photo almost onto the fold: its left edge snaps to the middle of the double page
	p, _ := b.Photos.Get(1)
	OnInput(b, ev(input.PointerDown, p.X, p.Y))
	if b.Selected != 1 || b.Mode != Move {
		t.Fatalf("the photo must be grabbed: %d %d", b.Selected, b.Mode)
	}
	dx := pageW - (p.X - p.W/2) + 1
	OnInput(b, ev(input.PointerMove, p.X+dx, p.Y))
	if out := OnInput(b, ev(input.PointerUp, p.X+dx, p.Y)); out.Len() != 1 {
		t.Fatal("a move must be saved")
	}
	p, _ = b.Photos.Get(1)
	if math.Abs(p.X-p.W/2-pageW) > 1e-9 {
		t.Fatalf("the left edge must snap to the fold: %v", p.X-p.W/2)
	}

	// scale by the corner to double size
	OnInput(b, ev(input.PointerDown, p.X+p.W/2, p.Y+p.H/2))
	if b.Mode != Scale {
		t.Fatalf("the corner must scale: %d", b.Mode)
	}
	OnInput(b, ev(input.PointerMove, p.X+p.W, p.Y+p.H))
	OnInput(b, ev(input.PointerUp, p.X+p.W, p.Y+p.H))
	q, _ := b.Photos.Get(1)
	if math.Abs(q.W-2*p.W) > 1e-6 || math.Abs(q.H-2*p.H) > 1e-6 {
		t.Fatalf("the photo must be twice as large: %v from %v", q, p)
	}

	// the selected photo goes a layer down and is deleted
	OnInput(b, input.Event{Type: input.KeyDown, Code: "BracketLeft"})
	if top, _ := b.Photos.Get(1); top.URL != "/a.jpg" || b.Selected != 0 {
		t.Fatalf("the layers must be swapped: %v %d", top, b.Selected)
	}
	OnInput(b, input.Event{Type: input.KeyDown, Code: "Delete"})
	if b.Photos.Len() != 1 || b.Selected != -1 {
		t.Fatalf("the photo must be deleted: %d %d", b.Photos.Len(), b.Selected)
	}
	Frame(b, 1)
}

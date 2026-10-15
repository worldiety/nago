// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"context"
	"math"
	"testing"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/native"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/shapes"
	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/list"
)

func drawing() shapes.Drawing {
	tags := dict.New[string, shapes.Color]()
	tags.Set("sky", shapes.Blue)
	tags.Set("grass", shapes.Green)
	tags.Set("sun", shapes.Red)
	return shapes.Drawing{
		Name: "scene",
		Shapes: list.Of[shapes.Shape](
			shapes.Circle{R: 1},
			shapes.Rect{W: 2, H: 3},
			shapes.Poly{Points: list.Of(shapes.Point{X: 0, Y: 0}, shapes.Point{X: 4, Y: 0}, shapes.Point{X: 0, Y: 4})},
		),
		Tags: tags,
	}
}

// TestUnions passes sums, maps and lists of sums through the typed client of both engines.
func TestUnions(t *testing.T) {
	b := compile(t, "internal/examples/shapes")
	if u := b.Unions["go.wdy.de/nago/pkg/gostrict/internal/examples/shapes.Shape"]; len(u.Variants) != 3 {
		t.Fatalf("unexpected union %+v", u)
	}

	ctx := context.Background()
	for name, c := range map[string]shapes.Client{
		"native": {Instance: mustLoad(t, native.Engine{Exports: shapes.Exports()}, b)},
		"goja":   {Instance: mustLoad(t, goEngine, b)},
	} {
		area, err := c.Area(ctx, shapes.Circle{R: 2})
		if err != nil || math.Abs(area-4*math.Pi) > 1e-9 {
			t.Fatal(name, area, err)
		}

		scaled, err := c.Scale(ctx, shapes.Rect{W: 2, H: 3}, 2)
		if err != nil || scaled != (shapes.Rect{W: 4, H: 6}) {
			t.Fatal(name, scaled, err)
		}

		total, summary, err := c.Summary(ctx, drawing())
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(total-(math.Pi+6+8)) > 1e-9 || summary != "scene: 3 shapes, sky=blue grass=green sun=red " {
			t.Fatalf("%s: %v %q", name, total, summary)
		}

		largest, err := c.Largest(ctx, drawing(), shapes.Circle{R: 0})
		if err != nil {
			t.Fatal(err)
		}
		if p, ok := largest.(shapes.Poly); !ok || p.Points.Len() != 3 {
			t.Fatalf("%s: largest %#v", name, largest)
		}
	}
}

type canvasBackend struct {
	totals []shapes.Total
}

func (b *canvasBackend) Total(ctx context.Context, m shapes.Total) error {
	b.totals = append(b.totals, m)
	return nil
}

// TestUnionMessages sends the variants of a union to one handler.
func TestUnionMessages(t *testing.T) {
	b := compile(t, "internal/examples/shapes")
	ctx := context.Background()
	for _, e := range []gostrict.Engine{native.Engine{Exports: shapes.Exports()}, goEngine} {
		in := mustLoad(t, e, b)
		res, err := in.Call(ctx, "NewCanvas")
		if err != nil {
			t.Fatal(err)
		}
		be := &canvasBackend{}
		s := gostrict.NewSession(in, b, res[0].(gostrict.Token), func(ctx context.Context, m gostrict.Message) error {
			return shapes.Dispatch(ctx, be, m)
		})
		for _, ev := range []shapes.Event{
			shapes.Added{Shape: shapes.Rect{W: 1, H: 2}},
			shapes.Added{Shape: shapes.Circle{R: 1}},
			shapes.Cleared{},
			shapes.Added{Shape: shapes.Rect{W: 3, H: 3}},
		} {
			if err := s.Deliver(ctx, shapes.EventMessage(ev)); err != nil {
				t.Fatal(err)
			}
		}
		if len(be.totals) != 4 || be.totals[1].Count != 2 || be.totals[2].Count != 0 || be.totals[3] != (shapes.Total{Area: 9, Count: 1}) {
			t.Fatalf("unexpected totals %v", be.totals)
		}
	}
}

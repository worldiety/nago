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

	gostrict.Variants[shapes.Shape](shapes.Circle{}, shapes.Rect{}, shapes.Poly{})

	ctx := context.Background()
	for name, in := range map[string]gostrict.Instance{
		"native": mustLoad(t, nativeEngine("shapes"), b),
		"goja":   mustLoad(t, goEngine, b),
	} {
		var area float64
		if err := gostrict.Call(ctx, in, "Area", []any{shapes.Circle{R: 2}}, &area); err != nil || math.Abs(area-4*math.Pi) > 1e-9 {
			t.Fatal(name, area, err)
		}

		var scaled shapes.Shape
		if err := gostrict.Call(ctx, in, "Scale", []any{shapes.Rect{W: 2, H: 3}, 2.0}, &scaled); err != nil || scaled != (shapes.Rect{W: 4, H: 6}) {
			t.Fatal(name, scaled, err)
		}

		var total float64
		var summary string
		if err := gostrict.Call(ctx, in, "Summary", []any{drawing()}, &total, &summary); err != nil {
			t.Fatal(err)
		}
		if math.Abs(total-(math.Pi+6+8)) > 1e-9 || summary != "scene: 3 shapes, sky=blue grass=green sun=red " {
			t.Fatalf("%s: %v %q", name, total, summary)
		}

		var largest shapes.Shape
		if err := gostrict.Call(ctx, in, "Largest", []any{drawing(), shapes.Circle{R: 0}}, &largest); err != nil {
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
	gostrict.Variants[shapes.Shape](shapes.Circle{}, shapes.Rect{}, shapes.Poly{})
	gostrict.Variants[shapes.Event](shapes.Added{}, shapes.Cleared{})
	for _, e := range []gostrict.Engine{nativeEngine("shapes"), goEngine} {
		in := mustLoad(t, e, b)
		res, err := in.Call(ctx, "NewCanvas")
		if err != nil {
			t.Fatal(err)
		}
		be := &canvasBackend{}
		router := gostrict.Handle(new(gostrict.Router), be.Total)
		s := gostrict.NewSession(in, b, res[0].(gostrict.Token), func(ctx context.Context, m gostrict.Message) error {
			return router.Dispatch(ctx, m)
		})
		for _, ev := range []shapes.Event{
			shapes.Added{Shape: shapes.Rect{W: 1, H: 2}},
			shapes.Added{Shape: shapes.Circle{R: 1}},
			shapes.Cleared{},
			shapes.Added{Shape: shapes.Rect{W: 3, H: 3}},
		} {
			if err := s.Send(ctx, ev); err != nil {
				t.Fatal(err)
			}
		}
		if len(be.totals) != 4 || be.totals[1].Count != 2 || be.totals[2].Count != 0 || be.totals[3] != (shapes.Total{Area: 9, Count: 1}) {
			t.Fatalf("unexpected totals %v", be.totals)
		}
	}
}

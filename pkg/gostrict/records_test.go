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
	"slices"
	"testing"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/native"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/orders"
	"go.wdy.de/nago/pkg/gostrict/std/list"
)

// TestRecordsAndState passes records with lists and a state through the typed client of both engines.
func TestRecordsAndState(t *testing.T) {
	b := compile(t, "internal/examples/orders")
	ctx := context.Background()

	if rec := b.Records["go.wdy.de/nago/pkg/gostrict/internal/examples/orders.Order"]; len(rec.Fields) != 3 || rec.Fields[1].Type.String() != "list.List[go.wdy.de/nago/pkg/gostrict/internal/examples/orders.Item]" {
		t.Fatalf("unexpected schema %+v", rec)
	}

	clients := map[string]orders.Client{
		"native": {Instance: mustLoad(t, native.Engine{Exports: orders.Exports()}, b)},
		"goja":   {Instance: mustLoad(t, goEngine, b)},
	}

	order := orders.Order{
		Customer: "Ada",
		Items:    list.Of(orders.Item{Name: "Bolt", Qty: 60, Price: 0.1}, orders.Item{Name: "Gear", Qty: 2, Price: 7.5}),
		Discount: 0.05,
	}

	type result struct {
		total   float64
		lines   []string
		biggest orders.Item
		err     string
		bounces int32
		game    orders.Game
	}
	results := map[string]result{}
	for name, c := range clients {
		var r result
		s, err, callErr := c.Summarize(ctx, order, false)
		if callErr != nil {
			t.Fatal(name, callErr)
		}
		if err != nil {
			t.Fatal(name, err)
		}
		r.total, r.lines, r.biggest = s.Total, slices.Collect(s.Lines.Values()), s.Biggest

		_, err, callErr = c.Summarize(ctx, orders.Order{Items: list.New[orders.Item]()}, true)
		if callErr != nil || err == nil {
			t.Fatal(name, callErr, err)
		}
		r.err = err.Error()

		game, err := c.NewGame(ctx, 10)
		if err != nil {
			t.Fatal(err)
		}
		for range 600 {
			if r.bounces, err = c.Step(ctx, game, 1.0/60); err != nil {
				t.Fatal(err)
			}
		}
		if r.game, err = c.Snapshot(ctx, game); err != nil {
			t.Fatal(err)
		}
		results[name] = r
	}

	n, g := results["native"], results["goja"]
	t.Logf("total %.4f, lines %v, biggest %v, error %q, %d bounces, trail %d", n.total, n.lines, n.biggest, n.err, n.bounces, n.game.Trail.Len())
	if math.Abs(n.total-g.total) > 1e-9 || !slices.Equal(n.lines, g.lines) || n.biggest != g.biggest || n.err != g.err {
		t.Fatalf("the summaries differ\nnative %+v\ngoja   %+v", n, g)
	}
	if n.bounces != g.bounces || n.bounces == 0 || math.Abs(n.game.Pos-g.game.Pos) > 1e-9 || n.game.Trail.Len() != g.game.Trail.Len() {
		t.Fatalf("the games differ\nnative %+v\ngoja   %+v", n.game, g.game)
	}
	if n.lines[0] != "60x Bolt = 600 ct" || n.biggest.Name != "Gear" {
		t.Fatalf("unexpected summary %+v", n)
	}

	// a state can not cross to another engine
	if _, err := clients["goja"].Step(ctx, gostrict.State[orders.Game]{Token: gostrict.NewToken("go.wdy.de/nago/pkg/gostrict/internal/examples/orders.Game", &orders.Game{})}, 1); err == nil {
		t.Fatal("a native state was accepted by goja")
	}
}

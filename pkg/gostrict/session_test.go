// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/native"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/arena"
	"go.wdy.de/nago/pkg/gostrict/internal/examples/arena/proto"
	"go.wdy.de/nago/pkg/gostrict/std/list"
)

// backend is the server of the arena in normal Go. It answers every tenth move with a hit and every 25th with
// the state of the world, both as messages, which the program did not request.
type backend struct {
	mu      sync.Mutex
	session *gostrict.Session
	moves   []proto.Move
	acks    []proto.Ack
}

func (b *backend) Move(ctx context.Context, m proto.Move) error {
	b.mu.Lock()
	b.moves = append(b.moves, m)
	n := len(b.moves)
	b.mu.Unlock()

	if n%10 == 0 {
		if err := b.session.Deliver(ctx, arena.HitMessage(proto.Hit{Target: m.ID, Damage: 7, Seq: int32(n)})); err != nil {
			return err
		}
	}
	if n%25 == 0 {
		players := list.New[proto.Pos]()
		for i := range n / 25 {
			players.Append(proto.Pos{ID: fmt.Sprint("bot", i), X: float64(i), Z: m.Z})
		}
		return b.session.Deliver(ctx, arena.WorldMessage(proto.World{Players: players}))
	}
	return nil
}

func (b *backend) Ack(ctx context.Context, m proto.Ack) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.acks = append(b.acks, m)
	return nil
}

// TestSession synchronizes the arena program with its backend through messages in both engines.
func TestSession(t *testing.T) {
	b := compile(t, "internal/examples/arena")
	if b.Handlers["go.wdy.de/nago/pkg/gostrict/internal/examples/arena/proto.Hit"] != "OnHit" {
		t.Fatalf("unexpected handlers %v", b.Handlers)
	}

	run := func(e gostrict.Engine) (*backend, []any) {
		ctx := context.Background()
		in := mustLoad(t, e, b)
		res, err := in.Call(ctx, "NewClient", "ada")
		if err != nil {
			t.Fatal(err)
		}

		be := &backend{}
		be.session = gostrict.NewSession(in, b, res[0].(gostrict.Token), func(ctx context.Context, m gostrict.Message) error {
			return arena.Dispatch(ctx, be, m)
		})

		// a push of another goroutine, while the frames run, is serialized with them
		var wg sync.WaitGroup
		wg.Go(func() {
			if err := be.session.Deliver(ctx, arena.HitMessage(proto.Hit{Target: "ada", Damage: 1, Seq: -1})); err != nil {
				t.Error(err)
			}
		})
		for i := range 100 {
			if _, err := be.session.Call(ctx, "Frame", float64(i)/10); err != nil {
				t.Fatal(err)
			}
		}
		wg.Wait()

		status, err := be.session.Call(ctx, "Status")
		if err != nil {
			t.Fatal(err)
		}
		return be, status
	}

	nb, ns := run(native.Engine{Exports: arena.Exports()})
	gb, gs := run(goEngine)

	t.Logf("status hp, moves, others: native %v, goja %v; %d moves and %d acks at the backend", ns, gs, len(nb.moves), len(nb.acks))
	if !gostrict.SameValues(ns, gs) || ns[0] != int32(100-10*7-1) || ns[1] != int32(100) || ns[2] != int32(4) {
		t.Fatalf("unexpected status: native %v, goja %v", ns, gs)
	}
	if len(nb.moves) != 100 || len(gb.moves) != 100 || len(nb.acks) != 11 || len(gb.acks) != 11 {
		t.Fatalf("unexpected messages: native %d/%d, goja %d/%d", len(nb.moves), len(nb.acks), len(gb.moves), len(gb.acks))
	}
	for i := range nb.moves {
		n, g := nb.moves[i], gb.moves[i]
		if !gostrict.CloseValues([]any{n.ID, n.X, n.Z, n.Seq}, []any{g.ID, g.X, g.Z, g.Seq}, 1e-9) {
			t.Fatalf("move %d differs: %v %v", i, n, g)
		}
	}
}

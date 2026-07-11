// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ndbinspector_test

import (
	"testing"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/inspector/ndb"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/ndb"
	"go.wdy.de/nago/pkg/ndb/msgstore"
)

func TestUseCasesMessages(t *testing.T) {
	root := t.TempDir()
	db := option.Must(ndb.Open(root, ndb.Options{}))
	defer db.Close()

	// seed two message streams via the msgstore engine
	eng, err := db.Engine("events", ndb.EngineOptions{Kind: msgstore.EngineKind, Config: msgstore.Options{}})
	if err != nil {
		t.Fatal(err)
	}
	m := eng.(ndb.MessageEngine).Messages()
	var tr ndb.TraceID
	for i := 0; i < 5; i++ {
		if _, err := m.Append("orders", tr, []byte("order")); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 3; i++ {
		if _, err := m.Append("users", tr, []byte("user")); err != nil {
			t.Fatal(err)
		}
	}

	const instPath = "/db/main"
	uc := ndbinspector.NewUseCases(func() []ndbinspector.Instance {
		return []ndbinspector.Instance{{Path: instPath, Name: "main", DB: db}}
	})
	su := user.SU()

	// instances
	instances, err := uc.Instances(su)
	if err != nil {
		t.Fatal(err)
	}
	if len(instances) != 1 || instances[0].Path != instPath {
		t.Fatalf("instances = %+v", instances)
	}

	// engines
	engines, err := uc.MessageEngines(su, instPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(engines) != 1 || engines[0].Name != "events" || engines[0].Kind != msgstore.EngineKind {
		t.Fatalf("engines = %+v", engines)
	}

	// types + stats
	types, err := uc.Types(su, instPath, "events")
	if err != nil {
		t.Fatal(err)
	}
	if len(types) != 2 {
		t.Fatalf("want 2 types, got %+v", types)
	}
	var orders ndbinspector.TypeInfo
	for _, ti := range types {
		if ti.Type == "orders" {
			orders = ti
		}
	}
	if orders.Type != "orders" || orders.Bytes <= 0 || orders.Segments < 1 {
		t.Fatalf("orders stat = %+v", orders)
	}

	// window: read all 5 orders
	rows, err := uc.Window(su, ndbinspector.WindowRequest{Instance: instPath, Engine: "events", Types: []ndb.TypeID{"orders"}, MinSeq: 0, Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 5 {
		t.Fatalf("want 5 order messages, got %d", len(rows))
	}
	for i := 1; i < len(rows); i++ {
		if rows[i].Seq <= rows[i-1].Seq {
			t.Fatalf("messages not strictly ascending by seq: %v", rows)
		}
	}
	if string(rows[0].Payload) != "order" {
		t.Fatalf("payload = %q", rows[0].Payload)
	}

	// window limit is enforced
	limited, err := uc.Window(su, ndbinspector.WindowRequest{Instance: instPath, Engine: "events", Types: []ndb.TypeID{"orders"}, Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(limited) != 2 {
		t.Fatalf("window limit not enforced: %d", len(limited))
	}

	// delete a single seq (tombstone) — the payload disappears from reads
	delSeq := rows[2].Seq
	if err := uc.DeleteSeq(su, instPath, "events", "orders", delSeq); err != nil {
		t.Fatal(err)
	}
	after, err := uc.Window(su, ndbinspector.WindowRequest{Instance: instPath, Engine: "events", Types: []ndb.TypeID{"orders"}, Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range after {
		if r.Seq == delSeq && !r.Tomb {
			t.Fatalf("seq %d should be a tombstone after delete", delSeq)
		}
	}

	// multi-type window: replay orders + users coherently in global Seq order
	multi, err := uc.Window(su, ndbinspector.WindowRequest{Instance: instPath, Engine: "events", Types: []ndb.TypeID{"orders", "users"}, Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	if len(multi) < 6 {
		t.Fatalf("multi-type window too small: %d", len(multi))
	}
	for i := 1; i < len(multi); i++ {
		if multi[i].Seq <= multi[i-1].Seq {
			t.Fatalf("multi-type window not seq-coherent at index %d: %v", i, multi)
		}
	}
	var sawOrders, sawUsers bool
	for _, r := range multi {
		switch r.Type {
		case "orders":
			sawOrders = true
		case "users":
			sawUsers = true
		}
	}
	if !sawOrders || !sawUsers {
		t.Fatalf("multi-type window missing a type: orders=%v users=%v", sawOrders, sawUsers)
	}

	// delete the whole users stream
	if err := uc.DeleteType(su, instPath, "events", "users"); err != nil {
		t.Fatal(err)
	}
	types2, err := uc.Types(su, instPath, "events")
	if err != nil {
		t.Fatal(err)
	}
	for _, ti := range types2 {
		if ti.Type == "users" {
			t.Fatalf("users stream should be gone: %+v", types2)
		}
	}
}

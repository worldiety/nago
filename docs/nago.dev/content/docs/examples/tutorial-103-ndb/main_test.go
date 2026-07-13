// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"context"
	"testing"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/ndb"
	"go.wdy.de/nago/pkg/ndb/msgstore"
	"go.wdy.de/nago/pkg/ndb/tsdb"
	"go.wdy.de/nago/pkg/timeseries"
)

// openApp builds the runtime objects directly on a temp ndb database (bypassing
// the Configurator/HTTP server) so the domain wiring can be tested headlessly.
func openApp(t *testing.T) (*app, *ndb.DB) {
	t.Helper()
	db := option.Must(ndb.Open(t.TempDir(), ndb.Options{}))

	msgEng := option.Must(db.Engine("accounts", ndb.EngineOptions{Kind: msgstore.EngineKind, Config: msgstore.Options{}}))
	msgs := msgEng.(ndb.MessageEngine).Messages()

	backend := evs.NewNDBBackend[AccEvt, *Account](msgs)
	handler := evs.NewHandler[*Account](backend, accountID, backend.Register)
	handler.RegisterEvents(MoneyDeposited{}, MoneyWithdrawn{})

	ledger := newLedgerProjection(msgs)
	ledger.Run()

	tsEng := option.Must(db.Engine("metrics", ndb.EngineOptions{Kind: tsdb.EngineKind, Config: tsdb.Options{}}))
	tdb := tsEng.(interface{ DB() *tsdb.DB }).DB()
	col := option.Must(tdb.Column(tsBucket, tsColumn, tsdb.Schema{Scheme: tsdb.SchemeDecimal, Decimals: 2}))
	colStr := option.Must(tdb.Column(tsBucket, tsColumnStr, tsdb.Schema{Scheme: tsdb.SchemeString}))

	a := &app{handler: handler, ledger: ledger, tsColumn: col, tsColumnStr: colStr}
	a.tsBaseMs = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	a.tsLastMs = a.tsBaseMs + int64(tsCount-1)*tsStepMs
	return a, db
}

func TestEventSourcingHandlerAndProjection(t *testing.T) {
	a, db := openApp(t)
	defer db.Close()
	su := user.SU()
	ctx := context.Background()

	// deposit 1000, withdraw 300 -> balance 700
	if _, err := a.handler.Handle(su, "acc-1", DepositCmd{ID: "acc-1", Amount: 1000}); err != nil {
		t.Fatal(err)
	}
	seq, err := a.handler.Handle(su, "acc-1", WithdrawCmd{ID: "acc-1", Amount: 300})
	if err != nil {
		t.Fatal(err)
	}

	// write model (handler) reflects the balance
	acc := option.Must(a.handler.Aggregate(ctx, "acc-1"))
	if acc.Balance != 700 {
		t.Fatalf("handler balance = %d, want 700", acc.Balance)
	}

	// invariant: overdraft is rejected by Decide
	if _, err := a.handler.Handle(su, "acc-1", WithdrawCmd{ID: "acc-1", Amount: 999999}); err == nil {
		t.Fatal("overdraft must be rejected")
	}

	// read model (projection) is eventually consistent; wait for the tail
	if err := a.ledger.WaitFor(ctx, ndb.Seq(seq)); err != nil {
		t.Fatal(err)
	}
	stats, ok := evs.Value(a.ledger)
	if !ok {
		t.Fatal("ledger projection has no value")
	}
	if stats.Deposits != 1000 || stats.Withdrawals != 300 || stats.Count != 2 {
		t.Fatalf("ledger = %+v, want deposits 1000 withdrawals 300 count 2", stats)
	}
}

func TestTimeseriesSeedAndM4(t *testing.T) {
	a, db := openApp(t)
	defer db.Close()

	if !isColumnEmpty(a.tsColumn, a.tsBaseMs, a.tsLastMs) {
		t.Fatal("fresh column should be empty")
	}
	if err := seedTimeseries(a.tsColumn, a.tsBaseMs); err != nil {
		t.Fatal(err)
	}
	if isColumnEmpty(a.tsColumn, a.tsBaseMs, a.tsLastMs) {
		t.Fatal("column should be populated after seeding")
	}

	// raw point count
	var raw int
	if err := a.tsColumn.ScanF64(a.tsBaseMs, a.tsLastMs, func(ts []int64, _ []float64) bool {
		raw += len(ts)
		return true
	}); err != nil {
		t.Fatal(err)
	}
	if raw != tsCount {
		t.Fatalf("raw points = %d, want %d", raw, tsCount)
	}

	// M4 reduces to at most 4*width points, far fewer than raw
	const width = 200
	rng := timeseries.NewRange(timeseries.UnixMilli(a.tsBaseMs), timeseries.UnixMilli(a.tsLastMs), time.UTC)
	var n int
	for range timeseries.M4(a.tsColumn.IterF64(a.tsBaseMs, a.tsLastMs), rng, width) {
		n++
	}
	if n == 0 {
		t.Fatal("M4 produced no points")
	}
	if n > 4*width {
		t.Fatalf("M4 produced %d points, must be <= %d", n, 4*width)
	}
	if n >= raw {
		t.Fatalf("M4 did not downsample: %d >= %d", n, raw)
	}
}

func TestStatusTimeseriesSeed(t *testing.T) {
	a, db := openApp(t)
	defer db.Close()

	strLast := a.tsBaseMs + int64(tsStrCount-1)*tsStrStepMs
	if !isColumnEmptyStr(a.tsColumnStr, a.tsBaseMs, strLast) {
		t.Fatal("fresh string column should be empty")
	}
	if err := seedStatusTimeseries(a.tsColumnStr, a.tsBaseMs); err != nil {
		t.Fatal(err)
	}
	if isColumnEmptyStr(a.tsColumnStr, a.tsBaseMs, strLast) {
		t.Fatal("string column should be populated after seeding")
	}

	// read back all values and check count + that several distinct states occur
	var count int
	seen := map[string]bool{}
	if err := a.tsColumnStr.ScanString(a.tsBaseMs, strLast, func(ts []int64, vals []string) bool {
		for i := range ts {
			count++
			seen[vals[i]] = true
		}
		return true
	}); err != nil {
		t.Fatal(err)
	}
	if count != tsStrCount {
		t.Fatalf("status points = %d, want %d", count, tsStrCount)
	}
	if !seen["running"] || !seen["error"] || !seen["warning"] {
		t.Fatalf("expected multiple distinct states, saw %v", seen)
	}
}

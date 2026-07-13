// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package main provides a stress/benchmark playground for the store inspector
// (admin center -> Inspektor -> Stores). It populates a single entity store
// with a large amount of entries (4 million by default), so that the paging,
// counting and selection behavior of the inspector can be investigated with a
// realistic, huge dataset.
//
// The store inspector is expected to be efficient because it iterates ids only
// and pages the result. This example makes it easy to reproduce and measure the
// actual runtime behavior with millions of entries.
package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	cfginspector "go.wdy.de/nago/application/inspector/cfg"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

// entryCount is the amount of sample entries which are populated into the
// benchmark store on the very first start. 4 million entries are enough to make
// any accidental O(n) or O(n^2) behavior in the inspector clearly visible.
const entryCount = 4_000_000

// progressEvery controls how often the population progress is logged.
const progressEvery = 100_000

// BenchRecordID is the identity type of our sample aggregate.
type BenchRecordID string

// BenchRecord is a small, but non-trivial aggregate. It is serialized as JSON
// into the entity store, so each entry has a realistic payload (a few hundred
// bytes) instead of being empty.
type BenchRecord struct {
	ID        BenchRecordID `json:"id,omitempty"`
	Index     int           `json:"index,omitempty"`
	Firstname string        `json:"firstname,omitempty"`
	Lastname  string        `json:"lastname,omitempty"`
	Email     string        `json:"email,omitempty"`
	Age       int           `json:"age,omitempty"`
	City      string        `json:"city,omitempty"`
	Note      string        `json:"note,omitempty"`
	CreatedAt time.Time     `json:"created_at,omitempty"`
}

func (r BenchRecord) Identity() BenchRecordID {
	return r.ID
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_105")

		cfg.Serve(vuejs.Dist())
		cfg.SetDecorator(cfg.NewScaffold().
			Login(true).
			Decorator())

		option.MustZero(cfg.StandardSystems())

		// The bootstrap admin (admin@localhost) automatically receives all
		// nago.* permissions, which includes nago.data.inspector. Thus, after
		// logging in you can directly open the store inspector without any
		// manual role setup.
		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))

		// Enable the store inspector: Admin Center -> Inspektor -> Stores.
		option.Must(cfginspector.Enable(cfg))

		// Populate the benchmark store. This happens in the background so that
		// the web server is available immediately. On subsequent starts the
		// store is already filled and population is skipped.
		populateBenchStore(cfg)

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.Text("Store Inspector Benchmark").Font(ui.Title),
				ui.Text(fmt.Sprintf("This app populates a store named %q with up to %d entries.", benchStoreName(), entryCount)),
				ui.Text("Open the Admin Center -> Inspektor -> Stores and select the store to investigate the inspector performance."),
				ui.Text("Login: admin@localhost / %6UbRsCuM8N$auy"),
			).Gap(ui.L16).Frame(ui.Frame{}.MatchScreen())
		})
	}).
		Run()
}

// benchStoreName returns the entity store bucket name that SloppyRepository
// derives from the reflected type name of BenchRecord. We expose it so the UI
// and logs can reference the exact store the user has to select in the
// inspector.
func benchStoreName() string {
	return "BenchRecord"
}

func populateBenchStore(cfg *application.Configurator) {
	// SloppyRepository uses the reflected type name ("BenchRecord") as the
	// bucket name, which is exactly the store that will show up in the
	// inspector.
	repo := application.SloppyRepository[BenchRecord, BenchRecordID](cfg)

	// Count() is O(1) on the tdb entity store (backed by an in-memory btree),
	// so this idempotency guard is cheap even for millions of entries.
	current := option.Must(repo.Count())
	if current >= entryCount {
		slog.Info("bench store already populated, skipping", "store", benchStoreName(), "count", current)
		return
	}

	slog.Info("populating bench store in background", "store", benchStoreName(), "target", entryCount, "existing", current)

	go func() {
		start := time.Now()

		// We save via an iterator, so at most one record is materialized at a
		// time. This keeps the populator at O(1) memory even for 4 million
		// entries.
		it := func(yield func(BenchRecord) bool) {
			for i := current; i < entryCount; i++ {
				if !yield(makeRecord(i)) {
					return
				}

				if (i+1)%progressEvery == 0 {
					elapsed := time.Since(start)
					perSec := float64(i+1-current) / elapsed.Seconds()
					slog.Info("populating bench store",
						"written", i+1,
						"target", entryCount,
						"elapsed", elapsed.Round(time.Second),
						"records_per_sec", int(perSec),
					)
				}
			}
		}

		if err := repo.SaveAll(it); err != nil {
			slog.Error("failed to populate bench store", "err", err.Error())
			return
		}

		slog.Info("finished populating bench store",
			"store", benchStoreName(),
			"count", entryCount,
			"duration", time.Since(start).Round(time.Second),
		)
	}()
}

var (
	firstnames = []string{"Max", "Anna", "Peter", "Julia", "Lukas", "Laura", "Felix", "Sophie", "Tobias", "Marie"}
	lastnames  = []string{"Müller", "Schmidt", "Schneider", "Fischer", "Weber", "Meyer", "Wagner", "Becker", "Hoffmann", "Schäfer"}
	cities     = []string{"Berlin", "Hamburg", "München", "Köln", "Frankfurt", "Stuttgart", "Düsseldorf", "Leipzig", "Dortmund", "Essen"}
)

// makeRecord builds a deterministic record for the given index. The id is
// zero-padded so that the natural lexicographic order of the store equals the
// numeric order, which makes paging in the inspector easy to follow and fully
// reproducible across runs.
func makeRecord(i int) BenchRecord {
	return BenchRecord{
		ID:        BenchRecordID(fmt.Sprintf("record-%012d", i)),
		Index:     i,
		Firstname: firstnames[i%len(firstnames)],
		Lastname:  lastnames[(i/len(firstnames))%len(lastnames)],
		Email:     fmt.Sprintf("user%d@example.com", i),
		Age:       18 + (i % 60),
		City:      cities[i%len(cities)],
		Note:      "This is a synthetic benchmark record used to stress test the nago store inspector paging and counting behavior.",
		CreatedAt: time.Unix(0, 0).UTC(),
	}
}

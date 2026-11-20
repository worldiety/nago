// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgai

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
)

// The benchmarks of the heads of the conversations in M27 besides those of sessions_bench_test.go. Those which measure
// the heap run in a process of their own, which no other benchmark shares, after BenchmarkHeadsStocks.

// stockS2P and stockS50P are S2 and S50 without heads, which no measurement ever opens with the ai management: the
// stocks of a former version of nago.
var (
	stockS2P  = benchStock{name: "s2p", foreign: 2_000}
	stockS50P = benchStock{name: "s50p", foreign: 50_000}
)

// benchHeap returns the heap after two collections.
func benchHeap() uint64 {
	runtime.GC()
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.HeapAlloc
}

// benchData are the data of a directory, opened like Enable does it, without enabling the ai management: the ReBAC
// database, the conversations and the store of the heads.
type benchData struct {
	cfg   *application.Configurator
	app   *core.Application
	repo  session.Repository
	rdb   *rebac.DB
	heads blob.Store
}

// benchOpenData opens the data of the directory with a fresh configurator. Unlike nagotest, nothing keeps it once it
// has been destroyed, so that a measurement of the heap does not see it.
func benchOpenData(b *testing.B, dir string) *benchData {
	b.Helper()

	cfg := application.NewConfigurator()
	cfg.SetApplicationID(benchAppID)
	cfg.SetDataDir(dir)
	app, err := cfg.NewInProcessApplication()
	if err != nil {
		b.Fatal(err)
	}

	return &benchData{
		cfg:   cfg,
		app:   app,
		rdb:   std.Must(cfg.RDB()),
		repo:  session.Repository(std.Must(application.JSONRepository[session.Session](cfg, string(session.Namespace)))),
		heads: benchHeads(b, cfg),
	}
}

// subject returns the measuring person.
func (d *benchData) subject(b *testing.B, seed benchSeed) auth.Subject {
	b.Helper()

	users := std.Must(d.cfg.UserManagement())
	return std.Must(users.UseCases.SubjectFromUser(user.SU(), seed.User)).Unwrap()
}

// wired registers the ReBAC rules of Enable.
func (d *benchData) wired() {
	d.rdb.RegisterStaticRule(rebac.StaticRule{Source: user.Namespace, Relation: rebac.Owner, Target: session.Namespace})
	for _, pid := range session.InstancePermissions {
		d.rdb.RegisterStaticRule(rebac.StaticRule{Source: user.Namespace, Relation: rebac.Relation(pid), Target: session.Namespace})
	}
}

// benchPrepared returns the seed of the stock, which must have been seeded by an earlier process, e.g. by
// BenchmarkHeadsStocks, so that no configurator of the seeding is around.
func benchPrepared(b *testing.B, stock benchStock) benchSeed {
	b.Helper()

	dir := stock.dir(b)
	if _, err := os.Stat(filepath.Join(dir, benchSeedFile)); err != nil {
		b.Fatalf("stock %s is not prepared, run BenchmarkHeadsStocks first: %v", stock.name, err)
	}

	return stock.seed(b, dir)
}

// BenchmarkHeadsStocks prepares the stocks of the measurements of the heap in a process of its own: it seeds S2P and
// S50P and builds the heads of S2 and S50. It measures nothing.
func BenchmarkHeadsStocks(b *testing.B) {
	for _, stock := range []benchStock{stockS2P, stockS50P} {
		stock.seed(b, stock.dir(b))
	}

	for _, stock := range []benchStock{stockS2, stockS50} {
		stock.built(b, 0)
	}
}

// BenchmarkHeadsMemory is G3 of M27: the heap after opening S2 and S50 with built heads ("with") against the same
// conversations, which never had heads ("without", the stocks S2P and S50P). Both open the data like Enable without
// enabling the ai management, which would build the heads of S2P and S50P. "with" then enables it and reports the heap
// once more (enabled-B) and after 1 000 lookups of the resumable conversation, 100 loads of the history and 1 000
// renames of different conversations (sequence-B). Run each with -test.benchtime=1x in a process of its own, after
// BenchmarkHeadsStocks.
func BenchmarkHeadsMemory(b *testing.B) {
	for _, c := range []struct {
		name       string
		with, pure benchStock
	}{{"s2", stockS2, stockS2P}, {"s50", stockS50, stockS50P}} {
		b.Run(c.name+"/without", func(b *testing.B) {
			seed := benchPrepared(b, c.pure)
			d := benchOpenData(b, c.pure.dir(b))
			defer d.app.Destroy()
			if n := std.Must(d.repo.Count()); n != seed.Sessions {
				b.Fatalf("%d conversations instead of %d", n, seed.Sessions)
			}
			if n := std.Must(blob.Count(b.Context(), d.heads)); n != 0 {
				b.Fatalf("the stock without heads has %d entries of heads", n)
			}

			b.ResetTimer()
			opened := benchHeap()
			b.StopTimer()

			b.ReportMetric(float64(opened), "opened-B")
			b.ReportMetric(float64(seed.Sessions), "sessions")
			runtime.KeepAlive(d)
		})

		b.Run(c.name+"/with", func(b *testing.B) {
			seed := benchPrepared(b, c.with)
			d := benchOpenData(b, c.with.dir(b))
			defer d.app.Destroy()
			if n := std.Must(d.repo.Count()); n != seed.Sessions {
				b.Fatalf("%d conversations instead of %d", n, seed.Sessions)
			}
			if done, n := benchHeadsState(b, d.heads); !done || n != seed.Sessions {
				b.Fatalf("heads built %v: %d heads for %d conversations, run BenchmarkHeadsStocks first", done, n, seed.Sessions)
			}

			b.ResetTimer()
			opened := benchHeap()

			mgmt := std.Must(Enable(d.cfg))
			enabled := benchHeap()

			subject := d.subject(b, seed)
			uc := mgmt.SessionUseCases
			opts := session.FindResumableOptions{Tags: benchTags, ProviderHint: seed.ProviderHint}
			for range 1_000 {
				if found := std.Must(uc.FindResumable(subject, opts)); found.IsSome() {
					b.Fatalf("found %s", found.Unwrap().ID)
				}
			}

			for range 100 {
				n := 0
				for _, err := range uc.FindSummaries(subject, session.FindAllOptions{Tags: benchTags, CreatedBy: seed.User}) {
					if err != nil {
						b.Fatal(err)
					}
					n++
				}
				if n != benchOwnSessions {
					b.Fatalf("history of %d conversations", n)
				}
			}

			ids := slices.Clone(seed.Own)
			for i := 0; len(ids) < 1_000; i++ {
				ids = append(ids, session.ID(benchIdent("foreign", i)))
			}
			for i, id := range ids {
				if err := uc.Rename(user.SU(), id, fmt.Sprintf("Umbenannt %d", i)); err != nil {
					b.Fatal(err)
				}
			}

			sequence := benchHeap()
			b.StopTimer()

			b.ReportMetric(float64(opened), "opened-B")
			b.ReportMetric(float64(enabled), "enabled-B")
			b.ReportMetric(float64(sequence), "sequence-B")
			b.ReportMetric(float64(int64(sequence)-int64(enabled)), "sequence-delta-B")
			b.ReportMetric(float64(seed.Sessions), "sessions")
			runtime.KeepAlive(d)
		})
	}
}

// BenchmarkHeadsStart is G4 of M27: the start of the session use cases with heads until the first result of the
// lookup of the resumable conversation in S50, without and with 10 marks of open writes, after the data have been
// opened. Each start constructs the use cases anew; the marks are written before each.
func BenchmarkHeadsStart(b *testing.B) {
	for _, marks := range []int{0, 10} {
		b.Run(fmt.Sprintf("marks%d", marks), func(b *testing.B) {
			seed := stockS50.built(b, 0)
			d := benchOpenData(b, stockS50.dir(b))
			defer d.app.Destroy()
			d.wired()
			subject := d.subject(b, seed)
			opts := session.FindResumableOptions{Tags: benchTags, ProviderHint: seed.ProviderHint}

			start := func() {
				uc := session.NewUseCases(d.repo, d.rdb, session.Heads(d.cfg.Context(), d.heads))
				if found := std.Must(uc.FindResumable(subject, opts)); found.IsSome() {
					b.Fatalf("found %s", found.Unwrap().ID)
				}
			}

			// the mark of an open write, see application/ai/session/heads.go
			mark := func() {
				for _, id := range seed.Own[:marks] {
					if err := blob.Put(d.heads, "o/"+string(id), []byte(`{"createdBy":"`+string(seed.User)+`"}`)); err != nil {
						b.Fatal(err)
					}
				}
			}

			unmarked := func() {
				for _, id := range seed.Own[:marks] {
					if std.Must(d.heads.Exists(b.Context(), "o/"+string(id))) {
						b.Fatalf("the start has left the mark of %s", id)
					}
				}
			}

			mark()
			first := benchTimed(start)
			unmarked()

			durations := make([]time.Duration, 0, b.N)
			b.ResetTimer()
			for range b.N {
				b.StopTimer()
				mark()
				b.StartTimer()
				durations = append(durations, benchTimed(start))
				b.StopTimer()
				unmarked()
				b.StartTimer()
			}
			b.StopTimer()

			b.ReportMetric(benchMillis(first), "first-ms")
			b.ReportMetric(benchMillis(benchMedian(durations)), "median-ms")
		})
	}
}

// BenchmarkHeadsBuild is G5 of M27: the build of the heads of S2 and S50, which never had heads, on a copy of S2P and
// S50P, from the enabling of the ai management until the build has finished, and once with an abort after half of the
// heads, after which a fresh configurator finishes the build. The time does not include opening and closing the data.
// It reports the heap afterwards as built-B, to compare with enabled-B of BenchmarkHeadsMemory. Run each with
// -test.benchtime=1x in a process of its own, after BenchmarkHeadsStocks.
func BenchmarkHeadsBuild(b *testing.B) {
	for _, c := range []struct {
		name  string
		stock benchStock
	}{{"s2", stockS2P}, {"s50", stockS50P}} {
		for _, abort := range []bool{false, true} {
			name := c.name + "/whole"
			if abort {
				name = c.name + "/aborted"
			}

			b.Run(name, func(b *testing.B) {
				seed := benchPrepared(b, c.stock)
				pure := c.stock.dir(b)
				dir := filepath.Join(filepath.Dir(pure), "work-"+c.stock.name)
				benchCopyDir(b, pure, dir)
				b.Cleanup(func() { _ = os.RemoveAll(dir) })

				var took time.Duration
				var d *benchData
				run := func(abortAt int) {
					d = benchOpenData(b, dir)
					start := time.Now()
					std.Must(Enable(d.cfg))
					// polled cheaply: the layout and the number of entries of the store, which counts the heads and the layout
					for {
						done := benchHeadsDone(b, d.heads)
						if done || (abortAt > 0 && std.Must(blob.Count(b.Context(), d.heads)) > int64(abortAt)) {
							took += time.Since(start)
							if abortAt > 0 && done {
								b.Fatalf("the build has finished before the abort")
							}
							if _, n := benchHeadsState(b, d.heads); done && n != seed.Sessions {
								b.Fatalf("%d heads for %d conversations", n, seed.Sessions)
							}
							return
						}
						time.Sleep(5 * time.Millisecond)
					}
				}

				b.ResetTimer()
				if abort {
					run(seed.Sessions / 2)
					d.app.Destroy()
				}
				run(0)
				b.StopTimer()

				built := benchHeap()
				d.app.Destroy()

				b.ReportMetric(took.Seconds(), "build-s")
				b.ReportMetric(float64(built), "built-B")
			})
		}
	}
}

// benchCopyDir copies the directory src to dst, which it replaces.
func benchCopyDir(b *testing.B, src, dst string) {
	b.Helper()

	if err := os.RemoveAll(dst); err != nil {
		b.Fatal(err)
	}

	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0700)
		}

		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}

		if _, err := io.Copy(out, in); err != nil {
			_ = out.Close()
			return err
		}

		return out.Close()
	})
	if err != nil {
		b.Fatal(err)
	}
}

// BenchmarkHeadsSave is G6 of M27: saving a conversation of about 26 KB through Rename, with heads against without,
// in a fresh directory beside the stocks. It reports the median of b.N saves and the growth of the log per save over the first 100 saves,
// before the log is compacted. Run with -test.benchtime=1000x.
func BenchmarkHeadsSave(b *testing.B) {
	for _, withHeads := range []bool{false, true} {
		name := "without"
		if withHeads {
			name = "with"
		}

		b.Run(name, func(b *testing.B) {
			// on the volume of the stocks like the other measurements, if there is one
			dir := filepath.Join(filepath.Dir(stockS2.dir(b)), "work-save-"+name)
			if err := os.RemoveAll(dir); err != nil {
				b.Fatal(err)
			}
			b.Cleanup(func() { _ = os.RemoveAll(dir) })

			d := benchOpenData(b, dir)
			defer d.app.Destroy()
			d.wired()

			sess := benchSession(session.ID(benchIdent("save", 0)), user.ID(benchIdent("person", 0)), "bench", 1, xtime.Now())
			benchSave(b, d.repo, d.rdb, sess)

			var opts []session.Option
			if withHeads {
				opts = append(opts, session.Heads(d.cfg.Context(), d.heads))
			}
			uc := session.NewUseCases(d.repo, d.rdb, opts...)
			if withHeads {
				for deadline := time.Now().Add(time.Minute); ; {
					if benchHeadsDone(b, d.heads) {
						break
					}
					if time.Now().After(deadline) {
						b.Fatal("the heads are not built")
					}
					time.Sleep(time.Millisecond)
				}
			}

			wal := filepath.Join(dir, "tdb", "tdb.wal")
			logSize := func() int64 {
				return std.Must(os.Stat(wal)).Size()
			}

			// the first run of the benchmark with b.N = 1 measures the log of its single save
			logged := min(100, b.N)

			i := 0
			save := func() {
				i++
				if err := uc.Rename(user.SU(), sess.ID, fmt.Sprintf("Rechnungen prüfen %d", 10_000+i)); err != nil {
					b.Fatal(err)
				}
			}

			var growth int64
			durations := make([]time.Duration, 0, b.N)
			b.ResetTimer()
			for n := range b.N {
				before := logSize()
				durations = append(durations, benchTimed(save))
				if n < logged {
					growth += logSize() - before
				}
			}
			b.StopTimer()

			if got := std.Must(d.repo.FindByID(sess.ID)).Unwrap().Title; got != fmt.Sprintf("Rechnungen prüfen %d", 10_000+b.N) {
				b.Fatalf("the last save is %q", got)
			}
			if withHeads {
				if done, n := benchHeadsState(b, d.heads); !done || n != 1 {
					b.Fatalf("heads built %v: %d heads", done, n)
				}
			}

			b.ReportMetric(benchMillis(benchMedian(durations)), "median-ms")
			b.ReportMetric(float64(growth)/float64(logged), "log-B/save")
			b.ReportMetric(float64(len(std.Must(json.Marshal(sess)))), "session-B")
		})
	}
}

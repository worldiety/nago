// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"go.wdy.de/nago/pkg/ndb/tsdb"
)

// cutModel is what a column must hold: secured points are on disk, unsecured ones may be lost by a crash of the process.
type cutModel struct {
	secured, unsecured map[int64]string
}

func (m cutModel) clone() cutModel {
	return cutModel{secured: cloneMap(m.secured), unsecured: cloneMap(m.unsecured)}
}

func cloneMap(m map[int64]string) map[int64]string {
	c := make(map[int64]string, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}

// holds says whether got agrees with the model for ts: a secured point is there with its value, an unsecured one is there
// with its value or missing, any other time is missing.
func (m cutModel) holds(got map[int64]string, ts int64) bool {
	v, present := got[ts]
	if want, ok := m.secured[ts]; ok {
		return present && v == want
	}
	if want, ok := m.unsecured[ts]; ok {
		return !present || v == want
	}
	return !present
}

type cutOp struct {
	kind string // put, sync, overwrite (out of order, into the head), delete
	ts   int64
	v    string
}

func (o cutOp) String() string {
	if o.kind == "sync" {
		return "Sync"
	}
	return fmt.Sprintf("%s %d=%s", o.kind, o.ts, o.v)
}

func (o cutOp) apply(t testing.TB, s surfaceScheme, c *tsdb.Column, m *cutModel) {
	t.Helper()
	switch o.kind {
	case "put":
		must(t, s.put(c, o.ts, o.v))
		m.unsecured[o.ts] = o.v
	case "sync":
		must(t, c.Sync())
		for ts, v := range m.unsecured {
			m.secured[ts] = v
		}
		clear(m.unsecured)
	case "overwrite":
		must(t, s.put(c, o.ts, o.v))
		m.secured[o.ts] = o.v
	case "delete":
		must(t, c.Delete(o.ts))
		delete(m.secured, o.ts)
	default:
		t.Fatalf("op %q", o.kind)
	}
}

// cutSteps is the sequence after the setup of a row: Syncs of one and of several points, a Sync without a new point,
// then corrections and a deletion of secured points through the head, a new point in a gap between them, and Syncs
// after them. The first point has the time from; no block fills up, because a block has 64 points.
func cutSteps(s surfaceScheme, from int64) []cutOp {
	p := func(i int) cutOp { return cutOp{kind: "put", ts: from + int64(i)*10, v: value(s, i, 0)} }
	sync := cutOp{kind: "sync"}
	return []cutOp{
		p(0), sync,
		p(1), p(2), sync,
		sync,
		p(3), p(4), p(5), sync,
		{kind: "overwrite", ts: from, v: value(s, 0, 1)},
		{kind: "delete", ts: from + 10},
		{kind: "overwrite", ts: from + 25, v: value(s, 2, 2)}, // into the gap between point 2 and 3
		p(6), sync,
		p(7), // not secured at the end
	}
}

// cutAt are the lengths, at which a file is cut that a step let grow from old to new bytes: every one, or with -short
// all of the step but its last byte, and an empty file. A cut at old with old > 0 is the state before the step, which the
// step before checks in full.
func cutAt(old, new int) []int {
	var res []int
	for n := old; n < new; n++ {
		if !testing.Short() || n == new-1 || n == 0 {
			res = append(res, n)
		}
	}
	return res
}

// TestSyncSurvivesACutAnywhere: a crash of the process stops the writes after an arbitrary byte. The test records the
// files of the column after each step and opens every state that a crash in the middle of a step can leave: each file
// the step changed, cut at every length between its size before and after the step, with the other files of the
// step before and after it. After the open, every point secured before the step is there with its value, no point is
// there twice or with a value it never had, and a point of the step is there as before or as after it. The full state
// after a Sync has every point written before it. An open after a cut continues the column: a point and a Sync later,
// a second crash has both. In every row some cut state lacks a point of the state after its step, which shows that the
// cuts reach the data. With -short each file is cut before the last byte of a step only (cutAt), and the column is
// continued after that cut of each Sync and after the last Sync; the full run before a push cuts every byte and
// continues after each state.
//
// Writing that only appends is all of these steps: the setup is done before the first state is recorded, and no block
// fills up. A step that writes into the middle of a file or removes one fails the premise.
func TestSyncSurvivesACutAnywhere(t *testing.T) {
	opts := quietOptions(64)
	setups := []struct {
		name  string
		setup func(t *testing.T, s surfaceScheme, dir string) (*tsdb.Column, cutModel, int64)
	}{
		{name: "a new column", setup: func(t *testing.T, s surfaceScheme, dir string) (*tsdb.Column, cutModel, int64) {
			db := openDB(t, dir, opts)
			return column(t, db, "b", "c", s.schema), cutModel{secured: map[int64]string{}, unsecured: map[int64]string{}}, 0
		}},
		{name: "after a sealed block", setup: func(t *testing.T, s surfaceScheme, dir string) (*tsdb.Column, cutModel, int64) {
			db := openDB(t, dir, opts)
			c := column(t, db, "b", "c", s.schema)
			m := cutModel{secured: map[int64]string{}, unsecured: map[int64]string{}}
			for i := range opts.BlockPoints {
				cutOp{kind: "put", ts: int64(i) * 10, v: value(s, i, 0)}.apply(t, s, c, &m)
			}
			cutOp{kind: "sync"}.apply(t, s, c, &m)
			return c, m, int64(opts.BlockPoints) * 10
		}},
		{name: "continued after Close", setup: func(t *testing.T, s surfaceScheme, dir string) (*tsdb.Column, cutModel, int64) {
			db := openDB(t, dir, opts)
			c := column(t, db, "b", "c", s.schema)
			m := cutModel{secured: map[int64]string{}, unsecured: map[int64]string{}}
			for i := range 5 {
				cutOp{kind: "put", ts: int64(i) * 10, v: value(s, i, 0)}.apply(t, s, c, &m)
			}
			closeDB(t, db) // Close secures them
			for ts, v := range m.unsecured {
				m.secured[ts] = v
			}
			clear(m.unsecured)
			return lookup(t, openDB(t, dir, opts), "b", "c"), m, 50
		}},
	}
	for _, s := range surfaceSchemes() {
		for _, setup := range setups {
			t.Run(s.name+"/"+setup.name, func(t *testing.T) {
				h := n22cCapture(t)
				dir := t.TempDir()
				c, m, from := setup.setup(t, s, dir)
				colDir := filepath.Join(dir, "b", "c")
				steps := cutSteps(s, from)
				snaps := []snapshot{readSnapshot(t, colDir)}
				models := []cutModel{m.clone()}
				for _, op := range steps {
					op.apply(t, s, c, &m)
					snaps = append(snaps, readSnapshot(t, colDir))
					models = append(models, m.clone())
				}

				var states, losses atomic.Int64
				scratch := t.TempDir()
				open := func(state snapshot, k int, touched []int64, full, again bool, where string) (map[int64]string, error) {
					dir := filepath.Join(scratch, fmt.Sprint(states.Add(1)))
					return openCutState(s, opts, dir, state, models[k], models[k+1], touched, full, again, where)
				}
				// the steps are checked in parallel, the states of one step one after the other
				var jobs []func() error
				lastSync := 0
				for k, op := range steps {
					if op.kind == "sync" {
						lastSync = k
					}
				}
				for k, op := range steps {
					before, after := snaps[k], snaps[k+1]
					changed := changedFiles(t, fmt.Sprintf("step %d (%s)", k+1, op), before, after)
					if op.kind == "sync" && len(changed) > 1 {
						t.Errorf("step %d (%s) wrote %d files of the column %v, at most one", k+1, op, len(changed), changed)
					}
					var touched []int64
					if op.kind != "sync" {
						touched = []int64{op.ts}
					}
					jobs = append(jobs, func() error {
						gotAfter, err := open(after, k, touched, true, !testing.Short() || k == lastSync, fmt.Sprintf("after step %d (%s)", k+1, op))
						if err != nil {
							return err
						}
						// The order in which a step writes several files is not observable. Its cuts must hold for one order.
						var errs []error
						for _, order := range orders(changed) {
							err := func() error {
								for i, name := range order {
									for _, n := range cutAt(len(before[name]), len(after[name])) {
										state := cloneSnapshot(before)
										for _, done := range order[:i] {
											state[done] = after[done]
										}
										state[name] = after[name][:n]
										// with -short the column is continued after the last cut of a Sync only
										again := !testing.Short() || op.kind == "sync" && n == len(after[name])-1
										where := fmt.Sprintf("step %d (%s): %s cut at %d of %d bytes, order %v", k+1, op, name, n, len(after[name]), order)
										got, err := open(state, k, touched, false, again, where)
										if err != nil {
											return err
										}
										for ts := range gotAfter {
											if _, ok := got[ts]; !ok {
												losses.Add(1)
												break
											}
										}
									}
								}
								return nil
							}()
							if err == nil {
								return nil
							}
							errs = append(errs, err)
						}
						if len(errs) > 0 {
							return fmt.Errorf("no order of the writes of step %d (%s) keeps the column: %v", k+1, op, errs)
						}
						return nil
					})
				}
				if err := runAll(jobs); err != nil {
					t.Fatal(err)
				}
				// wokoda N23 (R-DEC-TS-NAGO-DAMAGE): what a crash leaves is no damage, no open of a cut state reports one
				if w := h.warnings("b", "c"); len(w) != 0 {
					t.Errorf("%d warnings over the cut states, a crash alone reports nothing: %v", len(w), w)
				}
				if losses.Load() == 0 {
					t.Errorf("premise: none of the %d states lacks a point of the state after its step, so the cuts reach no data", states.Load())
				}
				t.Logf("%d states opened, %d of them lack a point of the state after their step", states.Load(), losses.Load())
			})
		}
	}
}

// runAll runs the independent checks fns on a few goroutines and returns the first error in the order of fns.
func runAll(fns []func() error) error {
	errs := make([]error, len(fns))
	next := atomic.Int64{}
	var wg sync.WaitGroup
	for range max(1, min(4, runtime.GOMAXPROCS(0)/2)) {
		wg.Go(func() {
			for i := next.Add(1) - 1; i < int64(len(fns)); i = next.Add(1) - 1 {
				errs[i] = fns[i]()
			}
		})
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

// orders are the orders in which a step may have written the files it changed.
func orders(files []string) [][]string {
	if len(files) <= 1 {
		return [][]string{files}
	}
	var res [][]string
	for i, f := range files {
		rest := append(slices.Clone(files[:i]), files[i+1:]...)
		for _, o := range orders(rest) {
			res = append(res, append([]string{f}, o...))
		}
	}
	return res
}

// changedFiles returns the files a step changed. A file it removed or wrote into the middle of fails the test: the cuts
// model writing that only appends.
func changedFiles(t *testing.T, step string, before, after snapshot) []string {
	t.Helper()
	var changed []string
	for name, b := range after {
		old, ok := before[name]
		switch {
		case !ok:
			changed = append(changed, name)
		case bytes.Equal(old, b):
		case bytes.HasPrefix(b, old):
			changed = append(changed, name)
		default:
			t.Fatalf("premise: %s wrote into the middle of %s (%d to %d bytes): Sync, Put and the head only append", step, name, len(old), len(b))
		}
	}
	for name := range before {
		if _, ok := after[name]; !ok {
			t.Fatalf("premise: %s removed %s", step, name)
		}
	}
	slices.Sort(changed)
	return changed
}

func cloneSnapshot(s snapshot) snapshot {
	c := make(snapshot, len(s))
	for k, v := range s {
		c[k] = v
	}
	return c
}

// openCutState opens a crash state and checks it against the models before and after its step. A full state is the one
// after the step and must agree with the model after it. With again, the column then continues: a point after all
// others and a Sync survive a second crash next to everything the first open read. It returns what the first open read.
func openCutState(s surfaceScheme, opts tsdb.Options, dir string, state snapshot, before, after cutModel,
	touched []int64, full, again bool, where string) (map[int64]string, error) {
	defer func() { _ = os.RemoveAll(dir) }()
	colDir := filepath.Join(dir, "b", "c")
	if err := os.MkdirAll(colDir, 0755); err != nil {
		return nil, err
	}
	for name, b := range state {
		if err := os.WriteFile(filepath.Join(colDir, name), b, 0644); err != nil {
			return nil, err
		}
	}
	got, unsynced, err := openAndRead(s, opts, dir)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", where, err)
	}
	if unsynced != 0 {
		return nil, fmt.Errorf("%s: %d points only in memory right after the open", where, unsynced)
	}
	if err := checkColumnNames(colDir, s.schema.Scheme == tsdb.SchemeEnum); err != nil {
		return nil, fmt.Errorf("%s: %w", where, err)
	}

	times := map[int64]bool{}
	for _, m := range []map[int64]string{got, before.secured, before.unsecured, after.secured, after.unsecured} {
		for ts := range m {
			times[ts] = true
		}
	}
	for ts := range times {
		var ok bool
		switch {
		case full:
			ok = after.holds(got, ts)
		case slices.Contains(touched, ts):
			ok = before.holds(got, ts) || after.holds(got, ts)
		default:
			ok = before.holds(got, ts)
		}
		if !ok {
			v, present := got[ts]
			return nil, fmt.Errorf("%s: the time %d reads %q (present %v); secured before %q, after %q",
				where, ts, v, present, before.secured[ts], after.secured[ts])
		}
	}
	if !again {
		return got, nil
	}

	// the column goes on from what the open took up
	next := int64(1_000_000)
	v := value(s, 999, 0)
	if err := continueAndCrash(s, opts, dir, next, v); err != nil {
		return nil, fmt.Errorf("%s: %w", where, err)
	}
	got2, _, err := openAndRead(s, opts, dir+".again")
	_ = os.RemoveAll(dir + ".again")
	if err != nil {
		return nil, fmt.Errorf("%s: after the second crash: %w", where, err)
	}
	want := cloneMap(got)
	want[next] = v
	if !mapsEqual(got2, want) {
		return nil, fmt.Errorf("%s: after a point, a Sync and a second crash the column reads %v, want %v", where, describeMap(got2), describeMap(want))
	}
	return got, nil
}

// openAndRead opens the database in dir, reads every point of b/c and closes it again.
func openAndRead(s surfaceScheme, opts tsdb.Options, dir string) (map[int64]string, int64, error) {
	db, err := tsdb.Open(dir, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("open: %w", err)
	}
	defer func() { _ = db.Close() }()
	c, ok, err := db.LookupColumn("b", "c")
	if err != nil || !ok {
		return nil, 0, fmt.Errorf("lookup: ok=%v err=%v", ok, err)
	}
	got, err := scanMap(s, c)
	return got, c.Stats().Unsynced, err
}

// continueAndCrash opens dir, writes the point next=v and a Sync, and leaves the disk as a crash after it in dir.again.
func continueAndCrash(s surfaceScheme, opts tsdb.Options, dir string, next int64, v string) error {
	db, err := tsdb.Open(dir, opts)
	if err != nil {
		return fmt.Errorf("open to continue: %w", err)
	}
	defer func() { _ = db.Close() }()
	c, ok, err := db.LookupColumn("b", "c")
	if err != nil || !ok {
		return fmt.Errorf("lookup to continue: ok=%v err=%v", ok, err)
	}
	if err := s.put(c, next, v); err != nil {
		return err
	}
	if err := c.Sync(); err != nil {
		return err
	}
	return os.CopyFS(dir+".again", os.DirFS(dir))
}

// scanMap reads every point of the column and fails if a time comes more than once or out of order.
func scanMap(s surfaceScheme, c *tsdb.Column) (map[int64]string, error) {
	m := map[int64]string{}
	last, first := int64(0), true
	add := func(ts int64, v string) error {
		if !first && ts <= last {
			return fmt.Errorf("the time %d follows %d: a point read twice or out of order", ts, last)
		}
		first, last = false, ts
		m[ts] = v
		return nil
	}
	var err error
	if s.schema.Scheme == tsdb.SchemeDecimal {
		serr := c.ScanI64(minTS, maxTS, func(ts, vs []int64) bool {
			for i := range ts {
				if err = add(ts[i], strconv.FormatInt(vs[i], 10)); err != nil {
					return false
				}
			}
			return true
		})
		return m, errors.Join(err, serr)
	}
	serr := c.ScanString(minTS, maxTS, func(ts []int64, vs []string) bool {
		for i := range ts {
			if err = add(ts[i], vs[i]); err != nil {
				return false
			}
		}
		return true
	})
	return m, errors.Join(err, serr)
}

// checkColumnNames is columnFiles without a testing.TB.
func checkColumnNames(dir string, enum bool) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	open := 0
	for _, e := range entries {
		switch name := e.Name(); {
		case name == "schema.json" || name == "head.wal" || enum && name == "enum.dict":
		case chunkName.MatchString(name):
		case openName.MatchString(name):
			open++
		default:
			return fmt.Errorf("the column holds %q, which is no file of a column", name)
		}
	}
	if open > 1 {
		return fmt.Errorf("the column holds %d open chunks", open)
	}
	return nil
}

func mapsEqual(a, b map[int64]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if w, ok := b[k]; !ok || v != w {
			return false
		}
	}
	return true
}

func describeMap(m map[int64]string) string {
	var sb strings.Builder
	for _, k := range sortedKeys(m) {
		fmt.Fprintf(&sb, "%d=%s ", k, m[k])
	}
	return strings.TrimSpace(sb.String())
}

// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"maps"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"go.wdy.de/nago/pkg/ndb/tsdb"
)

// The test below runs its own binary as a child process, which writes into a database in a loop, and kills it with
// SIGKILL at random places. The child reports each step on stdout before and after it, so the parent knows which steps
// were done and which one ran when the kill came.
const (
	killChildEnv = "TSDB_KILL_CHILD" // the directory of the database; set in the child only
	killSeedEnv  = "TSDB_KILL_SEED"  // the seed of the child, or of the whole run in the parent
	killsEnv     = "TSDB_KILLS"      // the number of kills; 5 with -short, 60 otherwise; 1000 once before a push
)

// killOptions make sealing, splitting and compactions frequent: blocks of 8 points, chunks of 40, a compaction of the head
// after 256 B of corrections or a quarter of tombstones, and a synchronous one at 64 entries.
func killOptions() tsdb.Options {
	return tsdb.Options{BlockPoints: 8, Split: tsdb.SplitByCount(40), FlushBytes: 256, CompactTombstoneRatio: 0.25, MaxHeadPoints: 64}
}

var killColumns = []struct {
	name   string
	schema tsdb.Schema
}{
	{name: "d", schema: tsdb.Schema{Scheme: tsdb.SchemeDecimal}},
	{name: "e", schema: tsdb.Schema{Scheme: tsdb.SchemeEnum}},
}

// killStep is one step of the child, as it reports it: kind, the column ("*" for all), and its points or times.
type killStep struct {
	kind   string // append, sync, overwrite, delete, deleteRange, close, flush
	col    string
	points map[int64]string // append, overwrite
	times  []int64          // delete; deleteRange: lo and hi
	// inMemory is ColumnStats.Unsynced of each column after the step: the newest points written that are only in memory.
	// Every older one is on disk, in a sealed block or secured by a Sync.
	inMemory map[string]int64
}

func (s killStep) String() string {
	var sb strings.Builder
	sb.WriteString(s.kind + " " + s.col)
	for _, ts := range sortedKeys(s.points) {
		fmt.Fprintf(&sb, " %d=%s", ts, s.points[ts])
	}
	for _, ts := range s.times {
		fmt.Fprintf(&sb, " %d", ts)
	}
	return sb.String()
}

func parseKillStep(fields []string) (killStep, error) {
	if len(fields) < 2 {
		return killStep{}, fmt.Errorf("step %v", fields)
	}
	s := killStep{kind: fields[0], col: fields[1], points: map[int64]string{}}
	for _, f := range fields[2:] {
		tsStr, v, isPoint := strings.Cut(f, "=")
		ts, err := strconv.ParseInt(tsStr, 10, 64)
		if err != nil {
			return killStep{}, err
		}
		if isPoint {
			s.points[ts] = v
		} else {
			s.times = append(s.times, ts)
		}
	}
	return s, nil
}

// killModel holds the points of the columns: secured ones must survive a crash, unsecured ones may be lost.
type killModel map[string]*cutModel

func newKillModel() killModel {
	m := killModel{}
	for _, c := range killColumns {
		m[c.name] = &cutModel{secured: map[int64]string{}, unsecured: map[int64]string{}}
	}
	return m
}

func (m killModel) clone() killModel {
	c := killModel{}
	for k, v := range m {
		cm := v.clone()
		c[k] = &cm
	}
	return c
}

func (m killModel) secure(col string) {
	for name, cm := range m {
		if col == "*" || col == name {
			maps.Copy(cm.secured, cm.unsecured)
			clear(cm.unsecured)
		}
	}
}

// apply is the effect of a finished step. A correction or deletion is in the head's log on disk when it returns.
func (m killModel) apply(s killStep) {
	cm := m[s.col]
	switch s.kind {
	case "append":
		maps.Copy(cm.unsecured, s.points)
	case "sync", "flush", "close":
		m.secure(s.col)
	case "overwrite":
		for ts, v := range s.points {
			cm.secured[ts] = v
			delete(cm.unsecured, ts)
		}
	case "delete":
		for _, ts := range s.times {
			delete(cm.secured, ts)
			delete(cm.unsecured, ts)
		}
	case "deleteRange":
		for _, mm := range []map[int64]string{cm.secured, cm.unsecured} {
			for ts := range mm {
				if ts >= s.times[0] && ts <= s.times[1] {
					delete(mm, ts)
				}
			}
		}
	}
	// what is not in memory any more is on disk: all but the newest inMemory points
	for col, n := range s.inMemory {
		cm := m[col]
		keys := sortedKeys(cm.unsecured)
		for _, ts := range keys[:max(0, len(keys)-int(n))] {
			cm.secured[ts] = cm.unsecured[ts]
			delete(cm.unsecured, ts)
		}
	}
}

// touched are the times of a column whose state a step changes in a way a crash may stop half way.
func (s killStep) touched(col string, m killModel) []int64 {
	if s.col != col {
		return nil
	}
	switch s.kind {
	case "append", "overwrite":
		return slices.Collect(maps.Keys(s.points))
	case "delete":
		return s.times
	case "deleteRange":
		var res []int64
		for _, mm := range []map[int64]string{m[col].secured, m[col].unsecured} {
			for ts := range mm {
				if ts >= s.times[0] && ts <= s.times[1] {
					res = append(res, ts)
				}
			}
		}
		return res
	}
	return nil
}

// TestCrashAnywhereKeepsTheSecuredPoints: a crash of the process at any place of an append, a Sync, the sealing of a
// block, a correction or deletion, a Close or Flush with their compaction, or a compaction of the head in the background
// loses no point that a Sync, Close or Flush secured before, that was in a chunk before (ColumnStats.Unsynced after each
// step says which of the newest points are still only in memory), or whose correction or deletion had returned before,
// and reads no point twice or with a value it never had. A step that ran when the kill came may
// have happened or not, for each of its points. After each open the column holds no file besides those of a column
// today. Sealing, Close and compaction write into the middle of files and over several files; for them this is the
// reasoned exception to cutting files at every byte (see TestSyncSurvivesACutAnywhere): the child is killed at random
// places, 5 times after at most 30 steps each with -short, otherwise 60 times or as often as TSDB_KILLS says (1000 once
// before a push), and the log names the steps the kills hit.
//
// wokoda N23 (R-DEC-TS-NAGO-DAMAGE, -DICT, -WAL): after each open no warning of the columns comes, because a crash of
// the process alone reports nothing; and every second database begins with the images of testdata/8ae1daaee, a WAL in
// the format before N23 (column d, image wal) and a dictionary in the format before N23 (column e, image dict), so that
// a kill can hit their rewrite; whether one did, the test does not see. The first database is one of them, so that
// -short has one.
func TestCrashAnywhereKeepsTheSecuredPoints(t *testing.T) {
	if dir := os.Getenv(killChildEnv); dir != "" {
		seed, _ := strconv.ParseUint(os.Getenv(killSeedEnv), 10, 64)
		killChild(dir, seed)
		return
	}
	kills := 60
	if testing.Short() {
		kills = 5
	}
	if s := os.Getenv(killsEnv); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 {
			t.Fatalf("%s=%q", killsEnv, s)
		}
		kills = n
	}
	seed := uint64(time.Now().UnixNano())
	if s := os.Getenv(killSeedEnv); s != "" {
		seed, _ = strconv.ParseUint(s, 10, 64)
	}
	t.Logf("%d kills, %s=%d", kills, killSeedEnv, seed)
	rng := rand.New(rand.NewPCG(seed, 1))
	h := n22cCapture(t)
	perDB := 25
	if kills < 2*perDB {
		perDB = (kills + 1) / 2
	}

	hits := map[string]int{}
	var dir string
	var model killModel
	for round := range kills {
		if round%perDB == 0 { // a new database now and then keeps the steps short
			dir = filepath.Join(t.TempDir(), "db")
			model = newKillModel()
			if (round/perDB)%2 == 0 {
				killFromImages(t, dir, model)
			}
		}
		mark := len(h.warnings("b", "d")) + len(h.warnings("b", "e"))
		steps, running := runAndKill(t, dir, seed+uint64(round)+1, rng)
		killLog = steps
		for _, s := range steps {
			model.apply(s)
		}
		hit := "between two steps"
		after := model
		if running != nil {
			hit = running.kind
			after = model.clone()
			after.apply(*running)
		}
		hits[hit]++
		model = checkAfterKill(t, dir, round, model, after, running)
		if w := append(h.warnings("b", "d"), h.warnings("b", "e")...); len(w) != mark {
			t.Fatalf("round %d, killed in %v: %d warnings after the open, a crash alone reports nothing: %v", round, running, len(w)-mark, w)
		}
	}
	keys := slices.Sorted(maps.Keys(hits))
	var sb strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&sb, " %s %d,", k, hits[k])
	}
	t.Logf("%d kills hit:%s", kills, strings.TrimSuffix(sb.String(), ","))
}

// runAndKill starts the child on dir and kills it after a random number of finished steps and a random pause. It
// returns the finished steps and the step that ran when the kill came, if any.
func runAndKill(t *testing.T, dir string, seed uint64, rng *rand.Rand) ([]killStep, *killStep) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestCrashAnywhereKeepsTheSecuredPoints$", "-test.count=1")
	cmd.Env = append(os.Environ(), killChildEnv+"="+dir, killSeedEnv+"="+strconv.FormatUint(seed, 10))
	stdout, err := cmd.StdoutPipe()
	must(t, err)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	must(t, cmd.Start())

	lines := make(chan string, 1024)
	go func() {
		sc := bufio.NewScanner(stdout)
		for sc.Scan() {
			lines <- sc.Text()
		}
		close(lines)
	}()
	killAfter := rng.IntN(120) // finished steps; 0 is right after the open
	if testing.Short() {
		killAfter = rng.IntN(30)
	}
	pause := time.Duration(rng.IntN(2000)) * time.Microsecond
	var log []string
	killed := false
	done := 0
	for line := range lines {
		log = append(log, line)
		if strings.HasPrefix(line, "N22 E ") {
			t.Fatalf("the child failed: %s\n%s", line, stderr.String())
		}
		if line == "N22 R" || strings.HasPrefix(line, "N22 D ") {
			if line != "N22 R" {
				done++
			}
			if !killed && done >= killAfter {
				time.Sleep(pause) // where the kill hits; nothing waits for its effect
				must(t, cmd.Process.Signal(syscall.SIGKILL))
				killed = true
			}
		}
	}
	err = cmd.Wait()
	if !killed {
		t.Fatalf("the child ended by itself: %v\n%s", err, stderr.String())
	}

	var steps []killStep
	var running *killStep
	begun := map[int]killStep{}
	for _, line := range log {
		f := strings.Fields(line)
		if len(f) < 3 || f[0] != "N22" {
			continue
		}
		k, err := strconv.Atoi(f[2])
		if err != nil {
			t.Fatalf("child line %q", line)
		}
		switch f[1] {
		case "B":
			s, err := parseKillStep(f[3:])
			if err != nil {
				t.Fatalf("child line %q: %v", line, err)
			}
			begun[k] = s
			running = &s
		case "D":
			s := begun[k]
			s.inMemory = map[string]int64{}
			for _, f := range f[3:] {
				col, n, _ := strings.Cut(f, "=")
				v, err := strconv.ParseInt(n, 10, 64)
				if err != nil {
					t.Fatalf("child line %q", line)
				}
				s.inMemory[col] = v
			}
			steps = append(steps, s)
			running = nil
		}
	}
	return steps, running
}

// killFromImages puts the images wal (as column d) and dict (as column e) of testdata/8ae1daaee into the database at dir
// and their points, as 8ae1daaee read them (expect.json), as secured into the model.
func killFromImages(t *testing.T, dir string, model killModel) {
	t.Helper()
	reads := n23FixtureReads(t)
	for col, img := range map[string]string{"d": "wal", "e": "dict"} {
		must(t, os.CopyFS(filepath.Join(dir, "b", col), os.DirFS(filepath.Join(n23FixtureRoot, "db", "n23", img))))
		for _, p := range reads[img] {
			tsStr, v, _ := strings.Cut(p, "=")
			ts, err := strconv.ParseInt(tsStr, 10, 64)
			must(t, err)
			if col == "e" {
				v = n23Value(v)
			}
			model[col].secured[ts] = v
		}
	}
}

// killLog are the finished steps of the last child, for the message of a failure.
var killLog []killStep

// describeKillLog lists the steps of the last child that concern col and the time ts, and every Sync, Close and Flush.
func describeKillLog(col string, ts int64) string {
	var sb strings.Builder
	sb.WriteString("the finished steps of the child that concern it:\n")
	for i, s := range killLog {
		_, isPoint := s.points[ts]
		if s.col == col && (isPoint || slices.Contains(s.times, ts) || s.kind == "deleteRange") ||
			s.kind == "sync" || s.kind == "close" || s.kind == "flush" {
			fmt.Fprintf(&sb, "  %d: %s\n", i, s)
		}
	}
	return sb.String()
}

// checkAfterKill opens the database after a kill and compares each column with the model: the finished steps before
// the kill give the model, the step that ran gives after. A time a running step touched may read as before or after it.
// It closes the database again and returns what the columns hold now, all of it secured.
func checkAfterKill(t *testing.T, dir string, round int, model, after killModel, running *killStep) killModel {
	t.Helper()
	db, err := tsdb.Open(dir, killOptions())
	if err != nil {
		t.Fatalf("round %d: open after the kill: %v", round, err)
	}
	next := newKillModel()
	for _, kc := range killColumns {
		s := decimalScheme()
		if kc.schema.Scheme == tsdb.SchemeEnum {
			s = enumScheme()
		}
		c, ok, err := db.LookupColumn("b", kc.name)
		if err != nil {
			t.Fatalf("round %d: lookup %s: %v", round, kc.name, err)
		}
		got := map[int64]string{}
		if ok {
			got, err = scanMap(s, c)
			if err != nil {
				t.Fatalf("round %d, column %s: %v", round, kc.name, err)
			}
			if u := c.Stats().Unsynced; u != 0 {
				t.Errorf("round %d, column %s: %d points only in memory right after the open", round, kc.name, u)
			}
			if err := checkColumnNames(filepath.Join(dir, "b", kc.name), kc.schema.Scheme == tsdb.SchemeEnum); err != nil {
				t.Errorf("round %d, column %s after the open: %v", round, kc.name, err)
			}
		}
		var touched []int64
		if running != nil {
			touched = running.touched(kc.name, model)
		}
		m, a := model[kc.name], after[kc.name]
		times := map[int64]bool{}
		for _, mm := range []map[int64]string{got, m.secured, m.unsecured, a.secured, a.unsecured} {
			for ts := range mm {
				times[ts] = true
			}
		}
		for ts := range times {
			ok := m.holds(got, ts) || slices.Contains(touched, ts) && a.holds(got, ts)
			if !ok {
				v, present := got[ts]
				t.Fatalf("round %d, column %s, killed in %v: the time %d reads %q (present %v); the model has it secured as %q, unsecured as %q\n%s",
					round, kc.name, running, ts, v, present, m.secured[ts], m.unsecured[ts], describeKillLog(kc.name, ts))
			}
		}
		next[kc.name].secured = got
	}
	if err := db.Close(); err != nil {
		t.Fatalf("round %d: close: %v", round, err)
	}
	return next
}

// killChild is the child: it writes steps into the columns until it is killed, and reports each before and after it.
func killChild(dir string, seed uint64) {
	report := func(format string, args ...any) {
		_, _ = io.WriteString(os.Stdout, fmt.Sprintf("N22 "+format+"\n", args...))
	}
	fail := func(err error) {
		report("E %v", strings.ReplaceAll(err.Error(), "\n", " "))
		os.Exit(1)
	}
	rng := rand.New(rand.NewPCG(seed, 2))
	opts := killOptions()
	open := func() (*tsdb.DB, map[string]*tsdb.Column) {
		db, err := tsdb.Open(dir, opts)
		if err != nil {
			fail(err)
		}
		cols := map[string]*tsdb.Column{}
		for _, kc := range killColumns {
			c, err := db.Column("b", kc.name, kc.schema)
			if err != nil {
				fail(err)
			}
			cols[kc.name] = c
		}
		return db, cols
	}
	db, cols := open()

	// what is on disk now is secured; the child picks the targets of corrections among the secured points
	secured := map[string]map[int64]bool{}
	unsecured := map[string]map[int64]bool{}
	next := map[string]int64{}
	for _, kc := range killColumns {
		secured[kc.name], unsecured[kc.name] = map[int64]bool{}, map[int64]bool{}
		s := decimalScheme()
		if kc.schema.Scheme == tsdb.SchemeEnum {
			s = enumScheme()
		}
		got, err := scanMap(s, cols[kc.name])
		if err != nil {
			fail(err)
		}
		for ts := range got {
			secured[kc.name][ts] = true
			next[kc.name] = max(next[kc.name], ts/10*10+10)
		}
	}
	secure := func(col string) {
		for name := range unsecured {
			if col == "*" || col == name {
				maps.Copy(secured[name], unsecured[name])
				clear(unsecured[name])
			}
		}
	}
	val := func(col string, gen int) string {
		if col == "e" {
			return fmt.Sprintf("v%d", rng.IntN(5)+gen*10)
		}
		return strconv.Itoa(rng.IntN(100000) + gen*1_000_000)
	}
	pick := func(col string) (int64, bool) {
		if len(secured[col]) == 0 {
			return 0, false
		}
		keys := slices.Sorted(maps.Keys(secured[col]))
		return keys[rng.IntN(len(keys))], true
	}
	put := func(col string, ts int64, v string) error {
		if col == "e" {
			return cols[col].PutString(ts, v)
		}
		n, _ := strconv.ParseInt(v, 10, 64)
		return cols[col].PutI64(ts, n)
	}

	report("R")
	for k := 0; k < 100_000; k++ {
		col := killColumns[rng.IntN(len(killColumns))].name
		s := killStep{col: col, points: map[int64]string{}}
		r := rng.IntN(100)
		switch {
		case r < 35:
			s.kind = "append"
			for range 1 + rng.IntN(12) {
				s.points[next[col]] = val(col, 0)
				next[col] += 10
			}
		case r < 50:
			s.kind = "sync"
			if rng.IntN(4) == 0 {
				s.col = "*"
			}
		case r < 68:
			s.kind = "overwrite"
			for range 1 + rng.IntN(3) {
				if ts, ok := pick(col); ok {
					if rng.IntN(3) == 0 && ts+5 < next[col]-10 {
						ts += 5 // a new point in a gap, before the newest point, so that it goes into the head
					}
					s.points[ts] = val(col, 1+rng.IntN(3))
				}
			}
		case r < 78:
			s.kind = "delete"
			for range 1 + rng.IntN(2) {
				if ts, ok := pick(col); ok {
					s.times = append(s.times, ts)
				}
			}
		case r < 80:
			s.kind = "deleteRange"
			if ts, ok := pick(col); ok {
				s.times = []int64{ts, ts + 25}
			}
		case r < 92:
			s.kind, s.col = "close", "*"
		default:
			s.kind = "flush"
		}
		if len(s.points) == 0 && (s.kind == "append" || s.kind == "overwrite") ||
			len(s.times) == 0 && (s.kind == "delete" || s.kind == "deleteRange") {
			continue
		}
		report("B %d %s", k, s)
		var err error
		switch s.kind {
		case "append", "overwrite":
			for _, ts := range sortedKeys(s.points) {
				if err = put(col, ts, s.points[ts]); err != nil {
					break
				}
			}
		case "sync":
			for name, c := range cols {
				if s.col == "*" || s.col == name {
					if err = c.Sync(); err != nil {
						break
					}
				}
			}
		case "delete":
			for _, ts := range s.times {
				if err = cols[col].Delete(ts); err != nil {
					break
				}
			}
		case "deleteRange":
			err = cols[col].DeleteRange(s.times[0], s.times[1])
		case "close":
			if err = db.Close(); err == nil {
				db, cols = open()
			}
		case "flush":
			err = cols[col].Flush()
		}
		if err != nil {
			fail(fmt.Errorf("step %d %s: %w", k, s, err))
		}
		switch s.kind {
		case "append":
			for ts := range s.points {
				unsecured[col][ts] = true
			}
		case "sync", "close", "flush":
			secure(s.col)
		case "overwrite":
			for ts := range s.points {
				secured[col][ts] = true
				delete(unsecured[col], ts)
			}
		case "delete":
			for _, ts := range s.times {
				delete(secured[col], ts)
				delete(unsecured[col], ts)
			}
		case "deleteRange":
			for _, mm := range []map[int64]bool{secured[col], unsecured[col]} {
				for ts := range mm {
					if ts >= s.times[0] && ts <= s.times[1] {
						delete(mm, ts)
					}
				}
			}
		}
		report("D %d d=%d e=%d", k, cols["d"].Stats().Unsynced, cols["e"].Stats().Unsynced)
	}
	os.Exit(0)
}

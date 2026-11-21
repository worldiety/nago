// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uischeduler_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/scheduler"
	uischeduler "go.wdy.de/nago/application/scheduler/ui"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

// The findings of the surface check of N29 in the dashboard, each beside its counter row. Times come from a test
// clock of the manager; the page itself redraws by the real clock, which only changes relative texts like "vor 3 s".

// testClock is a scheduler.Clock whose time only moves by set; its timers fire, when set reaches them.
type testClock struct {
	mu      sync.Mutex
	now     time.Time
	pending []*testTimer
}

type testTimer struct {
	c        *testClock
	ch       chan time.Time
	deadline time.Time
	done     bool
}

func (c *testClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *testClock) NewTimer(d time.Duration) scheduler.Timer {
	c.mu.Lock()
	defer c.mu.Unlock()
	tm := &testTimer{c: c, ch: make(chan time.Time, 1), deadline: c.now.Add(d)}
	if d <= 0 {
		tm.done = true
		tm.ch <- c.now
		return tm
	}

	c.pending = append(c.pending, tm)
	return tm
}

func (tm *testTimer) C() <-chan time.Time { return tm.ch }

func (tm *testTimer) Stop() bool {
	tm.c.mu.Lock()
	defer tm.c.mu.Unlock()
	if tm.done {
		return false
	}

	tm.done = true
	tm.c.pending = slices.DeleteFunc(tm.c.pending, func(p *testTimer) bool { return p == tm })
	return true
}

func (c *testClock) set(to time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = to
	var keep []*testTimer
	for _, tm := range c.pending {
		if tm.deadline.After(to) {
			keep = append(keep, tm)
			continue
		}

		tm.done = true
		tm.ch <- tm.deadline
	}

	c.pending = keep
}

// gate is a runner, which logs its messages and then waits until the test opens it or its context ends.
type gate struct {
	open    chan struct{}
	started chan struct{}
	logs    func(l *slog.Logger)
	result  error
}

func newGate(logs func(l *slog.Logger), result error) *gate {
	return &gate{open: make(chan struct{}), started: make(chan struct{}, 10), logs: logs, result: result}
}

func (g *gate) run(ctx context.Context) error {
	if g.logs != nil {
		g.logs(scheduler.LoggerFrom(ctx))
	}

	g.started <- struct{}{}
	select {
	case <-g.open:
		return g.result
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (g *gate) waitStarted(t *testing.T) {
	t.Helper()
	select {
	case <-g.started:
	case <-time.After(waitLimit):
		t.Fatal("der Runner wurde nicht aufgerufen")
	}
}

// stores are the stores of a manager, which survive its end like the data directory of a process.
type stores struct {
	settings, runs *mem.BlobStore
	logDir         string
}

func newStores(t *testing.T) stores {
	return stores{settings: mem.NewBlobStore("n29.settings"), runs: mem.NewBlobStore("n29.runs"), logDir: t.TempDir()}
}

func copyMem(t *testing.T, src *mem.BlobStore) *mem.BlobStore {
	dst := mem.NewBlobStore(src.Name())
	for key, err := range src.List(context.Background(), blob.ListOptions{}) {
		if err != nil {
			t.Fatal(err)
		}

		if buf, ok := src.Load(key); ok {
			dst.Store(key, buf)
		}
	}

	return dst
}

// crash returns copies of the stores as they are now, like after a crash of the process.
func (s stores) crash(t *testing.T) stores {
	logDir := t.TempDir()
	if err := os.CopyFS(logDir, os.DirFS(s.logDir)); err != nil {
		t.Fatal(err)
	}

	return stores{settings: copyMem(t, s.settings), runs: copyMem(t, s.runs), logDir: logDir}
}

func (s stores) manager(t *testing.T, clock scheduler.Clock) *scheduler.Manager {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return scheduler.NewManagerWith(ctx, scheduler.ManagerOptions{
		Settings: json.NewSloppyJSONRepository[scheduler.Settings, scheduler.ID](s.settings),
		Runs:     json.NewSloppyJSONRepository[scheduler.Run, scheduler.RunID](s.runs),
		LogDir:   s.logDir,
		Clock:    clock,
	})
}

// modelOf wires a dashboard to a manager like an embedder, which checks no rights.
func modelOf(m *scheduler.Manager, id scheduler.ID, loc *time.Location) uischeduler.DashboardModel {
	return uischeduler.DashboardModel{
		Status: func() (scheduler.StatusResult, error) { return scheduler.NewStatus(m)(user.SU(), id) },
		Runs:   func(offset, limit int) ([]scheduler.Run, int, error) { return m.RunsPage(id, offset, limit) },
		RunLog: func(run scheduler.RunID, q scheduler.LogQuery) (scheduler.LogPage, error) {
			return m.RunLog(id, run, q)
		},
		Location: loc,
		Config:   ui.Text("n29-config-marker"),
	}
}

// useCasesOf builds the use cases of the admin center on a manager.
func useCasesOf(m *scheduler.Manager, settings scheduler.SettingsRepository) scheduler.UseCases {
	return scheduler.UseCases{
		Configure:          scheduler.NewConfigure(m),
		Reconfigure:        scheduler.NewReconfigure(m),
		Remove:             scheduler.NewRemove(m),
		ViewLogs:           scheduler.NewViewLogs(m),
		Status:             scheduler.NewStatus(m),
		ExecuteNow:         scheduler.NewExecuteNow(m),
		ListSchedulers:     scheduler.NewListSchedulers(m),
		Stop:               scheduler.NewStop(m),
		Start:              scheduler.NewStart(m),
		FindSettingsByID:   scheduler.NewFindSettingsByID(settings),
		UpdateSettings:     scheduler.NewUpdateSettings(m, settings),
		DeleteSettingsByID: scheduler.NewDeleteSettingsByID(m, settings),
		ListRuns:           scheduler.NewListRuns(m),
		ListRunsPage:       scheduler.NewListRunsPage(m),
		ViewRunLog:         scheduler.NewViewRunLog(m),
	}
}

// pages serves dashboards, which a test registers by name before it opens them.
type pages struct {
	mu     sync.Mutex
	models map[string]uischeduler.DashboardModel
	url    string
}

func servePages(t *testing.T) *pages {
	p := &pages{models: map[string]uischeduler.DashboardModel{}}
	p.url = nagotest.Serve(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagotest.n29")
		cfg.RootView("n29-dashboard", func(wnd core.Window) core.View {
			p.mu.Lock()
			model, ok := p.models[wnd.Values()["case"]]
			p.mu.Unlock()
			if !ok {
				return ui.Text("n29: unbekannter Fall")
			}

			return uischeduler.Dashboard(wnd, model)
		})
	})

	return p
}

func (p *pages) open(t *testing.T, name string, model uischeduler.DashboardModel, opts ...nagotest.OpenOption) *nagotest.Window {
	p.mu.Lock()
	p.models[name] = model
	p.mu.Unlock()
	return nagotest.Dial(t, p.url, "n29-dashboard", append([]nagotest.OpenOption{nagotest.Values(core.Values{"case": name})}, opts...)...)
}

// texts returns the texts of all text views within s.
func texts(s nagotest.Selection) []string {
	all := s.FindAll(nagotest.Type[*proto.TextView]())
	res := make([]string, 0, all.Len())
	for i := range all.Len() {
		res = append(res, all.At(i).Text())
	}

	return res
}

// card returns the texts of the outermost component which contains the title, but none of the other titles.
func card(w *nagotest.Window, title string, others ...string) (string, bool) {
	cands := w.FindAll(nagotest.Containing(nagotest.Text(title)))
	for i := range cands.Len() {
		c := cands.At(i)
		foreign := false
		for _, o := range others {
			if c.FindAll(nagotest.Text(o)).Len() > 0 {
				foreign = true
			}
		}

		if !foreign {
			return strings.Join(texts(c), " | "), true
		}
	}

	return "", false
}

var kpiTitles = []string{"Nächster Lauf", "Zuletzt gestartet", "Zuletzt beendet", "Letzte Läufe", "Protokoll", "Konfiguration"}

// lastEnded returns the texts of the statistic "Zuletzt beendet".
func lastEnded(t *testing.T, w *nagotest.Window) string {
	t.Helper()
	var others []string
	for _, k := range kpiTitles {
		if k != "Zuletzt beendet" {
			others = append(others, k)
		}
	}

	s, ok := card(w, "Zuletzt beendet", others...)
	if !ok {
		t.Fatal("die Kennzahl „Zuletzt beendet“ fehlt")
	}

	return s
}

func wideTables(w *nagotest.Window) nagotest.Selection {
	return w.FindAll(nagotest.Where("Tabelle mit mindestens drei Spalten", func(n nagotest.Node) bool {
		tb, ok := n.Component.(*proto.Table)
		return ok && len(tb.Header.Columns) >= 3
	}))
}

// cellTexts returns the texts within the rows of a table, without the texts of its header.
func cellTexts(s nagotest.Selection) []string {
	tb := s.Node().Component.(*proto.Table)
	var header []string
	for _, c := range tb.Header.Columns {
		if tv, ok := c.Content.(*proto.TextView); ok {
			header = append(header, string(tv.Value))
		}
	}

	var res []string
	for _, txt := range texts(s) {
		if !slices.Contains(header, txt) && strings.TrimSpace(txt) != "" {
			res = append(res, txt)
		}
	}

	return res
}

func berlinZone(t *testing.T) *time.Location {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}

	return loc
}

// fixtureRuns returns a dashboard model on a manager with one ended run with 37 entries, 2 warnings and 1 error of
// 3.5 s and a log with an info, a warning and an error.
func fixtureRuns(t *testing.T, loc *time.Location, warnings, errs int) uischeduler.DashboardModel {
	clock := &testClock{now: time.Date(2026, 6, 1, 20, 17, 0, 0, time.UTC)}
	st := newStores(t)
	m := st.manager(t, clock)
	g := newGate(func(l *slog.Logger) {
		for i := range 37 - warnings - errs {
			l.Info(fmt.Sprintf("n29-info-%d", i))
		}

		for i := range warnings {
			l.Warn(fmt.Sprintf("n29-warn-%d", i))
		}

		for i := range errs {
			l.Error(fmt.Sprintf("n29-error-%d", i))
		}
	}, nil)

	const id scheduler.ID = "n29.job"
	if err := m.Configure(scheduler.Options{ID: id, Name: "n29-name", Kind: scheduler.Manual, Runner: g.run, Location: loc, Durable: true, Quiet: true}); err != nil {
		t.Fatal(err)
	}

	if err := m.StartNow(id, "Anna"); err != nil {
		t.Fatal(err)
	}

	g.waitStarted(t)
	clock.set(clock.Now().Add(3500 * time.Millisecond))
	close(g.open)
	waitEnded(t, m, id, 1)
	return modelOf(m, id, loc)
}

func waitEnded(t *testing.T, m *scheduler.Manager, id scheduler.ID, n int) []scheduler.Run {
	t.Helper()
	deadline := time.Now().Add(waitLimit)
	for {
		runs, _, err := m.RunsPage(id, 0, 100)
		if err != nil {
			t.Fatal(err)
		}

		ended := 0
		for _, r := range runs {
			if r.Outcome != scheduler.OutcomeRunning && !r.CompletedAt.IsZero() {
				ended++
			}
		}

		if ended == n {
			return runs
		}

		if time.Now().After(deadline) {
			t.Fatalf("%d beendete Läufe erwartet, vorhanden %+v", n, runs)
		}

		time.Sleep(time.Millisecond)
	}
}

// TestN29DashboardNarrow checks finding (1): below 768 px the latest runs and the log are no tables with three or more
// columns, but every value of their rows is still shown; from 768 px on they are tables as before.
//
// Zusagen: R-SCHEDULE-PAGE, R-DEC-SCHEDULE-NAGO
func TestN29DashboardNarrow(t *testing.T) {
	p := servePages(t)
	model := fixtureRuns(t, berlinZone(t), 2, 1)

	wide := p.open(t, "narrow", model, nagotest.Size(768, 900))
	wide.WaitFor(nagotest.Text("n29-error-0"), waitLimit)
	tables := wideTables(wide)
	if tables.Len() < 2 {
		t.Fatalf("Gegenzeile 768 px: Tabellen für „Letzte Läufe“ und „Protokoll“ erwartet, gefunden %d", tables.Len())
	}

	var values []string
	for i := range tables.Len() {
		values = append(values, cellTexts(tables.At(i))...)
	}

	if !slices.Contains(values, "n29-error-0") {
		t.Fatalf("Prämisse: die Tabellen enthalten das Protokoll, gefunden %q", values)
	}

	for _, width := range []int{767, 375} {
		t.Run(fmt.Sprintf("%d px", width), func(t *testing.T) {
			w := p.open(t, "narrow", model, nagotest.Size(width, 900))
			w.WaitFor(nagotest.Text("n29-error-0"), waitLimit)
			if n := wideTables(w).Len(); n != 0 {
				t.Errorf("unter 768 px keine Tabelle mit drei oder mehr Spalten erwartet, gefunden %d", n)
			}

			// a cell may stand as several named values, e.g. "Nr. 1 · von Hand: Anna" as "Nr." "1" and "von Hand: Anna"
			shown := textsOfTree(w)
			present := func(part string) bool {
				for i, s := range shown {
					if strings.Contains(s, part) || (i+1 < len(shown) && s+" "+shown[i+1] == part) {
						return true
					}
				}
				return false
			}

			for _, v := range values {
				for _, part := range strings.Split(v, " · ") {
					if !present(part) {
						t.Errorf("Wert %q der Tabellen (Teil %q) fehlt unter 768 px", v, part)
					}
				}
			}
		})
	}
}

// TestN29DashboardEntries checks findings (4) and (5): the amount of entries names warnings also beside errors, and
// durations and percentages have a decimal comma.
//
// Zusagen: R-SCHEDULE-PAGE, R-DEC-SCHEDULE-NAGO
func TestN29DashboardEntries(t *testing.T) {
	p := servePages(t)
	loc := berlinZone(t)
	errorsRe, warningsRe := regexp.MustCompile(`[1-9][0-9]* Fehler`), regexp.MustCompile(`[1-9][0-9]* Warnung`)

	tests := []struct {
		name                string
		warnings, errs      int
		wantWarn, wantError bool
	}{
		{"Warnungen neben Fehlern", 2, 1, true, true},
		{"Gegenzeile nur Warnungen", 2, 0, true, false},
		{"Gegenzeile nur Fehler", 0, 1, false, true},
	}

	for _, tt := range tests {
		for _, width := range []int{1280, 767} {
			t.Run(fmt.Sprintf("%s, %d px", tt.name, width), func(t *testing.T) {
				w := p.open(t, tt.name, fixtureRuns(t, loc, tt.warnings, tt.errs), nagotest.Size(width, 900))
				w.WaitFor(nagotest.Text("n29-info-0"), waitLimit)
				all := strings.Join(textsOfTree(w), " | ")

				if got := warningsRe.MatchString(all); got != tt.wantWarn {
					t.Errorf("Warnungen genannt %v, erwartet %v: %s", got, tt.wantWarn, all)
				}

				if got := errorsRe.MatchString(all); got != tt.wantError {
					t.Errorf("Fehler genannt %v, erwartet %v: %s", got, tt.wantError, all)
				}

				// (5) the run took 3.5 s
				if !strings.Contains(all, "3,5") || strings.Contains(all, "3.5") {
					t.Errorf("Dauer mit Dezimalkomma erwartet: %s", all)
				}
			})
		}
	}
}

// textsOfTree returns all texts of the window.
func textsOfTree(w *nagotest.Window) []string {
	all := w.FindAll(nagotest.Type[*proto.TextView]())
	res := make([]string, 0, all.Len())
	for i := range all.Len() {
		res = append(res, all.At(i).Text())
	}

	return res
}

// TestN29DashboardPercent checks finding (5) for the percentage of the statistics: 2 of 3 runs in 24 hours
// succeeded.
//
// Zusagen: R-SCHEDULE-PAGE, R-DEC-SCHEDULE-NAGO
func TestN29DashboardPercent(t *testing.T) {
	p := servePages(t)
	clock := &testClock{now: time.Date(2026, 6, 1, 20, 17, 0, 0, time.UTC)}
	st := newStores(t)
	m := st.manager(t, clock)
	const id scheduler.ID = "n29.job"
	var fail atomic.Bool
	runner := func(context.Context) error {
		if fail.Load() {
			return fmt.Errorf("n29 gescheitert")
		}
		return nil
	}

	if err := m.Configure(scheduler.Options{ID: id, Name: "n29-name", Kind: scheduler.Manual, Runner: runner, Location: berlinZone(t), Durable: true, Quiet: true}); err != nil {
		t.Fatal(err)
	}

	for i, f := range []bool{false, true, false} {
		fail.Store(f)
		clock.set(clock.Now().Add(time.Minute))
		if err := m.StartNow(id, "Anna"); err != nil {
			t.Fatal(err)
		}

		waitEnded(t, m, id, i+1)
	}

	w := p.open(t, "percent", modelOf(m, id, berlinZone(t)))
	w.WaitFor(nagotest.TextContains("66"), waitLimit)
	all := strings.Join(textsOfTree(w), " | ")
	if !strings.Contains(all, "66,7") || strings.Contains(all, "66.7") {
		t.Fatalf("Prozent mit Dezimalkomma erwartet: %s", all)
	}
}

// TestN29DashboardLastEnded checks findings (2) and (3): "Zuletzt beendet" takes time, duration and outcome from the
// latest ended run, while a run is running and after a restart, which closed an open run as "Neustart".
//
// Zusagen: R-SCHEDULE-PAGE, R-SCHEDULE-RESTART, R-DEC-SCHEDULE-NAGO
func TestN29DashboardLastEnded(t *testing.T) {
	p := servePages(t)
	loc := berlinZone(t)
	const id scheduler.ID = "n29.job"

	// run A failed after 2.5 s at 20:41 in Berlin, run B began at 22:17 in Berlin and is held
	begin := func(t *testing.T) (stores, *testClock, *scheduler.Manager, *gate) {
		clock := &testClock{now: time.Date(2026, 6, 1, 18, 41, 0, 0, time.UTC)}
		st := newStores(t)
		m := st.manager(t, clock)
		var current atomic.Pointer[gate]
		runner := func(ctx context.Context) error { return current.Load().run(ctx) }
		if err := m.Configure(scheduler.Options{ID: id, Name: "n29-name", Kind: scheduler.Cron, Defaults: scheduler.Settings{ID: id, CronHour: 3}, Runner: runner, Location: loc, Durable: true, Quiet: true}); err != nil {
			t.Fatal(err)
		}

		a := newGate(nil, fmt.Errorf("n29 gescheitert"))
		current.Store(a)
		if err := m.StartNow(id, "Anna"); err != nil {
			t.Fatal(err)
		}

		a.waitStarted(t)
		clock.set(time.Date(2026, 6, 1, 18, 41, 2, 500_000_000, time.UTC))
		close(a.open)
		waitEnded(t, m, id, 1)

		b := newGate(func(l *slog.Logger) { l.Info("n29-b") }, nil)
		current.Store(b)
		clock.set(time.Date(2026, 6, 1, 20, 17, 0, 0, time.UTC))
		if err := m.StartNow(id, "Anna"); err != nil {
			t.Fatal(err)
		}

		b.waitStarted(t)
		return st, clock, m, b
	}

	t.Run("während eines Laufs: der letzte beendete Lauf", func(t *testing.T) {
		_, _, m, b := begin(t)
		defer func() { close(b.open); waitEnded(t, m, id, 2) }()
		w := p.open(t, "running", modelOf(m, id, loc))
		w.WaitFor(nagotest.TextContains("22:17"), waitLimit)
		got := lastEnded(t, w)
		for _, want := range []string{"Fehler", "20:41", "2,5"} {
			if !strings.Contains(got, want) {
				t.Errorf("„Zuletzt beendet“ ohne %q des letzten beendeten Laufs: %s", want, got)
			}
		}

		for _, foreign := range []string{"läuft", "22:17"} {
			if strings.Contains(got, foreign) {
				t.Errorf("„Zuletzt beendet“ mit %q des laufenden Laufs: %s", foreign, got)
			}
		}
	})

	t.Run("Gegenzeile nach dem Ende des Laufs: dieser Lauf", func(t *testing.T) {
		_, clock, m, b := begin(t)
		clock.set(time.Date(2026, 6, 1, 20, 17, 1, 500_000_000, time.UTC))
		close(b.open)
		waitEnded(t, m, id, 2)
		w := p.open(t, "ended", modelOf(m, id, loc))
		w.WaitFor(nagotest.TextContains("22:17"), waitLimit)
		got := lastEnded(t, w)
		for _, want := range []string{"erfolgreich", "22:17", "1,5"} {
			if !strings.Contains(got, want) {
				t.Errorf("„Zuletzt beendet“ ohne %q des letzten Laufs: %s", want, got)
			}
		}

		for _, foreign := range []string{"20:41", "2,5"} {
			if strings.Contains(got, foreign) {
				t.Errorf("„Zuletzt beendet“ mit %q des vorigen Laufs: %s", foreign, got)
			}
		}
	})

	t.Run("nach einem Neustart, der den offenen Lauf als Neustart schloss: dieser Lauf", func(t *testing.T) {
		st, _, crashedManager, b := begin(t)
		crashed := st.crash(t)
		// the crashed manager ends its run before the test removes its log directory
		defer func() { close(b.open); waitEnded(t, crashedManager, id, 2) }()
		clock := &testClock{now: time.Date(2026, 6, 1, 20, 23, 0, 0, time.UTC)}
		m := crashed.manager(t, clock)
		if err := m.Configure(scheduler.Options{ID: id, Name: "n29-name", Kind: scheduler.Cron, Defaults: scheduler.Settings{ID: id, CronHour: 3}, Runner: func(context.Context) error { return nil }, Location: loc, Durable: true, Quiet: true}); err != nil {
			t.Fatal(err)
		}

		runs := waitEnded(t, m, id, 2)
		closed := runs[0]
		if closed.CanceledBy != scheduler.CanceledByRestart || closed.Outcome != scheduler.OutcomeCanceled {
			t.Fatalf("Prämisse: der offene Lauf ist als Neustart geschlossen: %+v", closed)
		}

		w := p.open(t, "restart", modelOf(m, id, loc))
		w.WaitFor(nagotest.TextContains("22:17"), waitLimit)
		got := lastEnded(t, w)
		for _, want := range []string{"abgebrochen", closed.CompletedAt.In(loc).Format("15:04")} {
			if !strings.Contains(got, want) {
				t.Errorf("„Zuletzt beendet“ ohne %q des geschlossenen Laufs: %s", want, got)
			}
		}

		for _, foreign := range []string{"20:41", "2,5", "noch nie"} {
			if strings.Contains(got, foreign) {
				t.Errorf("„Zuletzt beendet“ mit %q statt vom geschlossenen Lauf: %s", foreign, got)
			}
		}
	})
}

// TestN29PageOverviewZone checks finding (6): the admin center shows the mark and the times of a cron scheduler in its
// zone and names the zone, for two zones, so that the zone of the machine does not matter.
//
// Zusagen: R-FRAME-TIMES-BERLIN, R-DEC-SCHEDULE-NAGO
func TestN29PageOverviewZone(t *testing.T) {
	berlin := berlinZone(t)
	newYork, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		loc       *time.Location
		other     *time.Location
		start     string // begin of the run at 20:17 UTC in the zone
		otherTime string
	}{
		{"Europe/Berlin", berlin, newYork, "22:17", "16:17"},
		{"Gegenzeile America/New_York", newYork, berlin, "16:17", "22:17"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clock := &testClock{now: time.Date(2026, 6, 1, 20, 17, 0, 0, time.UTC)}
			st := newStores(t)
			m := st.manager(t, clock)
			settings := json.NewSloppyJSONRepository[scheduler.Settings, scheduler.ID](st.settings)
			uc := useCasesOf(m, settings)
			const id scheduler.ID = "n29.job"
			if err := uc.Configure(user.SU(), scheduler.Options{ID: id, Name: "n29-name", Kind: scheduler.Cron, Defaults: scheduler.Settings{ID: id, CronHour: 23, CronMinute: 45}, Runner: func(context.Context) error { return nil }, Location: tt.loc}); err != nil {
				t.Fatal(err)
			}

			if err := m.StartNow(id, "Anna"); err != nil {
				t.Fatal(err)
			}

			waitEnded(t, m, id, 1)
			su := asSU(uc)
			url := nagotest.Serve(t, func(cfg *application.Configurator) {
				cfg.SetApplicationID("de.worldiety.nagotest.n29")
				cfg.RootView("n29-overview", func(wnd core.Window) core.View { return uischeduler.PageOverview(wnd, su) })
			})

			w := nagotest.Dial(t, url, "n29-overview", nagotest.Values(core.Values{"id": string(id)}))
			w.WaitFor(nagotest.TextContains(tt.start), waitLimit)
			all := strings.Join(textsOfTree(w), " | ")
			for _, want := range []string{tt.loc.String(), "23:45", tt.start} {
				if !strings.Contains(all, want) {
					t.Errorf("%q fehlt: %s", want, all)
				}
			}

			for _, foreign := range []string{tt.other.String(), tt.otherTime} {
				if strings.Contains(all, foreign) {
					t.Errorf("%q aus einer anderen Zone: %s", foreign, all)
				}
			}
		})
	}
}

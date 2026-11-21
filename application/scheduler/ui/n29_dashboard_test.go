// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uischeduler_test

import (
	"context"
	"log/slog"
	"reflect"
	"sync"
	"testing"
	"time"
	_ "time/tzdata"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/scheduler"
	uischeduler "go.wdy.de/nago/application/scheduler/ui"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

// The dashboard redraws itself each second through a background task, thus a window within the process never
// settles. The tests use a websocket window (nagotest.Dial), which settles on the answer to its last action.

const waitLimit = 10 * time.Second

// count returns the amount of texts or labels which are exactly s, e.g. of buttons.
func count(w *nagotest.Window, s string) int {
	return w.FindAll(nagotest.Text(s)).Len() + w.FindAll(nagotest.Label(s)).Len()
}

// TestN29Dashboard checks the exported dashboard: it shows the statistics, the latest runs, the configuration of the
// embedding page and the log from the model, the hints above the statistics and the times in the zone of the model,
// but neither a head, breadcrumbs, buttons nor ui.H1. It reads the runs by pages.
//
// Zusagen: R-SCHEDULE-PAGE, R-FRAME-PAGE-HEAD, R-NFR-SCHEDULE-RUNS, R-DEC-SCHEDULE-NAGO
func TestN29Dashboard(t *testing.T) {
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}

	started := time.Date(2026, 6, 1, 20, 17, 0, 0, time.UTC)
	run := scheduler.Run{
		ID:          "n29.job#2026-06-01-00001",
		Scheduler:   "n29.job",
		Nr:          1,
		StartedAt:   started,
		CompletedAt: started.Add(3 * time.Second),
		Outcome:     scheduler.OutcomeSucceeded,
		Entries:     1,
		LogFile:     "n29.log",
		Trigger:     scheduler.TriggerManual,
		By:          "Anna",
		Due:         started,
	}

	status := scheduler.StatusResult{
		State:           scheduler.Paused,
		LastStartedAt:   run.StartedAt,
		LastCompletedAt: run.CompletedAt,
		NextPlannedAt:   started.Add(24 * time.Hour),
		Options:         scheduler.Options{ID: "n29.job", Name: "n29-name", Description: "n29-description", Kind: scheduler.Manual},
		LastRun:         std.Some(run),
		Stats:           scheduler.RunStats{Total24h: 1, LastDuration: 3 * time.Second, AvgDuration: 3 * time.Second},
	}

	page := scheduler.LogPage{Entries: []scheduler.LogEntry{{Level: slog.LevelInfo, Time: started.Add(time.Second), Msg: "n29-log-marker"}}, Total: 1}

	var mu sync.Mutex
	var limits []int
	model := func(loc *time.Location) uischeduler.DashboardModel {
		return uischeduler.DashboardModel{
			Status: func() (scheduler.StatusResult, error) { return status, nil },
			Runs: func(offset, limit int) ([]scheduler.Run, int, error) {
				mu.Lock()
				limits = append(limits, limit)
				mu.Unlock()
				if offset > 0 {
					return nil, 1, nil
				}

				return []scheduler.Run{run}, 1, nil
			},
			RunLog: func(id scheduler.RunID, q scheduler.LogQuery) (scheduler.LogPage, error) {
				if id != run.ID {
					return scheduler.LogPage{}, scheduler.ErrRunNotFound
				}

				return page, nil
			},
			Location: loc,
			Config:   ui.Text("n29-config-marker"),
			Hints:    []core.View{ui.Text("n29-hint-marker")},
		}
	}

	url := nagotest.Serve(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagotest.n29")
		cfg.RootView("n29-berlin", func(wnd core.Window) core.View { return uischeduler.Dashboard(wnd, model(berlin)) })
		cfg.RootView("n29-utc", func(wnd core.Window) core.View { return uischeduler.Dashboard(wnd, model(time.UTC)) })
	})

	tests := []struct {
		name          string
		path          core.NavigationPath
		shown, hidden string
	}{
		{"Zone Europe/Berlin", "n29-berlin", "22:17", "20:17"},
		{"Gegenzeile Zone UTC", "n29-utc", "20:17", "22:17"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := nagotest.Dial(t, url, tt.path)
			w.WaitFor(nagotest.Text("n29-log-marker"), waitLimit)

			for _, s := range []string{"Nächster Lauf", "Letzte Läufe", "Konfiguration", "Protokoll", "n29-config-marker", "n29-hint-marker"} {
				if w.FindAll(nagotest.Text(s)).Len() == 0 {
					t.Errorf("%q fehlt", s)
				}
			}

			// the hints stand above the statistics
			order := w.FindAll(nagotest.Where("Hinweis oder Kennzahl", func(n nagotest.Node) bool {
				tv, ok := n.Component.(*proto.TextView)
				return ok && (tv.Value == "n29-hint-marker" || tv.Value == "Nächster Lauf")
			}))
			if order.Len() < 2 {
				t.Errorf("Hinweis und Kennzahl erwartet, %d Treffer", order.Len())
			} else if first := order.At(0).Text(); first != "n29-hint-marker" {
				t.Errorf("Hinweis über den Kennzahlen erwartet, zuerst steht %q", first)
			}

			// the configuration of the embedding page stands in the card "Konfiguration"
			cards := w.FindAll(nagotest.And(nagotest.Containing(nagotest.Text("Konfiguration")), nagotest.Containing(nagotest.Text("n29-config-marker"))))
			inCard := false
			for i := range cards.Len() {
				c := cards.At(i)
				if c.FindAll(nagotest.Text("Letzte Läufe")).Len() == 0 && c.FindAll(nagotest.Text("Protokoll")).Len() == 0 {
					inCard = true
				}
			}

			if !inCard {
				t.Error("Inhalt des Einbettenden nicht in der Karte „Konfiguration“")
			}

			// no head, no breadcrumbs, no buttons, no ui.H1
			if n := w.FindAll(nagotest.Type[*proto.WindowTitle]()).Len(); n != 0 {
				t.Errorf("ui.H1 gefunden (%d Fenstertitel)", n)
			}

			for _, s := range []string{"n29-name", "n29-description", "Jetzt ausführen", "Starten", "Beenden", "Bearbeiten", "Weitere Aktionen", "Admin Center", "Hintergrundprozesse"} {
				if n := count(w, s); n != 0 {
					t.Errorf("%q gehört zum Kopf der Seite, gefunden %d", s, n)
				}
			}

			// times in the zone of the model
			if w.FindAll(nagotest.TextContains(tt.shown)).Len() == 0 {
				t.Errorf("Zeit %s fehlt", tt.shown)
			}

			if n := w.FindAll(nagotest.TextContains(tt.hidden)).Len(); n != 0 {
				t.Errorf("Zeit %s in einer anderen Zone gefunden (%d)", tt.hidden, n)
			}
		})
	}

	mu.Lock()
	defer mu.Unlock()
	if len(limits) == 0 {
		t.Fatal("Runs des Modells nie gelesen")
	}

	for _, l := range limits {
		if l < 1 || l > 100 {
			t.Fatalf("Runs mit limit %d gelesen, erwartet eine Seite von 1 bis 100", l)
		}
	}
}

// asSU replaces the subject of each use case by the super user, so that an anonymous websocket window may show the
// page of the admin center.
func asSU(uc scheduler.UseCases) scheduler.UseCases {
	subjectType := reflect.TypeFor[auth.Subject]()
	v := reflect.ValueOf(&uc).Elem()
	for i := range v.NumField() {
		f := v.Field(i)
		if f.Kind() != reflect.Func || f.IsNil() || f.Type().NumIn() == 0 || f.Type().In(0) != subjectType {
			continue
		}

		orig := reflect.ValueOf(f.Interface())
		f.Set(reflect.MakeFunc(f.Type(), func(args []reflect.Value) []reflect.Value {
			args[0] = reflect.ValueOf(user.SU())
			return orig.Call(args)
		}))
	}

	return uc
}

// TestN29PageOverviewKeepsHead is the counter row of TestN29Dashboard: the page of the admin center keeps its head
// with ui.H1, its breadcrumbs and its buttons around the dashboard.
//
// Zusagen: R-DEC-SCHEDULE-NAGO
func TestN29PageOverviewKeepsHead(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	uc := scheduler.NewUseCases(ctx, json.NewSloppyJSONRepository[scheduler.Settings, scheduler.ID](mem.NewBlobStore("n29.settings")))
	if err := uc.Configure(user.SU(), scheduler.Options{ID: "n29.job", Name: "n29-name", Kind: scheduler.Manual, Runner: func(context.Context) error { return nil }}); err != nil {
		t.Fatal(err)
	}

	su := asSU(uc)
	url := nagotest.Serve(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.nagotest.n29")
		cfg.RootView("n29-overview", func(wnd core.Window) core.View { return uischeduler.PageOverview(wnd, su) })
	})

	w := nagotest.Dial(t, url, "n29-overview", nagotest.Values(core.Values{"id": "n29.job"}))
	w.WaitFor(nagotest.Text("n29-name"), waitLimit)

	if w.FindAll(nagotest.Type[*proto.WindowTitle]()).Len() == 0 {
		t.Error("ui.H1 fehlt")
	}

	for _, s := range []string{"Jetzt ausführen", "Admin Center", "Hintergrundprozesse", "Nächster Lauf", "Letzte Läufe", "Konfiguration", "Protokoll"} {
		if count(w, s) == 0 {
			t.Errorf("%q fehlt", s)
		}
	}
}

// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion_test

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/ai/completion"
	uicompletion "go.wdy.de/nago/application/ai/completion/ui"
	"go.wdy.de/nago/application/ai/provider/echo"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

const (
	purposeRoute   core.NavigationPath = "rechnungen"
	noPurposeRoute core.NavigationPath = "notizen"
	invoicePurpose                     = "Offene Rechnungen prüfen und begleichen"
)

// windowContextRows are the same question for two routes of one application side by side: one registered with
// application.Purpose, one registered without. The second row shows that the first one is not answered for
// every route alike.
//
// The windows are anonymous, whose subject carries the context of the application, like a signed-in user does.
// user.SU carries context.Background, thus a window acting as SU sees no value of the application context at all.
var windowContextRows = []struct {
	name    string
	path    core.NavigationPath
	want    []string
	wantNot []string
}{
	{
		name:    "route with purpose",
		path:    purposeRoute,
		want:    []string{"Route: rechnungen\n", "Purpose of this screen: " + invoicePurpose + "\n"},
		wantNot: []string{"no description"},
	},
	{
		name:    "route without purpose",
		path:    noPurposeRoute,
		want:    []string{"Route: notizen", "no description was registered"},
		wantNot: []string{"Purpose of this screen"},
	},
}

// registerWindowContextRoutes registers both routes of the rows, only the first one with a purpose.
func registerWindowContextRoutes(cfg *application.Configurator, factory func(path core.NavigationPath) func(wnd core.Window) core.View) {
	cfg.SetApplicationID("de.worldiety.uicompletion.windowcontexttest")
	cfg.RootView(purposeRoute, factory(purposeRoute), application.Purpose(invoicePurpose))
	cfg.RootView(noPurposeRoute, factory(noPurposeRoute))
}

func checkContains(t *testing.T, where, got string, want, wantNot []string) {
	t.Helper()
	var failures []string
	for _, s := range want {
		if !strings.Contains(got, s) {
			failures = append(failures, "expected "+strconv.Quote(s))
		}
	}

	for _, s := range wantNot {
		if strings.Contains(got, s) {
			failures = append(failures, "did not expect "+strconv.Quote(s))
		}
	}

	if len(failures) > 0 {
		t.Errorf("%s in the %s, got:\n%s", strings.Join(failures, "; "), where, got)
	}
}

// recorded keeps the last string per route, written from within the window of that route.
type recorded struct {
	mutex  sync.Mutex
	byPath map[core.NavigationPath]string
}

func newRecorded() *recorded {
	return &recorded{byPath: map[core.NavigationPath]string{}}
}

func (r *recorded) put(path core.NavigationPath, s string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.byPath[path] = s
}

func (r *recorded) get(path core.NavigationPath) (string, bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	s, ok := r.byPath[path]
	return s, ok
}

// WindowContext names the purpose the current route was registered with, and says for a route without one
// that nothing was registered.
func TestWindowContext_NamesThePurposeRegisteredWithTheRoute(t *testing.T) {
	contexts := newRecorded()
	app := nagotest.New(t, func(cfg *application.Configurator) {
		registerWindowContextRoutes(cfg, func(path core.NavigationPath) func(wnd core.Window) core.View {
			return func(wnd core.Window) core.View {
				contexts.put(path, uicompletion.WindowContext(wnd))
				return ui.Text("Seite " + string(path))
			}
		})
	})

	for _, row := range windowContextRows {
		t.Run(row.name, func(t *testing.T) {
			w := app.Open(t, nil, row.path)
			w.Find(nagotest.Text("Seite " + string(row.path))).Exactly(1)

			got, ok := contexts.get(row.path)
			if !ok {
				t.Fatalf("the root view of %s has not been rendered", row.path)
			}

			checkContains(t, "window context", got, row.want, row.wantNot)
		})
	}
}

// An agent which embeds WindowContext into its system prompt sends the purpose of the route to the provider, as
// the documentation of WindowContext suggests.
func TestWindowContext_PurposeReachesTheSystemPrompt(t *testing.T) {
	const domainPrompt = "Du hilfst in der Buchhaltung."

	prompts := newRecorded()
	app := nagotest.New(t, func(cfg *application.Configurator) {
		registerWindowContextRoutes(cfg, func(path core.NavigationPath) func(wnd core.Window) core.View {
			fake := &chatFake{parent: func(_ context.Context, opts completion.Options) (completion.Result, error) {
				if strings.HasPrefix(opts.System, domainPrompt) {
					prompts.put(path, opts.System)
				}
				return text("erledigt"), nil
			}}

			return func(wnd core.Window) core.View {
				return uicompletion.Chat(wnd, uicompletion.ChatOptions{
					Completions:        fake,
					Provider:           echo.New("p", "p"),
					DisableCurrentTime: true,
					Agents: []uicompletion.Agent{{
						SystemPromptFunc: func() string {
							return domainPrompt + "\n\n" + uicompletion.WindowContext(wnd)
						},
					}},
				})
			}
		})
	})

	for _, row := range windowContextRows {
		t.Run(row.name, func(t *testing.T) {
			w := app.Open(t, nil, row.path)
			w.Type(w.Find(nagotest.Label("Nachricht")), "Was sehe ich hier?")
			w.Click(w.Find(nagotest.Text("Senden")))
			w.WaitFor(richText("erledigt"), 5*time.Second)

			got, ok := prompts.get(row.path)
			if !ok {
				t.Fatalf("no turn on %s was sent with the system prompt of the agent", row.path)
			}

			checkContains(t, "system prompt", got, row.want, row.wantNot)
		})
	}
}

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
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/provider/openai"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

var benchTags = []string{"bench-assistant"}

// BenchmarkAssistantWindow is G0 of M27: the build of a window with nago's floating assistant, once closed and once
// open. The window is decorated like an application does it with [Assistant.Decorate], with History and without a
// kept conversation, in S2 (see sessions_bench_test.go): 2 000 conversations of other persons and 200 of the person of
// the window, a real account with [RoleAssistantUser], of about 25 KB each with the same tags and provider in nago's
// store, one of the other persons waiting on a question, none of the own. The stock is opened with a fresh
// configurator.
//
// A build is a reload of the window with nagotest: the root view is destroyed and allocated again within the same
// scope, like a navigation to another page of the browser tab, until the window settled. The assistant stays open
// or closed across it. The first reload does not count, the following b.N are timed one by one and reported as
// median-ms. Run it with -test.benchtime=50x for the median over 50 builds per state.
func BenchmarkAssistantWindow(b *testing.B) {
	b.Run("closed", func(b *testing.B) { benchAssistantWindow(b, false) })
	b.Run("open", func(b *testing.B) { benchAssistantWindow(b, true) })
}

func benchAssistantWindow(b *testing.B, open bool) {
	llm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "list", "data": []any{map[string]any{"id": "bench-model", "object": "model"}}})
	}))
	b.Cleanup(llm.Close)

	var mgmt Management
	// builds counts the window builds, so that a reload which builds nothing cannot pass for a fast one
	var builds atomic.Int64
	opts := AssistantOptions{Label: "Assistent", History: true, Tags: benchTags}
	app, seed := stockS2.open(b, 0, func(c *application.Configurator) {
		mgmt = std.Must(Enable(c))
		c.SetDecorator(func(wnd core.Window, view core.View) core.View {
			return mgmt.Assistant.Decorate(wnd, view, opts)
		})
		c.RootViewWithDecoration("page", func(wnd core.Window) core.View {
			builds.Add(1)
			return ui.Text("Seite")
		})
	})
	cfg := app.Configurator()

	// the provider of the seeded conversations, now served by llm
	secrets := std.Must(cfg.SecretManagement())
	if err := secrets.UseCases.UpdateMyCredentials(user.SU(), seed.Secret, openai.Settings{Name: "Bench", BaseURL: llm.URL + "/v1"}); err != nil {
		b.Fatal(err)
	}

	for deadline := time.Now().Add(10 * time.Second); ; {
		if c, _ := mgmt.Assistant.Providers(user.SU()); len(c) == 1 && string(c[0].Provider.Identity()) == seed.ProviderHint {
			if models, err := mgmt.Assistant.Models(user.SU(), c[0]); err == nil && len(models) == 1 {
				break
			}
		}
		if time.Now().After(deadline) {
			b.Fatal("provider not loaded")
		}
		time.Sleep(10 * time.Millisecond)
	}

	users := std.Must(cfg.UserManagement())
	if err := users.UseCases.UpdateOtherRoles(user.SU(), seed.User, []role.ID{RoleAssistantUser}); err != nil {
		b.Fatal(err)
	}
	subject := seed.subject(b, app)

	w := app.Open(b, subject, "page", nagotest.Lean())
	if open {
		w.Click(w.Find(nagotest.Label("Assistent")))
	}

	// the state to measure, before and after
	state := func() {
		w.Find(nagotest.Text("Seite"))
		w.Find(nagotest.Label("Assistent"))
		if open {
			w.Find(nagotest.Label("Nachricht"))
		} else {
			w.FindAll(nagotest.Label("Nachricht")).None()
		}
	}

	w.Reload()
	state()

	durations := make([]time.Duration, 0, b.N)
	before := builds.Load()
	b.ResetTimer()
	for range b.N {
		start := time.Now()
		w.Reload()
		durations = append(durations, time.Since(start))
	}
	b.StopTimer()

	if got := builds.Load() - before; got < int64(b.N) {
		b.Fatalf("%d reloads built the window only %d times", b.N, got)
	}
	state()
	b.ReportMetric(float64(benchMedian(durations))/float64(time.Millisecond), "median-ms")
}

// benchMessages returns four rounds of a question, a tool call, its result and an answer, about 25 KB as JSON.
func benchMessages(i int) []completion.Message {
	sentence := fmt.Sprintf("Die Rechnung %d enthält drei Positionen, einen Rabatt und die Umsatzsteuer. ", 4700+i)
	text := func(n int) string { return strings.Repeat(sentence, n) }

	var msgs []completion.Message
	for round := range 4 {
		call := completion.ToolCall{
			ID:        fmt.Sprintf("call-%d-%d", i, round),
			Name:      "find_invoices",
			Arguments: std.Must(json.Marshal(map[string]any{"query": text(2), "limit": 20})),
		}
		msgs = append(msgs,
			completion.Message{Role: completion.User, Content: []completion.Content{completion.Text{Text: text(8)}}},
			completion.Message{Role: completion.Assistant, Content: []completion.Content{completion.Text{Text: text(1)}, call}},
			completion.Message{Role: completion.User, Content: []completion.Content{completion.ToolResult{ToolCallID: call.ID, Content: []completion.Content{completion.Text{Text: text(40)}}}}},
			completion.Message{Role: completion.Assistant, Content: []completion.Content{completion.Text{Text: text(25)}}},
		)
	}

	return msgs
}

func benchMedian(d []time.Duration) time.Duration {
	if len(d) == 0 {
		return 0
	}

	s := slices.Clone(d)
	slices.Sort(s)
	if len(s)%2 == 1 {
		return s[len(s)/2]
	}

	return (s[len(s)/2-1] + s[len(s)/2]) / 2
}

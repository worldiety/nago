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
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// The situation of BenchmarkAssistantWindow: the conversations of other persons in nago's store.
const benchForeignSessions = 2000

var benchTags = []string{"bench-assistant"}

// BenchmarkAssistantWindow measures the build of a window with nago's floating assistant, once closed and once
// open. The window is decorated like an application does it with [Assistant.Decorate], with History and without a
// kept conversation, in the situation of the lookup benchmark of uicompletion (BenchmarkFindResumable): 2 000
// conversations of other persons of about 25 KB each with the same tags and provider in nago's store, one of them
// waiting on a question, and the person of the window, a real account with [RoleAssistantUser], has none.
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
	app := nagotest.New(b, func(c *application.Configurator) {
		c.SetApplicationID("de.worldiety.assistantbench")
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

	secrets := std.Must(cfg.SecretManagement())
	sid := std.Must(secrets.UseCases.CreateSecret(user.SU(), openai.Settings{Name: "Bench", BaseURL: llm.URL + "/v1"}))
	if err := secrets.UseCases.UpdateMySecretGroups(user.SU(), sid, []group.ID{group.System}); err != nil {
		b.Fatal(err)
	}

	var providerHint string
	for deadline := time.Now().Add(10 * time.Second); providerHint == ""; {
		if c, _ := mgmt.Assistant.Providers(user.SU()); len(c) == 1 {
			providerHint = string(c[0].Provider.Identity())
			break
		}
		if time.Now().After(deadline) {
			b.Fatal("provider not loaded")
		}
		time.Sleep(10 * time.Millisecond)
	}

	users := std.Must(cfg.UserManagement())
	usr := std.Must(users.UseCases.Create(user.SU(), user.ShortRegistrationUser{
		Firstname:        "Mira",
		Lastname:         "Messend",
		Email:            "mira.messend@example.com",
		Password:         "Sup3r-Geheim!2026",
		PasswordRepeated: "Sup3r-Geheim!2026",
		Verified:         true,
	}))
	if err := users.UseCases.UpdateOtherRoles(user.SU(), usr.ID, []role.ID{RoleAssistantUser}); err != nil {
		b.Fatal(err)
	}
	subject := std.Must(users.UseCases.SubjectFromUser(user.SU(), usr.ID)).Unwrap()

	// the same store, which Enable has opened for the session use cases, written like by an earlier process
	repo := session.Repository(std.Must(application.JSONRepository[session.Session](cfg, string(session.Namespace))))
	seedForeignSessions(b, repo, std.Must(cfg.RDB()), benchTags, providerHint)

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

// seedForeignSessions writes benchForeignSessions conversations of 200 other persons into repo, each of about
// 25 KB of text and tool calls, the first one waiting on a question, and grants each its creator like
// session.Create does. It is the situation of BenchmarkFindResumable in uicompletion.
func seedForeignSessions(b *testing.B, repo session.Repository, rdb *rebac.DB, tags []string, providerHint string) {
	b.Helper()

	base := xtime.Now() - xtime.UnixMilliseconds(benchForeignSessions)*1000
	for i := range benchForeignSessions {
		s := session.Session{
			ID:           session.ID(fmt.Sprintf("bench-%05d", i)),
			Title:        fmt.Sprintf("Rechnungen prüfen %d", i),
			Model:        "bench-model",
			System:       "Du bist der Assistent einer Anwendung für Rechnungen.",
			ProviderHint: providerHint,
			Tags:         slices.Clone(tags),
			Messages:     benchMessages(i),
			CreatedBy:    user.ID(fmt.Sprintf("bench-person-%03d", i%200)),
			CreatedAt:    base + xtime.UnixMilliseconds(i)*1000,
			UpdatedAt:    base + xtime.UnixMilliseconds(i)*1000 + 500,
			Usage:        completion.Usage{InputTokens: 12_000, OutputTokens: 900},
		}

		if i == 0 {
			ask := completion.ToolCall{ID: "ask-0", Name: "ask_user", Arguments: json.RawMessage(`{"question":"Welche Rechnung soll ich prüfen?"}`)}
			s.Messages = append(s.Messages, completion.Message{Role: completion.Assistant, Content: []completion.Content{ask}})
			s.Pending = &completion.Continuation{Pending: []completion.PendingCall{{Kind: completion.PendingQuestion, Call: ask, Question: "Welche Rechnung soll ich prüfen?"}}}
			s.PendingRevision = 1
		}

		if err := repo.Save(s); err != nil {
			b.Fatal(err)
		}

		src := rebac.Entity{Namespace: user.Namespace, Instance: rebac.Instance(s.CreatedBy)}
		target := rebac.Entity{Namespace: session.Namespace, Instance: rebac.Instance(s.ID)}
		relations := []rebac.Relation{rebac.Owner}
		for _, pid := range session.InstancePermissions {
			relations = append(relations, rebac.Relation(pid))
		}
		for _, rel := range relations {
			if err := rdb.Put(rebac.Triple{Source: src, Relation: rel, Target: target}); err != nil {
				b.Fatal(err)
			}
		}
	}
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

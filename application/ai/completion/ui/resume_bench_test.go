// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xtime"
)

// The situation of BenchmarkFindResumable: the conversations of other persons in nago's store.
const (
	benchForeignSessions = 2000
	benchProviderHint    = "bench-provider"
)

var benchTags = []string{"bench-assistant"}

// BenchmarkFindResumable measures the lookup a freshly built floating assistant does to find the conversation it
// continues: nago's store of conversations holds 2 000 conversations of other persons of about 25 KB each with the
// same tags and provider, one of them waiting on a question, and the measuring person, a real account whose rights
// come from ReBAC, has none.
//
// The store is the one nago uses for nago.ai.session: the JSON repository of the configurator on its tdb in the
// data directory. Every run starts with fresh session use cases like a process after its start; the first lookup
// of a run is reported apart as first-ms, the following b.N lookups are timed one by one and reported as
// median-ms. Run it with -test.benchtime=50x for the median over 50 lookups after a first one, which does not
// count.
func BenchmarkFindResumable(b *testing.B) {
	app := nagotest.New(b, func(c *application.Configurator) {
		c.SetApplicationID("de.worldiety.uicompletion.bench")
	})
	cfg := app.Configurator()

	// the wiring of application/ai/cfg
	rdb := std.Must(cfg.RDB())
	rdb.RegisterStaticRule(rebac.StaticRule{Source: user.Namespace, Relation: rebac.Owner, Target: session.Namespace})
	for _, pid := range session.InstancePermissions {
		rdb.RegisterStaticRule(rebac.StaticRule{Source: user.Namespace, Relation: rebac.Relation(pid), Target: session.Namespace})
	}
	repo := session.Repository(std.Must(application.JSONRepository[session.Session](cfg, string(session.Namespace))))

	users := std.Must(cfg.UserManagement())
	usr := std.Must(users.UseCases.Create(user.SU(), user.ShortRegistrationUser{
		Firstname:        "Mira",
		Lastname:         "Messend",
		Email:            "mira.messend@example.com",
		Password:         "Sup3r-Geheim!2026",
		PasswordRepeated: "Sup3r-Geheim!2026",
		Verified:         true,
	}))
	subject := std.Must(users.UseCases.SubjectFromUser(user.SU(), usr.ID)).Unwrap()

	seedForeignSessions(b, repo, rdb, benchTags, benchProviderHint)

	uc := session.NewUseCases(repo, rdb)
	lookup := func() {
		if got := findResumable(subject, uc, benchTags, benchProviderHint); got != nil {
			b.Fatalf("the measuring person has no waiting conversation, but found %s", got.ID)
		}
	}

	first := timed(lookup)

	durations := make([]time.Duration, 0, b.N)
	b.ResetTimer()
	for range b.N {
		durations = append(durations, timed(lookup))
	}
	b.StopTimer()

	b.ReportMetric(millis(first), "first-ms")
	b.ReportMetric(millis(median(durations)), "median-ms")
}

// seedForeignSessions writes benchForeignSessions conversations of 200 other persons into repo, each of about
// 25 KB of text and tool calls, the first one waiting on a question, and grants each its creator like
// session.Create does.
func seedForeignSessions(b *testing.B, repo session.Repository, rdb *rebac.DB, tags []string, providerHint string) {
	b.Helper()

	base := xtime.Now() - xtime.UnixMilliseconds(benchForeignSessions)*1000
	var size int
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

		if i == 1 {
			buf := std.Must(json.Marshal(s))
			size = len(buf)
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

	b.Logf("seeded %d conversations of other persons, %d bytes of JSON each", benchForeignSessions, size)
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

func timed(fn func()) time.Duration {
	start := time.Now()
	fn()
	return time.Since(start)
}

func median(d []time.Duration) time.Duration {
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

func millis(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}

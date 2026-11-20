// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion_test

import (
	"context"
	"fmt"
	"iter"
	"reflect"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/ai/completion"
	uicompletion "go.wdy.de/nago/application/ai/completion/ui"
	"go.wdy.de/nago/application/ai/provider/echo"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
)

// historyReads records the conversations the use cases read from the store.
type historyReads struct {
	session.Repository
	mu    sync.Mutex
	reads []session.ID
}

func (r *historyReads) note(id session.ID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reads = append(r.reads, id)
}

func (r *historyReads) take() []session.ID {
	r.mu.Lock()
	defer r.mu.Unlock()
	res := r.reads
	r.reads = nil
	return res
}

func (r *historyReads) FindByID(id session.ID) (option.Opt[session.Session], error) {
	o, err := r.Repository.FindByID(id)
	if err == nil && o.IsSome() {
		r.note(id)
	}
	return o, err
}

func (r *historyReads) noted(seq iter.Seq2[session.Session, error]) iter.Seq2[session.Session, error] {
	return func(yield func(session.Session, error) bool) {
		for s, err := range seq {
			if err == nil {
				r.note(s.ID)
			}
			if !yield(s, err) {
				return
			}
		}
	}
}

func (r *historyReads) All() iter.Seq2[session.Session, error] { return r.noted(r.Repository.All()) }

func (r *historyReads) FindAllByPrefix(p session.ID) iter.Seq2[session.Session, error] {
	return r.noted(r.Repository.FindAllByPrefix(p))
}

func (r *historyReads) FindAllByID(ids iter.Seq[session.ID]) iter.Seq2[session.Session, error] {
	return r.noted(r.Repository.FindAllByID(ids))
}

// historyEntries reads the entries of the open history dialog, in their order: the texts of each row with a delete
// button.
func historyEntries(w *nagotest.Window) []string {
	rows := w.FindAll(nagotest.And(nagotest.TypeName("Stack"), nagotest.Containing(nagotest.Label("Verlauf löschen"))))
	var res []string
	for i := range rows.Len() {
		r := rows.At(i)
		if r.FindAll(nagotest.Label("Verlauf löschen")).Len() == 1 {
			res = append(res, r.Text())
		}
	}
	return res
}

// formerEntries are the entries the history dialog showed before the heads (history.go at c27dcff50), written out
// from the sessions of the person: the most recently updated first; the title, or the preview of the first message
// for an empty title or the title of the chat panel; the time of the last update; a note on a waiting session; the
// number of messages.
func formerEntries(sessions []session.Session, panelTitle string, loc *time.Location) []string {
	sessions = slices.Clone(sessions)
	slices.SortStableFunc(sessions, func(a, b session.Session) int { return int(b.UpdatedAt) - int(a.UpdatedAt) })
	var res []string
	for _, s := range sessions {
		label := s.Title
		if label == "" || label == panelTitle {
			label = s.Preview()
		}
		e := label + " " + s.UpdatedAt.Time(loc).Format("2006-01-02 15:04")
		if s.Pending != nil {
			e += " • wartet auf deine Antwort"
		}
		res = append(res, e+fmt.Sprintf(" %d Nachrichten", len(s.Messages)))
	}
	return res
}

// TestChat_HistoryFromHeads (wokoda R-DEC-ASSISTANT-SESSION-HEADS, „der Verlauf: die Köpfe unter h/<Person>/, keinen
// Inhalt“): the history dialog of a person shows the same entries as before the heads, with title or preview, time,
// waiting note and number of messages, also for a session titled like the chat panel, never a child or a session of
// another person; opening it, and picking an entry, whose messages the chat then shows, read no conversation of
// another person. Counter row: the same dialog with use cases without Heads shows the same entries and
// reads the conversations of the other person.
func TestChat_HistoryFromHeads(t *testing.T) {
	const panel = "Assistent"
	me := admin{Subject: user.SU(), id: "me"}
	pending := &completion.Continuation{Pending: []completion.PendingCall{{Kind: completion.PendingQuestion, Call: completion.ToolCall{ID: "q1", Name: completion.AskUserToolName}}}}
	msg := func(role completion.Role, text string) completion.Message {
		return completion.Message{Role: role, Content: []completion.Content{completion.Text{Text: text}}}
	}
	at := func(minutes int) xtime.UnixMilliseconds {
		return xtime.UnixMilliseconds(time.Date(2026, 11, 20, 9, minutes, 0, 0, time.UTC).UnixMilli())
	}
	stored := []session.Session{
		{ID: "m-titled", Title: "Rechnung prüfen", CreatedBy: "me", UpdatedAt: at(1), Messages: []completion.Message{msg(completion.User, "Bitte prüfe die Rechnung"), msg(completion.Assistant, "Gern")}},
		{ID: "m-untitled", CreatedBy: "me", UpdatedAt: at(3), Messages: []completion.Message{msg(completion.User, "Wie spät ist es in Tokio?")}},
		{ID: "m-panel", Title: panel, CreatedBy: "me", UpdatedAt: at(2), Messages: []completion.Message{msg(completion.User, "Alter Chat mit dem Titel des Panels"), msg(completion.Assistant, "a"), msg(completion.User, "b")}},
		{ID: "m-waiting", Title: "Welche Farbe?", CreatedBy: "me", UpdatedAt: at(4), Pending: pending, PendingRevision: 1, Messages: []completion.Message{msg(completion.User, "Male etwas")}},
		{ID: "m-child", ParentID: "m-titled", CreatedBy: "me", UpdatedAt: at(5), Messages: []completion.Message{msg(completion.User, "Unteragent")}},
	}
	for i := range 20 {
		stored = append(stored, session.Session{ID: session.ID(fmt.Sprintf("o%02d", i)), Title: "fremd", CreatedBy: "someone-else", UpdatedAt: at(10 + i), Messages: []completion.Message{msg(completion.User, "fremd")}})
	}
	mine := []session.ID{"m-titled", "m-untitled", "m-panel", "m-waiting", "m-child"}

	open := func(t *testing.T, heads bool) (*nagotest.Window, *historyReads, *time.Location) {
		t.Helper()
		content := mem.NewBlobStore("nago.ai.session.history")
		repo := &historyReads{Repository: session.Repository(json.NewSloppyJSONRepository[session.Session, session.ID](content))}
		for _, s := range stored {
			if err := repo.Repository.Save(s); err != nil {
				t.Fatal(err)
			}
		}
		rdb, err := rebac.NewDB(mem.NewBlobStore("rebac"))
		if err != nil {
			t.Fatal(err)
		}
		var opts []session.Option
		if heads {
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			opts = append(opts, session.Heads(ctx, mem.NewBlobStore("nago.ai.session.head.history")))
		}
		sessions := session.NewUseCases(repo, rdb, opts...)
		if heads {
			// the store comes from a former version: wait until its heads are built, i.e. the summaries of the person
			// read no conversation any more
			deadline := time.Now().Add(10 * time.Second)
			for {
				repo.take()
				for _, err := range sessions.FindSummaries(me, session.FindAllOptions{CreatedBy: "me"}) {
					if err != nil {
						t.Fatal(err)
					}
				}
				if len(repo.take()) == 0 {
					break
				}
				if time.Now().After(deadline) {
					t.Fatalf("the heads of the stored sessions are never built")
				}
				time.Sleep(time.Millisecond)
			}
		}

		var loc atomic.Pointer[time.Location]
		comps := &chatFake{parent: func(context.Context, completion.Options) (completion.Result, error) { return text("ok"), nil }}
		app := nagotest.New(t, func(cfg *application.Configurator) {
			cfg.SetApplicationID("de.worldiety.uicompletion.headshistorytest")
			cfg.RootView("chat", func(wnd core.Window) core.View {
				loc.Store(wnd.Location())
				return uicompletion.Chat(wnd, uicompletion.ChatOptions{
					Title:              panel,
					Sessions:           sessions,
					History:            true,
					Completions:        comps,
					Provider:           echo.New("p", "p"),
					DisableCurrentTime: true,
				})
			})
		})
		w := app.Open(t, me, "chat")
		return w, repo, loc.Load()
	}

	var entries [2][]string
	for i, c := range []struct {
		name    string
		heads   bool
		foreign bool
	}{
		{"with Heads", true, false},
		{"counter row: without Heads", false, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			w, repo, loc := open(t, c.heads)
			repo.take()
			w.Click(w.Find(nagotest.Text("Verlauf")))
			w.WaitFor(nagotest.Text("Rechnung prüfen"), 5*time.Second)
			reads := repo.take()
			foreign := slices.ContainsFunc(reads, func(id session.ID) bool { return !slices.Contains(mine, id) })
			if foreign != c.foreign {
				t.Errorf("opening the history read conversations of another person: %v, expected %v (%q)", foreign, c.foreign, reads)
			}

			var own []session.Session
			for _, s := range stored {
				if s.CreatedBy == "me" && s.ParentID == "" {
					own = append(own, s)
				}
			}
			entries[i] = historyEntries(w)
			if want := formerEntries(own, panel, loc); !reflect.DeepEqual(entries[i], want) {
				t.Errorf("the history shows\n\t%q\nexpected as before\n\t%q", entries[i], want)
			}

			w.Click(w.Find(nagotest.Text("Wie spät ist es in Tokio?")))
			w.WaitFor(richText("Wie spät ist es in Tokio?"), 5*time.Second)
			if got := repo.take(); c.heads && slices.ContainsFunc(got, func(id session.ID) bool { return !slices.Contains(mine, id) }) {
				t.Errorf("picking an entry read %q, expected no conversation of another person", got)
			}
		})
	}
	if !reflect.DeepEqual(entries[0], entries[1]) {
		t.Errorf("with and without Heads the history differs:\n\t%q\n\t%q", entries[0], entries[1])
	}
}

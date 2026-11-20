// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

// Acceptance tests of the heads of the sessions, wokoda decision R-DEC-ASSISTANT-SESSION-HEADS
// (requirements/dec/R-DEC-ASSISTANT-SESSION-HEADS/detail.md): with [Heads], FindResumable, FindSummaries, FindAll
// with CreatedBy and Delete give the same result as reading every session, also during a build, after a failed
// write and after a crash at every point of a write, and once the build has finished, they read only the sessions of
// the person, Delete only the deleted session and its children. The reference is always the same use cases without
// Heads on the same store (hfix.ref), which read every session. The instruments are in heads_probe_test.go. The
// restore is tested through the real wiring in application/ai/cfg.

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/xtime"
)

// TestSummary_OfASession (R-DEC-ASSISTANT-SESSION-HEADS): the summary carries what a list shows and a search needs,
// taken from the session. Each row stands next to its counterpart: waiting and not, ended at the turn limit and
// not, with and without title, child and parent.
func TestSummary_OfASession(t *testing.T) {
	waitingOne := legacy("w1", "alice", 1_000, true)
	waitingOne.Title = "Rechnung"
	waitingOne.Tags = []string{"ctx:a", "more"}
	waitingOne.CreatedAt = 900
	waitingOne.TurnLimited = true

	plain := legacy("p1", "bob", 2_000, false)
	plain.Messages = append(plain.Messages, answerResult("ok").Message, completion.Message{Role: completion.User, Content: userText("next")})

	child := legacy("c1", "carol", 3_000, false)
	child.ParentID = "p1"

	for _, s := range []Session{waitingOne, plain, child} {
		t.Run(string(s.ID), func(t *testing.T) {
			got := s.Summary()
			for _, c := range []struct {
				field     string
				got, want any
			}{
				{"ID", got.ID, s.ID},
				{"ParentID", got.ParentID, s.ParentID},
				{"CreatedBy", got.CreatedBy, s.CreatedBy},
				{"Title", got.Title, s.Title},
				{"Preview", got.Preview, s.Preview()},
				{"ProviderHint", got.ProviderHint, s.ProviderHint},
				{"Tags", fmt.Sprint(got.Tags), fmt.Sprint(s.Tags)},
				{"Pending", got.Pending, s.Pending != nil},
				{"TurnLimited", got.TurnLimited, s.TurnLimited},
				{"Messages", got.Messages, len(s.Messages)},
				{"CreatedAt", got.CreatedAt, s.CreatedAt},
				{"UpdatedAt", got.UpdatedAt, s.UpdatedAt},
			} {
				if c.got != c.want {
					t.Errorf("%s: %v, expected %v", c.field, c.got, c.want)
				}
			}
		})
	}
	if waitingOne.Summary().Pending == plain.Summary().Pending || waitingOne.Summary().Messages == plain.Summary().Messages {
		t.Errorf("premise: the rows do not differ in Pending and Messages")
	}
}

// TestHeads_SameAsReadingAll (R-DEC-ASSISTANT-SESSION-HEADS): after every step of a sequence of create, continue,
// wait, newer session, rename, resolve, dismiss, delete, children and a conversation of another person continued by
// alice, and after a restart, every search equals reading every session, for three persons, the system subject and a
// subject which is not signed in, in three contexts (other tags, other provider). Each step also names the session it
// expects for alice in ctx:a, so that the reference itself cannot drift into nothing. Delete removes the children,
// also those of another person. The store is empty at the start, so the heads are complete at once.
func TestHeads_SameAsReadingAll(t *testing.T) {
	f := newHFix(t, false)
	uc := f.start(false)
	if got := lookup(t, uc, f.alice, ctxA); got != "" {
		t.Fatalf("premise: empty store, found %q", got)
	}

	noChildOf := func(t *testing.T, parent ID) {
		t.Helper()
		all, _ := sessionsOf(f.ref, user.SU(), FindAllOptions{IncludeChildren: true})
		for _, s := range all {
			if s.ID == parent || s.ParentID == parent {
				t.Errorf("session %s is left after the deletion of %s", s.ID, parent)
			}
		}
	}

	var a, b, d, e, g Session
	var gKids []ID
	steps := []struct {
		name string
		do   func(t *testing.T)
		want func() []ID // what alice continues in ctx:a; more than one where a tie of the clock decides
	}{
		{"create without waiting", func(t *testing.T) { a = create(t, uc, f.alice, ctxA, ""); a = answer(t, uc, f.alice, a.ID) }, func() []ID { return []ID{""} }},
		{"wait", func(t *testing.T) { a = mustAsk(t, uc, f.alice, a.ID) }, func() []ID { return []ID{a.ID} }},
		{"newer waiting session", func(t *testing.T) { b = waiting(t, uc, f.alice, ctxA) }, func() []ID { return []ID{b.ID, a.ID} }},
		{"waiting session of another person", func(t *testing.T) { waiting(t, uc, f.bob, ctxA) }, func() []ID { return []ID{b.ID, a.ID} }},
		{"rename the older one", func(t *testing.T) {
			if err := uc.Rename(f.alice, a.ID, "renamed"); err != nil {
				t.Fatalf("rename: %v", err)
			}
		}, func() []ID { return []ID{a.ID, b.ID} }},
		{"resolve it", func(t *testing.T) {
			o, err := f.bare.FindByID(a.ID)
			if err != nil || o.IsNone() {
				t.Fatalf("setup: %v", err)
			}
			resolve(t, uc, f.alice, o.Unwrap())
		}, func() []ID { return []ID{b.ID} }},
		{"dismiss the other", func(t *testing.T) {
			if _, err := uc.Dismiss(f.alice, b.ID, b.PendingRevision); err != nil {
				t.Fatalf("dismiss: %v", err)
			}
		}, func() []ID { return []ID{""} }},
		{"waiting session with more tags", func(t *testing.T) { d = waiting(t, uc, f.alice, ctxMore) }, func() []ID { return []ID{""} }},
		{"delete it", func(t *testing.T) {
			if err := uc.Delete(f.alice, d.ID); err != nil {
				t.Fatalf("delete: %v", err)
			}
		}, func() []ID { return []ID{""} }},
		{"waiting child newer than its waiting parent", func(t *testing.T) {
			e = waiting(t, uc, f.alice, ctxA)
			child := create(t, uc, f.alice, ctxA, e.ID)
			if x := mustAsk(t, uc, f.alice, child.ID); x.ParentID != e.ID {
				t.Fatalf("setup: the child has no parent")
			}
		}, func() []ID { return []ID{e.ID} }},
		{"delete the parent with its child", func(t *testing.T) {
			if err := uc.Delete(f.alice, e.ID); err != nil {
				t.Fatalf("delete: %v", err)
			}
			noChildOf(t, e.ID)
		}, func() []ID { return []ID{""} }},
		{"a conversation of bob continued by alice", func(t *testing.T) {
			g = waiting(t, uc, f.bob, ctxA)
			grant(t, f.rdb, "alice", g.ID, PermAppend, PermFindByID)
			for range 2 {
				gKids = append(gKids, create(t, uc, f.alice, ctxA, g.ID).ID)
			}
		}, func() []ID { return []ID{""} }},
		{"bob deletes it with the children of alice", func(t *testing.T) {
			if err := uc.Delete(f.bob, g.ID); err != nil {
				t.Fatalf("delete: %v", err)
			}
			noChildOf(t, g.ID)
		}, func() []ID { return []ID{""} }},
		{"a waiting session of the system subject", func(t *testing.T) { waiting(t, uc, user.SU(), ctxA) }, func() []ID { return []ID{""} }},
		{"a waiting session again", func(t *testing.T) { e = waiting(t, uc, f.alice, ctxA) }, func() []ID { return []ID{e.ID} }},
	}
	for _, st := range steps {
		st.do(t)
		if got := lookup(t, uc, f.alice, ctxA); !slices.Contains(st.want(), got) {
			t.Errorf("%s: alice continues %q, expected one of %q", st.name, got, st.want())
		}
		f.same(st.name, uc)
	}
	if len(gKids) != 2 {
		t.Fatalf("premise: alice did not continue the conversation of bob")
	}

	// a restart: new use cases on the same stores trust the heads, without reading the conversations again
	restarted := f.restart()
	m := f.repo.mark()
	if got := lookup(t, restarted, f.alice, ctxA); got != e.ID {
		t.Errorf("after a restart: alice continues %q, expected %q", got, e.ID)
	}
	summariesOf(restarted, f.alice, FindAllOptions{CreatedBy: "alice"})
	if fr := foreign(f.repo.since(m), e.ID); len(fr) > 0 {
		t.Errorf("after a restart: the first searches of alice read %q, expected only her waiting session", fr)
	}
	f.same("after a restart", restarted)
}

// TestHeads_EquallyNewKeepsTheOrder (R-DEC-ASSISTANT-SESSION-HEADS): sessions with the same UpdatedAt, stored by a
// former version (thus found through the build of the heads), are decided like before: the first in the order of the
// store, whatever the order they were written in. Counter row: one newer by a millisecond wins. A waiting child,
// newer than everything, never; its counter row is the same session without parent.
func TestHeads_EquallyNewKeepsTheOrder(t *testing.T) {
	const at = xtime.UnixMilliseconds(1_000)
	stored := func(id ID, updated xtime.UnixMilliseconds, parent ID) Session {
		s := legacy(id, "alice", updated, true)
		s.ParentID = parent
		return s
	}
	for _, c := range []struct {
		name     string
		sessions []Session
		want     ID
	}{
		{"three equally new, written s3, s1, s2", []Session{stored("s3", at, ""), stored("s1", at, ""), stored("s2", at, "")}, "s1"},
		{"counter row: s2 newer by a millisecond", []Session{stored("s3", at, ""), stored("s1", at, ""), stored("s2", at+1, "")}, "s2"},
		{"a waiting child newer than all", []Session{stored("s1", at, ""), stored("s0", at+5, "s1")}, "s1"},
		{"counter row: the same session without parent", []Session{stored("s1", at, ""), stored("s0", at+5, "")}, "s0"},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := newHFix(t, false)
			f.stored(c.sessions...)
			uc := f.start(false)
			f.awaitBuilt(uc)
			got, ref := lookup(t, uc, f.alice, ctxA), readAll(t, f.ref, f.alice, ctxA)
			if got != c.want || ref != c.want {
				t.Errorf("FindResumable %q, reading every session %q, expected %q", got, ref, c.want)
			}
			f.same(c.name, uc)
		})
	}
}

// TestHeads_ReBACAndSubjects (R-DEC-ASSISTANT-SESSION-HEADS): the searches return only what the subject may read
// like through FindAll: with the grants of alice on her newest waiting session withdrawn, it is not found, the older
// one is (counter row: before the withdrawal the newest), her summaries lose it; with both withdrawn, none. A subject
// which is not signed in continues nothing (counter row: alice signed in); its summaries are those FindAll yields for
// it, see TestHeads_SameAsReadingAll. The system subject, which may read every
// session, continues only the session it created itself (counter row: its own one), never the one of alice.
func TestHeads_ReBACAndSubjects(t *testing.T) {
	f := newHFix(t, false)
	uc := f.start(false)
	older := waiting(t, uc, f.alice, ctxA)
	newer := waiting(t, uc, f.alice, ctxA)
	withdraw := func(id ID) {
		if err := f.rdb.DeleteByQuery(rebac.Select().Where().Target().Is(Namespace, rebacInstance(id))); err != nil {
			t.Fatal(err)
		}
	}

	type row struct {
		name      string
		subject   auth.Subject
		want      ID
		summaries int // of alice, as the subject sees them
	}
	check := func(rows ...row) {
		t.Helper()
		for _, r := range rows {
			got, ref := lookup(t, uc, r.subject, ctxA), readAll(t, f.ref, r.subject, ctxA)
			if got != r.want || ref != r.want {
				t.Errorf("%s: FindResumable %q, reading every session %q, expected %q", r.name, got, ref, r.want)
			}
			if n, _ := summariesOf(uc, r.subject, FindAllOptions{CreatedBy: "alice"}); len(n) != r.summaries {
				t.Errorf("%s: %d summaries of alice, expected %d", r.name, len(n), r.summaries)
			}
		}
		f.same("rebac", uc)
	}

	newestID, otherID := newer.ID, older.ID
	if older.UpdatedAt == newer.UpdatedAt && older.ID < newer.ID {
		newestID, otherID = older.ID, newer.ID // equally new: the first in the order of the store
	}

	check(
		row{"counter row: alice before the withdrawal", f.alice, newestID, 2},
		row{"not signed in", invalidSubject{f.alice}, "", 2}, // FindSummaries yields what FindAll yields
		row{"system subject without own session", user.SU(), "", 2},
	)
	withdraw(newestID)
	check(row{"alice without grants on the newest", f.alice, otherID, 1})
	withdraw(otherID)
	check(row{"alice without grants on both", f.alice, "", 0})

	own := waiting(t, uc, user.SU(), ctxA)
	check(
		row{"counter row: system subject with its own session", user.SU(), own.ID, 2},
		row{"alice does not find the session of the system subject", f.alice, "", 0},
	)
}

// TestHeads_RightAfterAFailedWrite (R-DEC-ASSISTANT-SESSION-HEADS): a write which fails, before or after the store
// has written the conversation, does not leave a search wrong: they equal reading every session at once, after the
// next write and after a restart, which clears the marks. Each failing row stands next to the same write failing on
// the other side of the store.
func TestHeads_RightAfterAFailedWrite(t *testing.T) {
	for _, c := range []struct {
		name string
		run  func(t *testing.T, f *hfix, uc UseCases) (expected ID)
	}{
		{"dismiss fails after it has been written", func(t *testing.T, f *hfix, uc UseCases) ID {
			x := waiting(t, uc, f.alice, ctxA)
			f.repo.failWrite = func(op string, s Session, id ID) (bool, bool) { return op == "save" && id == x.ID, true }
			if _, err := uc.Dismiss(f.alice, x.ID, x.PendingRevision); err == nil {
				t.Fatalf("premise: the dismiss did not fail")
			}
			return ""
		}},
		{"counter row: dismiss fails before it has been written", func(t *testing.T, f *hfix, uc UseCases) ID {
			x := waiting(t, uc, f.alice, ctxA)
			f.repo.failWrite = func(op string, s Session, id ID) (bool, bool) { return op == "save" && id == x.ID, false }
			if _, err := uc.Dismiss(f.alice, x.ID, x.PendingRevision); err == nil {
				t.Fatalf("premise: the dismiss did not fail")
			}
			return x.ID
		}},
		{"wait fails after it has been written", func(t *testing.T, f *hfix, uc UseCases) ID {
			x := create(t, uc, f.alice, ctxA, "")
			f.repo.failWrite = func(op string, s Session, id ID) (bool, bool) {
				return op == "save" && id == x.ID && s.Pending != nil, true
			}
			_, _ = ask(uc, f.alice, x.ID)
			if o, _ := f.bare.FindByID(x.ID); o.IsNone() || o.Unwrap().Pending == nil {
				t.Fatalf("premise: the waiting session has not been written")
			}
			return x.ID
		}},
		{"counter row: wait fails before it has been written", func(t *testing.T, f *hfix, uc UseCases) ID {
			x := create(t, uc, f.alice, ctxA, "")
			f.repo.failWrite = func(op string, s Session, id ID) (bool, bool) {
				return op == "save" && id == x.ID && s.Pending != nil, false
			}
			_, _ = ask(uc, f.alice, x.ID)
			if o, _ := f.bare.FindByID(x.ID); o.IsNone() || o.Unwrap().Pending != nil {
				t.Fatalf("premise: the waiting session has been written")
			}
			return ""
		}},
		{"delete fails after it has been deleted", func(t *testing.T, f *hfix, uc UseCases) ID {
			x := waiting(t, uc, f.alice, ctxA)
			f.repo.failWrite = func(op string, s Session, id ID) (bool, bool) { return op == "delete" && id == x.ID, true }
			if err := uc.Delete(f.alice, x.ID); err == nil {
				t.Fatalf("premise: the delete did not fail")
			}
			return ""
		}},
		{"counter row: delete fails before it has been deleted", func(t *testing.T, f *hfix, uc UseCases) ID {
			x := waiting(t, uc, f.alice, ctxA)
			f.repo.failWrite = func(op string, s Session, id ID) (bool, bool) { return op == "delete" && id == x.ID, false }
			if err := uc.Delete(f.alice, x.ID); err == nil {
				t.Fatalf("premise: the delete did not fail")
			}
			return x.ID
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := newHFix(t, false)
			uc := f.start(false)
			want := c.run(t, f, uc)
			if got := lookup(t, uc, f.alice, ctxA); got != want {
				t.Errorf("after the failed write: alice continues %q, expected %q", got, want)
			}
			f.same("after the failed write", uc)
			next := waiting(t, uc, f.bob, ctxA)
			if got := lookup(t, uc, f.bob, ctxA); got != next.ID {
				t.Errorf("next write: bob continues %q, expected %q", got, next.ID)
			}
			f.same("after the next write", uc)
			restarted := f.restart()
			eventually(t, "the start clears the marks", func() bool { return len(keys(t, f.heads, "o/")) == 0 })
			f.same("after a restart", restarted)
		})
	}
}

// TestHeads_ReadsNoForeignSession (R-DEC-ASSISTANT-SESSION-HEADS, row „A sucht ihr fortzusetzendes Gespräch und
// ihren Verlauf, B hat 1 000 Gespräche, Aufbau fertig“): once the heads are built, the lookup of the session to
// continue, the history (FindSummaries), FindAll with CreatedBy and Delete read no conversation of bob, the history
// none at all, Delete only the deleted session and its child. Counter rows: the same searches without Heads, and while
// a build is held in the middle of the conversations of bob, read his conversations, and still answer as reading
// every session. An empty store is ready at once: right after the start, nothing of bob is read.
func TestHeads_ReadsNoForeignSession(t *testing.T) {
	seed := func(f *hfix) {
		for i := range 1000 {
			if err := f.bare.Save(legacy(ID(fmt.Sprintf("b%04d", i)), "bob", xtime.UnixMilliseconds(10_000+i), i%2 == 0)); err != nil {
				t.Fatal(err)
			}
		}
		f.stored(legacy("a-legacy", "alice", 500, true))
	}
	type search struct {
		name    string
		run     func(uc UseCases)
		content bool // reads the content of hits at all
	}
	searches := func(f *hfix) []search {
		return []search{
			{"FindResumable", func(uc UseCases) { lookup(t, uc, f.alice, ctxA) }, true},
			{"FindSummaries with CreatedBy", func(uc UseCases) { summariesOf(uc, f.alice, FindAllOptions{CreatedBy: "alice"}) }, false},
			{"FindAll with CreatedBy", func(uc UseCases) { sessionsOf(uc, f.alice, FindAllOptions{CreatedBy: "alice"}) }, true},
		}
	}

	t.Run("heads built", func(t *testing.T) {
		f := newHFix(t, false)
		seed(f)
		uc := f.start(false)
		f.awaitBuilt(uc)
		w := waiting(t, uc, f.alice, ctxA)
		s := create(t, uc, f.alice, ctxA, "")
		c := create(t, uc, f.alice, ctxA, s.ID)
		mine := []ID{"a-legacy", w.ID, s.ID, c.ID}
		if got := lookup(t, uc, f.alice, ctxA); got != w.ID {
			t.Fatalf("premise: alice continues %q, expected %q", got, w.ID)
		}
		f.sameFor("heads built", uc, f.alice)

		for _, x := range searches(f) {
			m := f.repo.mark()
			x.run(uc)
			reads := f.repo.since(m)
			if fr := foreign(reads, mine...); len(fr) > 0 {
				t.Errorf("%s read %d conversations of others, e.g. %q", x.name, len(fr), fr[0])
			}
			if !x.content && len(reads) > 0 {
				t.Errorf("%s read %d conversations, expected only heads", x.name, len(reads))
			}
		}
		m := f.repo.mark()
		if err := uc.Delete(f.alice, s.ID); err != nil {
			t.Fatal(err)
		}
		if fr := foreign(f.repo.since(m), s.ID, c.ID); len(fr) > 0 {
			t.Errorf("Delete read %d other conversations than the deleted one and its child, e.g. %q", len(fr), fr[0])
		}
		if o, _ := f.bare.FindByID(c.ID); o.IsSome() {
			t.Errorf("the child of the deleted session is left")
		}
	})

	t.Run("counter row: without Heads", func(t *testing.T) {
		f := newHFix(t, false)
		seed(f)
		uc := NewUseCases(f.repo, f.rdb)
		for _, x := range searches(f) {
			if x.name == "FindSummaries with CreatedBy" {
				continue // it reads what FindAll reads
			}
			m := f.repo.mark()
			x.run(uc)
			if fr := foreign(f.repo.since(m), "a-legacy"); len(fr) == 0 {
				t.Errorf("%s without Heads read no conversation of bob; the probe does not see the reads", x.name)
			}
		}
	})

	t.Run("counter row: during a build", func(t *testing.T) {
		f := newHFix(t, false)
		seed(f)
		h := f.repo.holdRead("b0500")
		t.Cleanup(h.free)
		uc := f.start(false)
		h.wait(t, "the read of b0500 by the build")
		for _, x := range searches(f) {
			m := f.repo.mark()
			x.run(uc)
			if fr := foreign(f.repo.since(m), "a-legacy"); len(fr) == 0 {
				t.Errorf("%s during the build read no conversation above the cursor", x.name)
			}
		}
		f.sameFor("during the build", uc, f.alice)
		h.free()
		f.awaitBuilt(uc)
		for _, x := range searches(f) {
			m := f.repo.mark()
			x.run(uc)
			if fr := foreign(f.repo.since(m), "a-legacy"); len(fr) > 0 {
				t.Errorf("%s after the build read %d conversations of bob", x.name, len(fr))
			}
		}
	})

	t.Run("an empty store is ready at once", func(t *testing.T) {
		f := newHFix(t, false)
		uc := f.start(false)
		for range 3 {
			waiting(t, uc, f.bob, ctxA)
		}
		w := waiting(t, uc, f.alice, ctxA)
		for _, x := range searches(f) {
			m := f.repo.mark()
			x.run(uc)
			if fr := foreign(f.repo.since(m), w.ID); len(fr) > 0 {
				t.Errorf("%s right after the start read %d conversations of bob", x.name, len(fr))
			}
		}
	})
}

// TestHeads_ContentGoneBypassingUseCases (R-DEC-ASSISTANT-SESSION-HEADS, row „Inhalt an den Anwendungsfällen vorbei
// gelöscht“): a head whose conversation has been deleted past the use cases counts for no search. Counter row: the
// other session, whose conversation is there, appears in every search.
func TestHeads_ContentGoneBypassingUseCases(t *testing.T) {
	f := newHFix(t, false)
	uc := f.start(false)
	gone := waiting(t, uc, f.alice, ctxA)
	kept := create(t, uc, f.alice, ctxA, "")

	appears := func(id ID) (resumable, summary, findAll bool) {
		resumable = lookup(t, uc, f.alice, ctxA) == id
		all, _ := summariesOf(uc, f.alice, FindAllOptions{CreatedBy: "alice"})
		summary = slices.ContainsFunc(all, func(s Summary) bool { return s.ID == id })
		found, _ := sessionsOf(uc, f.alice, FindAllOptions{CreatedBy: "alice"})
		findAll = slices.Contains(idsOf(found), id)
		return
	}

	if r, s, a := appears(gone.ID); !r || !s || !a {
		t.Fatalf("premise: the waiting session does not appear: %v %v %v", r, s, a)
	}
	if err := f.bare.DeleteByID(gone.ID); err != nil {
		t.Fatal(err)
	}
	if !hasHeadOf(t, f.heads, gone.ID) {
		t.Fatalf("premise: the head of the deleted conversation is gone as well")
	}
	if r, s, a := appears(gone.ID); r || s || a {
		t.Errorf("a head without conversation appears: FindResumable %v, FindSummaries %v, FindAll %v", r, s, a)
	}
	if _, s, a := appears(kept.ID); !s || !a {
		t.Errorf("counter row: the session with conversation does not appear: FindSummaries %v, FindAll %v", s, a)
	}
	f.same("after the deletion past the use cases", uc)
}

// crashSetup is what every crash starts from: alice has a parent with a child and a waiting session, bob a waiting
// session.
type crashSetup struct{ p, c, w, b Session }

func crashFixture(t *testing.T) (*hfix, UseCases, crashSetup) {
	t.Helper()
	f := newHFix(t, true)
	uc := f.start(true)
	p := create(t, uc, f.alice, ctxA, "")
	c := create(t, uc, f.alice, ctxA, p.ID)
	w := waiting(t, uc, f.alice, ctxA)
	b := waiting(t, uc, f.bob, ctxA)
	return f, uc, crashSetup{p, c, w, b}
}

// TestHeads_CrashAtEveryWrite (R-DEC-ASSISTANT-SESSION-HEADS, row „Absturz nach dem Inhalt, vor dem Kopf“): the
// stores stop writing after the k-th write of a create, a wait, a dismiss, a rename, a new child and a deletion with
// its child, for every k from none to all, like a process which crashed there. The next start on the same stores
// clears every mark, reading only the marked conversations, and every search equals reading every session. Counter
// row: with the marks removed after the conversation has been written and before its head, the next start keeps
// the former state of the heads, and a search differs from reading every session.
func TestHeads_CrashAtEveryWrite(t *testing.T) {
	ops := []struct {
		name   string
		target func(s crashSetup) ID // "" for a session the operation creates
		run    func(t *testing.T, f *hfix, uc UseCases, s crashSetup)
		mark   bool // the counter row without marks applies
	}{
		{"create", func(crashSetup) ID { return "" }, func(t *testing.T, f *hfix, uc UseCases, s crashSetup) {
			_, _ = uc.Create(f.alice, CreateOptions{Model: "fake-model", Tags: ctxA.Tags, ProviderHint: ctxA.ProviderHint, Input: userText("new")})
		}, true},
		{"wait", func(s crashSetup) ID { return s.p.ID }, func(t *testing.T, f *hfix, uc UseCases, s crashSetup) {
			_, _ = ask(uc, f.alice, s.p.ID)
		}, true},
		{"dismiss", func(s crashSetup) ID { return s.w.ID }, func(t *testing.T, f *hfix, uc UseCases, s crashSetup) {
			_, _ = uc.Dismiss(f.alice, s.w.ID, s.w.PendingRevision)
		}, true},
		{"rename", func(s crashSetup) ID { return s.w.ID }, func(t *testing.T, f *hfix, uc UseCases, s crashSetup) {
			_ = uc.Rename(f.alice, s.w.ID, "renamed")
		}, true},
		{"new child", func(crashSetup) ID { return "" }, func(t *testing.T, f *hfix, uc UseCases, s crashSetup) {
			_, _ = uc.Create(f.alice, CreateOptions{Model: "fake-model", ParentID: s.w.ID})
		}, false},
		{"delete with its child", func(s crashSetup) ID { return s.p.ID }, func(t *testing.T, f *hfix, uc UseCases, s crashSetup) {
			_ = uc.Delete(f.alice, s.p.ID)
		}, false},
	}

	for _, op := range ops {
		t.Run(op.name, func(t *testing.T) {
			f, uc, s := crashFixture(t)
			f.log.arm(-1)
			op.run(t, f, uc, s)
			total := f.log.log()
			if len(total) == 0 {
				t.Fatalf("premise: %s wrote nothing", op.name)
			}
			target := op.target(s)
			afterContent := -1 // the k after which the conversation is written, and its head is not
			for i, w := range total {
				if strings.HasPrefix(w, "content set ") && (target == "" || w == "content set "+string(target)) {
					afterContent = i + 1
					break
				}
			}

			for k := 0; k <= len(total); k++ {
				t.Run(fmt.Sprintf("after %d of %d writes", k, len(total)), func(t *testing.T) {
					f, uc, s := crashFixture(t)
					f.log.arm(k)
					op.run(t, f, uc, s)
					f.cancel()

					var marked []ID
					for _, key := range keys(t, f.heads, "o/") {
						marked = append(marked, ID(strings.TrimPrefix(key, "o/")))
					}
					restarted := f.restart()
					eventually(t, "the start clears the marks", func() bool { return len(keys(t, f.heads, "o/")) == 0 })
					if fr := foreign(f.repo.since(0), marked...); len(fr) > 0 {
						t.Errorf("the start read %q, expected only the marked %q", fr, marked)
					}
					f.same(fmt.Sprintf("%s, crashed after %d writes, next start", op.name, k), restarted)
				})
			}

			if !op.mark {
				return
			}
			t.Run("counter row: without marks", func(t *testing.T) {
				if afterContent < 0 {
					t.Fatalf("premise: %s wrote no conversation: %q", op.name, total)
				}
				f, uc, s := crashFixture(t)
				f.log.arm(afterContent)
				op.run(t, f, uc, s)
				f.cancel()
				if len(keys(t, f.heads, "o/")) == 0 {
					t.Fatalf("premise: no mark after the conversation of %s has been written: %q", op.name, total[:afterContent])
				}
				for _, key := range keys(t, f.heads, "o/") {
					if err := f.heads.Delete(context.Background(), key); err != nil {
						t.Fatal(err)
					}
				}
				restarted := f.restart()
				if d := f.diffs(restarted); len(d) == 0 {
					t.Errorf("without the marks, every search still equals reading every session; the crash after %d writes does not separate the conversation from its head", afterContent)
				}
			})
		})
	}
}

// TestHeads_MarkAfterAFailedHead (R-DEC-ASSISTANT-SESSION-HEADS): when the head of a rename cannot be written, the mark
// stays, and every search of the same process equals reading every session. The next write of the session, or the
// next start, clears it. A start which cannot read the marked conversation keeps the mark, and the searches stay
// right; the start after it clears the mark. Counter row: the rename whose head has been written leaves no mark.
func TestHeads_MarkAfterAFailedHead(t *testing.T) {
	setup := func(t *testing.T, failHead bool) (*hfix, UseCases, Session) {
		f := newHFix(t, true)
		uc := f.start(true)
		w := waiting(t, uc, f.alice, ctxA)
		create(t, uc, f.alice, ctxA, "")
		if failHead {
			f.log.mu.Lock()
			f.log.failOnce = func(e string) bool {
				return strings.HasPrefix(e, "head set h/") && strings.HasSuffix(e, "/"+string(w.ID))
			}
			f.log.mu.Unlock()
		}
		_ = uc.Rename(f.alice, w.ID, "renamed")
		return f, uc, w
	}
	markOf := func(t *testing.T, f *hfix, id ID) bool {
		return slices.Contains(keys(t, f.heads, "o/"), "o/"+string(id))
	}

	t.Run("counter row: the head has been written", func(t *testing.T) {
		f, uc, w := setup(t, false)
		if markOf(t, f, w.ID) {
			t.Errorf("a mark is left after a rename which wrote its head")
		}
		f.same("after the rename", uc)
	})

	for _, c := range []struct {
		name string
		then func(t *testing.T, f *hfix, uc UseCases, w Session)
	}{
		{"the next write of the session clears it", func(t *testing.T, f *hfix, uc UseCases, w Session) {
			if err := uc.Rename(f.alice, w.ID, "again"); err != nil {
				t.Fatal(err)
			}
			if markOf(t, f, w.ID) {
				t.Errorf("the mark is left after the next write of the session")
			}
			f.same("after the next write", uc)
		}},
		{"the next start clears it", func(t *testing.T, f *hfix, uc UseCases, w Session) {
			restarted := f.restart()
			eventually(t, "the start clears the mark", func() bool { return !markOf(t, f, w.ID) })
			f.same("after the next start", restarted)
		}},
		{"a start which cannot read the session keeps it", func(t *testing.T, f *hfix, uc UseCases, w Session) {
			f.repo = newContentRepo(f.bare)
			failed := f.repo.failRead(w.ID, 0)
			restarted := f.start(false)
			select {
			case <-failed:
			case <-time.After(patience):
				t.Fatalf("premise: the start did not read the marked conversation")
			}
			if !markOf(t, f, w.ID) {
				t.Errorf("the mark is gone, although the start could not read the conversation")
			}
			f.same("after a start which could not read the session", restarted)
			again := f.restart()
			eventually(t, "the start after it clears the mark", func() bool { return !markOf(t, f, w.ID) })
			f.same("after the start after it", again)
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			f, uc, w := setup(t, true)
			if !markOf(t, f, w.ID) {
				t.Fatalf("the mark is gone after the head of the rename failed")
			}
			f.same("right after the failed head", uc)
			c.then(t, f, uc, w)
		})
	}
}

// TestHeads_BuildAgainstConcurrentWrites (R-DEC-ASSISTANT-SESSION-HEADS, row „Speichern gegen laufenden Aufbau“): a
// write which runs while the build has read the conversation s03 and has not yet written its head is not lost. The
// interleaving is forced: the store holds the build right after the read of s03; then the write runs until it has
// been written (a build without the lock of s03) or until it waits for the build (one with it); only then the build
// goes on. Afterwards head and conversation agree, and every search equals reading every session. Rows: s03 is
// renamed, starts to wait, is dismissed, is deleted.
func TestHeads_BuildAgainstConcurrentWrites(t *testing.T) {
	for _, c := range []struct {
		name    string
		waiting bool
		write   func(f *hfix, uc UseCases) error
		check   func(t *testing.T, s []Summary)
	}{
		{"rename", false, func(f *hfix, uc UseCases) error { return uc.Rename(f.alice, "s03", "renamed") }, func(t *testing.T, s []Summary) {
			if i := slices.IndexFunc(s, func(x Summary) bool { return x.ID == "s03" }); i < 0 || s[i].Title != "renamed" {
				t.Errorf("the summary of s03 does not carry the new title: %+v", s)
			}
		}},
		{"wait", false, func(f *hfix, uc UseCases) error { _, err := ask(uc, f.alice, "s03"); return err }, func(t *testing.T, s []Summary) {
			if i := slices.IndexFunc(s, func(x Summary) bool { return x.ID == "s03" }); i < 0 || !s[i].Pending {
				t.Errorf("the summary of s03 does not wait: %+v", s)
			}
		}},
		{"dismiss", true, func(f *hfix, uc UseCases) error { _, err := uc.Dismiss(f.alice, "s03", 1); return err }, func(t *testing.T, s []Summary) {
			if i := slices.IndexFunc(s, func(x Summary) bool { return x.ID == "s03" }); i < 0 || s[i].Pending {
				t.Errorf("the summary of s03 still waits: %+v", s)
			}
		}},
		{"delete", false, func(f *hfix, uc UseCases) error { return uc.Delete(f.alice, "s03") }, func(t *testing.T, s []Summary) {
			if slices.ContainsFunc(s, func(x Summary) bool { return x.ID == "s03" }) {
				t.Errorf("the deleted s03 is listed: %+v", s)
			}
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := newHFix(t, false)
			for i := 1; i <= 6; i++ {
				id := ID(fmt.Sprintf("s%02d", i))
				f.stored(legacy(id, "alice", xtime.UnixMilliseconds(1_000+i), id == "s03" && c.waiting))
			}
			h := f.repo.holdRead("s03")
			t.Cleanup(h.free)
			uc := f.start(false)
			h.wait(t, "the read of s03 by the build")

			var err error
			done := startWrite(t, func() { err = c.write(f, uc) })
			select {
			case <-done:
				t.Logf("the write finished while the build held s03")
			default:
				t.Logf("the write waited for the build")
			}
			h.free()
			select {
			case <-done:
			case <-time.After(patience):
				t.Fatalf("the write did not finish after the build went on")
			}
			if err != nil {
				t.Fatalf("the write failed: %v", err)
			}
			f.awaitBuilt(uc)
			all, _ := summariesOf(uc, f.alice, FindAllOptions{CreatedBy: "alice"})
			c.check(t, all)
			f.same(c.name, uc)
			f.same(c.name+", after a restart", f.restart())
		})
	}
}

// TestHeads_TwoUseCasesOnOneStore (R-DEC-ASSISTANT-SESSION-HEADS): two use cases on the same stores write the same
// session. The interleaving is forced: the first holds right after it has written the conversation, before its head;
// the second renames the session until it has written it (no lock shared between the two) or until it waits (a
// shared lock); then the first goes on. Afterwards the head carries the title of the conversation, for both use
// cases and after a restart, and no mark is left.
func TestHeads_TwoUseCasesOnOneStore(t *testing.T) {
	f := newHFix(t, false)
	first := f.start(false)
	x := create(t, first, f.alice, ctxA, "")

	h := newHold()
	t.Cleanup(h.free)
	var once sync.Once
	f.repo.mu.Lock()
	f.repo.afterWrite = func(op string, s Session, id ID) {
		if op == "save" && id == x.ID && s.Title == "one" {
			once.Do(func() {
				close(h.reached)
				<-h.release
			})
		}
	}
	f.repo.mu.Unlock()
	second := NewUseCases(Repository(json.NewSloppyJSONRepository[Session, ID](f.content)), f.rdb, Heads(f.ctx, f.heads))

	firstDone := make(chan error, 1)
	go func() { firstDone <- first.Rename(f.alice, x.ID, "one") }()
	h.wait(t, "the write of the conversation by the first use cases")
	var err2 error
	done := startWrite(t, func() { err2 = second.Rename(f.alice, x.ID, "two") })
	h.free()
	if err := <-firstDone; err != nil {
		t.Fatal(err)
	}
	<-done
	if err2 != nil {
		t.Fatal(err2)
	}

	o, _ := f.bare.FindByID(x.ID)
	if o.IsNone() || o.Unwrap().Title != "two" {
		t.Fatalf("premise: the conversation is not the one of the later rename: %+v", o)
	}
	for _, c := range []struct {
		name string
		uc   UseCases
	}{{"first", first}, {"second", second}, {"after a restart", f.restart()}} {
		all, _ := summariesOf(c.uc, f.alice, FindAllOptions{CreatedBy: "alice"})
		if len(all) != 1 || all[0].Title != "two" {
			t.Errorf("%s: the head says %+v, the conversation %q", c.name, all, "two")
		}
		f.same(c.name, c.uc)
	}
	if m := keys(t, f.heads, "o/"); len(m) > 0 {
		t.Errorf("marks are left: %q", m)
	}
}

// TestHeads_BuildResumesAtTheCursor (R-DEC-ASSISTANT-SESSION-HEADS): a build which is aborted after 550 of 600
// conversations resumes at the next start at its cursor: it reads none of the first 499 again, and afterwards every
// search equals reading every session. Counter row: a build aborted after 300, before its first cursor, reads the
// first conversation again, so that the probe sees what the next build reads.
func TestHeads_BuildResumesAtTheCursor(t *testing.T) {
	for _, c := range []struct {
		name     string
		abortAt  int32
		resumed  bool
		premised int // heads of the first conversations which the first build must have written
	}{
		{"aborted after 550", 550, true, 500},
		{"counter row: aborted after 300", 300, false, 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := newHFix(t, false)
			var ids []ID
			for i := range 600 {
				id := ID(fmt.Sprintf("s%04d", i))
				ids = append(ids, id)
				by := []user.ID{"alice", "bob"}[i%2]
				if err := f.bare.Save(legacy(id, by, xtime.UnixMilliseconds(1_000+i), i%7 == 0)); err != nil {
					t.Fatal(err)
				}
			}

			var n atomic.Int32
			aborted := make(chan uint64, 1)
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			f.repo.setOnRead(func(id ID) error {
				if goid() != f.testG && n.Add(1) == c.abortAt {
					cancel()
					aborted <- goid()
				}
				return nil
			})
			NewUseCases(f.repo, f.rdb, Heads(ctx, f.heads))
			select {
			case g := <-aborted:
				awaitGone(t, g, "the aborted build")
			case <-time.After(patience):
				t.Fatalf("premise: the build did not read %d conversations", c.abortAt)
			}
			for _, id := range ids[:c.premised] {
				if !hasHeadOf(t, f.heads, id) {
					t.Fatalf("premise: the first build left %s without head", id)
				}
			}

			uc := f.restart()
			f.awaitBuilt(uc)
			again := byOthers(f.repo.since(0), f.testG)
			if c.resumed {
				for _, id := range ids[:499] {
					if slices.Contains(again, id) {
						t.Errorf("the next build read %s again, below its cursor", id)
						break
					}
				}
			} else if !slices.Contains(again, ids[0]) {
				t.Errorf("counter row: the next build did not read %s again: %d reads", ids[0], len(again))
			}
			for _, id := range ids {
				if !hasHeadOf(t, f.heads, id) {
					t.Errorf("%s has no head after the build", id)
				}
			}
			f.sameFor(c.name, uc, user.SU(), f.alice)
		})
	}
}

// logRecorder keeps every record of the default logger.
type logRecorder struct {
	mu    sync.Mutex
	lines []string
}

type recordingHandler struct {
	rec   *logRecorder
	attrs []slog.Attr
}

func (h recordingHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h recordingHandler) Handle(_ context.Context, r slog.Record) error {
	var b strings.Builder
	b.WriteString(r.Message)
	for _, a := range h.attrs {
		b.WriteString(" " + a.String())
	}
	r.Attrs(func(a slog.Attr) bool {
		b.WriteString(" " + a.String())
		return true
	})
	h.rec.mu.Lock()
	defer h.rec.mu.Unlock()
	h.rec.lines = append(h.rec.lines, b.String())
	return nil
}

func (h recordingHandler) WithAttrs(as []slog.Attr) slog.Handler {
	return recordingHandler{rec: h.rec, attrs: append(slices.Clone(h.attrs), as...)}
}

func (h recordingHandler) WithGroup(string) slog.Handler { return h }

func recordLogs(t *testing.T) *logRecorder {
	rec := &logRecorder{}
	prev := slog.Default()
	slog.SetDefault(slog.New(recordingHandler{rec: rec}))
	t.Cleanup(func() {
		slog.SetDefault(prev)
		if t.Failed() {
			rec.mu.Lock()
			defer rec.mu.Unlock()
			for _, l := range rec.lines {
				t.Log("log: " + l)
			}
		}
	})
	return rec
}

func (r *logRecorder) mentions(s string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.ContainsFunc(r.lines, func(l string) bool { return strings.Contains(l, s) })
}

// TestHeads_BuildPassesOverAnUndecodableSession (R-DEC-ASSISTANT-SESSION-HEADS, row „ein Gespräch nicht
// dekodierbar“): the build passes over a conversation which cannot be decoded, because it is no JSON or carries a
// kind of content nago does not know, e.g. of a newer version, logs its id and finishes; every other conversation has
// its head, and the searches equal reading every session except for the undecodable one. Counter row: the same
// conversation decodable gets its head.
func TestHeads_BuildPassesOverAnUndecodableSession(t *testing.T) {
	for _, c := range []struct {
		name string
		raw  string // the stored conversation s03, decodable if empty
	}{
		{"s03 is no JSON", `{"id":"s03","createdBy":"alice","title":`},
		{"s03 has an unknown kind of content", `{"id":"s03","createdBy":"alice","messages":[{"role":"user","content":[{"type":"no-such-kind"}]}]}`},
		{"counter row: s03 can be decoded", ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			bad := c.raw != ""
			f := newHFix(t, false)
			for _, id := range []ID{"s01", "s02", "s04", "s05"} {
				f.stored(legacy(id, "alice", 1_000, false))
			}
			if bad {
				f.content.Store("s03", []byte(c.raw))
				if _, err := f.bare.FindByID("s03"); err == nil {
					t.Fatalf("premise: s03 can be decoded")
				}
				grant(t, f.rdb, "alice", "s03", InstancePermissions...)
			} else {
				f.stored(legacy("s03", "alice", 1_000, false))
			}
			logs := recordLogs(t)
			uc := f.start(false)
			f.awaitBuilt(uc)

			for _, id := range []ID{"s01", "s02", "s04", "s05"} {
				if !hasHeadOf(t, f.heads, id) {
					t.Errorf("%s has no head", id)
				}
			}
			if got := hasHeadOf(t, f.heads, "s03"); got == bad {
				t.Errorf("s03 has a head: %v, expected %v", got, !bad)
			}
			if bad && !logs.mentions("s03") {
				t.Errorf("no log record names s03")
			}
			f.same(c.name, uc)
		})
	}
}

// buildToS04 lets a first build write the heads of s01 to s03 of six conversations of alice and stops it at its read
// of s04, before its first cursor; it returns once the build has ended.
func buildToS04(t *testing.T) *hfix {
	t.Helper()
	f := newHFix(t, false)
	for i := 1; i <= 6; i++ {
		f.stored(legacy(ID(fmt.Sprintf("s%02d", i)), "alice", xtime.UnixMilliseconds(1_000+i), i == 2))
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	stopped := make(chan uint64, 1)
	var once sync.Once
	f.repo.setOnRead(func(id ID) error {
		if id == "s04" && goid() != f.testG {
			once.Do(func() {
				cancel()
				stopped <- goid()
			})
		}
		return nil
	})
	NewUseCases(f.repo, f.rdb, Heads(ctx, f.heads))
	select {
	case g := <-stopped:
		awaitGone(t, g, "the first build")
	case <-time.After(patience):
		t.Fatalf("premise: the first build did not read s04")
	}
	for _, id := range []ID{"s01", "s02", "s03"} {
		if !hasHeadOf(t, f.heads, id) {
			t.Fatalf("premise: the first build left %s without head", id)
		}
	}
	return f
}

// idleOrIn reports whether the goroutine g has ended or waits outside of the probe, and whether it waits inside the
// probe, i.e. in a read after the one which failed.
func idleOrIn(g uint64) (idle, inProbe bool) {
	for _, s := range strings.Split(allStacks(), "\n\n") {
		if !strings.HasPrefix(s, fmt.Sprintf("goroutine %d [", g)) {
			continue
		}
		if strings.Contains(s, "session.(*contentRepo).read") {
			return false, true
		}
		head, _, _ := strings.Cut(s, "\n")
		_, state, _ := strings.Cut(head, "[")
		for _, p := range []string{"sync.", "semacquire", "chan ", "select", "sleep"} {
			if strings.HasPrefix(state, p) {
				return true, false
			}
		}
		return false, false
	}
	return true, false
}

// TestHeads_ReadErrorAbortsTheBuild (R-DEC-ASSISTANT-SESSION-HEADS, row „Lesefehler während des Aufbaus“): a build
// which cannot read s02 stops there, reads no further conversation and takes no head away, neither of s02 nor of s01
// and s03, which a former build wrote; until the next start the searches read the conversations and equal reading
// every session, and the next start finishes the build. Counter row: without the error, the same build finishes.
func TestHeads_ReadErrorAbortsTheBuild(t *testing.T) {
	t.Run("counter row: without an error", func(t *testing.T) {
		f := buildToS04(t)
		uc := f.restart()
		f.awaitBuilt(uc)
		for i := 1; i <= 6; i++ {
			if id := ID(fmt.Sprintf("s%02d", i)); !hasHeadOf(t, f.heads, id) {
				t.Errorf("%s has no head", id)
			}
		}
		f.same("built", uc)
	})

	t.Run("s02 cannot be read", func(t *testing.T) {
		f := buildToS04(t)
		f.repo = newContentRepo(f.bare)
		// the read of s02 by the build fails; every later read of the same goroutine waits here, so that a build
		// which goes on is seen, not raced
		var g atomic.Uint64
		var continued atomic.Bool
		failed := make(chan struct{})
		release := make(chan struct{})
		t.Cleanup(func() { close(release) })
		f.repo.setOnRead(func(id ID) error {
			me := goid()
			if me == f.testG {
				return nil
			}
			if gg := g.Load(); gg != 0 {
				if me == gg {
					continued.Store(true)
					<-release
				}
				return nil
			}
			if id == "s02" && g.CompareAndSwap(0, me) {
				close(failed)
				return errReadProbe
			}
			return nil
		})
		uc := f.start(false)
		select {
		case <-failed:
		case <-time.After(patience):
			t.Fatalf("premise: the build did not read s02")
		}
		eventually(t, "the end of the build or its next read", func() bool {
			idle, in := idleOrIn(g.Load())
			return idle || in
		})
		if continued.Load() {
			t.Fatalf("the build went on after it could not read s02")
		}
		for _, id := range []ID{"s01", "s02", "s03"} {
			if !hasHeadOf(t, f.heads, id) {
				t.Errorf("the head of %s is gone after the read error", id)
			}
		}
		f.same("after the read error", uc)
		nobody := newRebacSubject("nobody", f.rdb)
		m := f.repo.mark()
		summariesOf(uc, nobody, FindAllOptions{CreatedBy: "nobody"})
		if len(f.repo.since(m)) == 0 {
			t.Errorf("after the read error the build counts as finished: a search reads no conversation")
		}

		next := f.restart()
		f.awaitBuilt(next)
		for i := 1; i <= 6; i++ {
			if id := ID(fmt.Sprintf("s%02d", i)); !hasHeadOf(t, f.heads, id) {
				t.Errorf("after the next start, %s has no head", id)
			}
		}
		f.same("after the next start", next)
	})
}

// TestHeads_BuildRemovesHeadsWithoutContent (R-DEC-ASSISTANT-SESSION-HEADS): a build which finds a head whose
// conversation has been deleted past the use cases removes it at its end. Counter row: the head of s01, whose
// conversation is there, stays.
func TestHeads_BuildRemovesHeadsWithoutContent(t *testing.T) {
	f := buildToS04(t)
	if err := f.bare.DeleteByID("s02"); err != nil {
		t.Fatal(err)
	}
	uc := f.restart()
	f.awaitBuilt(uc)
	if hasHeadOf(t, f.heads, "s02") {
		t.Errorf("the head of the deleted s02 is left after the build")
	}
	if !hasHeadOf(t, f.heads, "s01") {
		t.Errorf("counter row: the head of s01 is gone")
	}
	f.same("after the build", uc)
}

// TestHeads_LockLeavesNothingBehindWithHeads (R-DEC-ASSISTANT-SESSION-HEADS, row „1 000 Umbenennungen verschiedener
// Gespräche“): with Heads, nago's lock per session is empty after 1 000 renames of different sessions; counter row:
// during a rename it has one entry. See TestLocker_LeavesNothingBehind for the use cases without Heads.
func TestHeads_LockLeavesNothingBehindWithHeads(t *testing.T) {
	f := newHFix(t, false)
	for i := range 1000 {
		if err := f.bare.Save(legacy(ID(fmt.Sprintf("s%04d", i)), "alice", 1_000, false)); err != nil {
			t.Fatal(err)
		}
	}
	uc := f.start(false)
	f.awaitBuilt(uc)
	lockBacklog(t, uc, 1000, func(id ID) (<-chan struct{}, func()) {
		h := newHold()
		var once sync.Once
		f.repo.mu.Lock()
		defer f.repo.mu.Unlock()
		f.repo.afterWrite = func(op string, s Session, got ID) {
			if op == "save" && got == id {
				once.Do(func() {
					close(h.reached)
					<-h.release
				})
			}
		}
		return h.reached, h.free
	})
}

// TestHeads_RestartWithoutReconciliation (R-DEC-ASSISTANT-SESSION-HEADS, „ohne Abgleich der Köpfe mit den Inhalten
// bei jedem Start“): a start on built heads reads no conversation: neither the start itself nor its first searches.
// Counter row: a start on a store of a former version reads them, because it builds.
func TestHeads_RestartWithoutReconciliation(t *testing.T) {
	for _, c := range []struct {
		name  string
		built bool
	}{{"heads built", true}, {"counter row: a store of a former version", false}} {
		t.Run(c.name, func(t *testing.T) {
			f := newHFix(t, false)
			f.stored(legacy("s1", "bob", 1_000, true), legacy("s2", "bob", 1_001, false))
			if c.built {
				f.awaitBuilt(f.start(false))
				f.cancel()
			}
			uc := f.restart()
			summariesOf(uc, f.alice, FindAllOptions{CreatedBy: "alice"})
			lookup(t, uc, f.alice, ctxA)
			if n := len(f.repo.since(0)); (n == 0) != c.built {
				t.Errorf("the start and the first searches of alice read %d conversations", n)
			}
		})
	}
}

// keep the imports of the probes used by the tests above
var _ = runtime.Gosched

// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

// Acceptance tests of the index of the waiting sessions behind [UseCases.FindResumable]. They prove the wokoda
// decision R-DEC-ASSISTANT-RESUME-INDEX: the lookup reads every session once at its first call since the start, then
// follows every create, append, resolve, dismiss, rename and delete, and its result equals reading every session
// (readAll below): the newest waiting session without parent, created by the subject, with the same provider hint and
// exactly the same tags.
//
// The tests observe through the use cases and a store under them (probeRepo), never through the index itself.

import (
	"errors"
	"iter"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/xtime"
)

// readAll is the reference: the session a chat of the subject continues, found by reading every session through
// FindAll, as the chat did before the index (application/ai/completion/ui/chat.go before d7dd885c7): the newest one,
// and among equally new ones the first in the order of the repository. The rule is written out from the decision,
// not taken from [Session.ResumableBy].
func readAll(t *testing.T, uc UseCases, subject auth.Subject, opts FindResumableOptions) ID {
	t.Helper()
	if subject == nil || !subject.Valid() {
		return ""
	}

	var best *Session
	for s, err := range uc.FindAll(subject, FindAllOptions{}) {
		if err != nil {
			t.Fatalf("reference: FindAll: %v", err)
		}
		if s.Pending == nil || s.ParentID != "" || s.CreatedBy != subject.ID() || s.ProviderHint != opts.ProviderHint || !sameTagsAs(s.Tags, opts.Tags) {
			continue
		}
		if best == nil || s.UpdatedAt > best.UpdatedAt {
			best = &s
		}
	}

	if best == nil {
		return ""
	}
	return best.ID
}

func sameTagsAs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, x := range b {
		if !slices.Contains(a, x) {
			return false
		}
	}
	return true
}

// lookup is the id FindResumable returns, "" for none.
func lookup(t *testing.T, uc UseCases, subject auth.Subject, opts FindResumableOptions) ID {
	t.Helper()
	found, err := uc.FindResumable(subject, opts)
	if err != nil {
		t.Fatalf("FindResumable: %v", err)
	}
	if found.IsNone() {
		return ""
	}
	return found.Unwrap().ID
}

// probeRepo is the store under the use cases. It counts the full reads, can let the next matching write fail (before
// or after it has been written, as a store which cannot tell) and can hold the next full read after its last session
// until the test releases it.
type probeRepo struct {
	Repository

	mu       sync.Mutex
	failNext func(op string, s Session, id ID) (fail, written bool) // one shot, see failOnce
	afterOp  func(op string, s Session, id ID)                      // after a write has been written
	hold     *readHold                                              // one shot, see holdNextRead

	fullReads atomic.Int32
}

var errProbe = errors.New("probe: the store failed")

type readHold struct {
	reached chan struct{}
	release chan struct{}
}

func newProbeRepo() *probeRepo {
	return &probeRepo{Repository: Repository(json.NewSloppyJSONRepository[Session, ID](mem.NewBlobStore(string(Namespace))))}
}

// failOnce lets the next write matching fn fail.
func (r *probeRepo) failOnce(fn func(op string, s Session, id ID) (fail, written bool)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.failNext = fn
}

// holdNextRead holds the next full read after its last session until release is closed.
func (r *probeRepo) holdNextRead() *readHold {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hold = &readHold{reached: make(chan struct{}), release: make(chan struct{})}
	return r.hold
}

func (r *probeRepo) takeHold() *readHold {
	r.mu.Lock()
	defer r.mu.Unlock()
	h := r.hold
	r.hold = nil
	return h
}

// write runs do as the write op of s or id, failing it as failOnce asked.
func (r *probeRepo) write(op string, s Session, id ID, do func() error) error {
	r.mu.Lock()
	fn := r.failNext
	fail, written := false, false
	if fn != nil {
		if fail, written = fn(op, s, id); fail {
			r.failNext = nil
		}
	}
	after := r.afterOp
	r.mu.Unlock()

	if fail && !written {
		return errProbe
	}
	if err := do(); err != nil {
		return err
	}
	if after != nil {
		after(op, s, id)
	}
	if fail {
		return errProbe
	}
	return nil
}

func (r *probeRepo) Save(s Session) error {
	return r.write("save", s, s.ID, func() error { return r.Repository.Save(s) })
}

func (r *probeRepo) SaveAll(it iter.Seq[Session]) error {
	for s := range it {
		if err := r.Save(s); err != nil {
			return err
		}
	}
	return nil
}

func (r *probeRepo) DeleteByID(id ID) error {
	return r.write("delete", Session{}, id, func() error { return r.Repository.DeleteByID(id) })
}

func (r *probeRepo) DeleteByEntity(s Session) error {
	return r.DeleteByID(s.ID)
}

func (r *probeRepo) DeleteAllByID(ids iter.Seq[ID]) error {
	for id := range ids {
		if err := r.DeleteByID(id); err != nil {
			return err
		}
	}
	return nil
}

func (r *probeRepo) Delete(predicate func(Session) (bool, error)) error {
	return r.write("delete-by-predicate", Session{}, "", func() error { return r.Repository.Delete(predicate) })
}

// held yields all of seq, then waits on a hold taken for this read.
func held[T any](r *probeRepo, seq iter.Seq2[T, error]) iter.Seq2[T, error] {
	r.fullReads.Add(1)
	h := r.takeHold()
	return func(yield func(T, error) bool) {
		// read completely first, so that no lock of the store is held while the read waits
		type item struct {
			v   T
			err error
		}
		var items []item
		for v, err := range seq {
			items = append(items, item{v, err})
		}
		for _, it := range items {
			if !yield(it.v, it.err) {
				return
			}
		}
		if h != nil {
			close(h.reached)
			<-h.release
		}
	}
}

func (r *probeRepo) All() iter.Seq2[Session, error] { return held(r, r.Repository.All()) }

func (r *probeRepo) Identifiers() iter.Seq2[ID, error] { return held(r, r.Repository.Identifiers()) }

func (r *probeRepo) FindAllByPrefix(prefix ID) iter.Seq2[Session, error] {
	return held(r, r.Repository.FindAllByPrefix(prefix))
}

func (r *probeRepo) IdentifiersByPrefix(prefix ID) iter.Seq2[ID, error] {
	return held(r, r.Repository.IdentifiersByPrefix(prefix))
}

// invalidSubject is a subject which is not signed in.
type invalidSubject struct{ rebacSubject }

func (invalidSubject) Valid() bool { return false }

// fixture is a store, its ReBAC database, the use cases on both and two persons.
type fixture struct {
	repo       *probeRepo
	rdb        *rebac.DB
	uc         UseCases
	alice, bob rebacSubject
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{repo: newProbeRepo(), rdb: newTestRDB(t)}
	f.uc = NewUseCases(f.repo, f.rdb)
	f.alice = newRebacSubject("alice", f.rdb, PermCreate)
	f.bob = newRebacSubject("bob", f.rdb, PermCreate)
	return f
}

var (
	ctxA     = FindResumableOptions{Tags: []string{"ctx:a"}, ProviderHint: "p"}
	ctxMore  = FindResumableOptions{Tags: []string{"ctx:a", "more"}, ProviderHint: "p"}
	ctxOther = FindResumableOptions{Tags: []string{"ctx:a"}, ProviderHint: "q"}
)

// create creates a session of the subject in the context, as child of parent unless it is empty.
func (f *fixture) create(t *testing.T, s auth.Subject, c FindResumableOptions, parent ID) Session {
	t.Helper()
	x, err := f.uc.Create(s, CreateOptions{Model: "fake-model", Tags: c.Tags, ProviderHint: c.ProviderHint, ParentID: parent})
	if err != nil {
		t.Fatalf("setup: create: %v", err)
	}
	return x
}

// ask lets the session wait on a question.
func (f *fixture) ask(t *testing.T, s auth.Subject, id ID) (Session, error) {
	t.Helper()
	return f.uc.Append(s, id, AppendOptions{
		Completions: &scriptedCompletions{results: []completion.Result{askResult()}},
		Input:       userText("go"),
		Tools:       []completion.Tool{completion.NewAskUserTool()},
	})
}

// answer appends a turn which ends with an answer, without waiting.
func (f *fixture) answer(t *testing.T, s auth.Subject, id ID) Session {
	t.Helper()
	x, err := f.uc.Append(s, id, AppendOptions{Completions: &scriptedCompletions{results: []completion.Result{answerResult("ok")}}, Input: userText("hi")})
	if err != nil {
		t.Fatalf("setup: answer: %v", err)
	}
	return x
}

// waiting creates a session of the subject in the context which waits on a question.
func (f *fixture) waiting(t *testing.T, s auth.Subject, c FindResumableOptions) Session {
	t.Helper()
	x := f.create(t, s, c, "")
	x, err := f.ask(t, s, x.ID)
	if err != nil || x.Pending == nil {
		t.Fatalf("setup: the session does not wait: %v", err)
	}
	return x
}

// resolve answers the question of a waiting session; the run ends with an answer.
func (f *fixture) resolve(t *testing.T, s auth.Subject, x Session) {
	t.Helper()
	_, err := f.uc.Resolve(s, x.ID, ResolveOptions{
		Run:         AppendOptions{Completions: &scriptedCompletions{results: []completion.Result{answerResult("thanks")}}, Tools: []completion.Tool{completion.NewAskUserTool()}},
		Revision:    x.PendingRevision,
		Resolutions: []completion.Resolution{{CallID: "q1", Answer: "the red one"}},
	})
	if err != nil {
		t.Fatalf("setup: resolve: %v", err)
	}
}

// pair is a person in a context.
type pair struct {
	name    string
	subject auth.Subject
	opts    FindResumableOptions
}

// sameAsReadAll compares FindResumable with the reference for every pair.
func sameAsReadAll(t *testing.T, step string, uc UseCases, pairs []pair) {
	t.Helper()
	for _, p := range pairs {
		if got, want := lookup(t, uc, p.subject, p.opts), readAll(t, uc, p.subject, p.opts); got != want {
			t.Errorf("%s, %s: FindResumable %q, reading every session %q", step, p.name, got, want)
		}
	}
}

// TestFindResumable_FollowsEveryWrite (R-DEC-ASSISTANT-RESUME-INDEX): after every step of a sequence of create, wait,
// newer session, rename, resolve, dismiss and delete, and after a restart, FindResumable equals reading every session,
// for two persons in three contexts (other tags, other provider). Each step also names the session it expects for
// alice in ctx:a, so that the reference itself cannot drift into nothing. The index is built by the first lookup on
// the empty store, so that every later session reaches it only through a write.
func TestFindResumable_FollowsEveryWrite(t *testing.T) {
	f := newFixture(t)
	var pairs []pair
	for _, s := range []struct {
		name string
		s    auth.Subject
	}{{"alice", f.alice}, {"bob", f.bob}} {
		for _, c := range []struct {
			name string
			o    FindResumableOptions
		}{{"ctx:a", ctxA}, {"ctx:a+more", ctxMore}, {"other provider", ctxOther}} {
			pairs = append(pairs, pair{s.name + " in " + c.name, s.s, c.o})
		}
	}

	if got := lookup(t, f.uc, f.alice, ctxA); got != "" {
		t.Fatalf("premise: empty store, found %q", got)
	}

	var a, b, d, e Session
	steps := []struct {
		name string
		do   func(t *testing.T)
		want func() []ID // what alice continues in ctx:a; more than one where a tie of the clock decides
	}{
		{"create without waiting", func(t *testing.T) { a = f.create(t, f.alice, ctxA, ""); a = f.answer(t, f.alice, a.ID) }, func() []ID { return []ID{""} }},
		{"wait", func(t *testing.T) {
			var err error
			if a, err = f.ask(t, f.alice, a.ID); err != nil || a.Pending == nil {
				t.Fatalf("setup: a does not wait: %v", err)
			}
		}, func() []ID { return []ID{a.ID} }},
		{"newer waiting session", func(t *testing.T) { b = f.waiting(t, f.alice, ctxA) }, func() []ID { return []ID{b.ID, a.ID} }},
		{"waiting session of another person", func(t *testing.T) { f.waiting(t, f.bob, ctxA) }, func() []ID { return []ID{b.ID, a.ID} }},
		{"rename the older one", func(t *testing.T) {
			if err := f.uc.Rename(f.alice, a.ID, "renamed"); err != nil {
				t.Fatalf("rename: %v", err)
			}
		}, func() []ID { return []ID{a.ID, b.ID} }},
		{"resolve it", func(t *testing.T) {
			a = f.mustFind(t, a.ID)
			f.resolve(t, f.alice, a)
		}, func() []ID { return []ID{b.ID} }},
		{"dismiss the other", func(t *testing.T) {
			if _, err := f.uc.Dismiss(f.alice, b.ID, b.PendingRevision); err != nil {
				t.Fatalf("dismiss: %v", err)
			}
		}, func() []ID { return []ID{""} }},
		{"waiting session with more tags", func(t *testing.T) { d = f.waiting(t, f.alice, ctxMore) }, func() []ID { return []ID{""} }},
		{"delete it", func(t *testing.T) {
			if err := f.uc.Delete(f.alice, d.ID); err != nil {
				t.Fatalf("delete: %v", err)
			}
		}, func() []ID { return []ID{""} }},
		{"waiting child newer than its waiting parent", func(t *testing.T) {
			e = f.waiting(t, f.alice, ctxA)
			child := f.create(t, f.alice, ctxA, e.ID)
			if x, err := f.ask(t, f.alice, child.ID); err != nil || x.Pending == nil || x.ParentID != e.ID {
				t.Fatalf("setup: the child does not wait: %v", err)
			}
		}, func() []ID { return []ID{e.ID} }},
		{"delete the parent with its child", func(t *testing.T) {
			if err := f.uc.Delete(f.alice, e.ID); err != nil {
				t.Fatalf("delete: %v", err)
			}
		}, func() []ID { return []ID{""} }},
		{"a waiting session again", func(t *testing.T) { e = f.waiting(t, f.alice, ctxA) }, func() []ID { return []ID{e.ID} }},
	}
	for _, st := range steps {
		st.do(t)
		got := lookup(t, f.uc, f.alice, ctxA)
		if !slices.Contains(st.want(), got) {
			t.Errorf("%s: alice continues %q, expected one of %q", st.name, got, st.want())
		}
		sameAsReadAll(t, st.name, f.uc, pairs)
	}

	// a restart: new use cases on the same store read it again
	restarted := NewUseCases(f.repo, f.rdb)
	if got := lookup(t, restarted, f.alice, ctxA); got != e.ID {
		t.Errorf("after a restart: alice continues %q, expected %q", got, e.ID)
	}
	sameAsReadAll(t, "after a restart", restarted, pairs)
}

func (f *fixture) mustFind(t *testing.T, id ID) Session {
	t.Helper()
	o, err := f.repo.FindByID(id)
	if err != nil || o.IsNone() {
		t.Fatalf("setup: session %s: %v", id, err)
	}
	return o.Unwrap()
}

// TestFindResumable_ReadsEverySessionOnlyOnce (R-DEC-ASSISTANT-RESUME-INDEX): the first lookup since the start reads
// the store once, every later lookup does not read it in full again, also after writes of the use cases. Counter
// row: the reference, which reads every session, reads the store once per call; the probe sees full reads.
func TestFindResumable_ReadsEverySessionOnlyOnce(t *testing.T) {
	f := newFixture(t)
	x := f.waiting(t, f.alice, ctxA)
	f.waiting(t, f.bob, ctxA)

	reads := func(fn func()) int32 {
		before := f.repo.fullReads.Load()
		fn()
		return f.repo.fullReads.Load() - before
	}

	if n := reads(func() { lookup(t, f.uc, f.alice, ctxA) }); n != 1 {
		t.Errorf("first lookup: %d full reads, expected 1", n)
	}
	for i, write := range []func(){
		func() {},
		func() { _ = f.uc.Rename(f.alice, x.ID, "renamed") },
		func() { f.waiting(t, f.alice, ctxMore) },
	} {
		write()
		if n := reads(func() { lookup(t, f.uc, f.alice, ctxA) }); n != 0 {
			t.Errorf("lookup %d after the first: %d full reads, expected none", i+2, n)
		}
	}
	if n := reads(func() { readAll(t, f.uc, f.alice, ctxA) }); n != 1 {
		t.Errorf("counter row: reading every session: %d full reads, expected 1 (the probe does not see full reads)", n)
	}
}

// TestFindResumable_EquallyNewKeepsTheOrder (R-DEC-ASSISTANT-RESUME-INDEX): sessions with the same UpdatedAt, stored
// before the start (thus read by the first lookup), are decided like before: the first in the order of the store,
// whatever the order they were written in. Counter row: one newer by a millisecond wins. A waiting child, newer than
// everything, never; its counter row is the same session without parent.
func TestFindResumable_EquallyNewKeepsTheOrder(t *testing.T) {
	const at = xtime.UnixMilliseconds(1_000)
	pending := &completion.Continuation{Pending: []completion.PendingCall{{Kind: completion.PendingQuestion, Call: completion.ToolCall{ID: "q1"}}}}
	stored := func(id ID, updated xtime.UnixMilliseconds, parent ID) Session {
		return Session{ID: id, CreatedBy: "alice", ProviderHint: ctxA.ProviderHint, Tags: ctxA.Tags, Pending: pending, PendingRevision: 1, UpdatedAt: updated, ParentID: parent}
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
			repo, rdb := newProbeRepo(), newTestRDB(t)
			for _, s := range c.sessions {
				if err := repo.Repository.Save(s); err != nil {
					t.Fatal(err)
				}
				grant(t, rdb, "alice", s.ID, InstancePermissions...)
			}
			uc := NewUseCases(repo, rdb)
			alice := newRebacSubject("alice", rdb)
			got, ref := lookup(t, uc, alice, ctxA), readAll(t, uc, alice, ctxA)
			if got != c.want || ref != c.want {
				t.Errorf("FindResumable %q, reading every session %q, expected %q", got, ref, c.want)
			}
		})
	}
}

// TestFindResumable_ReBACAndSubjects (R-DEC-ASSISTANT-RESUME-INDEX): FindResumable returns only what the subject may
// read like through FindAll: with the grants of alice on her newest waiting session withdrawn, it is not found, the
// older one is (counter row: before the withdrawal the newest); with both withdrawn, none. A subject which is not
// signed in finds nothing (counter row: alice signed in). The system subject, which may read every session, finds
// only the sessions it created itself (counter row: its own one), never the one of alice.
func TestFindResumable_ReBACAndSubjects(t *testing.T) {
	f := newFixture(t)
	older := f.waiting(t, f.alice, ctxA)
	newer := f.waiting(t, f.alice, ctxA)
	if lookup(t, f.uc, f.alice, ctxA) == "" {
		t.Fatalf("premise: alice finds a waiting session")
	}
	withdraw := func(id ID) {
		if err := f.rdb.DeleteByQuery(rebac.Select().Where().Target().Is(Namespace, rebacInstance(id))); err != nil {
			t.Fatal(err)
		}
	}

	type row struct {
		name    string
		subject auth.Subject
		want    ID
	}
	check := func(rows ...row) {
		t.Helper()
		for _, r := range rows {
			got, ref := lookup(t, f.uc, r.subject, ctxA), readAll(t, f.uc, r.subject, ctxA)
			if got != r.want || ref != r.want {
				t.Errorf("%s: FindResumable %q, reading every session %q, expected %q", r.name, got, ref, r.want)
			}
		}
	}

	newestID := newer.ID
	if older.UpdatedAt == newer.UpdatedAt && older.ID < newer.ID {
		newestID = older.ID // equally new: the first in the order of the store
	}
	otherID := older.ID
	if newestID == older.ID {
		otherID = newer.ID
	}

	check(
		row{"counter row: alice before the withdrawal", f.alice, newestID},
		row{"not signed in", invalidSubject{f.alice}, ""},
		row{"system subject without own session", user.SU(), ""},
	)
	withdraw(newestID)
	check(row{"alice without grants on the newest", f.alice, otherID})
	withdraw(otherID)
	check(row{"alice without grants on both", f.alice, ""})

	own := f.waiting(t, user.SU(), ctxA)
	check(
		row{"counter row: system subject with its own session", user.SU(), own.ID},
		row{"alice does not find the session of the system subject", f.alice, ""},
	)
}

// TestFindResumable_RightAfterAFailedWrite (R-DEC-ASSISTANT-RESUME-INDEX): a write which fails, before or after the
// store has written it, does not leave the lookup wrong: it equals reading every session at once and after the next
// write. Each failing row stands next to the same write failing on the other side of the store.
func TestFindResumable_RightAfterAFailedWrite(t *testing.T) {
	for _, c := range []struct {
		name string
		run  func(t *testing.T, f *fixture) (expected ID)
	}{
		{"dismiss fails after it has been written", func(t *testing.T, f *fixture) ID {
			x := f.waiting(t, f.alice, ctxA)
			lookup(t, f.uc, f.alice, ctxA)
			f.repo.failOnce(func(op string, s Session, id ID) (bool, bool) { return op == "save" && id == x.ID, true })
			_, err := f.uc.Dismiss(f.alice, x.ID, x.PendingRevision)
			if err == nil {
				t.Fatalf("premise: the dismiss did not fail")
			}
			return ""
		}},
		{"counter row: dismiss fails before it has been written", func(t *testing.T, f *fixture) ID {
			x := f.waiting(t, f.alice, ctxA)
			lookup(t, f.uc, f.alice, ctxA)
			f.repo.failOnce(func(op string, s Session, id ID) (bool, bool) { return op == "save" && id == x.ID, false })
			_, err := f.uc.Dismiss(f.alice, x.ID, x.PendingRevision)
			if err == nil {
				t.Fatalf("premise: the dismiss did not fail")
			}
			return x.ID
		}},
		{"wait fails after it has been written", func(t *testing.T, f *fixture) ID {
			x := f.create(t, f.alice, ctxA, "")
			lookup(t, f.uc, f.alice, ctxA)
			f.repo.failOnce(func(op string, s Session, id ID) (bool, bool) {
				return op == "save" && id == x.ID && s.Pending != nil, true
			})
			_, _ = f.ask(t, f.alice, x.ID)
			if f.mustFind(t, x.ID).Pending == nil {
				t.Fatalf("premise: the waiting session has not been written")
			}
			return x.ID
		}},
		{"counter row: wait fails before it has been written", func(t *testing.T, f *fixture) ID {
			x := f.create(t, f.alice, ctxA, "")
			lookup(t, f.uc, f.alice, ctxA)
			f.repo.failOnce(func(op string, s Session, id ID) (bool, bool) {
				return op == "save" && id == x.ID && s.Pending != nil, false
			})
			_, _ = f.ask(t, f.alice, x.ID)
			if f.mustFind(t, x.ID).Pending != nil {
				t.Fatalf("premise: the waiting session has been written")
			}
			return ""
		}},
		{"delete fails after it has been deleted", func(t *testing.T, f *fixture) ID {
			x := f.waiting(t, f.alice, ctxA)
			lookup(t, f.uc, f.alice, ctxA)
			f.repo.failOnce(func(op string, s Session, id ID) (bool, bool) { return op == "delete" && id == x.ID, true })
			if err := f.uc.Delete(f.alice, x.ID); err == nil {
				t.Fatalf("premise: the delete did not fail")
			}
			return ""
		}},
		{"counter row: delete fails before it has been deleted", func(t *testing.T, f *fixture) ID {
			x := f.waiting(t, f.alice, ctxA)
			lookup(t, f.uc, f.alice, ctxA)
			f.repo.failOnce(func(op string, s Session, id ID) (bool, bool) { return op == "delete" && id == x.ID, false })
			if err := f.uc.Delete(f.alice, x.ID); err == nil {
				t.Fatalf("premise: the delete did not fail")
			}
			return x.ID
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t)
			want := c.run(t, f)
			pairs := []pair{{"alice", f.alice, ctxA}, {"bob", f.bob, ctxA}}
			if got := lookup(t, f.uc, f.alice, ctxA); got != want {
				t.Errorf("after the failed write: alice continues %q, expected %q", got, want)
			}
			sameAsReadAll(t, "after the failed write", f.uc, pairs)
			next := f.waiting(t, f.bob, ctxA)
			if got := lookup(t, f.uc, f.bob, ctxA); got != next.ID {
				t.Errorf("next write: bob continues %q, expected %q", got, next.ID)
			}
			sameAsReadAll(t, "after the next write", f.uc, pairs)
		})
	}
}

// concurrentWrite runs write; its name marks the goroutine in the stacks read by parkedOn.
//
//go:noinline
func concurrentWrite(write func(), done chan<- struct{}) {
	defer close(done)
	write()
}

// parkedOn reports whether the goroutine running concurrentWrite is parked on a lock or a channel, read from the
// stacks of all goroutines: it waits for something another goroutine holds.
func parkedOn() bool {
	buf := make([]byte, 1<<16)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			buf = buf[:n]
			break
		}
		buf = make([]byte, 2*len(buf))
	}
	for _, g := range strings.Split(string(buf), "\n\n") {
		if !strings.Contains(g, "session.concurrentWrite(") {
			continue
		}
		head, _, _ := strings.Cut(g, "\n")
		_, state, _ := strings.Cut(head, "[")
		for _, p := range []string{"sync.", "semacquire", "chan ", "select"} {
			if strings.HasPrefix(state, p) {
				return true
			}
		}
	}
	return false
}

// TestFindResumable_FirstBuildAgainstConcurrentWrites (R-DEC-ASSISTANT-RESUME-INDEX): a write which runs while the
// first lookup reads every session is not lost. The interleaving is forced: the store holds the read of the first
// lookup after its last session, before the lookup can take over what it read; then the write runs until it has
// been written (an index which does not keep writes away from a running build) or until it waits for the lookup (one
// which does); only then the read goes on. The lookup itself may answer as before or after the write, the next
// lookup must answer as after it. Rows: a session starts to wait (the write a stale build would lose), is dismissed,
// an older one is renamed to the newest, the newest is deleted.
func TestFindResumable_FirstBuildAgainstConcurrentWrites(t *testing.T) {
	for _, c := range []struct {
		name  string
		setup func(t *testing.T, f *fixture) (write func(), decisive func(op string, s Session, id ID) bool)
	}{
		{"a session starts to wait", func(t *testing.T, f *fixture) (func(), func(string, Session, ID) bool) {
			x := f.create(t, f.alice, ctxA, "")
			return func() { _, _ = f.ask(t, f.alice, x.ID) }, func(op string, s Session, id ID) bool { return op == "save" && id == x.ID && s.Pending != nil }
		}},
		{"the waiting session is dismissed", func(t *testing.T, f *fixture) (func(), func(string, Session, ID) bool) {
			x := f.waiting(t, f.alice, ctxA)
			return func() { _, _ = f.uc.Dismiss(f.alice, x.ID, x.PendingRevision) }, func(op string, s Session, id ID) bool { return op == "save" && id == x.ID && s.Pending == nil }
		}},
		{"an older waiting session is renamed to the newest", func(t *testing.T, f *fixture) (func(), func(string, Session, ID) bool) {
			older := f.waiting(t, f.alice, ctxA)
			f.waiting(t, f.alice, ctxA)
			return func() { _ = f.uc.Rename(f.alice, older.ID, "renamed") }, func(op string, s Session, id ID) bool { return op == "save" && id == older.ID && s.Title == "renamed" }
		}},
		{"the newest waiting session is deleted", func(t *testing.T, f *fixture) (func(), func(string, Session, ID) bool) {
			f.waiting(t, f.alice, ctxA)
			newest := f.waiting(t, f.alice, ctxA)
			if err := f.uc.Rename(f.alice, newest.ID, "newest"); err != nil {
				t.Fatal(err)
			}
			return func() { _ = f.uc.Delete(f.alice, newest.ID) }, func(op string, s Session, id ID) bool { return op == "delete" && id == newest.ID }
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t)
			write, decisive := c.setup(t, f)
			before := readAll(t, f.uc, f.alice, ctxA)

			landed := make(chan struct{})
			var once sync.Once
			f.repo.mu.Lock()
			f.repo.afterOp = func(op string, s Session, id ID) {
				if decisive(op, s, id) {
					once.Do(func() { close(landed) })
				}
			}
			f.repo.mu.Unlock()

			hold := f.repo.holdNextRead()
			var during ID
			lookupDone := make(chan struct{})
			go func() {
				defer close(lookupDone)
				found, err := f.uc.FindResumable(f.alice, ctxA)
				if err == nil && found.IsSome() {
					during = found.Unwrap().ID
				}
			}()
			select {
			case <-hold.reached:
			case <-lookupDone:
				t.Fatalf("premise: the first lookup did not read the sessions through All, Identifiers or FindAllByPrefix; the interleaving cannot be forced")
			}

			writeDone := make(chan struct{})
			go concurrentWrite(write, writeDone)
		wait:
			for {
				select {
				case <-landed:
					break wait
				case <-writeDone:
					break wait
				default:
				}
				if parkedOn() {
					break wait
				}
				runtime.Gosched()
			}
			close(hold.release)
			<-lookupDone
			<-writeDone

			after := readAll(t, f.uc, f.alice, ctxA)
			if during != before && during != after {
				t.Errorf("the lookup during the write answered %q, expected %q (before) or %q (after)", during, before, after)
			}
			if got := lookup(t, f.uc, f.alice, ctxA); got != after {
				t.Errorf("the next lookup answered %q, reading every session %q (before the write %q)", got, after, before)
			}
		})
	}
}

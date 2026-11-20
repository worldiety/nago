// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package session

// The instruments of the acceptance tests of the heads (heads_test.go): a repository which records every read of a
// conversation and can hold or fail it, stores which stop writing like a crashed process, the reference of reading
// every session and the comparison of all searches with it. They observe through the use cases and the stores under
// them, never through the heads themselves.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/xtime"
)

// patience bounds every wait for a result. It is no measure of time: a test waits for an observed result, and only a
// result which never comes ends the wait.
const patience = 10 * time.Second

// storeSeq names the stores of every fixture on their own, so that nothing a test leaves running meets another
// test, whatever the use cases key their locks and builds by.
var storeSeq atomic.Int64

func newStorePair() (content, heads *mem.BlobStore) {
	n := storeSeq.Add(1)
	return mem.NewBlobStore(fmt.Sprintf("nago.ai.session.t%d", n)), mem.NewBlobStore(fmt.Sprintf("nago.ai.session.head.t%d", n))
}

// goid returns the id of the calling goroutine, read from its stack.
func goid() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	f := strings.Fields(string(buf[:n]))
	if len(f) < 2 {
		return 0
	}
	id, _ := strconv.ParseUint(f[1], 10, 64)
	return id
}

func allStacks() string {
	buf := make([]byte, 1<<16)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			return string(buf[:n])
		}
		buf = make([]byte, 2*len(buf))
	}
}

// awaitGone waits until the goroutine has ended.
func awaitGone(t *testing.T, id uint64, what string) {
	t.Helper()
	deadline := time.Now().Add(patience)
	for strings.Contains(allStacks(), fmt.Sprintf("goroutine %d [", id)) {
		if time.Now().After(deadline) {
			t.Fatalf("%s: the goroutine %d did not end", what, id)
		}
		runtime.Gosched()
		time.Sleep(time.Millisecond)
	}
}

// eventually waits until cond holds.
func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(patience)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("%s did not happen", what)
		}
		runtime.Gosched()
		time.Sleep(time.Millisecond)
	}
}

// headWriter runs a write; its name marks the goroutine in the stacks read by writerParked.
//
//go:noinline
func headWriter(write func(), done chan<- struct{}) {
	defer close(done)
	write()
}

// writerParked reports whether a goroutine running headWriter is parked on a lock or a channel: it waits for
// something another goroutine holds.
func writerParked() bool {
	for _, g := range strings.Split(allStacks(), "\n\n") {
		if !strings.Contains(g, "session.headWriter(") {
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

// startWrite runs write in its own goroutine and returns once it has finished or waits for another goroutine, so
// that an interleaving is forced instead of left to the scheduler.
func startWrite(t *testing.T, write func()) (done chan struct{}) {
	t.Helper()
	done = make(chan struct{})
	go headWriter(write, done)
	deadline := time.Now().Add(patience)
	for {
		select {
		case <-done:
			return done
		default:
		}
		if writerParked() {
			return done
		}
		if time.Now().After(deadline) {
			t.Fatalf("the write neither finished nor waited")
		}
		runtime.Gosched()
	}
}

// readRecord is a read of the conversation id by the goroutine g.
type readRecord struct {
	id ID
	g  uint64
}

// hold stops the goroutine which reads a conversation until release is closed.
type hold struct {
	reached chan struct{}
	release chan struct{}
	g       atomic.Uint64
	once    sync.Once
}

func newHold() *hold { return &hold{reached: make(chan struct{}), release: make(chan struct{})} }

func (h *hold) wait(t *testing.T, what string) {
	t.Helper()
	select {
	case <-h.reached:
	case <-time.After(patience):
		t.Fatalf("premise: %s was never reached", what)
	}
}

func (h *hold) free() { h.once.Do(func() { close(h.release) }) }

// errReadProbe is a read error of the store, as the file of the store gives it, not a conversation which cannot be
// decoded.
var errReadProbe = &fs.PathError{Op: "pread", Path: "tdb.wal", Err: syscall.EIO}
var errWriteProbe = errors.New("probe: the store failed")

// contentRepo is the repository of the conversations under the use cases. It records every conversation it hands
// out, whoever reads it and through which method, and runs onRead before handing it out, which may hold or fail the
// read. Writes can fail before or after the store has written them, and afterWrite runs once a write landed.
type contentRepo struct {
	Repository

	mu         sync.Mutex
	reads      []readRecord
	onRead     func(id ID) error
	failWrite  func(op string, s Session, id ID) (fail, written bool) // one shot
	afterWrite func(op string, s Session, id ID)
}

func newContentRepo(r Repository) *contentRepo { return &contentRepo{Repository: r} }

func (r *contentRepo) setOnRead(fn func(id ID) error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onRead = fn
}

// holdRead holds the next read of id.
func (r *contentRepo) holdRead(id ID) *hold {
	h := newHold()
	var fired atomic.Bool
	r.setOnRead(func(got ID) error {
		if got != id || !fired.CompareAndSwap(false, true) {
			return nil
		}
		h.g.Store(goid())
		close(h.reached)
		<-h.release
		return nil
	})
	return h
}

// failRead fails the next read of id by a goroutine other than except (0 excepts none), e.g. of a build, and
// reports the goroutine.
func (r *contentRepo) failRead(id ID, except uint64) (failed <-chan uint64) {
	ch := make(chan uint64, 1)
	var fired atomic.Bool
	r.setOnRead(func(got ID) error {
		if got != id || (except != 0 && goid() == except) || !fired.CompareAndSwap(false, true) {
			return nil
		}
		ch <- goid()
		return errReadProbe
	})
	return ch
}

func (r *contentRepo) read(id ID) error {
	r.mu.Lock()
	r.reads = append(r.reads, readRecord{id: id, g: goid()})
	fn := r.onRead
	r.mu.Unlock()
	if fn != nil {
		return fn(id)
	}
	return nil
}

func (r *contentRepo) mark() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.reads)
}

// since returns the reads after mark.
func (r *contentRepo) since(mark int) []readRecord {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.reads[mark:])
}

func (r *contentRepo) FindByID(id ID) (option.Opt[Session], error) {
	o, err := r.Repository.FindByID(id)
	if err != nil || o.IsNone() {
		return o, err
	}
	if err := r.read(id); err != nil {
		return option.None[Session](), err
	}
	return o, nil
}

func (r *contentRepo) wrap(seq iter.Seq2[Session, error]) iter.Seq2[Session, error] {
	return func(yield func(Session, error) bool) {
		for s, err := range seq {
			if err == nil {
				if rerr := r.read(s.ID); rerr != nil {
					s, err = Session{}, rerr
				}
			}
			if !yield(s, err) {
				return
			}
		}
	}
}

func (r *contentRepo) All() iter.Seq2[Session, error] { return r.wrap(r.Repository.All()) }

func (r *contentRepo) FindAllByPrefix(prefix ID) iter.Seq2[Session, error] {
	return r.wrap(r.Repository.FindAllByPrefix(prefix))
}

func (r *contentRepo) FindAllByID(ids iter.Seq[ID]) iter.Seq2[Session, error] {
	return r.wrap(r.Repository.FindAllByID(ids))
}

func (r *contentRepo) write(op string, s Session, id ID, do func() error) error {
	r.mu.Lock()
	fn := r.failWrite
	fail, written := false, false
	if fn != nil {
		if fail, written = fn(op, s, id); fail {
			r.failWrite = nil
		}
	}
	after := r.afterWrite
	r.mu.Unlock()

	if fail && !written {
		return errWriteProbe
	}
	if err := do(); err != nil {
		return err
	}
	if after != nil {
		after(op, s, id)
	}
	if fail {
		return errWriteProbe
	}
	return nil
}

func (r *contentRepo) Save(s Session) error {
	return r.write("save", s, s.ID, func() error { return r.Repository.Save(s) })
}

func (r *contentRepo) SaveAll(it iter.Seq[Session]) error {
	for s := range it {
		if err := r.Save(s); err != nil {
			return err
		}
	}
	return nil
}

func (r *contentRepo) DeleteByID(id ID) error {
	return r.write("delete", Session{}, id, func() error { return r.Repository.DeleteByID(id) })
}

func (r *contentRepo) DeleteByEntity(s Session) error { return r.DeleteByID(s.ID) }

func (r *contentRepo) DeleteAllByID(ids iter.Seq[ID]) error {
	for id := range ids {
		if err := r.DeleteByID(id); err != nil {
			return err
		}
	}
	return nil
}

func (r *contentRepo) Delete(predicate func(Session) (bool, error)) error {
	return r.write("delete-by-predicate", Session{}, "", func() error {
		return r.Repository.Delete(func(s Session) (bool, error) {
			if err := r.read(s.ID); err != nil {
				return false, err
			}
			return predicate(s)
		})
	})
}

// crashLog is shared by the stores of one process. Once armed, it lets limit writes through and then stops every
// write, like a process which has crashed after its limit-th write; a negative limit never stops. failOnce instead
// fails the first write it matches and lets every other one through.
type crashLog struct {
	mu       sync.Mutex
	armed    bool
	limit    int
	writes   []string
	failOnce func(entry string) bool
}

var errCrashed = errors.New("probe: the process has crashed")

func (c *crashLog) arm(limit int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.armed, c.limit, c.writes = true, limit, nil
}

func (c *crashLog) allow(entry string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.failOnce != nil && c.failOnce(entry) {
		c.failOnce = nil
		return false
	}
	if !c.armed {
		return true
	}
	if c.limit >= 0 && len(c.writes) >= c.limit {
		return false
	}
	c.writes = append(c.writes, entry)
	return true
}

func (c *crashLog) log() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return slices.Clone(c.writes)
}

// crashStore lets a write through only as its crashLog allows. A write is one Set (a closed writer) or one Delete.
type crashStore struct {
	blob.Store
	log *crashLog
	tag string
}

func (s crashStore) NewWriter(ctx context.Context, key string) (io.WriteCloser, error) {
	return &crashWriter{s: s, key: key, ctx: ctx}, nil
}

func (s crashStore) Delete(ctx context.Context, key string) error {
	if !s.log.allow(s.tag + " del " + key) {
		return errCrashed
	}
	return s.Store.Delete(ctx, key)
}

func (s crashStore) Count(ctx context.Context) (int64, error) { return blob.Count(ctx, s.Store) }

func (s crashStore) Stat(ctx context.Context, key string) (option.Opt[blob.Info], error) {
	return blob.Stat(ctx, s.Store, key)
}

type crashWriter struct {
	s   crashStore
	key string
	ctx context.Context
	buf bytes.Buffer
}

func (w *crashWriter) Write(p []byte) (int, error) { return w.buf.Write(p) }

func (w *crashWriter) Close() error {
	if w.ctx != nil && w.ctx.Err() != nil {
		return w.ctx.Err()
	}
	if !w.s.log.allow(w.s.tag + " set " + w.key) {
		return errCrashed
	}
	return blob.Put(w.s.Store, w.key, w.buf.Bytes())
}

// keys lists the keys of a store with the prefix.
func keys(t *testing.T, store blob.Store, prefix string) []string {
	t.Helper()
	var res []string
	for k, err := range store.List(context.Background(), blob.ListOptions{Prefix: prefix}) {
		if err != nil {
			t.Fatalf("cannot list %s: %v", store.Name(), err)
		}
		res = append(res, k)
	}
	return res
}

// headKeyOf is the key of the head of s, as the decision names it.
func headKeyOf(s Session) string {
	switch {
	case s.ParentID != "":
		return "k/" + string(s.ParentID) + "/" + string(s.ID)
	case s.Pending != nil:
		return "h/" + string(s.CreatedBy) + "/w/" + string(s.ID)
	default:
		return "h/" + string(s.CreatedBy) + "/n/" + string(s.ID)
	}
}

// hasHeadOf reports whether the store has a head key of the session id, under whatever state.
func hasHeadOf(t *testing.T, heads blob.Store, id ID) bool {
	t.Helper()
	for _, k := range append(keys(t, heads, "h/"), keys(t, heads, "k/")...) {
		if strings.HasSuffix(k, "/"+string(id)) {
			return true
		}
	}
	return false
}

// hfix is a store of conversations with its heads, the use cases on both and a reference which reads every
// session: the same use cases without Heads, on the same store.
type hfix struct {
	t       *testing.T
	content *mem.BlobStore
	heads   *mem.BlobStore
	log     *crashLog
	bare    Repository   // the store of the conversations, unwrapped
	repo    *contentRepo // the store under the use cases
	rdb     *rebac.DB
	ctx     context.Context
	cancel  context.CancelFunc
	uc      UseCases // with Heads
	ref     UseCases // without Heads: reads every session

	alice, bob, carol rebacSubject
	testG             uint64
}

// newHFix creates the stores. Seed them, then start the use cases with start. With crash, the use cases write
// through crashStores.
func newHFix(t *testing.T, crash bool) *hfix {
	t.Helper()
	content, heads := newStorePair()
	f := &hfix{t: t, content: content, heads: heads, log: &crashLog{}, rdb: newTestRDB(t), testG: goid()}
	f.bare = Repository(json.NewSloppyJSONRepository[Session, ID](content))
	var under blob.Store = content
	if crash {
		under = crashStore{Store: content, log: f.log, tag: "content"}
	}
	f.repo = newContentRepo(Repository(json.NewSloppyJSONRepository[Session, ID](under)))
	f.ref = NewUseCases(f.bare, f.rdb)
	f.alice = newRebacSubject("alice", f.rdb, PermCreate)
	f.bob = newRebacSubject("bob", f.rdb, PermCreate)
	f.carol = newRebacSubject("carol", f.rdb, PermCreate)
	return f
}

// headStore is the store of the heads the use cases of the fixture write through.
func (f *hfix) headStore(crash bool) blob.Store {
	if crash {
		return crashStore{Store: f.heads, log: f.log, tag: "head"}
	}
	return f.heads
}

// start creates the use cases with Heads, as a start of the process does.
func (f *hfix) start(crash bool) UseCases {
	f.t.Helper()
	if f.cancel != nil {
		f.cancel()
	}
	f.ctx, f.cancel = context.WithCancel(context.Background())
	f.t.Cleanup(f.cancel)
	f.uc = NewUseCases(f.repo, f.rdb, Heads(f.ctx, f.headStore(crash)))
	return f.uc
}

// restart creates new use cases with Heads on the unwrapped stores, as the next start of the process does, and
// records their reads in a new contentRepo.
func (f *hfix) restart() UseCases {
	f.t.Helper()
	f.repo = newContentRepo(f.bare)
	return f.start(false)
}

// stored writes sessions directly into the store, as a former version left them, and grants their creators.
func (f *hfix) stored(sessions ...Session) {
	f.t.Helper()
	for _, s := range sessions {
		if err := f.bare.Save(s); err != nil {
			f.t.Fatal(err)
		}
		if s.CreatedBy != "" {
			grant(f.t, f.rdb, s.CreatedBy, s.ID, InstancePermissions...)
		}
	}
}

// awaitBuilt waits until the searches no longer read conversations: the build of the heads has finished.
func (f *hfix) awaitBuilt(uc UseCases) {
	f.t.Helper()
	nobody := newRebacSubject("nobody", f.rdb)
	eventually(f.t, "the end of the build of the heads", func() bool {
		// during a build, the searches read the conversations above the cursor and yield the error of one which
		// cannot be decoded, as reading every session does
		m := f.repo.mark()
		failed := false
		for _, err := range uc.FindSummaries(nobody, FindAllOptions{CreatedBy: "nobody"}) {
			failed = failed || err != nil
		}
		if _, err := uc.FindResumable(nobody, FindResumableOptions{}); err != nil {
			failed = true
		}
		return !failed && len(f.repo.since(m)) == 0
	})
}

var (
	ctxA     = FindResumableOptions{Tags: []string{"ctx:a"}, ProviderHint: "p"}
	ctxMore  = FindResumableOptions{Tags: []string{"ctx:a", "more"}, ProviderHint: "p"}
	ctxOther = FindResumableOptions{Tags: []string{"ctx:a"}, ProviderHint: "q"}
)

// invalidSubject is a subject which is not signed in.
type invalidSubject struct{ rebacSubject }

func (invalidSubject) Valid() bool { return false }

// legacy is a session as a former version stored it: no heads.
func legacy(id ID, by user.ID, updated xtime.UnixMilliseconds, waiting bool) Session {
	s := Session{ID: id, Model: "fake-model", CreatedBy: by, ProviderHint: ctxA.ProviderHint, Tags: ctxA.Tags, CreatedAt: updated, UpdatedAt: updated,
		Messages: []completion.Message{{Role: completion.User, Content: userText("question " + string(id))}}}
	if waiting {
		ask := askResult().Message
		s.Messages = append(s.Messages, ask)
		s.Pending = &completion.Continuation{Pending: []completion.PendingCall{{Kind: completion.PendingQuestion, Call: ask.Content[0].(completion.ToolCall)}}}
		s.PendingRevision = 1
	}
	return s
}

// create creates a session of the subject in the context, as child of parent unless it is empty.
func create(t *testing.T, uc UseCases, s auth.Subject, c FindResumableOptions, parent ID) Session {
	t.Helper()
	x, err := uc.Create(s, CreateOptions{Model: "fake-model", Tags: c.Tags, ProviderHint: c.ProviderHint, ParentID: parent, Input: userText("hello")})
	if err != nil {
		t.Fatalf("setup: create: %v", err)
	}
	return x
}

// ask lets the session wait on a question.
func ask(uc UseCases, s auth.Subject, id ID) (Session, error) {
	return uc.Append(s, id, AppendOptions{
		Completions: &scriptedCompletions{results: []completion.Result{askResult()}},
		Input:       userText("go"),
		Tools:       []completion.Tool{completion.NewAskUserTool()},
	})
}

func mustAsk(t *testing.T, uc UseCases, s auth.Subject, id ID) Session {
	t.Helper()
	x, err := ask(uc, s, id)
	if err != nil || x.Pending == nil {
		t.Fatalf("setup: the session does not wait: %v", err)
	}
	return x
}

// waiting creates a session of the subject in the context which waits on a question.
func waiting(t *testing.T, uc UseCases, s auth.Subject, c FindResumableOptions) Session {
	t.Helper()
	return mustAsk(t, uc, s, create(t, uc, s, c, "").ID)
}

// answer appends a turn which ends with an answer, without waiting.
func answer(t *testing.T, uc UseCases, s auth.Subject, id ID) Session {
	t.Helper()
	x, err := uc.Append(s, id, AppendOptions{Completions: &scriptedCompletions{results: []completion.Result{answerResult("ok")}}, Input: userText("hi")})
	if err != nil {
		t.Fatalf("setup: answer: %v", err)
	}
	return x
}

// resolve answers the question of a waiting session; the run ends with an answer.
func resolve(t *testing.T, uc UseCases, s auth.Subject, x Session) {
	t.Helper()
	_, err := uc.Resolve(s, x.ID, ResolveOptions{
		Run:         AppendOptions{Completions: &scriptedCompletions{results: []completion.Result{answerResult("thanks")}}, Tools: []completion.Tool{completion.NewAskUserTool()}},
		Revision:    x.PendingRevision,
		Resolutions: []completion.Resolution{{CallID: "q1", Answer: "the red one"}},
	})
	if err != nil {
		t.Fatalf("setup: resolve: %v", err)
	}
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

// readAll is the reference of FindResumable: the session a chat of the subject continues, found by reading every
// session through FindAll of use cases without Heads, as the chat did before (application/ai/completion/ui/chat.go
// before d7dd885c7): the newest one, and among equally new ones the first in the order of the repository. The rule
// is written out from the decision, not taken from [Session.ResumableBy].
func readAll(t *testing.T, ref UseCases, subject auth.Subject, opts FindResumableOptions) ID {
	t.Helper()
	if subject == nil || !subject.Valid() {
		return ""
	}

	var best *Session
	for s, err := range ref.FindAll(subject, FindAllOptions{}) {
		if err != nil {
			continue // a session which cannot be decoded is excepted
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

// sessionsOf collects FindAll; errors are counted, not collected.
func sessionsOf(uc UseCases, subject auth.Subject, opts FindAllOptions) (res []Session, errs int) {
	for s, err := range uc.FindAll(subject, opts) {
		if err != nil {
			errs++
			continue
		}
		res = append(res, s)
	}
	return res, errs
}

func summariesOf(uc UseCases, subject auth.Subject, opts FindAllOptions) (res []Summary, errs int) {
	for s, err := range uc.FindSummaries(subject, opts) {
		if err != nil {
			errs++
			continue
		}
		res = append(res, normalized(s))
	}
	return res, errs
}

// normalized lets no tags compare equal, whether they are nil or empty.
func normalized(s Summary) Summary {
	if len(s.Tags) == 0 {
		s.Tags = nil
	}
	return s
}

// refSummaries is the reference of FindSummaries: the summaries of what FindAll of the use cases without Heads
// yields with the same options, with CreatedBy in the order of the ids, otherwise by creator and id.
func refSummaries(ref UseCases, subject auth.Subject, opts FindAllOptions) []Summary {
	all, _ := sessionsOf(ref, subject, opts)
	res := make([]Summary, 0, len(all))
	for _, s := range all {
		res = append(res, normalized(s.Summary()))
	}
	slices.SortStableFunc(res, func(a, b Summary) int {
		if opts.CreatedBy == "" {
			if c := strings.Compare(string(a.CreatedBy), string(b.CreatedBy)); c != 0 {
				return c
			}
		}
		return strings.Compare(string(a.ID), string(b.ID))
	})
	return res
}

func idsOf(all []Session) []ID {
	res := make([]ID, 0, len(all))
	for _, s := range all {
		res = append(res, s.ID)
	}
	return res
}

// persons are the subjects every comparison asks for: three persons, the system subject, which may read every
// session and has created its own, and a subject which is not signed in.
func (f *hfix) persons() []auth.Subject {
	return []auth.Subject{f.alice, f.bob, f.carol, user.SU(), invalidSubject{f.alice}}
}

// diffs compares every search of uc with the reference on the same store and returns the differences.
func (f *hfix) diffs(uc UseCases) []string {
	f.t.Helper()
	return f.diffsFor(uc, f.persons())
}

// diffsFor compares the searches of the given persons.
func (f *hfix) diffsFor(uc UseCases, persons []auth.Subject) []string {
	f.t.Helper()
	var res []string
	add := func(format string, args ...any) { res = append(res, fmt.Sprintf(format, args...)) }

	for _, p := range persons {
		name := string(p.ID())
		if !p.Valid() {
			name = "not signed in"
		} else if name == "" {
			name = "system"
		}
		for _, c := range []FindResumableOptions{ctxA, ctxMore, ctxOther} {
			found, err := uc.FindResumable(p, c)
			if err != nil {
				add("%s: FindResumable %v: %v", name, c, err)
				continue
			}
			got := ID("")
			if found.IsSome() {
				got = found.Unwrap().ID
			}
			if want := readAll(f.t, f.ref, p, c); got != want {
				add("%s: FindResumable %v: %q, reading every session %q", name, c, got, want)
			}
		}

		var variants []FindAllOptions
		for _, by := range []user.ID{"alice", "bob", "carol"} {
			variants = append(variants, FindAllOptions{CreatedBy: by}, FindAllOptions{CreatedBy: by, Tags: ctxA.Tags})
		}
		for _, opts := range variants {
			got, gotErrs := sessionsOf(uc, p, opts)
			want, wantErrs := sessionsOf(f.ref, p, opts)
			if gotErrs > wantErrs {
				add("%s: FindAll %+v: %d errors, reading every session %d", name, opts, gotErrs, wantErrs)
			}
			if !reflect.DeepEqual(idsOf(got), idsOf(want)) {
				add("%s: FindAll %+v: %q, reading every session %q", name, opts, idsOf(got), idsOf(want))
			} else if !reflect.DeepEqual(got, want) {
				add("%s: FindAll %+v: other contents than reading every session", name, opts)
			}
		}

		for _, opts := range append(variants, FindAllOptions{}, FindAllOptions{IncludeChildren: true}) {
			got, gotErrs := summariesOf(uc, p, opts)
			want := refSummaries(f.ref, p, opts)
			if _, wantErrs := sessionsOf(f.ref, p, opts); gotErrs > wantErrs {
				add("%s: FindSummaries %+v: %d errors, reading every session %d", name, opts, gotErrs, wantErrs)
			}
			if len(got) == 0 && len(want) == 0 {
				continue
			}
			if !reflect.DeepEqual(got, want) {
				add("%s: FindSummaries %+v:\n\t%+v\nreading every session:\n\t%+v", name, opts, got, want)
			}
		}
	}

	return res
}

// same fails the test where a search of uc differs from reading every session.
func (f *hfix) same(step string, uc UseCases) {
	f.t.Helper()
	f.sameFor(step, uc, f.persons()...)
}

func (f *hfix) sameFor(step string, uc UseCases, persons ...auth.Subject) {
	f.t.Helper()
	for _, d := range f.diffsFor(uc, persons) {
		f.t.Errorf("%s: %s", step, d)
	}
}

// foreign returns the reads of conversations which do not belong to the allowed ones.
func foreign(reads []readRecord, allowed ...ID) []ID {
	var res []ID
	for _, r := range reads {
		if !slices.Contains(allowed, r.id) {
			res = append(res, r.id)
		}
	}
	return res
}

// byOthers returns the reads of goroutines other than g, e.g. of a build in the background.
func byOthers(reads []readRecord, g uint64) []ID {
	var res []ID
	for _, r := range reads {
		if r.g != g {
			res = append(res, r.id)
		}
	}
	return res
}

func idsOwnedBy(t *testing.T, ref UseCases, by user.ID) []ID {
	t.Helper()
	all, _ := sessionsOf(ref, user.SU(), FindAllOptions{IncludeChildren: true})
	var res []ID
	for _, s := range all {
		if s.CreatedBy == by {
			res = append(res, s.ID)
		}
	}
	return res
}

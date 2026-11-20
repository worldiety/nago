// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgai

// Acceptance tests of the heads of the sessions through the real wiring (wokoda R-DEC-ASSISTANT-SESSION-HEADS,
// requirements/dec/R-DEC-ASSISTANT-SESSION-HEADS/detail.md, „Rücksicherung“): a Configurator with Enable keeps the
// heads in nago.ai.session.head, the backup leaves them out, and after a restore, succeeded, aborted or crashed in the
// middle, every search equals reading every session and the heads are built again from the restored conversations.
// They use the API of nago before the heads only, so that they run against the former version and fail there at
// their statement; the names of the message and the mark of a restore are not known to them.

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	stdjson "encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/application/backup"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
)

const headStoreName = "nago.ai.session.head"

// headsPatience bounds every wait for a result; only a result which never comes ends the wait.
const headsPatience = 10 * time.Second

// headsPerson is a signed in user who may read every session, like an administrator, with an id of its own.
type headsPerson struct {
	user.Subject
	id user.ID
}

func (p headsPerson) ID() user.ID { return p.id }

func as(id user.ID) auth.Subject { return headsPerson{Subject: user.SU(), id: id} }

// headsApp is a Configurator with the AI management on a data directory, the store of the conversations and use
// cases without Heads on it, which read every session: the reference.
type headsApp struct {
	t       *testing.T
	cfg     *application.Configurator
	app     *core.Application
	mgmt    Management
	content blob.Store
	ref     session.UseCases
	once    sync.Once
}

func openHeadsApp(t *testing.T, dir string) *headsApp {
	t.Helper()
	cfg := application.NewConfigurator()
	cfg.SetDataDir(dir)
	cfg.SetApplicationID("de.worldiety.ai.headstest")
	mgmt, err := Enable(cfg)
	if err != nil {
		t.Fatal(err)
	}
	app, err := cfg.NewInProcessApplication()
	if err != nil {
		t.Fatal(err)
	}
	a := &headsApp{t: t, cfg: cfg, app: app, mgmt: mgmt}
	t.Cleanup(a.close)
	if a.content, err = cfg.EntityStore(string(session.Namespace)); err != nil {
		t.Fatal(err)
	}
	rdb, err := cfg.RDB()
	if err != nil {
		t.Fatal(err)
	}
	a.ref = session.NewUseCases(session.Repository(json.NewSloppyJSONRepository[session.Session, session.ID](a.content)), rdb)
	return a
}

func (a *headsApp) close() { a.once.Do(a.app.Destroy) }

// heads returns the store of the heads, which the AI management opened; it fails the test if there is none.
func (a *headsApp) heads() blob.Store {
	a.t.Helper()
	stores, err := a.cfg.Stores()
	if err != nil {
		a.t.Fatal(err)
	}
	o, err := stores.Get(headStoreName)
	if err != nil {
		a.t.Fatal(err)
	}
	if o.IsNone() {
		a.t.Fatalf("the AI management keeps no heads in %s", headStoreName)
	}
	return o.Unwrap()
}

func (a *headsApp) restore(ctx context.Context, zipped []byte) error {
	a.t.Helper()
	bm, err := a.cfg.BackupManagement()
	if err != nil {
		a.t.Fatal(err)
	}
	return bm.UseCases.Restore(ctx, user.SU(), bytes.NewReader(zipped))
}

// all reads every session, children included.
func (a *headsApp) all() []session.Session {
	a.t.Helper()
	var res []session.Session
	for s, err := range a.ref.FindAll(user.SU(), session.FindAllOptions{IncludeChildren: true}) {
		if err == nil {
			res = append(res, s)
		}
	}
	return res
}

// headKey is the key of the head of s, as the decision names it.
func headKey(s session.Session) string {
	switch {
	case s.ParentID != "":
		return "k/" + string(s.ParentID) + "/" + string(s.ID)
	case s.Pending != nil:
		return "h/" + string(s.CreatedBy) + "/w/" + string(s.ID)
	default:
		return "h/" + string(s.CreatedBy) + "/n/" + string(s.ID)
	}
}

func headKeys(t *testing.T, store blob.Store) []string {
	t.Helper()
	var res []string
	for _, prefix := range []string{"h/", "k/"} {
		for k, err := range store.List(context.Background(), blob.ListOptions{Prefix: prefix}) {
			if err != nil {
				t.Fatal(err)
			}
			res = append(res, k)
		}
	}
	slices.Sort(res)
	return res
}

// wantHeads are the heads of the conversations in the store, one for each, none else.
func (a *headsApp) wantHeads() []string {
	var res []string
	for _, s := range a.all() {
		res = append(res, headKey(s))
	}
	slices.Sort(res)
	return res
}

// awaitHeads waits until the heads are those of the conversations: the build after a restore has finished.
func (a *headsApp) awaitHeads(what string) {
	a.t.Helper()
	heads := a.heads()
	deadline := time.Now().Add(headsPatience)
	for {
		got, want := headKeys(a.t, heads), a.wantHeads()
		if slices.Equal(got, want) {
			return
		}
		if time.Now().After(deadline) {
			a.t.Fatalf("%s: the heads are\n\t%q\nnot those of the conversations\n\t%q", what, got, want)
		}
		runtime.Gosched()
		time.Sleep(time.Millisecond)
	}
}

var (
	headsCtxA     = session.FindResumableOptions{Tags: []string{"ctx:a"}, ProviderHint: "p"}
	headsCtxOther = session.FindResumableOptions{Tags: []string{"ctx:a"}, ProviderHint: "q"}
)

// readingAll is the reference of FindResumable: the newest waiting session without parent of the subject in the
// context, found by reading every session, the first in the order of the store among equally new ones.
func readingAll(ref session.UseCases, subject auth.Subject, opts session.FindResumableOptions) session.ID {
	var best *session.Session
	for s, err := range ref.FindAll(subject, session.FindAllOptions{}) {
		if err != nil || s.Pending == nil || s.ParentID != "" || s.CreatedBy != subject.ID() || s.ProviderHint != opts.ProviderHint || len(s.Tags) != len(opts.Tags) {
			continue
		}
		if !slices.ContainsFunc(opts.Tags, func(t string) bool { return !slices.Contains(s.Tags, t) }) && (best == nil || s.UpdatedAt > best.UpdatedAt) {
			best = &s
		}
	}
	if best == nil {
		return ""
	}
	return best.ID
}

func idsOfAll(uc session.UseCases, subject auth.Subject, opts session.FindAllOptions) []session.ID {
	var res []session.ID
	for s, err := range uc.FindAll(subject, opts) {
		if err == nil {
			res = append(res, s.ID)
		}
	}
	return res
}

// same fails the test where a search of the AI management differs from reading every session, for every creator
// of a session in the store and alice and bob.
func (a *headsApp) same(step string) {
	a.t.Helper()
	ids := []user.ID{"alice", "bob"}
	for _, s := range a.all() {
		if !slices.Contains(ids, s.CreatedBy) {
			ids = append(ids, s.CreatedBy)
		}
	}
	uc := a.mgmt.SessionUseCases
	for _, id := range ids {
		p := as(id)
		for _, c := range []session.FindResumableOptions{headsCtxA, headsCtxOther} {
			found, err := uc.FindResumable(p, c)
			if err != nil {
				a.t.Errorf("%s: %s: FindResumable: %v", step, id, err)
				continue
			}
			got := session.ID("")
			if found.IsSome() {
				got = found.Unwrap().ID
			}
			if want := readingAll(a.ref, p, c); got != want {
				a.t.Errorf("%s: %s continues %q in %v, reading every session %q", step, id, got, c, want)
			}
		}
		for _, opts := range []session.FindAllOptions{{CreatedBy: id}, {CreatedBy: id, Tags: headsCtxA.Tags}} {
			if got, want := idsOfAll(uc, p, opts), idsOfAll(a.ref, p, opts); !slices.Equal(got, want) {
				a.t.Errorf("%s: FindAll %+v: %q, reading every session %q", step, opts, got, want)
			}
		}
	}
}

func waitingSession(id session.ID, by user.ID, updated int, parent session.ID) session.Session {
	s := plainSession(id, by, updated, parent)
	call := completion.ToolCall{ID: "q1", Name: completion.AskUserToolName, Arguments: stdjson.RawMessage(`{"question":"which?"}`)}
	s.Messages = append(s.Messages, completion.Message{Role: completion.Assistant, Content: []completion.Content{call}})
	s.Pending = &completion.Continuation{Pending: []completion.PendingCall{{Kind: completion.PendingQuestion, Call: call}}}
	s.PendingRevision = 1
	return s
}

func plainSession(id session.ID, by user.ID, updated int, parent session.ID) session.Session {
	return session.Session{ID: id, Model: "fake-model", CreatedBy: by, ParentID: parent, ProviderHint: headsCtxA.ProviderHint, Tags: headsCtxA.Tags,
		CreatedAt: xtime.UnixMilliseconds(updated), UpdatedAt: xtime.UnixMilliseconds(updated),
		Messages: []completion.Message{{Role: completion.User, Content: []completion.Content{completion.Text{Text: "question " + string(id)}}}}}
}

// backupOf is a backup of the store of the conversations with the given sessions, in this order, in the format of
// application/backup.
func backupOf(t *testing.T, sessions ...session.Session) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	st := backup.Store{Name: string(session.Namespace), Stereotype: backup.StereotypeDocument}
	for i, s := range sessions {
		b, err := stdjson.Marshal(&s)
		if err != nil {
			t.Fatal(err)
		}
		path := fmt.Sprintf("data/%d", i+1)
		w, err := zw.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(b); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b)
		st.Blobs = append(st.Blobs, backup.Blob{ID: string(s.ID), Size: int64(len(b)), Sha256: hex.EncodeToString(sum[:]), Path: path})
	}
	idx, err := stdjson.Marshal(backup.Index{CreatedAt: time.Now(), Stores: []backup.Store{st}})
	if err != nil {
		t.Fatal(err)
	}
	w, err := zw.Create("index.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(idx); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// storesOfBackup returns the names of the stores in a backup.
func storesOfBackup(t *testing.T, zipped []byte) []string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(zipped), int64(len(zipped)))
	if err != nil {
		t.Fatal(err)
	}
	f, err := zr.Open("index.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var idx backup.Index
	if err := stdjson.NewDecoder(f).Decode(&idx); err != nil {
		t.Fatal(err)
	}
	var res []string
	for _, s := range idx.Stores {
		res = append(res, s.Name)
	}
	return res
}

// oldSessions writes conversations of alice and bob through the AI management.
func (a *headsApp) oldSessions() {
	a.t.Helper()
	for _, by := range []user.ID{"alice", "alice", "bob"} {
		if _, err := a.mgmt.SessionUseCases.Create(as(by), session.CreateOptions{Model: "fake-model", Tags: headsCtxA.Tags, ProviderHint: headsCtxA.ProviderHint, Input: []completion.Content{completion.Text{Text: "old"}}}); err != nil {
			a.t.Fatal(err)
		}
	}
}

// restoredSessions are the conversations of a backup: waiting ones of alice and bob, an answered one of alice with a
// child which bob created by continuing it.
func restoredSessions() []session.Session {
	return []session.Session{
		waitingSession("n1", "alice", 1_000, ""),
		plainSession("n2", "alice", 1_001, ""),
		waitingSession("n3", "bob", 1_002, ""),
		plainSession("n4", "bob", 1_003, "n2"),
	}
}

// TestHeads_BackupLeavesOutTheHeads (R-DEC-ASSISTANT-SESSION-HEADS, „die Sicherung lässt sie aus“): the AI management
// keeps a head of every conversation it writes in nago.ai.session.head, and a backup of the application contains the
// conversations, not the heads. Counter row: nago.ai.session is in the backup.
func TestHeads_BackupLeavesOutTheHeads(t *testing.T) {
	a := openHeadsApp(t, t.TempDir())
	a.oldSessions()
	if got, want := headKeys(t, a.heads()), a.wantHeads(); !slices.Equal(got, want) {
		t.Fatalf("premise: the heads are %q, not those of the conversations %q", got, want)
	}

	bm, err := a.cfg.BackupManagement()
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := bm.UseCases.Backup(context.Background(), user.SU(), &buf); err != nil {
		t.Fatal(err)
	}
	stores := storesOfBackup(t, buf.Bytes())
	if slices.Contains(stores, headStoreName) {
		t.Errorf("the backup contains %s: %q", headStoreName, stores)
	}
	if !slices.Contains(stores, string(session.Namespace)) {
		t.Errorf("counter row: the backup lacks %s: %q", session.Namespace, stores)
	}
}

// stopCtx runs check on every Done and Err, i.e. at the points where the restore asks its context.
type stopCtx struct {
	context.Context
	check func()
}

func (c stopCtx) Done() <-chan struct{} {
	c.check()
	return c.Context.Done()
}

func (c stopCtx) Err() error {
	c.check()
	return c.Context.Err()
}

// written counts the given conversations in the store.
func written(store blob.Store, sessions []session.Session) int {
	n := 0
	for _, s := range sessions {
		if ok, _ := store.Exists(context.Background(), string(s.ID)); ok {
			n++
		}
	}
	return n
}

// TestHeads_RestoreRebuildsTheHeads (R-DEC-ASSISTANT-SESSION-HEADS, row „Rücksicherung gelungen und abgebrochen“):
// after a restore which succeeded, of a backup of another application, and after one which was aborted after its
// first conversation, every search equals reading every session, at once and after the heads have been built again,
// and the heads are those of the restored conversations: none is left of the former ones. Counter row: before the
// restore, the heads are those of the former conversations, which the searches find.
func TestHeads_RestoreRebuildsTheHeads(t *testing.T) {
	t.Run("succeeded", func(t *testing.T) {
		src := openHeadsApp(t, t.TempDir())
		for _, s := range restoredSessions() {
			if err := session.Repository(json.NewSloppyJSONRepository[session.Session, session.ID](src.content)).Save(s); err != nil {
				t.Fatal(err)
			}
		}
		bm, err := src.cfg.BackupManagement()
		if err != nil {
			t.Fatal(err)
		}
		var zipped bytes.Buffer
		if err := bm.UseCases.Backup(context.Background(), user.SU(), &zipped); err != nil {
			t.Fatal(err)
		}

		a := openHeadsApp(t, t.TempDir())
		a.oldSessions()
		a.awaitHeads("counter row: before the restore")
		a.same("counter row: before the restore")
		if err := a.restore(context.Background(), zipped.Bytes()); err != nil {
			t.Fatal(err)
		}
		if n := written(a.content, restoredSessions()); n != len(restoredSessions()) {
			t.Fatalf("premise: %d of %d conversations restored", n, len(restoredSessions()))
		}
		a.same("right after the restore")
		a.awaitHeads("after the restore")
		a.same("after the build")
	})

	t.Run("aborted", func(t *testing.T) {
		a := openHeadsApp(t, t.TempDir())
		a.oldSessions()
		a.awaitHeads("counter row: before the restore")
		restored := restoredSessions()
		inner, cancel := context.WithCancel(context.Background())
		defer cancel()
		ctx := stopCtx{Context: inner, check: func() {
			if written(a.content, restored) > 0 {
				cancel()
			}
		}}
		if err := a.restore(ctx, backupOf(t, restored...)); err == nil {
			t.Fatalf("premise: the restore was not aborted")
		}
		if n := written(a.content, restored); n == 0 || n == len(restored) {
			t.Fatalf("premise: the restore was not aborted in the middle: %d of %d conversations restored", n, len(restored))
		}
		a.same("right after the aborted restore")
		a.awaitHeads("after the aborted restore")
		a.same("after the build")
	})
}

const crashHelperEnv = "NAGO_HEADS_CRASH_HELPER"

// TestHeads_HelperProcessCrashWhileRestoring is the process TestHeads_CrashWhileRestoring kills; it is skipped
// otherwise. It writes conversations through the AI management, then restores a backup and stops for good after the
// first restored conversation, telling so on its standard output.
func TestHeads_HelperProcessCrashWhileRestoring(t *testing.T) {
	dir, zipPath, ok := strings.Cut(os.Getenv(crashHelperEnv), string(os.PathListSeparator))
	if !ok {
		t.Skip("only run as the process TestHeads_CrashWhileRestoring kills")
	}
	zipped, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	a := openHeadsApp(t, dir)
	a.oldSessions()
	a.awaitHeads("before the restore")
	restored := restoredSessions()
	ctx := stopCtx{Context: context.Background(), check: func() {
		if n := written(a.content, restored); n > 0 && n < len(restored) {
			fmt.Println("HEADS-PAUSED")
			_ = os.Stdout.Sync()
			select {}
		}
	}}
	err = a.restore(ctx, zipped)
	fmt.Println("HEADS-NOT-PAUSED", err)
}

// TestHeads_CrashWhileRestoring (R-DEC-ASSISTANT-SESSION-HEADS, row „Absturz mitten im Leeren einer Rücksicherung“): a
// process which crashes during a restore, after it has cleared the store of the conversations and written the first
// restored one, leaves a store whose former heads would hide the restored conversation. The next start on the same
// directory searches like reading every session at once, and it builds the heads of the restored conversations,
// none of the former ones. The crash is real: the test kills a process of its own binary, which stopped there.
func TestHeads_CrashWhileRestoring(t *testing.T) {
	dir := t.TempDir()
	zipPath := dir + ".zip"
	if err := os.WriteFile(zipPath, backupOf(t, restoredSessions()...), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(zipPath) })

	cmd := exec.Command(os.Args[0], "-test.run=^TestHeads_HelperProcessCrashWhileRestoring$", "-test.count=1")
	cmd.Env = append(os.Environ(), crashHelperEnv+"="+dir+string(os.PathListSeparator)+zipPath)
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	paused := make(chan string, 1)
	go func() {
		sc := bufio.NewScanner(out)
		var seen []string
		for sc.Scan() {
			line := sc.Text()
			if strings.HasPrefix(line, "HEADS-") {
				paused <- line
				break
			}
			seen = append(seen, line)
			if len(seen) > 50 {
				seen = seen[1:]
			}
		}
		paused <- "EOF " + strings.Join(seen, "\n")
		_, _ = io.Copy(io.Discard, out)
	}()
	var line string
	select {
	case line = <-paused:
	case <-time.After(2 * headsPatience):
		line = "no answer"
	}
	_ = cmd.Process.Kill()
	_ = cmd.Wait()
	if line != "HEADS-PAUSED" {
		t.Fatalf("premise: the helper process did not stop in the middle of the restore: %s", line)
	}

	a := openHeadsApp(t, dir)
	if n := written(a.content, restoredSessions()); n == 0 || n == len(restoredSessions()) {
		t.Fatalf("premise: the crash did not happen in the middle of the restore: %d of %d conversations restored", n, len(restoredSessions()))
	}
	a.same("right after the start after the crash")
	a.awaitHeads("after the start after the crash")
	a.same("after the build")
}

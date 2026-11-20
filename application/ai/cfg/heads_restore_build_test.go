// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgai

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data/json"
)

// heldSave holds the save of one conversation right after it has been written.
type heldSave struct {
	session.Repository
	id      session.ID
	reached chan struct{}
	release chan struct{}
	once    sync.Once
	freed   sync.Once
}

func (r *heldSave) Save(s session.Session) error {
	if err := r.Repository.Save(s); err != nil {
		return err
	}
	if s.ID == r.id {
		r.once.Do(func() {
			close(r.reached)
			<-r.release
		})
	}
	return nil
}

func (r *heldSave) free() { r.freed.Do(func() { close(r.release) }) }

// restoreInBackground runs a restore; its name marks the goroutine in the stacks read by restoreParked.
//
//go:noinline
func restoreInBackground(restore func() error, done chan<- error) {
	done <- restore()
}

// restoreParked reports whether the goroutine running restoreInBackground waits for another goroutine, and its stack.
func restoreParked() (bool, string) {
	buf := make([]byte, 1<<20)
	n := runtime.Stack(buf, true)
	for _, g := range strings.Split(string(buf[:n]), "\n\n") {
		if !strings.Contains(g, "ai/cfg.restoreInBackground(") {
			continue
		}
		head, _, _ := strings.Cut(g, "\n")
		_, state, _ := strings.Cut(head, "[")
		for _, p := range []string{"sync.", "semacquire", "chan ", "select", "sleep"} {
			if strings.HasPrefix(state, p) {
				return true, g
			}
		}
		return false, g
	}
	return false, ""
}

func (a *headsApp) hasHead(id session.ID) bool {
	for _, k := range headKeys(a.t, a.heads()) {
		if strings.HasSuffix(k, "/"+string(id)) {
			return true
		}
	}
	return false
}

// TestHeads_RestoreDuringABuild (R-DEC-ASSISTANT-SESSION-HEADS, row „Rücksicherung während eines Aufbaus“): a restore
// while the build of the heads stands at s0550 of 600 conversations, beyond its first cursor, is followed by a build
// from the start: afterwards the heads are exactly those of the newly restored conversations r000 to r009, which lie
// below the former cursor, and every search equals reading every session, also while the first build stands.
//
// The interleaving is forced through the lock of a conversation, which use cases on the same store share: further use
// cases of the test hold a rename of s0550 after its conversation and before its head, so that the build after a
// first restore waits at s0550. Then the second restore runs until it has finished or waits for another goroutine,
// e.g. for the end of the build, and only then the rename goes on. Counter row of the decision, „Fortsetzen am alten
// Cursor: Köpfe darunter fehlen“: a build which went on at its former cursor would leave r000 to r009 without heads.
func TestHeads_RestoreDuringABuild(t *testing.T) {
	a := openHeadsApp(t, t.TempDir())
	heads := a.heads()
	rdb, err := a.cfg.RDB()
	if err != nil {
		t.Fatal(err)
	}

	// s0550 exists before the first restore, so that the further use cases can hold its lock
	bare := session.Repository(json.NewSloppyJSONRepository[session.Session, session.ID](a.content))
	if err := bare.Save(plainSession("s0550", "alice", 1, "")); err != nil {
		t.Fatal(err)
	}
	held := &heldSave{Repository: bare, id: "s0550", reached: make(chan struct{}), release: make(chan struct{})}
	t.Cleanup(held.free)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	further := session.NewUseCases(held, rdb, session.Heads(ctx, heads))
	renamed := make(chan error, 1)
	go func() { renamed <- further.Rename(as("alice"), "s0550", "held") }()
	select {
	case <-held.reached:
	case <-time.After(headsPatience):
		t.Fatalf("premise: the rename of s0550 did not write its conversation")
	}

	var first []session.Session
	for i := range 600 {
		by := []string{"alice", "bob"}[i%2]
		if i%9 == 0 {
			first = append(first, waitingSession(session.ID(fmt.Sprintf("s%04d", i)), user.ID(by), 1_000+i, ""))
		} else {
			first = append(first, plainSession(session.ID(fmt.Sprintf("s%04d", i)), user.ID(by), 1_000+i, ""))
		}
	}
	if err := a.restore(context.Background(), backupOf(t, first...)); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(headsPatience)
	for !a.hasHead("s0549") {
		if time.Now().After(deadline) {
			t.Fatalf("premise: the build after the first restore did not reach s0549")
		}
		runtime.Gosched()
		time.Sleep(time.Millisecond)
	}
	if a.hasHead("s0551") {
		t.Fatalf("premise: the build did not wait at the lock of s0550, which further use cases on the same store hold")
	}
	a.same("while the build stands at s0550")

	var second []session.Session
	for i := range 10 {
		by := []string{"alice", "bob"}[i%2]
		id := session.ID(fmt.Sprintf("r%03d", i))
		switch {
		case i == 9:
			second = append(second, plainSession(id, user.ID(by), 5_000+i, "r001"))
		case i%3 == 0:
			second = append(second, waitingSession(id, user.ID(by), 5_000+i, ""))
		default:
			second = append(second, plainSession(id, user.ID(by), 5_000+i, ""))
		}
	}
	zipped := backupOf(t, second...)
	done := make(chan error, 1)
	ended := make(chan struct{})
	go restoreInBackground(func() error {
		defer close(ended)
		return a.restore(context.Background(), zipped)
	}, done)
	// registered after the application, thus run before it is destroyed: the restore must not outlive it
	t.Cleanup(func() {
		held.free()
		select {
		case <-ended:
		case <-time.After(headsPatience):
			t.Errorf("the second restore did not end")
		}
	})
	var restoreErr error
	finished := false
	deadline = time.Now().Add(headsPatience)
wait:
	for {
		select {
		case restoreErr = <-done:
			finished = true
			break wait
		default:
		}
		parked, stack := restoreParked()
		if parked {
			break wait
		}
		if time.Now().After(deadline) {
			t.Fatalf("the second restore neither finished nor waited:\n%s", stack)
		}
		runtime.Gosched()
	}
	t.Logf("the second restore finished before the rename went on: %v", finished)
	held.free()
	if err := <-renamed; err != nil {
		t.Logf("the held rename: %v", err)
	}
	if !finished {
		select {
		case restoreErr = <-done:
		case <-time.After(headsPatience):
			t.Fatalf("the second restore did not finish")
		}
	}
	if restoreErr != nil {
		t.Fatal(restoreErr)
	}
	if n := written(a.content, second); n != len(second) {
		t.Fatalf("premise: %d of %d conversations of the second restore written", n, len(second))
	}

	a.same("right after the second restore")
	a.awaitHeads("after the second restore")
	a.same("after the build")
	for _, s := range second {
		if !a.hasHead(s.ID) {
			t.Errorf("%s, below the former cursor, has no head", s.ID)
		}
	}
	deadline = time.Now().Add(headsPatience)
	for {
		marks := 0
		for _, err := range heads.List(context.Background(), blob.ListOptions{Prefix: "o/"}) {
			if err == nil {
				marks++
			}
		}
		if marks == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("%d marks are left", marks)
		}
		time.Sleep(time.Millisecond)
	}
}

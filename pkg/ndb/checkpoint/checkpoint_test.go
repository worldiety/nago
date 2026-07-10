// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package checkpoint_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/pkg/ndb"
	"go.wdy.de/nago/pkg/ndb/checkpoint"
	"go.wdy.de/nago/pkg/ndb/msgstore"
)

const typeID ndb.TypeID = "1"

// openMessages opens a fresh msgstore-backed Messages capability in a temp dir.
func openMessages(t *testing.T) (ndb.Messages, func()) {
	t.Helper()
	root := t.TempDir()
	db := option.Must(ndb.Open(root, ndb.Options{}))
	eng, err := db.Engine("events", ndb.EngineOptions{Kind: msgstore.EngineKind, Config: msgstore.Options{Compress: msgstore.NoCompression}})
	if err != nil {
		t.Fatalf("open engine: %v", err)
	}
	me, ok := eng.(ndb.MessageEngine)
	if !ok {
		t.Fatal("expected message engine")
	}
	return me.Messages(), func() { option.MustZero(db.Close()) }
}

// countingStore wraps a Store to count Save calls and expose the last value.
type countingStore struct {
	inner checkpoint.Store
	saves int
	last  ndb.Seq
}

func (c *countingStore) Load() (ndb.Seq, error) { return c.inner.Load() }
func (c *countingStore) Save(seq ndb.Seq) error {
	c.saves++
	c.last = seq
	return c.inner.Save(seq)
}

func appendN(t *testing.T, m ndb.Messages, n int) {
	t.Helper()
	var trace [16]byte
	for i := 0; i < n; i++ {
		option.Must(m.Append(typeID, trace, []byte("x")))
	}
}

// drainHistory ranges the stream until it has seen `want` messages, committing
// each, then returns. It relies on the fact that a msgstore Tail delivers all
// history before blocking on the live edge; we stop after `want`.
func drainHistory(t *testing.T, stream func(func(ndb.TypeID, ndb.Message) bool), cp *checkpoint.Committer, want int) []ndb.Seq {
	t.Helper()
	var seqs []ndb.Seq
	stream(func(_ ndb.TypeID, msg ndb.Message) bool {
		seqs = append(seqs, msg.Seq)
		if err := cp.Commit(msg.Seq); err != nil {
			t.Fatalf("commit: %v", err)
		}
		return len(seqs) < want
	})
	return seqs
}

func TestResumeFromCommittedCursor(t *testing.T) {
	m, closeDB := openMessages(t)
	defer closeDB()

	appendN(t, m, 5)

	store := checkpoint.NewBlobStore(mem.NewBlobStore("cp"), "cursor")

	// First pass: process seqs 1..5, commit each immediately.
	stream, cp, err := checkpoint.Tail(m, []ndb.TypeID{typeID}, store, checkpoint.Options{SaveEvery: 1})
	if err != nil {
		t.Fatalf("tail: %v", err)
	}
	got := drainHistory(t, stream, cp, 5)
	if len(got) != 5 || got[0] != 1 || got[4] != 5 {
		t.Fatalf("first pass seqs = %v, want 1..5", got)
	}
	if err := cp.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	if committed, _ := store.Load(); committed != 5 {
		t.Fatalf("committed = %d, want 5", committed)
	}

	// More events arrive.
	appendN(t, m, 3) // seqs 6,7,8

	// Second pass (simulated restart): must resume at 6, NOT re-deliver 1..5.
	stream2, cp2, err := checkpoint.Tail(m, []ndb.TypeID{typeID}, store, checkpoint.Options{SaveEvery: 1})
	if err != nil {
		t.Fatalf("tail2: %v", err)
	}
	got2 := drainHistory(t, stream2, cp2, 3)
	if len(got2) != 3 || got2[0] != 6 || got2[2] != 8 {
		t.Fatalf("second pass seqs = %v, want 6..8", got2)
	}
}

func TestAtLeastOnceOnCrash(t *testing.T) {
	m, closeDB := openMessages(t)
	defer closeDB()

	appendN(t, m, 5)

	store := checkpoint.NewBlobStore(mem.NewBlobStore("cp"), "cursor")

	// First pass: process 1..5 but only commit up to 3 durably (simulate a crash
	// after the side effect of 3 but before committing 4 and 5).
	stream, cp, err := checkpoint.Tail(m, []ndb.TypeID{typeID}, store, checkpoint.Options{SaveEvery: 1})
	if err != nil {
		t.Fatalf("tail: %v", err)
	}
	got := drainHistory(t, stream, cp, 3) // commits 1,2,3
	if got[len(got)-1] != 3 {
		t.Fatalf("expected to stop at 3, got %v", got)
	}
	// No Flush of 4/5: crash.

	// Restart: must re-deliver from 4 (at-least-once), so 4 and 5 reappear.
	stream2, cp2, err := checkpoint.Tail(m, []ndb.TypeID{typeID}, store, checkpoint.Options{SaveEvery: 1})
	if err != nil {
		t.Fatalf("tail2: %v", err)
	}
	got2 := drainHistory(t, stream2, cp2, 2)
	if len(got2) != 2 || got2[0] != 4 || got2[1] != 5 {
		t.Fatalf("restart seqs = %v, want 4,5", got2)
	}
}

func TestBatchingReducesSaves(t *testing.T) {
	m, closeDB := openMessages(t)
	defer closeDB()

	appendN(t, m, 10)

	cs := &countingStore{inner: checkpoint.NewBlobStore(mem.NewBlobStore("cp"), "cursor")}

	// SaveEvery=4, interval disabled: expect flushes at 4 and 8 during the loop
	// (2 saves), then a final Flush persists 10 (3rd save).
	stream, cp, err := checkpoint.Tail(m, []ndb.TypeID{typeID}, cs, checkpoint.Options{SaveEvery: 4, SaveInterval: -1})
	if err != nil {
		t.Fatalf("tail: %v", err)
	}
	drainHistory(t, stream, cp, 10)

	if cs.saves != 2 {
		t.Fatalf("saves during loop = %d, want 2", cs.saves)
	}
	if err := cp.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	if cs.saves != 3 || cs.last != 10 {
		t.Fatalf("after flush saves=%d last=%d, want 3 and 10", cs.saves, cs.last)
	}
}

func TestBlobStoreRoundTrip(t *testing.T) {
	store := checkpoint.NewBlobStore(mem.NewBlobStore("cp"), "cursor")

	// empty store => 0
	if got, err := store.Load(); err != nil || got != 0 {
		t.Fatalf("empty load = %d, %v; want 0, nil", got, err)
	}

	if err := store.Save(ndb.Seq(42)); err != nil {
		t.Fatalf("save: %v", err)
	}
	if got, err := store.Load(); err != nil || got != 42 {
		t.Fatalf("load = %d, %v; want 42, nil", got, err)
	}

	// overwrite
	if err := store.Save(ndb.Seq(1 << 40)); err != nil {
		t.Fatalf("save2: %v", err)
	}
	if got, err := store.Load(); err != nil || got != ndb.Seq(1<<40) {
		t.Fatalf("load2 = %d, %v; want %d", got, err, ndb.Seq(1<<40))
	}
}

func TestCommitMonotonic(t *testing.T) {
	cs := &countingStore{inner: checkpoint.NewBlobStore(mem.NewBlobStore("cp"), "cursor")}
	cp := commitOnlyTail(t, cs)

	if err := cp.Commit(5); err != nil {
		t.Fatal(err)
	}
	if err := cp.Commit(3); err != nil { // lower: ignored
		t.Fatal(err)
	}
	if err := cp.Commit(5); err != nil { // equal: ignored
		t.Fatal(err)
	}
	if err := cp.Flush(); err != nil {
		t.Fatal(err)
	}
	if cs.last != 5 {
		t.Fatalf("committed last = %d, want 5", cs.last)
	}
}

// commitOnlyTail builds a Committer over an empty in-memory source purely to
// exercise Commit/Flush without needing live messages.
func commitOnlyTail(t *testing.T, store checkpoint.Store) *checkpoint.Committer {
	t.Helper()
	m, closeDB := openMessages(t)
	t.Cleanup(closeDB)
	_, cp, err := checkpoint.Tail(m, []ndb.TypeID{typeID}, store, checkpoint.Options{SaveEvery: 1, SaveInterval: -1})
	if err != nil {
		t.Fatalf("tail: %v", err)
	}
	return cp
}

func TestRunResumesAfterRestart(t *testing.T) {
	m, closeDB := openMessages(t)
	defer closeDB()

	appendN(t, m, 4) // seqs 1..4

	// A single shared blob store survives the simulated restart.
	blobs := mem.NewBlobStore("cp")
	const key = "cursor"

	var seen1 []ndb.Seq
	ctx1, cancel1 := context.WithCancel(context.Background())
	err := checkpoint.Run(ctx1, m, []ndb.TypeID{typeID}, blobs, key,
		checkpoint.Options{SaveEvery: 1},
		func(_ ndb.TypeID, msg ndb.Message) error {
			seen1 = append(seen1, msg.Seq)
			if msg.Seq == 4 {
				cancel1() // stop after processing all history
			}
			return nil
		})
	cancel1()
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("first run err = %v, want context.Canceled", err)
	}
	if len(seen1) != 4 || seen1[0] != 1 || seen1[3] != 4 {
		t.Fatalf("first run seqs = %v, want 1..4", seen1)
	}

	appendN(t, m, 2) // seqs 5,6

	// Restart: same blob store/key must resume at 5, not re-deliver 1..4.
	var seen2 []ndb.Seq
	ctx2, cancel2 := context.WithCancel(context.Background())
	err = checkpoint.Run(ctx2, m, []ndb.TypeID{typeID}, blobs, key,
		checkpoint.Options{SaveEvery: 1},
		func(_ ndb.TypeID, msg ndb.Message) error {
			seen2 = append(seen2, msg.Seq)
			if msg.Seq == 6 {
				cancel2()
			}
			return nil
		})
	cancel2()
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("second run err = %v, want context.Canceled", err)
	}
	if len(seen2) != 2 || seen2[0] != 5 || seen2[1] != 6 {
		t.Fatalf("second run seqs = %v, want 5,6", seen2)
	}
}

func TestRunHandlerErrorStopsBeforeCommit(t *testing.T) {
	m, closeDB := openMessages(t)
	defer closeDB()

	appendN(t, m, 5)

	blobs := mem.NewBlobStore("cp")
	const key = "cursor"
	boom := errors.New("boom")

	// Fail on seq 3: 1 and 2 commit, 3 does not.
	err := checkpoint.Run(context.Background(), m, []ndb.TypeID{typeID}, blobs, key,
		checkpoint.Options{SaveEvery: 1},
		func(_ ndb.TypeID, msg ndb.Message) error {
			if msg.Seq == 3 {
				return boom
			}
			return nil
		})
	if !errors.Is(err, boom) {
		t.Fatalf("run err = %v, want boom", err)
	}

	// Restart with a well-behaved handler: 3,4,5 must be re-delivered (3 was
	// never committed = at-least-once).
	var seen []ndb.Seq
	ctx, cancel := context.WithCancel(context.Background())
	_ = checkpoint.Run(ctx, m, []ndb.TypeID{typeID}, blobs, key,
		checkpoint.Options{SaveEvery: 1},
		func(_ ndb.TypeID, msg ndb.Message) error {
			seen = append(seen, msg.Seq)
			if msg.Seq == 5 {
				cancel()
			}
			return nil
		})
	cancel()
	if len(seen) != 3 || seen[0] != 3 || seen[2] != 5 {
		t.Fatalf("restart seqs = %v, want 3,4,5", seen)
	}
}

func TestRunCtxCancelOnSilentEdge(t *testing.T) {
	m, closeDB := openMessages(t)
	defer closeDB()

	appendN(t, m, 2)

	blobs := mem.NewBlobStore("cp")
	const key = "cursor"

	processed := make(chan struct{}, 8)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- checkpoint.Run(ctx, m, []ndb.TypeID{typeID}, blobs, key,
			checkpoint.Options{SaveEvery: 1},
			func(_ ndb.TypeID, msg ndb.Message) error {
				processed <- struct{}{}
				return nil
			})
	}()

	// Wait until both history events are processed; Run now blocks on the edge.
	for i := 0; i < 2; i++ {
		select {
		case <-processed:
		case <-time.After(2 * time.Second):
			t.Fatal("Run did not process history")
		}
	}

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run err = %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after ctx cancel on a silent edge")
	}

	// The cursor must have been flushed to 2 on exit.
	store := checkpoint.NewBlobStore(blobs, key)
	if got, _ := store.Load(); got != 2 {
		t.Fatalf("committed cursor = %d, want 2", got)
	}
}

func TestConsumerWaitForReadYourWrite(t *testing.T) {
	m, closeDB := openMessages(t)
	defer closeDB()

	blobs := mem.NewBlobStore("cp")
	c := checkpoint.NewConsumer(m, []ndb.TypeID{typeID}, blobs, "cursor",
		checkpoint.Options{SaveEvery: 1},
		func(_ ndb.TypeID, _ ndb.Message) error { return nil })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = c.Run(ctx) }()

	// Append a fresh event and learn its Seq (read-your-write target).
	var trace [16]byte
	seq := option.Must(m.Append(typeID, trace, []byte("x")))

	wctx, wcancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer wcancel()
	if err := c.WaitFor(wctx, seq); err != nil {
		t.Fatalf("WaitFor(%d) = %v, want nil", seq, err)
	}
	if got := c.Processed(); got < seq {
		t.Fatalf("Processed = %d after WaitFor(%d)", got, seq)
	}
}

func TestConsumerWaitForCtxCancel(t *testing.T) {
	m, closeDB := openMessages(t)
	defer closeDB()

	blobs := mem.NewBlobStore("cp")
	// Handler blocks forever, so the watermark never reaches the target.
	block := make(chan struct{})
	c := checkpoint.NewConsumer(m, []ndb.TypeID{typeID}, blobs, "cursor",
		checkpoint.Options{SaveEvery: 1},
		func(_ ndb.TypeID, _ ndb.Message) error { <-block; return nil })

	runCtx, runCancel := context.WithCancel(context.Background())
	go func() { _ = c.Run(runCtx) }()

	var trace [16]byte
	option.Must(m.Append(typeID, trace, []byte("x")))

	wctx, wcancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer wcancel()
	if err := c.WaitFor(wctx, 1); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WaitFor err = %v, want DeadlineExceeded", err)
	}

	close(block)
	runCancel()
}

func TestConsumerWatermarkStopsAtHandlerError(t *testing.T) {
	m, closeDB := openMessages(t)
	defer closeDB()

	appendN(t, m, 5)

	blobs := mem.NewBlobStore("cp")
	boom := errors.New("boom")
	c := checkpoint.NewConsumer(m, []ndb.TypeID{typeID}, blobs, "cursor",
		checkpoint.Options{SaveEvery: 1},
		func(_ ndb.TypeID, msg ndb.Message) error {
			if msg.Seq == 3 {
				return boom
			}
			return nil
		})

	if err := c.Run(context.Background()); !errors.Is(err, boom) {
		t.Fatalf("Run err = %v, want boom", err)
	}

	// Watermark advanced to 2 (seqs 1,2 succeeded) but NOT to 3 (handler failed).
	if got := c.Processed(); got != 2 {
		t.Fatalf("Processed = %d, want 2 (must not include the failed seq 3)", got)
	}
}

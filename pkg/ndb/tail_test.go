package ndb_test

import (
	"context"
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/ndb"
	"go.wdy.de/nago/pkg/ndb/msgstore"
)

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

// collector drains a Tail stream on its own goroutine into a slice of seqs.
type collector struct {
	mu   sync.Mutex
	seqs []ndb.Seq
}

func (c *collector) snapshot() []ndb.Seq {
	c.mu.Lock()
	defer c.mu.Unlock()
	return slices.Clone(c.seqs)
}

func (c *collector) waitForLen(t *testing.T, n int, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		c.mu.Lock()
		got := len(c.seqs)
		c.mu.Unlock()
		if got >= n {
			return true
		}
		time.Sleep(time.Millisecond)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.seqs) >= n
}

func TestTailReplayThenLive(t *testing.T) {
	m, closeDB := openMessages(t)
	defer closeDB()

	const typeID ndb.TypeID = "1"
	var trace [16]byte

	// historical events before tailing
	const hist = 5
	for i := 0; i < hist; i++ {
		option.Must(m.Append(typeID, trace, []byte("h")))
	}

	c := &collector{}
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, msg := range ndb.Tail(m, []ndb.TypeID{typeID}, ndb.TailOptions{FromSeq: 1}) {
			c.mu.Lock()
			c.seqs = append(c.seqs, msg.Seq)
			c.mu.Unlock()
			select {
			case <-stop:
				return
			default:
			}
		}
	}()

	// catch-up must surface the history
	if !c.waitForLen(t, hist, 2*time.Second) {
		t.Fatalf("tail did not catch up history: got %v", c.snapshot())
	}

	// now write live events; tail must pick them up
	const live = 5
	for i := 0; i < live; i++ {
		option.Must(m.Append(typeID, trace, []byte("l")))
	}

	if !c.waitForLen(t, hist+live, 2*time.Second) {
		t.Fatalf("tail did not deliver live events: got %v", c.snapshot())
	}

	close(stop)
	// nudge the loop so it observes stop and returns
	option.Must(m.Append(typeID, trace, []byte("nudge")))
	wg.Wait()

	got := c.snapshot()
	// strictly ascending, no duplicates, starting at 1
	for i := 1; i < len(got); i++ {
		if got[i] <= got[i-1] {
			t.Fatalf("seqs not strictly ascending / duplicated: %v", got)
		}
	}
	if got[0] != 1 {
		t.Fatalf("expected to start at seq 1, got %v", got)
	}
	// must contain the first hist+live sequences
	want := make([]ndb.Seq, 0, hist+live)
	for s := ndb.Seq(1); s <= ndb.Seq(hist+live); s++ {
		want = append(want, s)
	}
	if !slices.Equal(got[:hist+live], want) {
		t.Fatalf("missing/extra events: got %v want prefix %v", got, want)
	}
}

func TestTailFromSeqSkipsOlder(t *testing.T) {
	m, closeDB := openMessages(t)
	defer closeDB()

	const typeID ndb.TypeID = "1"
	var trace [16]byte
	for i := 0; i < 10; i++ {
		option.Must(m.Append(typeID, trace, []byte("x")))
	}

	c := &collector{}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, msg := range ndb.Tail(m, nil, ndb.TailOptions{FromSeq: 6}) {
			c.mu.Lock()
			c.seqs = append(c.seqs, msg.Seq)
			n := len(c.seqs)
			c.mu.Unlock()
			if n >= 5 {
				return
			}
		}
	}()

	if !c.waitForLen(t, 5, 2*time.Second) {
		t.Fatalf("expected 5 events from seq 6, got %v", c.snapshot())
	}
	wg.Wait()

	got := c.snapshot()
	if got[0] != 6 {
		t.Fatalf("expected first delivered seq 6, got %v", got)
	}
}

func TestTailMultipleTypesOrdered(t *testing.T) {
	m, closeDB := openMessages(t)
	defer closeDB()

	const a ndb.TypeID = "1"
	const b ndb.TypeID = "2"
	var trace [16]byte

	c := &collector{}
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, msg := range ndb.Tail(m, nil, ndb.TailOptions{}) {
			c.mu.Lock()
			c.seqs = append(c.seqs, msg.Seq)
			c.mu.Unlock()
			select {
			case <-stop:
				return
			default:
			}
		}
	}()

	// interleave writes across two types
	const each = 10
	for i := 0; i < each; i++ {
		option.Must(m.Append(a, trace, []byte("a")))
		option.Must(m.Append(b, trace, []byte("b")))
	}

	if !c.waitForLen(t, 2*each, 3*time.Second) {
		t.Fatalf("tail missed interleaved events: got %d", len(c.snapshot()))
	}

	close(stop)
	option.Must(m.Append(a, trace, []byte("nudge")))
	wg.Wait()

	got := c.snapshot()
	for i := 1; i < len(got); i++ {
		if got[i] <= got[i-1] {
			t.Fatalf("global Seq order violated across types: %v", got)
		}
	}
}

func TestTailStopsOnBreak(t *testing.T) {
	m, closeDB := openMessages(t)
	defer closeDB()

	const typeID ndb.TypeID = "1"
	var trace [16]byte
	option.Must(m.Append(typeID, trace, []byte("one")))

	// Break after the first event. The defer-unsubscribe inside Tail must run,
	// so a subsequent Close does not hang and no goroutine leaks.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range ndb.Tail(m, nil, ndb.TailOptions{FromSeq: 1}) {
			break
		}
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Tail did not return after break")
	}
}

func TestTailCtxCancelOnSilentEdge(t *testing.T) {
	m, closeDB := openMessages(t)
	defer closeDB()

	const typeID ndb.TypeID = "1"
	var trace [16]byte

	// One historical event; after replaying it the tail blocks on a silent live
	// edge (no further writes). A range-break cannot fire there because the loop
	// body is never re-entered. Ctx cancellation must unblock and return.
	option.Must(m.Append(typeID, trace, []byte("one")))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	replayed := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		first := true
		for range ndb.Tail(m, nil, ndb.TailOptions{FromSeq: 1, Ctx: ctx}) {
			if first {
				first = false
				close(replayed) // seen the historical event; now blocking on the edge
			}
		}
	}()

	select {
	case <-replayed:
	case <-time.After(2 * time.Second):
		t.Fatal("Tail did not deliver the historical event")
	}

	// Give the tail a moment to settle into the blocking wait, then cancel.
	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Tail did not return after ctx cancel on a silent live edge")
	}
}

// tailCollect starts a Tail in the background, feeding a collector until the
// returned cancel func is called. It is the shared harness for the gap tests
// below, which all follow the same shape: write, delete/overwrite, observe.
func tailCollect(t *testing.T, m ndb.Messages, types []ndb.TypeID, opts ndb.TailOptions) (*collector, func()) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	opts.Ctx = ctx

	c := &collector{}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, msg := range ndb.Tail(m, types, opts) {
			c.mu.Lock()
			c.seqs = append(c.seqs, msg.Seq)
			c.mu.Unlock()
		}
	}()

	return c, func() {
		cancel()
		wg.Wait()
	}
}

// TestTailStepsOverTombstone is the regression test for a follower stalling on
// a deleted message. DeleteSeq turns a slot into a tombstone, and replay filters
// tombstones out, so the Seq is missing from every later replay. A follower that
// insists on a gap-free run of sequence numbers waits for it forever and never
// delivers anything that comes after.
func TestTailStepsOverTombstone(t *testing.T) {
	m, closeDB := openMessages(t)
	defer closeDB()

	const typeID ndb.TypeID = "1"
	var trace [16]byte

	s1 := option.Must(m.Append(typeID, trace, []byte("a")))
	s2 := option.Must(m.Append(typeID, trace, []byte("b")))
	s3 := option.Must(m.Append(typeID, trace, []byte("c")))

	// punch a permanent hole in the middle of the history
	if err := m.DeleteSeq(typeID, s2); err != nil {
		t.Fatalf("delete seq: %v", err)
	}

	c, cancel := tailCollect(t, m, []ndb.TypeID{typeID}, ndb.TailOptions{FromSeq: 1})
	defer cancel()

	if !c.waitForLen(t, 2, 2*time.Second) {
		t.Fatalf("tail stalled on tombstone: got %v, want %v", c.snapshot(), []ndb.Seq{s1, s3})
	}

	// and it must keep following past the hole
	s4 := option.Must(m.Append(typeID, trace, []byte("d")))
	s5 := option.Must(m.Append(typeID, trace, []byte("e")))

	if !c.waitForLen(t, 4, 2*time.Second) {
		t.Fatalf("tail did not resume after tombstone: got %v", c.snapshot())
	}
	if got, want := c.snapshot(), []ndb.Seq{s1, s3, s4, s5}; !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestTailStepsOverRetainedOverwrite covers the same stall without any explicit
// deletion. Put allocates a fresh global Seq and physically overwrites the
// previous value, so the superseded Seq vanishes with no tombstone left behind.
// Two Puts on one type are enough to wedge a naive follower.
func TestTailStepsOverRetainedOverwrite(t *testing.T) {
	m, closeDB := openMessages(t)
	defer closeDB()

	const typeID ndb.TypeID = "retained"
	var trace [16]byte

	option.Must(m.Put(typeID, trace, []byte("v1"))) // superseded below
	s2 := option.Must(m.Put(typeID, trace, []byte("v2")))

	c, cancel := tailCollect(t, m, []ndb.TypeID{typeID}, ndb.TailOptions{FromSeq: 1})
	defer cancel()

	if !c.waitForLen(t, 1, 2*time.Second) {
		t.Fatalf("tail stalled on overwritten retained value: got %v", c.snapshot())
	}

	s3 := option.Must(m.Put(typeID, trace, []byte("v3")))
	if !c.waitForLen(t, 2, 2*time.Second) {
		t.Fatalf("tail did not follow further puts: got %v", c.snapshot())
	}
	if got, want := c.snapshot(), []ndb.Seq{s2, s3}; !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestTailRetainedTypeDoesNotBlockAppendType pins down the blast radius. Tail
// holds a single GLOBAL Seq watermark, so a hole burned by one type stalls
// delivery for every type. A retained type being overwritten in the background
// must not stop a follower of an unrelated, purely appending type.
func TestTailRetainedTypeDoesNotBlockAppendType(t *testing.T) {
	m, closeDB := openMessages(t)
	defer closeDB()

	const retained ndb.TypeID = "retained"
	const events ndb.TypeID = "events"
	var trace [16]byte

	c, cancel := tailCollect(t, m, []ndb.TypeID{events}, ndb.TailOptions{FromSeq: 1})
	defer cancel()

	var want []ndb.Seq
	for i := range 5 {
		// interleave: every append is preceded by a retained overwrite, so the
		// event stream is riddled with holes from the other type
		option.Must(m.Put(retained, trace, []byte("v")))
		want = append(want, option.Must(m.Append(events, trace, []byte{byte(i)})))
	}

	if !c.waitForLen(t, len(want), 3*time.Second) {
		t.Fatalf("append-only type blocked by retained type: got %v, want %v", c.snapshot(), want)
	}
	if got := c.snapshot(); !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestTailStepsOverDeletedType covers the third source of permanent holes:
// DeleteType removes a whole directory, so an entire range of sequence numbers
// disappears at once.
func TestTailStepsOverDeletedType(t *testing.T) {
	m, closeDB := openMessages(t)
	defer closeDB()

	const doomed ndb.TypeID = "doomed"
	const kept ndb.TypeID = "kept"
	var trace [16]byte

	first := option.Must(m.Append(kept, trace, []byte("a")))
	for range 5 {
		option.Must(m.Append(doomed, trace, []byte("x")))
	}

	if err := m.DeleteType(doomed); err != nil {
		t.Fatalf("delete type: %v", err)
	}

	c, cancel := tailCollect(t, m, []ndb.TypeID{kept}, ndb.TailOptions{FromSeq: 1})
	defer cancel()

	last := option.Must(m.Append(kept, trace, []byte("b")))

	if !c.waitForLen(t, 2, 2*time.Second) {
		t.Fatalf("tail stalled on deleted type: got %v", c.snapshot())
	}
	if got, want := c.snapshot(), []ndb.Seq{first, last}; !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestTailWaitsForInFlightWrite is the counterweight to the tests above: a gap
// must NOT be skipped just because it is a gap.
//
// The engine allocates a Seq before the write completes, so a concurrently
// written message is briefly missing from a replay exactly like a deleted one.
// Skipping it would silently drop a live event. This test hammers many parallel
// writers across several types while a follower runs, then asserts that every
// single appended Seq was delivered, in order.
func TestTailWaitsForInFlightWrite(t *testing.T) {
	m, closeDB := openMessages(t)
	defer closeDB()

	const writers = 8
	const perWriter = 50
	var trace [16]byte

	types := make([]ndb.TypeID, writers)
	for i := range types {
		types[i] = ndb.TypeID("t" + strconv.Itoa(i))
	}

	c, cancel := tailCollect(t, m, nil, ndb.TailOptions{FromSeq: 1})
	defer cancel()

	var mu sync.Mutex
	var appended []ndb.Seq

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(writers)
	for w := range writers {
		go func() {
			defer wg.Done()
			<-start
			for range perWriter {
				seq := option.Must(m.Append(types[w], trace, []byte("p")))
				mu.Lock()
				appended = append(appended, seq)
				mu.Unlock()
			}
		}()
	}
	close(start)
	wg.Wait()

	total := writers * perWriter
	if !c.waitForLen(t, total, 5*time.Second) {
		t.Fatalf("tail delivered %d of %d events: %v", len(c.snapshot()), total, c.snapshot())
	}

	mu.Lock()
	slices.Sort(appended)
	mu.Unlock()

	got := c.snapshot()
	if !slices.Equal(got, appended) {
		t.Fatalf("tail dropped or reordered in-flight writes:\n got %v\nwant %v", got, appended)
	}
}

package ndb_test

import (
	"slices"
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
	db := option.Must(ndb.Open(root, ndb.Options{DefaultKind: "msgstore"}))
	eng, err := db.Engine("events", ndb.EngineOptions{Config: msgstore.Options{Compress: msgstore.NoCompression}})
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

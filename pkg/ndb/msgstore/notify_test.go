package msgstore_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/ndb"
	"go.wdy.de/nago/pkg/ndb/msgstore"
)

// waitFor polls cond until it is true or the timeout elapses.
func waitFor(t *testing.T, timeout time.Duration, cond func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(time.Millisecond)
	}
	return cond()
}

func TestSubscribeAppendDelivers(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{Compress: msgstore.NoCompression}))
	defer func() { option.MustZero(db.Close()) }()

	const typeID msgstore.TypeID = 7
	var trace [16]byte
	trace[0] = 0xAB

	var mu sync.Mutex
	var got []ndb.Notification
	closeSub := db.Subscribe([]msgstore.TypeID{typeID}, func(n ndb.Notification) {
		mu.Lock()
		got = append(got, n)
		mu.Unlock()
	})
	defer closeSub()

	seq := option.Must(db.Append(typeID, trace, []byte("hello")))

	ok := waitFor(t, time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(got) == 1
	})
	if !ok {
		t.Fatalf("notification not delivered")
	}

	mu.Lock()
	defer mu.Unlock()
	n := got[0]
	if n.Type != typeID {
		t.Errorf("type: got %d want %d", n.Type, typeID)
	}
	if n.Seq != seq {
		t.Errorf("seq: got %d want %d", n.Seq, seq)
	}
	if n.TraceID != trace {
		t.Errorf("traceID not propagated: got %x", n.TraceID)
	}
	if n.TimeNano == 0 {
		t.Error("timeNano not set")
	}
}

func TestSubscribePutDelivers(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{Compress: msgstore.NoCompression}))
	defer func() { option.MustZero(db.Close()) }()

	const typeID msgstore.TypeID = 3
	var trace [16]byte

	var count atomic.Int64
	closeSub := db.Subscribe(nil, func(n ndb.Notification) {
		count.Add(1)
	})
	defer closeSub()

	option.Must(db.Put(typeID, trace, []byte("v1")))
	option.Must(db.Put(typeID, trace, []byte("v2")))

	if !waitFor(t, time.Second, func() bool { return count.Load() == 2 }) {
		t.Fatalf("expected 2 put notifications, got %d", count.Load())
	}
}

func TestSubscribeTypeFilter(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{Compress: msgstore.NoCompression}))
	defer func() { option.MustZero(db.Close()) }()

	const wanted msgstore.TypeID = 1
	const other msgstore.TypeID = 2
	var trace [16]byte

	var wantedCount, otherCount atomic.Int64
	closeSub := db.Subscribe([]msgstore.TypeID{wanted}, func(n ndb.Notification) {
		switch n.Type {
		case wanted:
			wantedCount.Add(1)
		case other:
			otherCount.Add(1)
		}
	})
	defer closeSub()

	option.Must(db.Append(other, trace, []byte("a")))
	option.Must(db.Append(wanted, trace, []byte("b")))
	option.Must(db.Append(other, trace, []byte("c")))

	if !waitFor(t, time.Second, func() bool { return wantedCount.Load() == 1 }) {
		t.Fatalf("expected 1 notification for wanted type, got %d", wantedCount.Load())
	}
	// give any erroneous deliveries a chance to arrive
	time.Sleep(50 * time.Millisecond)
	if otherCount.Load() != 0 {
		t.Fatalf("subscriber received %d notifications for filtered-out type", otherCount.Load())
	}
}

func TestSubscribeCloseStopsDelivery(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{Compress: msgstore.NoCompression}))
	defer func() { option.MustZero(db.Close()) }()

	const typeID msgstore.TypeID = 5
	var trace [16]byte

	var count atomic.Int64
	closeSub := db.Subscribe(nil, func(n ndb.Notification) {
		count.Add(1)
	})

	option.Must(db.Append(typeID, trace, []byte("before")))
	if !waitFor(t, time.Second, func() bool { return count.Load() == 1 }) {
		t.Fatalf("expected first notification")
	}

	closeSub()
	closeSub() // idempotent: must not panic

	before := count.Load()
	for i := 0; i < 10; i++ {
		option.Must(db.Append(typeID, trace, []byte("after")))
	}
	time.Sleep(50 * time.Millisecond)
	if count.Load() != before {
		t.Fatalf("delivery continued after close: %d -> %d", before, count.Load())
	}
}

func TestSubscribeSynchronousDelivery(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{Compress: msgstore.NoCompression}))
	defer func() { option.MustZero(db.Close()) }()

	const typeID msgstore.TypeID = 9
	var trace [16]byte

	// Delivery is synchronous: by the time Append returns, the callback has run.
	var delivered atomic.Bool
	closeSub := db.Subscribe(nil, func(n ndb.Notification) {
		delivered.Store(true)
	})
	defer closeSub()

	option.Must(db.Append(typeID, trace, []byte("x")))
	if !delivered.Load() {
		t.Fatal("expected synchronous delivery before Append returned")
	}
}

func TestSubscribeOrderMatchesAppend(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{Compress: msgstore.NoCompression}))
	defer func() { option.MustZero(db.Close()) }()

	const typeID msgstore.TypeID = 4
	var trace [16]byte

	var mu sync.Mutex
	var seqs []ndb.Seq
	closeSub := db.Subscribe(nil, func(n ndb.Notification) {
		mu.Lock()
		seqs = append(seqs, n.Seq)
		mu.Unlock()
	})
	defer closeSub()

	const n = 50
	for i := 0; i < n; i++ {
		option.Must(db.Append(typeID, trace, []byte("x")))
	}

	mu.Lock()
	defer mu.Unlock()
	if len(seqs) != n {
		t.Fatalf("expected %d synchronous notifications, got %d", n, len(seqs))
	}
	for i := 1; i < len(seqs); i++ {
		if seqs[i] <= seqs[i-1] {
			t.Fatalf("notifications out of order: %v", seqs)
		}
	}
}

func TestSubscribePanicIsContained(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{Compress: msgstore.NoCompression}))
	defer func() { option.MustZero(db.Close()) }()

	const typeID msgstore.TypeID = 11
	var trace [16]byte

	// A panicking subscriber must not break the writer or other subscribers.
	closePanic := db.Subscribe(nil, func(n ndb.Notification) {
		panic("boom")
	})
	defer closePanic()

	var good atomic.Int64
	closeGood := db.Subscribe(nil, func(n ndb.Notification) {
		good.Add(1)
	})
	defer closeGood()

	// Append must succeed despite the panicking subscriber.
	if _, err := db.Append(typeID, trace, []byte("x")); err != nil {
		t.Fatalf("append failed due to panicking subscriber: %v", err)
	}
	if good.Load() != 1 {
		t.Fatalf("well-behaved subscriber did not run: %d", good.Load())
	}
}

func TestSubscribeAfterCloseIsNoop(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{Compress: msgstore.NoCompression}))

	// Close the store, then subscribe: must return a usable no-op closer and
	// never deliver anything.
	option.MustZero(db.Close())

	var count atomic.Int64
	closeSub := db.Subscribe(nil, func(n ndb.Notification) { count.Add(1) })
	closeSub() // must not panic
	if count.Load() != 0 {
		t.Fatalf("unexpected delivery after close: %d", count.Load())
	}
}

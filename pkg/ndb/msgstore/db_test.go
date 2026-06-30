package msgstore_test

import (
	"math"
	"slices"
	"testing"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/ndb/msgstore"
)

func TestOpenClose(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{}))
	option.MustZero(db.Close())
}

func TestAppendAndReplay(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))
	defer func() { option.MustZero(db.Close()) }()

	traceID := msgstore.NewTraceID()

	const eventType msgstore.TypeID = "42"
	const count = 100

	// append events
	for i := range count {
		payload := []byte("hello world " + string(rune('A'+i%26)))
		seq := option.Must(db.Append(eventType, traceID, payload))
		if seq != msgstore.Seq(i+1) {
			t.Fatalf("expected seq %d, got %d", i+1, seq)
		}
	}

	// replay all
	var replayed int
	for _, msg := range db.Replay(nil, 1, math.MaxUint64) {
		_ = msg
		replayed++
	}
	if replayed != count {
		t.Fatalf("expected %d messages, got %d", count, replayed)
	}

	// replay filtered by type
	replayed = 0
	for _, msg := range db.Replay([]msgstore.TypeID{eventType}, 1, math.MaxUint64) {
		_ = msg
		replayed++
	}
	if replayed != count {
		t.Fatalf("expected %d messages for type %q, got %d", count, eventType, replayed)
	}

	// replay non-existent type
	replayed = 0
	for _, msg := range db.Replay([]msgstore.TypeID{"999"}, 1, math.MaxUint64) {
		_ = msg
		replayed++
	}
	if replayed != 0 {
		t.Fatalf("expected 0 messages for type 999, got %d", replayed)
	}
}

func TestAppendWithS2Compression(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.AlwaysS2,
	}))
	defer func() { option.MustZero(db.Close()) }()

	traceID := msgstore.NewTraceID()

	payload := make([]byte, 4096)
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	seq := option.Must(db.Append("1", traceID, payload))
	if seq != 1 {
		t.Fatalf("expected seq 1, got %d", seq)
	}

	// replay and verify decompression happens
	var found bool
	for _, msg := range db.Replay([]msgstore.TypeID{"1"}, 1, 1) {
		found = true
		if msg.Encoding != msgstore.EncodingS2 {
			t.Fatalf("expected S2 encoding, got %d", msg.Encoding)
		}
		decompressed := option.Must(msgstore.Decompress(msg.Encoding, slices.Clone(msg.Payload), msg.UncompressedLen))
		if len(decompressed) != len(payload) {
			t.Fatalf("expected %d bytes, got %d", len(payload), len(decompressed))
		}
	}
	if !found {
		t.Fatal("no message found in replay")
	}
}

func TestSplitByCount(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress:    msgstore.NoCompression,
		ShouldSplit: msgstore.SplitByCount(5),
	}))
	defer func() { option.MustZero(db.Close()) }()

	var traceID [16]byte
	const eventType msgstore.TypeID = "1"

	for i := range 12 {
		option.Must(db.Append(eventType, traceID, []byte("msg")))
		_ = i
	}

	// replay all and verify all 12 messages are readable
	var count int
	for _, msg := range db.Replay(nil, 1, math.MaxUint64) {
		_ = msg
		count++
	}
	if count != 12 {
		t.Fatalf("expected 12, got %d", count)
	}
}

func TestExclusiveLock(t *testing.T) {
	dir := t.TempDir()
	db1 := option.Must(msgstore.Open(dir, msgstore.Options{}))
	defer func() { option.MustZero(db1.Close()) }()

	// lockedfile.Create blocks until the lock is released, so we use a
	// goroutine with a timeout to verify it does not acquire immediately.
	done := make(chan error, 1)
	go func() {
		db2, err := msgstore.Open(dir, msgstore.Options{})
		if err != nil {
			done <- err
			return
		}
		option.MustZero(db2.Close())
		done <- nil
	}()

	select {
	case err := <-done:
		// If it returned immediately without error, locking failed
		if err == nil {
			t.Fatal("expected second Open to block, but it succeeded immediately")
		}
		// An error is also acceptable (means locking worked)
	case <-time.After(200 * time.Millisecond):
		// Timed out = the second Open is blocking as expected
	}
}

func TestReopenAndBootstrap(t *testing.T) {
	dir := t.TempDir()

	// first session: write some events
	db := option.Must(msgstore.Open(dir, msgstore.Options{Compress: msgstore.NoCompression}))

	var traceID [16]byte
	for i := range 10 {
		option.Must(db.Append("1", traceID, []byte("event")))
		_ = i
	}
	option.MustZero(db.Close())

	// second session: verify bootstrap picks up correct nextSeq
	db2 := option.Must(msgstore.Open(dir, msgstore.Options{Compress: msgstore.NoCompression}))
	defer func() { option.MustZero(db2.Close()) }()

	seq := option.Must(db2.Append("1", traceID, []byte("after reopen")))
	if seq != 11 {
		t.Fatalf("expected seq 11 after reopen, got %d", seq)
	}
}

func TestMultipleEventTypes(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{Compress: msgstore.NoCompression}))
	defer func() { option.MustZero(db.Close()) }()

	var traceID [16]byte
	for i := range 5 {
		option.Must(db.Append(msgstore.TypeID("1"), traceID, []byte("type1")))
		option.Must(db.Append(msgstore.TypeID("2"), traceID, []byte("type2")))
		_ = i
	}

	// replay only type 1
	var count int
	for _, _ = range db.Replay([]msgstore.TypeID{"1"}, 1, math.MaxUint64) {
		count++
	}
	if count != 5 {
		t.Fatalf("expected 5 messages for type 1, got %d", count)
	}

	// replay only type 2
	count = 0
	for _, _ = range db.Replay([]msgstore.TypeID{"2"}, 1, math.MaxUint64) {
		count++
	}
	if count != 5 {
		t.Fatalf("expected 5 messages for type 2, got %d", count)
	}
}

func TestReplayGlobalSequenceOrder(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{Compress: msgstore.NoCompression}))
	defer func() { option.MustZero(db.Close()) }()

	var traceID [16]byte

	// interleave Login (type 1) and Logout (type 2) events
	option.Must(db.Append(msgstore.TypeID("1"), traceID, []byte("login-alice")))   // seq 1
	option.Must(db.Append(msgstore.TypeID("1"), traceID, []byte("login-bob")))     // seq 2
	option.Must(db.Append(msgstore.TypeID("2"), traceID, []byte("logout-alice")))  // seq 3
	option.Must(db.Append(msgstore.TypeID("1"), traceID, []byte("login-charlie"))) // seq 4
	option.Must(db.Append(msgstore.TypeID("2"), traceID, []byte("logout-bob")))    // seq 5

	// replay both types: must come out in strict sequence order 1,2,3,4,5
	var seqIDs []msgstore.Seq
	var typeIDs []msgstore.TypeID
	for tid, msg := range db.Replay(nil, 1, math.MaxUint64) {
		seqIDs = append(seqIDs, msg.Seq)
		typeIDs = append(typeIDs, tid)
	}

	if len(seqIDs) != 5 {
		t.Fatalf("expected 5 messages, got %d", len(seqIDs))
	}

	// verify strict ascending sequence order
	for i := 1; i < len(seqIDs); i++ {
		if seqIDs[i] <= seqIDs[i-1] {
			t.Fatalf("sequence order violated at index %d: %d <= %d", i, seqIDs[i], seqIDs[i-1])
		}
	}

	// verify expected type interleaving: 1,1,2,1,2
	expectedTypes := []msgstore.TypeID{"1", "1", "2", "1", "2"}
	for i, tid := range typeIDs {
		if tid != expectedTypes[i] {
			t.Fatalf("wrong type at index %d: got %q, want %q", i, tid, expectedTypes[i])
		}
	}
}

func TestPutAndGet(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{Compress: msgstore.NoCompression}))
	defer func() { option.MustZero(db.Close()) }()

	var traceID [16]byte
	const eventType msgstore.TypeID = "100"

	// Get on empty type returns an empty option (no error)
	empty, err := db.Get(eventType)
	if err != nil {
		t.Fatalf("unexpected error on empty type: %v", err)
	}
	if empty.IsSome() {
		t.Fatal("expected empty option on empty type")
	}

	// Put multiple times – each overwrites the previous
	var lastSeq msgstore.Seq
	for i := range 10 {
		payload := []byte("value-" + string(rune('A'+i)))
		seq := option.Must(db.Put(eventType, traceID, payload))
		if seq <= lastSeq {
			t.Fatalf("sequence ID must increase: got %d, previous %d", seq, lastSeq)
		}
		lastSeq = seq

		// Get must always return the latest value
		msg := option.Must(db.Get(eventType)).Unwrap()
		if string(msg.Payload) != string(payload) {
			t.Fatalf("Get after Put %d: got %q, want %q", i, msg.Payload, payload)
		}
		if msg.Seq != seq {
			t.Fatalf("Get seqID: got %d, want %d", msg.Seq, seq)
		}
	}
}

func TestPutShrinkPayload(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{Compress: msgstore.NoCompression}))
	defer func() { option.MustZero(db.Close()) }()

	var traceID [16]byte
	const eventType msgstore.TypeID = "200"

	// Put a large payload
	bigPayload := make([]byte, 4096)
	for i := range bigPayload {
		bigPayload[i] = byte(i % 256)
	}
	option.Must(db.Put(eventType, traceID, bigPayload))

	// Put a small payload – overwrites the large one
	smallPayload := []byte("tiny")
	seq := option.Must(db.Put(eventType, traceID, smallPayload))

	// Get must return the small payload, not stale data
	msg := option.Must(db.Get(eventType)).Unwrap()
	if string(msg.Payload) != string(smallPayload) {
		t.Fatalf("expected %q, got %q", smallPayload, msg.Payload)
	}
	if msg.Seq != seq {
		t.Fatalf("expected seq %d, got %d", seq, msg.Seq)
	}
}

func TestPutReopenBootstrap(t *testing.T) {
	dir := t.TempDir()

	var traceID [16]byte
	const eventType msgstore.TypeID = "300"

	// first session: Put a value
	db := option.Must(msgstore.Open(dir, msgstore.Options{Compress: msgstore.NoCompression}))
	seq1 := option.Must(db.Put(eventType, traceID, []byte("retained-value")))
	option.MustZero(db.Close())

	// second session: Get must return the value, next seq must continue
	db2 := option.Must(msgstore.Open(dir, msgstore.Options{Compress: msgstore.NoCompression}))
	defer func() { option.MustZero(db2.Close()) }()

	msg := option.Must(db2.Get(eventType)).Unwrap()
	if string(msg.Payload) != "retained-value" {
		t.Fatalf("expected %q after reopen, got %q", "retained-value", msg.Payload)
	}
	if msg.Seq != seq1 {
		t.Fatalf("expected seq %d after reopen, got %d", seq1, msg.Seq)
	}

	// next Put must get a higher sequence ID
	seq2 := option.Must(db2.Put(eventType, traceID, []byte("updated")))
	if seq2 <= seq1 {
		t.Fatalf("expected seq > %d, got %d", seq1, seq2)
	}
}

func TestGetAfterAppend(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{Compress: msgstore.NoCompression}))
	defer func() { option.MustZero(db.Close()) }()

	var traceID [16]byte
	const eventType msgstore.TypeID = "400"

	// Append a history of events
	for i := range 20 {
		option.Must(db.Append(eventType, traceID, []byte("event-"+string(rune('A'+i%26)))))
	}

	// Get must return the last appended event
	msg := option.Must(db.Get(eventType)).Unwrap()
	if msg.Seq != 20 {
		t.Fatalf("expected seq 20, got %d", msg.Seq)
	}
	if string(msg.Payload) != "event-T" {
		t.Fatalf("expected %q, got %q", "event-T", msg.Payload)
	}
}

func TestReplayAfterPut(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{Compress: msgstore.NoCompression}))
	defer func() { option.MustZero(db.Close()) }()

	var traceID [16]byte
	const eventType msgstore.TypeID = "500"

	// Put several times – only last value is physically present
	for i := range 5 {
		option.Must(db.Put(eventType, traceID, []byte("v"+string(rune('0'+i)))))
	}

	// Replay must yield exactly one message (the current retained value)
	var count int
	var lastMsg msgstore.Message
	for _, msg := range db.Replay([]msgstore.TypeID{eventType}, 0, math.MaxUint64) {
		lastMsg = msg
		lastMsg.Payload = slices.Clone(msg.Payload)
		count++
	}
	if count != 1 {
		t.Fatalf("expected 1 message in replay after Put, got %d", count)
	}
	if string(lastMsg.Payload) != "v4" {
		t.Fatalf("expected %q, got %q", "v4", lastMsg.Payload)
	}
}

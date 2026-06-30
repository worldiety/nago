package msgstore_test

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"

	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/ndb/msgstore"
)

// ---------- DeleteByType ----------

func TestDeleteByType(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))
	defer func() { option.MustZero(db.Close()) }()

	var traceID [16]byte
	const typeA msgstore.TypeID = "1"
	const typeB msgstore.TypeID = "2"

	// append events to two different types
	for i := range 5 {
		option.Must(db.Append(typeA, traceID, []byte("typeA-"+strconv.Itoa(i))))
		option.Must(db.Append(typeB, traceID, []byte("typeB-"+strconv.Itoa(i))))
	}

	// sanity check: both types present
	if c := countReplayForType(db, typeA); c != 5 {
		t.Fatalf("before delete: expected 5 typeA messages, got %d", c)
	}
	if c := countReplayForType(db, typeB); c != 5 {
		t.Fatalf("before delete: expected 5 typeB messages, got %d", c)
	}

	// delete typeA
	option.MustZero(db.DeleteType(typeA))

	// typeA should be gone
	if c := countReplayForType(db, typeA); c != 0 {
		t.Fatalf("after delete: expected 0 typeA messages, got %d", c)
	}

	// typeB should be unaffected
	if c := countReplayForType(db, typeB); c != 5 {
		t.Fatalf("after delete: expected 5 typeB messages, got %d", c)
	}

	// the type directory should no longer exist on disk
	typeDirA := filepath.Join(dir, "events", string(typeA))
	if _, err := os.Stat(typeDirA); !os.IsNotExist(err) {
		t.Fatalf("expected type dir to be removed, stat returned: %v", err)
	}
}

func TestDeleteByTypeAllowsReappend(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))
	defer func() { option.MustZero(db.Close()) }()

	var traceID [16]byte
	const typeID msgstore.TypeID = "1"

	option.Must(db.Append(typeID, traceID, []byte("before-delete")))
	option.MustZero(db.DeleteType(typeID))

	// appending to the same type after delete should work
	seq := option.Must(db.Append(typeID, traceID, []byte("after-delete")))
	if seq == 0 {
		t.Fatal("expected non-zero sequence ID after re-append")
	}

	if c := countReplayForType(db, typeID); c != 1 {
		t.Fatalf("expected 1 message after re-append, got %d", c)
	}
}

func TestDeleteByTypeNonExistent(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))
	defer func() { option.MustZero(db.Close()) }()

	// deleting a type that was never written should not error
	if err := db.DeleteType("999"); err != nil {
		t.Fatalf("expected no error for non-existent type, got: %v", err)
	}
}

func TestDeleteByTypeWithSplitSegments(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress:    msgstore.NoCompression,
		ShouldSplit: msgstore.SplitByCount(3),
	}))
	defer func() { option.MustZero(db.Close()) }()

	var traceID [16]byte
	const typeID msgstore.TypeID = "1"

	// write enough messages to trigger multiple splits
	for i := range 10 {
		option.Must(db.Append(typeID, traceID, []byte("msg-"+strconv.Itoa(i))))
	}

	if c := countReplayForType(db, typeID); c != 10 {
		t.Fatalf("before delete: expected 10 messages, got %d", c)
	}

	option.MustZero(db.DeleteType(typeID))

	if c := countReplayForType(db, typeID); c != 0 {
		t.Fatalf("after delete: expected 0 messages, got %d", c)
	}
}

// ---------- DeleteByID ----------

func TestDeleteByID(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))
	defer func() { option.MustZero(db.Close()) }()

	var traceID [16]byte
	const typeID msgstore.TypeID = "1"

	var seqIDs []msgstore.Seq
	for i := range 5 {
		seq := option.Must(db.Append(typeID, traceID, []byte("msg-"+strconv.Itoa(i))))
		seqIDs = append(seqIDs, seq)
	}

	// delete the middle message
	option.MustZero(db.DeleteSeq(typeID, seqIDs[2]))

	// replay should return 4 messages (tombstones are filtered)
	var remaining []msgstore.Seq
	for _, msg := range db.Replay([]msgstore.TypeID{typeID}, 1, math.MaxUint64) {
		remaining = append(remaining, msg.Seq)
	}
	if len(remaining) != 4 {
		t.Fatalf("expected 4 messages after delete, got %d", len(remaining))
	}

	// the deleted seqID must not be in the result
	for _, s := range remaining {
		if s == seqIDs[2] {
			t.Fatal("deleted message still appeared in replay")
		}
	}
}

func TestDeleteByIDPayloadZeroed(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))

	var traceID [16]byte
	const typeID msgstore.TypeID = "1"

	payload := []byte("sensitive-data-that-must-be-erased")
	seq := option.Must(db.Append(typeID, traceID, payload))
	option.MustZero(db.Close())

	// reopen, delete, close
	db = option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))
	option.MustZero(db.DeleteSeq(typeID, seq))
	option.MustZero(db.Close())

	// read the raw segment file and verify the payload region is zeroed
	segPath := findSegmentFile(t, dir, typeID)
	data, err := os.ReadFile(segPath)
	if err != nil {
		t.Fatal(err)
	}

	// the original payload bytes should no longer be present
	for i := 0; i <= len(data)-len(payload); i++ {
		if slices.Equal(data[i:i+len(payload)], payload) {
			t.Fatal("original payload still found in segment file after tombstone")
		}
	}
}

func TestDeleteByIDNotFound(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))
	defer func() { option.MustZero(db.Close()) }()

	var traceID [16]byte
	const typeID msgstore.TypeID = "1"

	option.Must(db.Append(typeID, traceID, []byte("hello")))

	err := db.DeleteSeq(typeID, 9999)
	if !errors.Is(err, msgstore.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestDeleteByIDFirstAndLast(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))
	defer func() { option.MustZero(db.Close()) }()

	var traceID [16]byte
	const typeID msgstore.TypeID = "1"

	var seqIDs []msgstore.Seq
	for i := range 5 {
		seq := option.Must(db.Append(typeID, traceID, []byte("msg-"+strconv.Itoa(i))))
		seqIDs = append(seqIDs, seq)
	}

	// delete first and last
	option.MustZero(db.DeleteSeq(typeID, seqIDs[0]))
	option.MustZero(db.DeleteSeq(typeID, seqIDs[4]))

	var remaining []msgstore.Seq
	for _, msg := range db.Replay([]msgstore.TypeID{typeID}, 1, math.MaxUint64) {
		remaining = append(remaining, msg.Seq)
	}
	if len(remaining) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(remaining))
	}

	// verify the deleted IDs are absent
	for _, s := range remaining {
		if s == seqIDs[0] || s == seqIDs[4] {
			t.Fatalf("deleted seqID %d still in replay", s)
		}
	}
}

func TestDeleteByIDWithSplitSegments(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress:    msgstore.NoCompression,
		ShouldSplit: msgstore.SplitByCount(3),
	}))
	defer func() { option.MustZero(db.Close()) }()

	var traceID [16]byte
	const typeID msgstore.TypeID = "1"

	var seqIDs []msgstore.Seq
	for i := range 10 {
		seq := option.Must(db.Append(typeID, traceID, []byte("msg-"+strconv.Itoa(i))))
		seqIDs = append(seqIDs, seq)
	}

	// delete one message from the first segment and one from a later segment
	option.MustZero(db.DeleteSeq(typeID, seqIDs[1]))
	option.MustZero(db.DeleteSeq(typeID, seqIDs[7]))

	var remaining []msgstore.Seq
	for _, msg := range db.Replay([]msgstore.TypeID{typeID}, 1, math.MaxUint64) {
		remaining = append(remaining, msg.Seq)
	}
	if len(remaining) != 8 {
		t.Fatalf("expected 8 messages after deleting 2, got %d", len(remaining))
	}

	for _, s := range remaining {
		if s == seqIDs[1] || s == seqIDs[7] {
			t.Fatalf("deleted seqID %d still in replay", s)
		}
	}
}

func TestDeleteByIDIdempotent(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))
	defer func() { option.MustZero(db.Close()) }()

	var traceID [16]byte
	const typeID msgstore.TypeID = "1"

	seq := option.Must(db.Append(typeID, traceID, []byte("hello")))

	// first delete succeeds
	option.MustZero(db.DeleteSeq(typeID, seq))

	// second delete on already-tombstoned message returns not found
	// (SequenceID is now 0, so no match)
	err := db.DeleteSeq(typeID, seq)
	if !errors.Is(err, msgstore.ErrNotFound) {
		t.Fatalf("expected ErrNotFound on second delete, got: %v", err)
	}
}

func TestDeleteByIDSurvivesReopen(t *testing.T) {
	dir := t.TempDir()

	var traceID [16]byte
	const typeID msgstore.TypeID = "1"

	// write and delete
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))
	var seqIDs []msgstore.Seq
	for i := range 5 {
		seq := option.Must(db.Append(typeID, traceID, []byte("msg-"+strconv.Itoa(i))))
		seqIDs = append(seqIDs, seq)
	}
	option.MustZero(db.DeleteSeq(typeID, seqIDs[2]))
	option.MustZero(db.Close())

	// reopen and verify the tombstone persisted
	db2 := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))
	defer func() { option.MustZero(db2.Close()) }()

	var remaining []msgstore.Seq
	for _, msg := range db2.Replay([]msgstore.TypeID{typeID}, 1, math.MaxUint64) {
		remaining = append(remaining, msg.Seq)
	}
	if len(remaining) != 4 {
		t.Fatalf("expected 4 messages after reopen, got %d", len(remaining))
	}
	for _, s := range remaining {
		if s == seqIDs[2] {
			t.Fatal("tombstoned message reappeared after reopen")
		}
	}
}

// ---------- helpers ----------

func countReplayForType(db *msgstore.DB, typeID msgstore.TypeID) int {
	var count int
	for _, _ = range db.Replay([]msgstore.TypeID{typeID}, 1, math.MaxUint64) {
		count++
	}
	return count
}

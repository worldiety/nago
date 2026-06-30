package msgstore_test

import (
	"math"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/ndb/msgstore"
)

// TestRebuildTimeIndexBasic verifies that RebuildTimeIndex produces a
// functional time index that can resolve timestamps to sequence IDs.
func TestRebuildTimeIndexBasic(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))

	var traceID [16]byte
	const typeID msgstore.TypeID = 1

	before := time.Now().UnixNano()

	for i := range 10 {
		option.Must(db.Append(typeID, traceID, []byte("msg-"+strconv.Itoa(i))))
	}

	// lookup should work before rebuild
	seq, err := db.SeqForTime(before)
	if err != nil {
		t.Fatalf("SeqForTimestamp before rebuild: %v", err)
	}
	if seq < 1 || seq > 10 {
		t.Fatalf("unexpected seq before rebuild: %d", seq)
	}

	// rebuild
	option.MustZero(db.RebuildTimeIndex())

	// lookup should still work after rebuild
	seqAfter, err := db.SeqForTime(before)
	if err != nil {
		t.Fatalf("SeqForTimestamp after rebuild: %v", err)
	}
	if seqAfter < 1 || seqAfter > 10 {
		t.Fatalf("unexpected seq after rebuild: %d", seqAfter)
	}

	option.MustZero(db.Close())
}

// TestRebuildTimeIndexAfterDeleteByID verifies that RebuildTimeIndex excludes
// tombstoned messages so that the rebuilt index only contains live entries.
func TestRebuildTimeIndexAfterDeleteByID(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))

	var traceID [16]byte
	const typeID msgstore.TypeID = 1

	var seqIDs []msgstore.Seq
	for i := range 5 {
		seq := option.Must(db.Append(typeID, traceID, []byte("msg-"+strconv.Itoa(i))))
		seqIDs = append(seqIDs, seq)
	}

	// delete two messages
	option.MustZero(db.DeleteSeq(typeID, seqIDs[1]))
	option.MustZero(db.DeleteSeq(typeID, seqIDs[3]))

	// rebuild
	option.MustZero(db.RebuildTimeIndex())

	// replay should still return 3 messages
	var remaining []msgstore.Seq
	for _, msg := range db.Replay([]msgstore.TypeID{typeID}, 1, math.MaxUint64) {
		remaining = append(remaining, msg.Seq)
	}
	if len(remaining) != 3 {
		t.Fatalf("expected 3 messages after rebuild, got %d", len(remaining))
	}
	for _, s := range remaining {
		if s == seqIDs[1] || s == seqIDs[3] {
			t.Fatalf("tombstoned seqID %d still in replay after rebuild", s)
		}
	}

	option.MustZero(db.Close())
}

// TestRebuildTimeIndexAfterDeleteByType verifies that RebuildTimeIndex
// correctly handles a type that was entirely deleted.
func TestRebuildTimeIndexAfterDeleteByType(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))

	var traceID [16]byte
	const typeA msgstore.TypeID = 1
	const typeB msgstore.TypeID = 2

	for i := range 5 {
		option.Must(db.Append(typeA, traceID, []byte("a-"+strconv.Itoa(i))))
		option.Must(db.Append(typeB, traceID, []byte("b-"+strconv.Itoa(i))))
	}

	// delete all events of typeA
	option.MustZero(db.DeleteType(typeA))

	// rebuild
	option.MustZero(db.RebuildTimeIndex())

	// only typeB messages should exist
	var count int
	for _, _ = range db.Replay(nil, 1, math.MaxUint64) {
		count++
	}
	if count != 5 {
		t.Fatalf("expected 5 messages after rebuild, got %d", count)
	}

	option.MustZero(db.Close())
}

// TestRebuildTimeIndexMultipleTypes verifies that the rebuilt time index
// correctly covers messages from multiple event types.
func TestRebuildTimeIndexMultipleTypes(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))

	var traceID [16]byte

	before := time.Now().UnixNano()

	// interleave two types
	option.Must(db.Append(msgstore.TypeID(1), traceID, []byte("t1-a")))
	option.Must(db.Append(msgstore.TypeID(2), traceID, []byte("t2-a")))
	option.Must(db.Append(msgstore.TypeID(1), traceID, []byte("t1-b")))
	option.Must(db.Append(msgstore.TypeID(2), traceID, []byte("t2-b")))

	// rebuild
	option.MustZero(db.RebuildTimeIndex())

	// lookup should return a valid sequence ID
	seq, err := db.SeqForTime(before)
	if err != nil {
		t.Fatalf("SeqForTimestamp after rebuild: %v", err)
	}
	if seq < 1 || seq > 4 {
		t.Fatalf("unexpected seq after rebuild: %d", seq)
	}

	option.MustZero(db.Close())
}

// TestRebuildTimeIndexEmpty verifies that rebuilding an empty store succeeds
// without error.
func TestRebuildTimeIndexEmpty(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))

	option.MustZero(db.RebuildTimeIndex())

	option.MustZero(db.Close())
}

// TestRebuildTimeIndexCleansOldFiles verifies that the old time index files
// are actually removed and replaced by a fresh set.
func TestRebuildTimeIndexCleansOldFiles(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))

	var traceID [16]byte
	before := time.Now().UnixNano()
	for i := range 5 {
		option.Must(db.Append(msgstore.TypeID(1), traceID, []byte("msg-"+strconv.Itoa(i))))
	}

	// force flush of the time index buffer by performing a lookup
	_, _ = db.SeqForTime(before)

	// count time index files before rebuild
	timesDir := filepath.Join(dir, "times")
	beforeFiles := countFilesRecursive(t, timesDir)
	if beforeFiles == 0 {
		t.Fatal("expected at least one time index file before rebuild")
	}

	// delete all messages, then rebuild
	option.MustZero(db.DeleteType(msgstore.TypeID(1)))
	option.MustZero(db.RebuildTimeIndex())

	// with all messages deleted, the rebuilt index should have no files
	afterFiles := countFilesRecursive(t, timesDir)
	if afterFiles != 0 {
		t.Fatalf("expected 0 time index files after rebuild with no messages, got %d", afterFiles)
	}

	option.MustZero(db.Close())
}

// TestRebuildTimeIndexWithSplitSegments verifies that the rebuild works
// correctly when events span multiple segments.
func TestRebuildTimeIndexWithSplitSegments(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress:    msgstore.NoCompression,
		ShouldSplit: msgstore.SplitByCount(3),
	}))

	var traceID [16]byte
	const typeID msgstore.TypeID = 1

	before := time.Now().UnixNano()

	for i := range 12 {
		option.Must(db.Append(typeID, traceID, []byte("msg-"+strconv.Itoa(i))))
	}

	// rebuild
	option.MustZero(db.RebuildTimeIndex())

	// lookup should still work
	seq, err := db.SeqForTime(before)
	if err != nil {
		t.Fatalf("SeqForTimestamp after rebuild: %v", err)
	}
	if seq < 1 || seq > 12 {
		t.Fatalf("unexpected seq after rebuild: %d", seq)
	}

	// all 12 messages should still be replayable
	var count int
	for _, _ = range db.Replay(nil, 1, math.MaxUint64) {
		count++
	}
	if count != 12 {
		t.Fatalf("expected 12 messages after rebuild, got %d", count)
	}

	option.MustZero(db.Close())
}

// TestRebuildTimeIndexAppendAfterRebuild verifies that appending new messages
// after a rebuild works correctly and the time index stays consistent.
func TestRebuildTimeIndexAppendAfterRebuild(t *testing.T) {
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))

	var traceID [16]byte
	const typeID msgstore.TypeID = 1

	for i := range 5 {
		option.Must(db.Append(typeID, traceID, []byte("before-"+strconv.Itoa(i))))
	}

	option.MustZero(db.RebuildTimeIndex())

	afterRebuild := time.Now().UnixNano()

	// append more messages after rebuild
	for i := range 5 {
		option.Must(db.Append(typeID, traceID, []byte("after-"+strconv.Itoa(i))))
	}

	// lookup for the post-rebuild timestamp should find a sequence >= 6
	seq, err := db.SeqForTime(afterRebuild)
	if err != nil {
		t.Fatalf("SeqForTimestamp after rebuild+append: %v", err)
	}
	if seq < 6 || seq > 10 {
		t.Fatalf("unexpected seq for post-rebuild timestamp: %d", seq)
	}

	// total messages should be 10
	var count int
	for _, _ = range db.Replay(nil, 1, math.MaxUint64) {
		count++
	}
	if count != 10 {
		t.Fatalf("expected 10 messages total, got %d", count)
	}

	option.MustZero(db.Close())
}

// countFilesRecursive counts regular files under root recursively.
func countFilesRecursive(t *testing.T, root string) int {
	t.Helper()
	var count int
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			count++
		}
		return nil
	})
	return count
}

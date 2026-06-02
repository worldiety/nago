package msgstore_test

import (
	"crypto/rand"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/ndb/msgstore"
)

// writeTestEvents appends n events to a fresh DB and returns the dir.
func writeTestEvents(t *testing.T, n int, typeID msgstore.TypeID) string {
	t.Helper()
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))

	var traceID [16]byte
	for i := range n {
		payload := []byte("test-payload-" + strconv.Itoa(i))
		option.Must(db.Append(typeID, traceID, payload))
	}
	option.MustZero(db.Close())
	return dir
}

// countReplay counts the number of messages returned by Replay.
func countReplay(db *msgstore.DB) int {
	var count int
	for _, _ = range db.Replay(nil, 1, math.MaxUint64) {
		count++
	}
	return count
}

// findSegmentFile returns the path to the first .bin file in events/<typeID>/.
func findSegmentFile(t *testing.T, dir string, typeID msgstore.TypeID) string {
	t.Helper()
	typeDir := filepath.Join(dir, "events", strconv.FormatUint(uint64(typeID), 10))
	entries, err := os.ReadDir(typeDir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".bin") {
			return filepath.Join(typeDir, e.Name())
		}
	}
	t.Fatalf("no segment file found in %s", typeDir)
	return ""
}

var syncMarkerBytes = []byte{0xDE, 0xAD, 0x4E, 0x45, 0x56, 0x53, 0xBE, 0xEF}

func matchBytes(data, pattern []byte) bool {
	if len(data) < len(pattern) {
		return false
	}
	for i, b := range pattern {
		if data[i] != b {
			return false
		}
	}
	return true
}

func findAllSyncMarkers(data []byte) []int {
	var offsets []int
	for i := 0; i <= len(data)-len(syncMarkerBytes); i++ {
		if matchBytes(data[i:], syncMarkerBytes) {
			offsets = append(offsets, i)
		}
	}
	return offsets
}

func TestBitrotMiddleMessage(t *testing.T) {
	const typeID msgstore.TypeID = 1
	const total = 10
	dir := writeTestEvents(t, total, typeID)

	segPath := findSegmentFile(t, dir, typeID)
	data, err := os.ReadFile(segPath)
	if err != nil {
		t.Fatal(err)
	}

	// find the 5th sync marker and corrupt the message body after it
	markerOffsets := findAllSyncMarkers(data)
	if len(markerOffsets) < 5 {
		t.Fatalf("expected at least 5 sync markers, got %d", len(markerOffsets))
	}

	// corrupt bytes inside message #5 (after the frame header)
	corruptOffset := markerOffsets[4] + 12 // skip past sync+totalLen
	for j := 0; j < 10 && corruptOffset+j < len(data); j++ {
		data[corruptOffset+j] = 0xFF
	}

	if err := os.WriteFile(segPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	// reopen and replay – the corrupt message should be skipped
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))
	defer func() { option.MustZero(db.Close()) }()

	count := countReplay(db)
	if count != total-1 {
		t.Fatalf("expected %d messages (1 skipped), got %d", total-1, count)
	}
}

func TestBitrotCorruptPayloadLen(t *testing.T) {
	const typeID msgstore.TypeID = 1
	const total = 10
	dir := writeTestEvents(t, total, typeID)

	segPath := findSegmentFile(t, dir, typeID)
	data, err := os.ReadFile(segPath)
	if err != nil {
		t.Fatal(err)
	}

	// find the 3rd sync marker and corrupt the TotalLen field
	markerOffsets := findAllSyncMarkers(data)
	if len(markerOffsets) < 3 {
		t.Fatalf("expected at least 3 sync markers, got %d", len(markerOffsets))
	}

	// corrupt the TotalLen field (bytes 8..12 from marker start)
	off := markerOffsets[2]
	data[off+8] = 0xFF
	data[off+9] = 0xFF
	data[off+10] = 0xFF
	data[off+11] = 0xFF

	if err := os.WriteFile(segPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))
	defer func() { option.MustZero(db.Close()) }()

	count := countReplay(db)
	// message 3 is corrupt, messages 4-10 should be recovered via sync scan
	if count < total-1 {
		t.Fatalf("expected at least %d messages, got %d", total-1, count)
	}
}

func TestTailTruncation(t *testing.T) {
	const typeID msgstore.TypeID = 1
	const total = 10
	dir := writeTestEvents(t, total, typeID)

	segPath := findSegmentFile(t, dir, typeID)
	fi, err := os.Stat(segPath)
	if err != nil {
		t.Fatal(err)
	}

	// truncate the file to cut into the last message (remove last 20 bytes)
	newSize := fi.Size() - 20
	if err := os.Truncate(segPath, newSize); err != nil {
		t.Fatal(err)
	}

	// reopen – should repair tail and recover 9 messages
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))
	defer func() { option.MustZero(db.Close()) }()

	count := countReplay(db)
	if count != total-1 {
		t.Fatalf("expected %d messages after tail truncation, got %d", total-1, count)
	}

	// verify that appending still works with correct sequence IDs
	var traceID [16]byte
	seq := option.Must(db.Append(typeID, traceID, []byte("after-repair")))
	if seq < 10 {
		t.Fatalf("expected seqID >= 10 after repair, got %d", seq)
	}
}

func TestTailGarbageAppend(t *testing.T) {
	const typeID msgstore.TypeID = 1
	const total = 5
	dir := writeTestEvents(t, total, typeID)

	segPath := findSegmentFile(t, dir, typeID)

	// append random garbage at the end
	f, err := os.OpenFile(segPath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		t.Fatal(err)
	}
	garbage := make([]byte, 37) // odd size, not aligned to anything
	rand.Read(garbage)
	if _, err := f.Write(garbage); err != nil {
		t.Fatal(err)
	}
	f.Close()

	// reopen – should truncate garbage and recover all 5 messages
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))
	defer func() { option.MustZero(db.Close()) }()

	count := countReplay(db)
	if count != total {
		t.Fatalf("expected %d messages after garbage cleanup, got %d", total, count)
	}
}

func TestSyncMarkerInPayload(t *testing.T) {
	// write a message whose payload contains the sync marker pattern
	dir := t.TempDir()
	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))

	var traceID [16]byte

	// payload that embeds the sync marker multiple times
	payload := make([]byte, 100)
	copy(payload[10:], syncMarkerBytes)
	copy(payload[50:], syncMarkerBytes)
	copy(payload[80:], syncMarkerBytes)

	const typeID msgstore.TypeID = 1
	option.Must(db.Append(typeID, traceID, payload))
	option.Must(db.Append(typeID, traceID, []byte("after-marker-payload")))
	option.MustZero(db.Close())

	// reopen and verify both messages are readable
	db2 := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))
	defer func() { option.MustZero(db2.Close()) }()

	count := countReplay(db2)
	if count != 2 {
		t.Fatalf("expected 2 messages, got %d", count)
	}
}

func TestMultipleBitrotRegions(t *testing.T) {
	// corrupt two separate messages and verify the rest is recovered
	const typeID msgstore.TypeID = 1
	const total = 20
	dir := writeTestEvents(t, total, typeID)

	segPath := findSegmentFile(t, dir, typeID)
	data, err := os.ReadFile(segPath)
	if err != nil {
		t.Fatal(err)
	}

	markerOffsets := findAllSyncMarkers(data)
	if len(markerOffsets) < 15 {
		t.Fatalf("expected at least 15 sync markers, got %d", len(markerOffsets))
	}

	// corrupt message #3 and #10 (0-indexed: 2 and 9)
	for _, idx := range []int{2, 9} {
		off := markerOffsets[idx] + 12 // past frame header
		for j := 0; j < 10 && off+j < len(data); j++ {
			data[off+j] = 0xAA
		}
	}

	if err := os.WriteFile(segPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	db := option.Must(msgstore.Open(dir, msgstore.Options{
		Compress: msgstore.NoCompression,
	}))
	defer func() { option.MustZero(db.Close()) }()

	count := countReplay(db)
	if count != total-2 {
		t.Fatalf("expected %d messages (2 corrupted), got %d", total-2, count)
	}
}

package msgstore

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// TimeIndexEntry represents a single entry in the time index file.
// Each entry maps a monotonically increasing timestamp to a sequence ID.
type TimeIndexEntry struct {
	// TimestampNano is the unix timestamp in nanoseconds.
	TimestampNano int64
	// SequenceID is the global monotonic sequence ID assigned to this event.
	SequenceID uint64
}

// timeIndexEntrySize is the on-disk size of a single TimeIndexEntry: 8 + 8 = 16 bytes.
const timeIndexEntrySize = 16

// timeIndexFlushSize is the buffer threshold at which buffered entries are
// flushed to disk in a single pwrite. 1 MiB ≈ 65 536 entries.
const timeIndexFlushSize = 1 << 20

// Marshal encodes the entry into a 16-byte buffer.
func (e TimeIndexEntry) Marshal(buf *[timeIndexEntrySize]byte) {
	binary.BigEndian.PutUint64(buf[0:8], uint64(e.TimestampNano))
	binary.BigEndian.PutUint64(buf[8:16], e.SequenceID)
}

// UnmarshalTimeIndexEntry decodes a 16-byte buffer into a TimeIndexEntry.
func UnmarshalTimeIndexEntry(buf *[timeIndexEntrySize]byte) TimeIndexEntry {
	return TimeIndexEntry{
		TimestampNano: int64(binary.BigEndian.Uint64(buf[0:8])),
		SequenceID:    binary.BigEndian.Uint64(buf[8:16]),
	}
}

// timeIndex manages the times/ directory structure for timestamp→seqID lookups.
//
// Writes are buffered in memory and flushed to disk either when the buffer
// reaches timeIndexFlushSize (1 MiB ≈ 65k entries), on a day boundary change,
// when LookupMinSeq is called, or on Close. This eliminates the per-entry
// pwrite syscall that was the main write-throughput bottleneck under concurrent
// load, replacing it with a cheap slice-append (~5 ns under mutex).
//
// Since LookupMinSeq is expected to be rare (exploratory / operator use),
// the flush-on-read cost is acceptable.
type timeIndex struct {
	mu   sync.Mutex // protects all mutable state
	dir  string     // root times/ directory
	pool *FilePool

	// cached state for the current day file
	curPath   string // current day file path (empty = not yet initialised)
	curOffset int64  // next write offset within curPath (on-disk end)

	// write buffer: serialized 16-byte entries waiting to be flushed
	buf []byte
}

func newTimeIndex(dir string, pool *FilePool) *timeIndex {
	return &timeIndex{dir: dir, pool: pool}
}

// dayPath returns the file path for a given timestamp's day file.
// Format: times/YYYY/MM_DD.bin
func (ti *timeIndex) dayPath(ts time.Time) string {
	utc := ts.UTC()
	yearDir := filepath.Join(ti.dir, fmt.Sprintf("%04d", utc.Year()))
	return filepath.Join(yearDir, fmt.Sprintf("%02d_%02d.bin", utc.Month(), utc.Day()))
}

// ensureDayFile switches the cached day state to path if necessary.
// Flushes any pending buffer for the previous day before switching.
// Caller must hold ti.mu.
func (ti *timeIndex) ensureDayFile(path string) error {
	if ti.curPath == path {
		return nil // same day – nothing to do
	}

	// day boundary crossed: flush buffer for the old day first
	if err := ti.flushLocked(); err != nil {
		return err
	}

	// set up new day file
	parentDir := filepath.Dir(path)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("msgstore: timeindex mkdir: %w", err)
	}

	var offset int64
	fi, err := os.Stat(path)
	if err == nil {
		offset = fi.Size()
	}

	ti.curPath = path
	ti.curOffset = offset
	return nil
}

// flushLocked writes all buffered entries to disk in a single pwrite call.
// Caller must hold ti.mu.
func (ti *timeIndex) flushLocked() error {
	if len(ti.buf) == 0 {
		return nil
	}

	if ti.curPath == "" {
		ti.buf = ti.buf[:0]
		return nil
	}

	if _, err := ti.pool.WriteAt(ti.curPath, ti.buf, ti.curOffset); err != nil {
		return fmt.Errorf("msgstore: timeindex flush: %w", err)
	}

	ti.curOffset += int64(len(ti.buf))
	ti.buf = ti.buf[:0]
	return nil
}

// Append adds a timestamp→seqID entry to the in-memory buffer.
// The entry is NOT immediately written to disk. It will be flushed when the
// buffer reaches timeIndexFlushSize, on a day boundary change, on
// LookupMinSeq, or on Flush.
// Thread-safe: serialized via ti.mu.
func (ti *timeIndex) Append(tsNano int64, seqID uint64) error {
	ts := time.Unix(0, tsNano)
	path := ti.dayPath(ts)

	ti.mu.Lock()
	defer ti.mu.Unlock()

	if err := ti.ensureDayFile(path); err != nil {
		return err
	}

	// append serialized entry to buffer (~5 ns, no syscall)
	var entry [timeIndexEntrySize]byte
	binary.BigEndian.PutUint64(entry[0:8], uint64(tsNano))
	binary.BigEndian.PutUint64(entry[8:16], seqID)
	ti.buf = append(ti.buf, entry[:]...)

	// flush if buffer exceeds threshold
	if len(ti.buf) >= timeIndexFlushSize {
		return ti.flushLocked()
	}

	return nil
}

// Flush writes all buffered entries to disk. Thread-safe.
func (ti *timeIndex) Flush() error {
	ti.mu.Lock()
	defer ti.mu.Unlock()
	return ti.flushLocked()
}

// readEntryAt reads a single TimeIndexEntry at the given index via pool.ReadAt.
func (ti *timeIndex) readEntryAt(path string, idx int) (TimeIndexEntry, error) {
	var buf [timeIndexEntrySize]byte
	_, err := ti.pool.ReadAt(path, buf[:], int64(idx)*timeIndexEntrySize)
	if err != nil {
		return TimeIndexEntry{}, err
	}
	return UnmarshalTimeIndexEntry(&buf), nil
}

// LookupMinSeq returns the smallest sequence ID that has a timestamp >= tsNano.
// Uses binary search via random-access pool.ReadAt on the day file → O(log n)
// I/O with O(1) memory, no full file read required.
//
// Any buffered entries are flushed to disk before the lookup so that recent
// appends are visible to the search.
func (ti *timeIndex) LookupMinSeq(tsNano int64) (uint64, error) {
	// flush buffer so the lookup sees all appended entries
	if err := ti.Flush(); err != nil {
		return 0, fmt.Errorf("msgstore: timeindex flush before lookup: %w", err)
	}

	ts := time.Unix(0, tsNano)
	path := ti.dayPath(ts)

	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("msgstore: no time index for %s", ts.UTC().Format("2006-01-02"))
		}
		return 0, err
	}

	n := int(fi.Size() / timeIndexEntrySize)
	if n == 0 {
		return 0, fmt.Errorf("msgstore: empty time index for %s", ts.UTC().Format("2006-01-02"))
	}

	// binary search: each probe reads exactly 16 bytes via pool.ReadAt
	var searchErr error
	idx := sort.Search(n, func(i int) bool {
		if searchErr != nil {
			return true // short-circuit on prior error
		}
		entry, err := ti.readEntryAt(path, i)
		if err != nil {
			searchErr = fmt.Errorf("msgstore: timeindex read at %d: %w", i, err)
			return true
		}
		return entry.TimestampNano >= tsNano
	})
	if searchErr != nil {
		return 0, searchErr
	}

	if idx >= n {
		entry, err := ti.readEntryAt(path, n-1)
		if err != nil {
			return 0, fmt.Errorf("msgstore: timeindex read last: %w", err)
		}
		return entry.SequenceID, nil
	}

	entry, err := ti.readEntryAt(path, idx)
	if err != nil {
		return 0, fmt.Errorf("msgstore: timeindex read result: %w", err)
	}
	return entry.SequenceID, nil
}

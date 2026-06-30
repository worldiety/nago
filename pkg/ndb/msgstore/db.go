package msgstore

import (
	"container/heap"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"iter"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rogpeppe/go-internal/lockedfile"
	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/ndb"
)

// DB is the main event store handle. Only one process may hold it open at a time.
type DB struct {
	dir  string
	opts Options

	nextSeq atomic.Uint64 // next sequence ID to assign (atomic for lock-free allocation)

	typesMu sync.RWMutex          // protects the types map
	types   map[TypeID]*typeState // per event-type state

	lockFile *lockedfile.File // exclusive cross-platform file lock
	tindex   *timeIndex
	pool     *FilePool
	notify   *notifyRegistry // live append/put subscribers
}

// Compile-time proof that the engine implements the full neutral contract as
// well as each disjoint capability on its own. If a method signature drifts
// from the ndb contract, this fails to build.
var (
	_ ndb.Messages   = (*DB)(nil)
	_ ndb.History    = (*DB)(nil)
	_ ndb.Retained   = (*DB)(nil)
	_ ndb.TimeLookup = (*DB)(nil)
	_ ndb.Pruner     = (*DB)(nil)
	_ ndb.Notifier   = (*DB)(nil)
)

// typeState tracks the pending segment for a single event type.
// All file I/O goes through db.pool – no *os.File is held here.
type typeState struct {
	mu            sync.Mutex // per-type lock for concurrent append
	dir           string
	seg           *segmentFile
	writeOffset   int64 // current append position in the pending segment
	lastSeq       uint64
	lastMsgOffset int64  // file offset where the last written message starts (for O(1) Get)
	writeBuf      []byte // reusable marshal buffer, avoids per-append heap allocation
}

// Open opens (or creates) the event store at dir.
// The store is exclusively locked via lockedfile; a second Open on the same dir will block.
func Open(dir string, opts Options) (*DB, error) {
	opts.resolve()

	eventsDir := filepath.Join(dir, "events")
	timesDir := filepath.Join(dir, "times")
	if err := os.MkdirAll(eventsDir, 0755); err != nil {
		return nil, fmt.Errorf("msgstore: create events dir: %w", err)
	}
	if err := os.MkdirAll(timesDir, 0755); err != nil {
		return nil, fmt.Errorf("msgstore: create times dir: %w", err)
	}

	// acquire exclusive cross-platform file lock
	lockPath := filepath.Join(dir, ".lock")
	lockFile, err := lockedfile.Create(lockPath)
	if err != nil {
		return nil, fmt.Errorf("msgstore: store already locked by another process: %w", err)
	}

	pool := opts.FilePool

	db := &DB{
		dir:      dir,
		opts:     opts,
		lockFile: lockFile,
		pool:     pool,
		tindex:   newTimeIndex(timesDir, pool),
		types:    make(map[TypeID]*typeState),
		notify:   newNotifyRegistry(),
	}

	// bootstrap: find the global max sequence ID across all event type directories
	if err := db.bootstrap(eventsDir); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// bootstrap scans event type directories to find the next sequence ID.
func (db *DB) bootstrap(eventsDir string) error {
	entries, err := os.ReadDir(eventsDir)
	if err != nil {
		return fmt.Errorf("msgstore: read events dir: %w", err)
	}

	var globalMax uint64
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		_, err := strconv.ParseUint(e.Name(), 10, 64)
		if err != nil {
			slog.Warn("msgstore: skipping non-numeric event type dir", "name", e.Name())
			continue
		}

		typeDir := filepath.Join(eventsDir, e.Name())
		maxSeq := findMaxSeqInDir(db.pool, typeDir, db.opts.MaxMessageSize)
		if maxSeq > globalMax {
			globalMax = maxSeq
		}
	}

	db.nextSeq.Store(globalMax + 1)
	slog.Info("msgstore: bootstrap complete", "nextSeq", db.nextSeq.Load())
	return nil
}

// getTypeState returns (or lazily creates) the pending segment for a type.
// Uses double-checked locking: fast path with RLock, slow path with Lock.
func (db *DB) getTypeState(typeID TypeID) (*typeState, error) {
	// fast path: type already exists
	db.typesMu.RLock()
	ts, ok := db.types[typeID]
	db.typesMu.RUnlock()
	if ok {
		return ts, nil
	}

	// slow path: create new typeState under write lock
	db.typesMu.Lock()
	defer db.typesMu.Unlock()

	// double-check after acquiring write lock
	ts, ok = db.types[typeID]
	if ok {
		return ts, nil
	}

	typeDir := filepath.Join(db.dir, "events", strconv.FormatUint(uint64(typeID), 10))
	if err := os.MkdirAll(typeDir, 0755); err != nil {
		return nil, fmt.Errorf("msgstore: create type dir: %w", err)
	}

	// find existing pending segment or determine minSeq for new one
	segments, err := listSegments(typeDir)
	if err != nil {
		return nil, err
	}

	var minSeq uint64
	if len(segments) > 0 {
		last := segments[len(segments)-1]
		if last.isPending() {
			minSeq = last.minSeq
		} else {
			minSeq = last.maxSeq + 1
		}
	}

	seg, err := openOrCreatePending(db.pool, typeDir, minSeq, db.opts.MaxMessageSize)
	if err != nil {
		return nil, err
	}

	// rebuild info for existing pending segments
	var lastSeq uint64
	var lastMsgOff int64
	if seg.info.ByteSize > segHeaderSize {
		var count uint64
		var firstSeq uint64
		var firstTS int64
		off := int64(segHeaderSize)
		for msg, msgErr := range readMessages(db.pool, seg.path, db.opts.MaxMessageSize) {
			if msgErr != nil {
				break
			}
			count++
			if firstSeq == 0 {
				firstSeq = uint64(msg.Seq)
				firstTS = msg.TimeNano
			}
			lastMsgOff = off
			lastSeq = uint64(msg.Seq)
			off += int64(msgFrameOverhead + msgFixedSize + len(msg.Payload))
		}
		seg.info.MessageCount = count
		seg.info.FirstSeqID = firstSeq
		seg.info.FirstTimestamp = firstTS
	}

	ts = &typeState{
		dir:           typeDir,
		seg:           seg,
		writeOffset:   seg.info.ByteSize,
		lastSeq:       lastSeq,
		lastMsgOffset: lastMsgOff,
	}
	db.types[typeID] = ts
	return ts, nil
}

// Append writes a new event to the store. It assigns a globally unique,
// strictly monotonic [Seq]. Returns the assigned sequence number.
//
// Concurrent appends to different event types proceed in parallel – only
// appends to the same type are serialized via the per-type mutex. The global
// sequence counter is allocated lock-free via atomic increment.
//
// This implements [ndb.History].
func (db *DB) Append(typeID TypeID, traceID TraceID, payload []byte) (Seq, error) {
	if int64(len(payload)) > db.opts.MaxMessageSize {
		return 0, fmt.Errorf("%w: %d bytes", ErrPayloadTooLarge, len(payload))
	}

	ts, err := db.getTypeState(typeID)
	if err != nil {
		return 0, err
	}

	// allocate sequence ID atomically (lock-free)
	seqID := db.nextSeq.Add(1) - 1

	now := time.Now().UnixNano()

	// compress based on type
	enc, compressed := db.opts.Compress(typeID, payload)

	msg := Message{
		Type:            typeID,
		Seq:             Seq(seqID),
		TimeNano:        now,
		TraceID:         traceID,
		Encoding:        enc,
		UncompressedLen: uint32(len(payload)),
		Payload:         compressed,
	}

	// per-type lock: marshal, write, update segment info
	ts.mu.Lock()

	// check split condition before writing
	if ts.seg.info.MessageCount > 0 && db.opts.ShouldSplit(ts.seg.info) {
		if err := db.splitSegment(ts); err != nil {
			ts.mu.Unlock()
			return 0, fmt.Errorf("msgstore: split segment: %w", err)
		}
	}

	data := MarshalInto(&msg, ts.writeBuf)
	ts.writeBuf = data // keep for reuse on next append

	// write via pool at tracked offset (pwrite semantics)
	n, err := db.pool.WriteAt(ts.seg.path, data, ts.writeOffset)
	if err != nil {
		ts.mu.Unlock()
		return 0, fmt.Errorf("msgstore: write message: %w", err)
	}

	// update segment info
	ts.lastMsgOffset = ts.writeOffset
	ts.seg.info.MessageCount++
	ts.seg.info.ByteSize += int64(n)
	ts.writeOffset += int64(n)
	if ts.seg.info.FirstSeqID == 0 {
		ts.seg.info.FirstSeqID = seqID
		ts.seg.info.FirstTimestamp = now
	}
	ts.lastSeq = seqID

	ts.mu.Unlock()

	// update time index (has its own internal mutex)
	if err := db.tindex.Append(now, seqID); err != nil {
		slog.Warn("msgstore: time index append failed", "err", err)
	}

	// notify live subscribers (non-blocking; never stalls the writer)
	db.notify.publish(Notification{Type: typeID, Seq: Seq(seqID), TimeNano: now, TraceID: traceID})

	return Seq(seqID), nil
}

// splitSegment finalizes the current pending segment and opens a new one.
// The typeState is updated in-place. Caller must hold ts.mu.
func (db *DB) splitSegment(ts *typeState) error {
	if ts.lastSeq > 0 {
		if err := ts.seg.finalize(db.pool, ts.lastSeq); err != nil {
			return err
		}
	}

	newMinSeq := ts.lastSeq + 1
	if newMinSeq == 1 {
		newMinSeq = db.nextSeq.Load()
	}

	seg, err := openOrCreatePending(db.pool, ts.dir, newMinSeq, db.opts.MaxMessageSize)
	if err != nil {
		return err
	}

	// update in-place – no map modification needed
	ts.seg = seg
	ts.writeOffset = seg.info.ByteSize
	ts.lastMsgOffset = 0
	// writeBuf is kept for reuse
	return nil
}

// Replay iterates over all messages for the given event types within the
// sequence ID range [minSeq, maxSeq]. If types is empty, all types are included.
//
// Messages are yielded in strict global sequence ID order across all requested
// types (k-way merge). This guarantees correct state reconstruction even when
// multiple correlated event types (e.g. Login + Logout) are replayed together.
//
// Pending (not yet finalized) segments are always included so that the most
// recent events are never silently omitted.
//
// Performance: the iterator uses zero-copy deserialization internally. The
// yielded Message.Payload is a view into a reusable read buffer and is only
// valid until the next iteration step. Callers that need to keep the payload
// beyond the current step must copy it (e.g. slices.Clone(msg.Payload)).
//
// The yielded Message.Type is populated with the type the message belongs to,
// so it is consistent with the TypeID returned as the iterator's first value.
//
// This implements [ndb.History].
func (db *DB) Replay(types []TypeID, minSeq, maxSeq Seq) iter.Seq2[TypeID, Message] {
	return func(yield func(TypeID, Message) bool) {
		eventsDir := filepath.Join(db.dir, "events")
		maxMsgSize := db.opts.MaxMessageSize
		pool := db.pool

		typeDirs := db.resolveTypeDirs(eventsDir, types)

		// Build one pull-style cursor per type.
		// Each cursor chains through all segments of that type in order.
		var cursors cursorHeap
		var stops []func()
		defer func() {
			for _, s := range stops {
				s()
			}
		}()

		for _, td := range typeDirs {
			segments, err := listSegments(td.dir)
			if err != nil {
				slog.Warn("msgstore: replay list segments", "type", td.id, "err", err)
				continue
			}

			seq := replayType(pool, segments, maxMsgSize, uint64(minSeq), uint64(maxSeq))
			next, stop := iter.Pull2(seq)
			stops = append(stops, stop)

			c := &replayCursor{typeID: td.id, next: next, stop: stop}
			if c.advance() {
				cursors = append(cursors, c)
			}
		}

		// k-way merge: always pop the cursor with the smallest SequenceID
		heap.Init(&cursors)
		for cursors.Len() > 0 {
			c := heap.Pop(&cursors).(*replayCursor)
			if !yield(c.typeID, c.msg) {
				return
			}
			if c.advance() {
				heap.Push(&cursors, c)
			}
		}
	}
}

type typeDir struct {
	id  TypeID
	dir string
}

func (db *DB) resolveTypeDirs(eventsDir string, types []TypeID) []typeDir {
	if len(types) == 0 {
		entries, err := os.ReadDir(eventsDir)
		if err != nil {
			slog.Error("msgstore: replay readdir", "err", err)
			return nil
		}
		var result []typeDir
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			id, err := strconv.ParseUint(e.Name(), 10, 64)
			if err != nil {
				continue
			}
			result = append(result, typeDir{TypeID(id), filepath.Join(eventsDir, e.Name())})
		}
		return result
	}

	var result []typeDir
	for _, t := range types {
		d := filepath.Join(eventsDir, strconv.FormatUint(uint64(t), 10))
		if _, err := os.Stat(d); err == nil {
			result = append(result, typeDir{t, d})
		}
	}
	return result
}

// SeqForTime returns the smallest [Seq] with a timestamp >= tsNano.
//
// This implements [ndb.TimeLookup].
func (db *DB) SeqForTime(tsNano int64) (Seq, error) {
	s, err := db.tindex.LookupMinSeq(tsNano)
	return Seq(s), err
}

// Put writes or overwrites the single retained event for the given type.
// It behaves like MQTT retained messages: only the latest value is kept,
// previous values are physically overwritten.
//
// The message is always written at offset segHeaderSize (the very first
// message slot in the pending segment). No segment splitting is performed
// and the time index is not updated, because Put-based types are intended
// for "last known value" semantics without history.
//
// The sequence ID is still allocated from the global monotonic counter so
// that consumers can detect updates via sequence comparison.
//
// No file truncation is needed when the new message is smaller than the
// previous one: the stale tail bytes start in the middle of the old frame
// and therefore cannot form a valid sync marker + CRC combination.
// readMessages / Replay will stop at the invalid boundary. On the next
// DB reopen, repairTailTruncation cleans up the stale tail once.
// Get reads directly via lastMsgOffset + TotalLen and never sees the tail.
//
// This implements [ndb.Retained].
func (db *DB) Put(typeID TypeID, traceID TraceID, payload []byte) (Seq, error) {
	if int64(len(payload)) > db.opts.MaxMessageSize {
		return 0, fmt.Errorf("%w: %d bytes", ErrPayloadTooLarge, len(payload))
	}

	ts, err := db.getTypeState(typeID)
	if err != nil {
		return 0, err
	}

	// allocate sequence ID atomically (lock-free)
	seqID := db.nextSeq.Add(1) - 1

	now := time.Now().UnixNano()

	// compress based on type
	enc, compressed := db.opts.Compress(typeID, payload)

	msg := Message{
		Type:            typeID,
		Seq:             Seq(seqID),
		TimeNano:        now,
		TraceID:         traceID,
		Encoding:        enc,
		UncompressedLen: uint32(len(payload)),
		Payload:         compressed,
	}

	// per-type lock: marshal, overwrite at fixed offset, update segment info
	ts.mu.Lock()

	data := MarshalInto(&msg, ts.writeBuf)
	ts.writeBuf = data // keep for reuse

	// always write at the first message slot – overwrites the previous value
	n, err := db.pool.WriteAt(ts.seg.path, data, segHeaderSize)
	if err != nil {
		ts.mu.Unlock()
		return 0, fmt.Errorf("msgstore: put message: %w", err)
	}

	// update segment info – always exactly one message
	ts.seg.info.MessageCount = 1
	ts.seg.info.ByteSize = segHeaderSize + int64(n)
	ts.seg.info.FirstSeqID = seqID
	ts.seg.info.FirstTimestamp = now
	ts.writeOffset = segHeaderSize + int64(n)
	ts.lastMsgOffset = segHeaderSize
	ts.lastSeq = seqID

	ts.mu.Unlock()

	// deliberately no time index update – Put types have no history

	// notify live subscribers (non-blocking; never stalls the writer)
	db.notify.publish(Notification{Type: typeID, Seq: Seq(seqID), TimeNano: now, TraceID: traceID})

	return Seq(seqID), nil
}

// Get returns the last message written to the given event type.
// This works efficiently for both Put-based (retain) and Append-based
// (history) types by reading directly from the cached lastMsgOffset
// in a single ReadAt call – no segment scanning required.
//
// Returns an empty option (with a nil error) if no message has been written
// for the type yet. The returned Message.Payload is an independent copy safe
// to retain, and Message.Type is set to typeID.
//
// This implements [ndb.Retained].
func (db *DB) Get(typeID TypeID) (option.Opt[Message], error) {
	ts, err := db.getTypeState(typeID)
	if err != nil {
		return option.None[Message](), err
	}

	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.lastMsgOffset == 0 {
		return option.None[Message](), nil
	}

	pool := db.pool
	path := ts.seg.path
	maxMsgSize := db.opts.MaxMessageSize

	// read the frame header to determine the total message size
	hdr := make([]byte, prePayloadSize)
	if _, err := pool.ReadAt(path, hdr, ts.lastMsgOffset); err != nil {
		return option.None[Message](), fmt.Errorf("msgstore: get read header: %w", err)
	}

	if [8]byte(hdr[0:8]) != syncMarker {
		return option.None[Message](), fmt.Errorf("msgstore: get: %w", ErrInvalidSyncMarker)
	}

	innerLen := binary.BigEndian.Uint32(hdr[8:12])
	framedTotal := int64(msgFrameOverhead) + int64(innerLen)

	// read the complete framed message
	buf := make([]byte, framedTotal)
	if _, err := pool.ReadAt(path, buf, ts.lastMsgOffset); err != nil {
		return option.None[Message](), fmt.Errorf("msgstore: get read message: %w", err)
	}

	msg, _, err := UnmarshalMessage(buf, maxMsgSize)
	if err != nil {
		return option.None[Message](), fmt.Errorf("msgstore: get unmarshal: %w", err)
	}
	msg.Type = typeID

	return option.Some(msg), nil
}

// Subscribe registers fn for messages newly written via Append or Put for the
// given types (empty = all types). fn is invoked synchronously on the writer's
// goroutine right after the message becomes durable, so it must be fast and must
// not write back into the same store (see [ndb.Notifier]).
//
// The returned close function unsubscribes and is idempotent. All subscriptions
// are also dropped by [DB.Close].
//
// This implements [ndb.Notifier]. For a convenient "replay then follow" stream,
// see [ndb.Tail].
func (db *DB) Subscribe(types []TypeID, fn func(Notification)) (close func()) {
	return db.notify.subscribe(types, fn)
}

// Close releases all resources: the file pool and the lock file.
func (db *DB) Close() error {
	// drop live subscribers first
	if db.notify != nil {
		db.notify.closeAll()
	}

	db.typesMu.Lock()
	db.types = nil
	db.typesMu.Unlock()

	var firstErr error

	// flush buffered time index entries to disk
	if db.tindex != nil {
		if err := db.tindex.Flush(); err != nil {
			firstErr = err
		}
	}

	if db.pool != nil {
		if err := db.pool.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if db.lockFile != nil {
		db.lockFile.Close()
		db.lockFile = nil
	}

	return firstErr
}

// RebuildTimeIndex deletes the existing time index and rebuilds it from
// scratch by scanning all segments across all event types in global sequence
// order. Tombstoned messages are excluded from the rebuilt index.
//
// This is useful when the time index has become stale due to message
// deletions or missed index entries (e.g. after a crash).
//
// The caller must ensure that no concurrent Append calls are in progress
// during the rebuild.
func (db *DB) RebuildTimeIndex() error {
	// flush pending time index entries before tearing down
	if err := db.tindex.Flush(); err != nil {
		return fmt.Errorf("msgstore: flush time index before rebuild: %w", err)
	}

	// evict all time index files from the pool and remove the directory
	timesDir := filepath.Join(db.dir, "times")
	_ = filepath.WalkDir(timesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		db.pool.Evict(path)
		return nil
	})

	if err := os.RemoveAll(timesDir); err != nil {
		return fmt.Errorf("msgstore: remove times dir: %w", err)
	}
	if err := os.MkdirAll(timesDir, 0755); err != nil {
		return fmt.Errorf("msgstore: recreate times dir: %w", err)
	}

	// replace in-memory time index with a fresh instance
	db.tindex = newTimeIndex(timesDir, db.pool)

	// iterate all messages in global sequence order and rebuild the index
	var count uint64
	for _, msg := range db.Replay(nil, 0, ^Seq(0)) {
		if err := db.tindex.Append(msg.TimeNano, uint64(msg.Seq)); err != nil {
			return fmt.Errorf("msgstore: rebuild time index append: %w", err)
		}
		count++
	}

	// flush the rebuilt index to disk
	if err := db.tindex.Flush(); err != nil {
		return fmt.Errorf("msgstore: flush rebuilt time index: %w", err)
	}

	slog.Info("msgstore: time index rebuild complete", "entries", count)
	return nil
}

// DeleteType removes all events for the given event type by deleting the
// entire type directory. The in-memory state for the type is also cleaned up.
// Time index entries are intentionally kept for performance reasons.
//
// This implements [ndb.Pruner].
func (db *DB) DeleteType(typeID TypeID) error {
	// remove from in-memory map first
	db.typesMu.Lock()
	ts, ok := db.types[typeID]
	if ok {
		delete(db.types, typeID)
	}
	db.typesMu.Unlock()

	typeDir := filepath.Join(db.dir, "events", strconv.FormatUint(uint64(typeID), 10))

	// evict the current pending segment handle from the pool
	if ok {
		ts.mu.Lock()
		if ts.seg != nil {
			db.pool.Evict(ts.seg.path)
		}
		ts.mu.Unlock()
	}

	// evict any finalized segment handles that might still be in the pool
	segments, _ := listSegments(typeDir)
	for _, seg := range segments {
		db.pool.Evict(seg.path)
	}

	// remove the entire type directory
	if err := os.RemoveAll(typeDir); err != nil {
		return fmt.Errorf("msgstore: remove type dir: %w", err)
	}

	return nil
}

// DeleteSeq soft-deletes a single message by overwriting it as a tombstone.
// The Seq is set to 0 and the payload is zeroed out. The message frame
// size and time index entries are preserved for performance reasons.
// Returns ErrNotFound if no message with the given seq exists for the type.
//
// This implements [ndb.Pruner].
func (db *DB) DeleteSeq(typeID TypeID, seq Seq) error {
	seqID := uint64(seq)
	typeDir := filepath.Join(db.dir, "events", strconv.FormatUint(uint64(typeID), 10))

	segments, err := listSegments(typeDir)
	if err != nil {
		return fmt.Errorf("msgstore: list segments: %w", err)
	}

	maxMsgSize := db.opts.MaxMessageSize
	pool := db.pool

	for _, seg := range segments {
		// skip finalized segments where seqID is outside range
		if !seg.isPending() && (seqID < seg.minSeq || seqID > seg.maxSeq) {
			continue
		}

		found, err := tombstoneMessage(pool, seg.path, seqID, maxMsgSize)
		if err != nil {
			return err
		}
		if found {
			return nil
		}
	}

	return ErrNotFound
}

// tombstoneMessage scans a segment file for a message with the given seqID
// and overwrites it in-place as a tombstone: SequenceID is set to 0, the
// payload is zeroed, encoding is set to raw, and the CRC is recomputed.
// Returns true if the message was found and overwritten.
func tombstoneMessage(pool *FilePool, path string, seqID uint64, maxMsgSize int64) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("msgstore: stat segment: %w", err)
	}

	fileSize := fi.Size()
	if fileSize <= segHeaderSize {
		return false, nil
	}

	var readBuf []byte
	offset := int64(segHeaderSize)

	for offset < fileSize {
		remaining := fileSize - offset
		if remaining < int64(msgFrameOverhead) {
			break
		}

		// read frame header to determine message size
		if cap(readBuf) < prePayloadSize {
			readBuf = make([]byte, max(4096, prePayloadSize))
		}
		readBuf = readBuf[:prePayloadSize]

		readLen := int64(prePayloadSize)
		if readLen > remaining {
			break
		}
		if _, err := pool.ReadAt(path, readBuf[:readLen], offset); err != nil {
			break
		}

		// validate sync marker
		if [8]byte(readBuf[0:8]) != syncMarker {
			break
		}

		innerLen := binary.BigEndian.Uint32(readBuf[8:12])
		if int64(innerLen) < int64(msgFixedSize) || int64(innerLen) > maxMsgSize+int64(msgFixedSize) {
			break
		}

		framedTotal := int64(msgFrameOverhead) + int64(innerLen)
		if remaining < framedTotal {
			break
		}

		// read full framed message
		if int64(cap(readBuf)) < framedTotal {
			readBuf = make([]byte, framedTotal)
		} else {
			readBuf = readBuf[:framedTotal]
		}
		if _, err := pool.ReadAt(path, readBuf, offset); err != nil {
			break
		}

		inner := readBuf[msgFrameOverhead:]
		msgSeqID := binary.BigEndian.Uint64(inner[0:8])

		if msgSeqID == seqID {
			// set SequenceID to 0 (tombstone marker)
			binary.BigEndian.PutUint64(inner[0:8], 0)

			// zero out payload bytes
			payloadLen := binary.BigEndian.Uint32(inner[33:37])
			clear(inner[41 : 41+payloadLen])

			// reset encoding and uncompressed length
			inner[32] = byte(EncodingRaw)
			binary.BigEndian.PutUint32(inner[37:41], 0)

			// recompute CRC over the modified inner body
			newCRC := crc32.ChecksumIEEE(inner[:int(innerLen)-4])
			binary.BigEndian.PutUint32(inner[int(innerLen)-4:], newCRC)

			// write the modified inner portion back at the same offset
			if _, err := pool.WriteAt(path, inner[:innerLen], offset+int64(msgFrameOverhead)); err != nil {
				return false, fmt.Errorf("msgstore: write tombstone: %w", err)
			}

			return true, nil
		}

		offset += framedTotal
	}

	return false, nil
}

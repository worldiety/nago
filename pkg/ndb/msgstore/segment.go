package msgstore

import (
	"encoding/binary"
	"fmt"
	"iter"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// segmentFile represents a single .bin segment file on disk.
type segmentFile struct {
	path   string
	minSeq uint64
	maxSeq uint64 // 0 means pending (open) segment
	info   SegmentInfo
}

// isPending returns true if this is a pending (not yet finalized) segment.
func (s *segmentFile) isPending() bool {
	return s.maxSeq == 0
}

// openOrCreatePending initialises the pending segment file via the pool.
// If the file does not exist it is created and the segment header is written.
// For existing files, trailing corrupt or incomplete data is truncated
// (crash recovery) before returning.
// No raw *os.File is returned – all subsequent I/O goes through the pool.
func openOrCreatePending(pool *FilePool, dir string, minSeq uint64, maxMsgSize int64) (*segmentFile, error) {
	name := fmt.Sprintf("%d_.bin", minSeq)
	path := filepath.Join(dir, name)

	seg := &segmentFile{
		path:   path,
		minSeq: minSeq,
	}

	fi, statErr := os.Stat(path)
	if statErr != nil && !os.IsNotExist(statErr) {
		return nil, fmt.Errorf("msgstore: stat pending segment: %w", statErr)
	}

	if statErr != nil || fi.Size() == 0 {
		// new file – write header via pool (WriteAt creates the file)
		header := marshalSegHeader()
		if _, err := pool.WriteAt(path, header, 0); err != nil {
			return nil, fmt.Errorf("msgstore: write segment header: %w", err)
		}
		seg.info.ByteSize = segHeaderSize
	} else {
		// existing file – validate header
		hdr := make([]byte, segHeaderSize)
		if _, err := pool.ReadAt(path, hdr, 0); err != nil {
			return nil, fmt.Errorf("msgstore: read segment header: %w", err)
		}
		if err := validateSegHeader(hdr); err != nil {
			return nil, err
		}

		// repair any trailing garbage from a crash / partial write
		validEnd, err := repairTailTruncation(pool, path, maxMsgSize)
		if err != nil {
			return nil, fmt.Errorf("msgstore: repair tail: %w", err)
		}
		seg.info.ByteSize = validEnd
	}

	return seg, nil
}

// finalize renames a pending segment to its finalized name with maxSeq.
// The old path is evicted from the pool before the rename.
func (s *segmentFile) finalize(pool *FilePool, maxSeq uint64) error {
	newName := fmt.Sprintf("%d_%d.bin", s.minSeq, maxSeq)
	newPath := filepath.Join(filepath.Dir(s.path), newName)

	// evict the old handle before renaming so the pool stays consistent
	pool.Evict(s.path)

	if err := os.Rename(s.path, newPath); err != nil {
		return fmt.Errorf("msgstore: finalize segment: %w", err)
	}

	s.path = newPath
	s.maxSeq = maxSeq
	return nil
}

// parseSegmentName parses "123_456.bin" or "123_.bin" into min/max seq IDs.
func parseSegmentName(name string) (minSeq, maxSeq uint64, pending bool, err error) {
	if !strings.HasSuffix(name, ".bin") {
		return 0, 0, false, fmt.Errorf("not a segment file: %s", name)
	}
	base := strings.TrimSuffix(name, ".bin")
	parts := strings.SplitN(base, "_", 2)
	if len(parts) != 2 {
		return 0, 0, false, fmt.Errorf("invalid segment name: %s", name)
	}

	minSeq, err = strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, 0, false, fmt.Errorf("invalid minSeq in %s: %w", name, err)
	}

	if parts[1] == "" {
		return minSeq, 0, true, nil
	}

	maxSeq, err = strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return 0, 0, false, fmt.Errorf("invalid maxSeq in %s: %w", name, err)
	}

	return minSeq, maxSeq, false, nil
}

// listSegments returns all segment files in a directory, sorted by minSeq ascending.
func listSegments(dir string) ([]segmentFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var segments []segmentFile
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".bin") {
			continue
		}
		minSeq, maxSeq, _, err := parseSegmentName(e.Name())
		if err != nil {
			slog.Warn("msgstore: skipping unrecognized file", "file", e.Name(), "err", err)
			continue
		}
		segments = append(segments, segmentFile{
			path:   filepath.Join(dir, e.Name()),
			minSeq: minSeq,
			maxSeq: maxSeq,
		})
	}

	sort.Slice(segments, func(i, j int) bool {
		return segments[i].minSeq < segments[j].minSeq
	})

	return segments, nil
}

// prePayloadSize is the frame overhead plus the fixed portion of the inner
// message before the variable-length payload:
// SyncMarker(8)+TotalLen(4) + SeqID(8)+Timestamp(8)+TraceID(16)+Encoding(1)+PayloadLen(4)+UncompressedLen(4) = 53.
const prePayloadSize = msgFrameOverhead + 8 + 8 + 16 + 1 + 4 + 4

// scanForNextSync searches for the next syncMarker in the file starting at
// offset. It reads in overlapping chunks to handle markers that span chunk
// boundaries. Returns the file offset where the sync marker starts, or -1 if
// none is found before EOF.
func scanForNextSync(pool *FilePool, path string, offset int64, fileSize int64) int64 {
	const chunkSize = 32 * 1024
	// overlap by len(syncMarker)-1 to catch markers at chunk boundaries
	buf := make([]byte, chunkSize+len(syncMarker)-1)

	for offset+int64(len(syncMarker)) <= fileSize {
		readSize := fileSize - offset
		if readSize > int64(len(buf)) {
			readSize = int64(len(buf))
		}
		if readSize < int64(len(syncMarker)) {
			return -1 // not enough bytes left for a sync marker
		}
		n, err := pool.ReadAt(path, buf[:readSize], offset)
		if err != nil || n < len(syncMarker) {
			return -1
		}

		for i := 0; i <= n-len(syncMarker); i++ {
			if [8]byte(buf[i:i+8]) == syncMarker {
				return offset + int64(i)
			}
		}

		// advance past this chunk, minus overlap
		advance := int64(n) - int64(len(syncMarker)-1)
		if advance <= 0 {
			return -1 // cannot make progress
		}
		offset += advance
	}
	return -1
}

// readMessages iterates over all valid messages in a segment file.
// All I/O is performed through the FilePool via per-message ReadAt calls,
// so no additional file handles are opened or leaked.
//
// When a corrupt message is encountered (CRC mismatch, invalid sync marker,
// implausible lengths), the iterator scans forward for the next valid sync
// marker and continues from there. This allows recovery from bitrot or
// partial corruption in the middle of a segment file.
//
// Truncated data at the end of the file is silently ignored (crash recovery).
//
// The yielded Message.Payload is a zero-copy view into a shared read buffer.
// It is only valid until the iterator advances to the next message. Callers
// that need to retain the payload must copy it before continuing iteration.
func readMessages(pool *FilePool, path string, maxMsgSize int64) iter.Seq2[Message, error] {
	return func(yield func(Message, error) bool) {
		fi, err := os.Stat(path)
		if err != nil {
			yield(Message{}, fmt.Errorf("msgstore: stat segment %s: %w", path, err))
			return
		}

		fileSize := fi.Size()
		if fileSize < segHeaderSize {
			yield(Message{}, fmt.Errorf("msgstore: segment too small: %s", path))
			return
		}

		// validate header
		hdr := make([]byte, segHeaderSize)
		if _, err := pool.ReadAt(path, hdr, 0); err != nil {
			yield(Message{}, fmt.Errorf("msgstore: read header %s: %w", path, err))
			return
		}
		if err := validateSegHeader(hdr); err != nil {
			yield(Message{}, err)
			return
		}

		// reusable read buffer – grown as needed, never shrunk within one iteration.
		// This eliminates per-message heap allocations completely.
		var readBuf []byte

		offset := int64(segHeaderSize)
		for offset < fileSize {
			remaining := fileSize - offset

			// not enough room for even the frame header → tail truncation
			if remaining < int64(msgFrameOverhead) {
				slog.Warn("msgstore: truncated frame at EOF", "file", path, "offset", offset)
				return
			}

			// ensure readBuf can hold at least the pre-payload header
			if cap(readBuf) < prePayloadSize {
				readBuf = make([]byte, max(4096, prePayloadSize))
			}
			readBuf = readBuf[:prePayloadSize]

			// read frame header + fixed fields to determine total message size
			readLen := int64(prePayloadSize)
			if readLen > remaining {
				readLen = remaining
			}
			readBuf = readBuf[:readLen]
			if _, err := pool.ReadAt(path, readBuf, offset); err != nil {
				slog.Warn("msgstore: read frame header", "file", path, "offset", offset, "err", err)
				return
			}

			// validate sync marker
			if len(readBuf) < len(syncMarker) || [8]byte(readBuf[0:8]) != syncMarker {
				slog.Warn("msgstore: invalid sync marker, scanning forward", "file", path, "offset", offset)
				nextOff := scanForNextSync(pool, path, offset+1, fileSize)
				if nextOff < 0 {
					slog.Warn("msgstore: no further sync marker found", "file", path, "offset", offset)
					return
				}
				offset = nextOff
				continue
			}

			// not enough data for the full pre-payload header → tail truncation
			if readLen < int64(prePayloadSize) {
				slog.Warn("msgstore: truncated message header at EOF", "file", path, "offset", offset)
				return
			}

			// read TotalLen from frame header
			innerLen := binary.BigEndian.Uint32(readBuf[8:12])

			// plausibility check on innerLen
			if int64(innerLen) < int64(msgFixedSize) || int64(innerLen) > maxMsgSize+int64(msgFixedSize) {
				slog.Warn("msgstore: implausible frame length, scanning forward",
					"file", path, "offset", offset, "innerLen", innerLen)
				nextOff := scanForNextSync(pool, path, offset+1, fileSize)
				if nextOff < 0 {
					return
				}
				offset = nextOff
				continue
			}

			framedTotal := int64(msgFrameOverhead) + int64(innerLen)
			if remaining < framedTotal {
				slog.Warn("msgstore: truncated message at EOF", "file", path, "offset", offset)
				return
			}

			// grow readBuf if needed, preserving already-read header bytes
			if int64(cap(readBuf)) < framedTotal {
				newBuf := make([]byte, framedTotal)
				copy(newBuf, readBuf)
				readBuf = newBuf
			} else {
				readBuf = readBuf[:framedTotal]
			}

			// read remaining bytes after the pre-payload header
			if tailSize := framedTotal - int64(prePayloadSize); tailSize > 0 {
				if _, err := pool.ReadAt(path, readBuf[prePayloadSize:], offset+int64(prePayloadSize)); err != nil {
					slog.Warn("msgstore: read message body", "file", path, "offset", offset, "err", err)
					return
				}
			}

			// zero-copy unmarshal: msg.Payload is a view into readBuf
			msg, n, err := UnmarshalMessageNoCopy(readBuf, maxMsgSize)
			if err != nil {
				slog.Warn("msgstore: corrupt message, skipping", "file", path, "offset", offset, "err", err)
				// try to skip past this frame using innerLen if plausible,
				// otherwise scan for next sync marker
				skipTo := offset + framedTotal
				if skipTo <= offset || skipTo > fileSize {
					nextOff := scanForNextSync(pool, path, offset+1, fileSize)
					if nextOff < 0 {
						return
					}
					offset = nextOff
				} else {
					offset = skipTo
				}
				continue
			}
			if !yield(msg, nil) {
				return
			}
			offset += int64(n)
		}
	}
}

// repairTailTruncation scans all messages in a segment file and truncates
// any trailing garbage after the last valid message. This handles crash
// recovery where a partial write left incomplete data at the end of the file.
// Returns the byte offset of the validated end (= new file size).
func repairTailTruncation(pool *FilePool, path string, maxMsgSize int64) (int64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	fileSize := fi.Size()
	if fileSize <= segHeaderSize {
		return fileSize, nil
	}

	// iterate over all messages tracking the end offset of the last valid one
	lastValidEnd := int64(segHeaderSize)
	for msg, msgErr := range readMessages(pool, path, maxMsgSize) {
		if msgErr != nil {
			break
		}
		// compute the wire size of this message to advance our end tracker
		wireSize := int64(msgFrameOverhead + msgFixedSize + len(msg.Payload))
		lastValidEnd += wireSize
	}

	// The above approach double-counts when readMessages skips corrupt
	// messages internally. We need a more reliable approach: re-scan
	// ourselves keeping track of offset.
	lastValidEnd = int64(segHeaderSize)
	var readBuf []byte
	offset := int64(segHeaderSize)
	for offset < fileSize {
		remaining := fileSize - offset
		if remaining < int64(msgFrameOverhead) {
			break
		}

		if cap(readBuf) < msgFrameOverhead {
			readBuf = make([]byte, msgFrameOverhead)
		}
		readBuf = readBuf[:msgFrameOverhead]
		if _, err := pool.ReadAt(path, readBuf, offset); err != nil {
			break
		}

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

		// read the full framed message to validate CRC
		if int64(cap(readBuf)) < framedTotal {
			readBuf = make([]byte, framedTotal)
		} else {
			readBuf = readBuf[:framedTotal]
		}
		if _, err := pool.ReadAt(path, readBuf, offset); err != nil {
			break
		}

		_, n, err := UnmarshalMessageNoCopy(readBuf, maxMsgSize)
		if err != nil {
			break
		}

		offset += int64(n)
		lastValidEnd = offset
	}

	if lastValidEnd < fileSize {
		slog.Warn("msgstore: truncating corrupt/incomplete tail",
			"file", path, "validEnd", lastValidEnd, "fileSize", fileSize)
		pool.Evict(path)
		if err := os.Truncate(path, lastValidEnd); err != nil {
			return 0, fmt.Errorf("msgstore: truncate: %w", err)
		}
	}

	return lastValidEnd, nil
}

// findMaxSeqInDir scans all segment files in dir and returns the highest
// sequence ID found. Returns 0 if the directory is empty or has no messages.
func findMaxSeqInDir(pool *FilePool, dir string, maxMsgSize int64) uint64 {
	segments, err := listSegments(dir)
	if err != nil || len(segments) == 0 {
		return 0
	}

	last := segments[len(segments)-1]
	if !last.isPending() {
		return last.maxSeq
	}

	var maxSeq uint64
	for msg, err := range readMessages(pool, last.path, maxMsgSize) {
		if err != nil {
			break
		}
		if uint64(msg.Seq) > maxSeq {
			maxSeq = uint64(msg.Seq)
		}
	}

	if maxSeq == 0 && len(segments) > 1 {
		secondLast := segments[len(segments)-2]
		if !secondLast.isPending() {
			return secondLast.maxSeq
		}
	}

	return maxSeq
}

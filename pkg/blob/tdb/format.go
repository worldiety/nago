// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tdb

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"slices"
)

// A record consists of a header and a payload. The header holds the kind (1 byte), the size of the payload and the
// CRC32 (IEEE) of the payload (each 4 bytes, little endian). The payload holds the transaction as uvarint, then the
// bucket and the key, each as uvarint length and bytes, and for a set the value the same way. A remove has no value.
// [Node] encodes the same format.
const (
	headerSize     = 1 + 4 + 4
	maxPayloadSize = 64 * 1024 * 1024 // a reasonable OOM limit for keys and values, which Set accepts
)

var (
	errTornRecord    = errors.New("incomplete record at the end of the file")
	errHiddenRecords = errors.New("valid records follow a bad record")
)

// appendRecord appends the encoded record to buf. It returns the buffer and the offset of the value within the record.
func appendRecord(buf []byte, kind nodeKind, tx uint64, bucket, key string, val []byte) ([]byte, int) {
	start := len(buf)
	payload := uvarintLen(tx) + uvarintLen(uint64(len(bucket))) + len(bucket) + uvarintLen(uint64(len(key))) + len(key)
	if kind == setKeyValue {
		payload += uvarintLen(uint64(len(val))) + len(val)
	}

	buf = slices.Grow(buf, headerSize+payload)
	buf = append(buf, byte(kind), 0, 0, 0, 0, 0, 0, 0, 0)
	binary.LittleEndian.PutUint32(buf[start+1:], uint32(payload))
	startOfPayload := len(buf)
	buf = binary.AppendUvarint(buf, tx)
	buf = binary.AppendUvarint(buf, uint64(len(bucket)))
	buf = append(buf, bucket...)
	buf = binary.AppendUvarint(buf, uint64(len(key)))
	buf = append(buf, key...)
	valOffset := 0
	if kind == setKeyValue {
		buf = binary.AppendUvarint(buf, uint64(len(val)))
		valOffset = len(buf) - start
		buf = append(buf, val...)
	}

	binary.LittleEndian.PutUint32(buf[start+5:], crc32.ChecksumIEEE(buf[startOfPayload:]))
	return buf, valOffset
}

// recordSize returns the encoded size of a set record with the given lengths.
func recordSize(tx uint64, bucketLen, keyLen int, valLen uint32) int64 {
	return int64(headerSize + uvarintLen(tx) + uvarintLen(uint64(bucketLen)) + bucketLen + uvarintLen(uint64(keyLen)) + keyLen +
		uvarintLen(uint64(valLen)) + int(valLen))
}

func uvarintLen(v uint64) int {
	n := 1
	for v >= 0x80 {
		v >>= 7
		n++
	}

	return n
}

// record is a decoded record. Bucket and key refer to the buffer of the scan and are only valid within the callback.
type record struct {
	kind   nodeKind
	tx     uint64
	bucket []byte
	key    []byte
	valOff int64 // offset of the value within the file
	valLen uint32
}

// scanRecords reads the complete records of f between from and to sequentially. It calls fn with each record, its
// offset and its encoded bytes, which are only valid within the call. It returns the offset after the last complete
// record. If the section ends within a record, whatever its header holds, or with zero bytes, as a crash may leave
// them, it returns that offset together with errTornRecord. The section limits the size of a record, thus it reads
// records larger than maxPayloadSize, which former versions have written.
func scanRecords(f *os.File, from, to int64, fn func(rec record, off int64, raw []byte) error) (int64, error) {
	// a large buffer saves syscalls, but most files of a database are small
	r := bufio.NewReaderSize(io.NewSectionReader(f, from, to-from), int(min(max(to-from, 4096), 1024*1024)))
	off := from
	var buf []byte
	for {
		buf = slices.Grow(buf[:0], headerSize)[:headerSize]
		if _, err := io.ReadFull(r, buf); err != nil {
			if err == io.EOF {
				return off, nil
			}

			if err == io.ErrUnexpectedEOF {
				return off, errTornRecord
			}

			return off, err
		}

		kind := nodeKind(buf[0])
		size := binary.LittleEndian.Uint32(buf[1:])
		crc := binary.LittleEndian.Uint32(buf[5:])
		if kind == 0 && size == 0 && crc == 0 && onlyZeros(r) {
			return off, errTornRecord
		}

		if int64(size) > to-off-headerSize {
			return off, errTornRecord
		}

		if kind != setKeyValue && kind != removeKeyValue {
			return off, fmt.Errorf("record at %d of %s: %w", off, f.Name(), InvalidNodeType)
		}

		buf = slices.Grow(buf, int(size))[:headerSize+int(size)]
		if _, err := io.ReadFull(r, buf[headerSize:]); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return off, errTornRecord
			}

			return off, err
		}

		payload := buf[headerSize:]
		if crc32.ChecksumIEEE(payload) != crc {
			return off, fmt.Errorf("record at %d of %s: %w", off, f.Name(), InvalidNodeCRC)
		}

		rec, err := decodePayload(kind, payload)
		if err != nil {
			return off, fmt.Errorf("record at %d of %s: %w", off, f.Name(), err)
		}

		rec.valOff += off + headerSize
		if err := fn(rec, off, buf); err != nil {
			return off, err
		}

		off += int64(len(buf))
	}
}

// onlyZeros consumes the reader and tells whether it holds nothing but zero bytes.
func onlyZeros(r *bufio.Reader) bool {
	for {
		b, err := r.ReadByte()
		if err != nil {
			return true
		}

		if b != 0 {
			return false
		}
	}
}

var errCorruptPayload = errors.New("corrupt record payload")

// decodePayload decodes the payload without copying. The offset of the value is relative to the payload.
func decodePayload(kind nodeKind, p []byte) (record, error) {
	rec := record{kind: kind}
	tx, n := binary.Uvarint(p)
	if n <= 0 {
		return rec, errCorruptPayload
	}

	rec.tx = tx
	pos := n
	slice := func() ([]byte, bool) {
		l, n := binary.Uvarint(p[pos:])
		if n <= 0 || l > uint64(len(p)-pos-n) {
			return nil, false
		}

		pos += n
		s := p[pos : pos+int(l)]
		pos += int(l)
		return s, true
	}

	var ok bool
	if rec.bucket, ok = slice(); !ok {
		return rec, errCorruptPayload
	}

	if rec.key, ok = slice(); !ok {
		return rec, errCorruptPayload
	}

	if kind == setKeyValue {
		val, ok := slice()
		if !ok {
			return rec, errCorruptPayload
		}

		rec.valLen = uint32(len(val))
		rec.valOff = int64(pos - len(val))
	}

	// nothing follows, otherwise a flipped kind would turn a set into a removal
	if pos != len(p) {
		return rec, errCorruptPayload
	}

	return rec, nil
}

// isBadRecord tells whether scanRecords has stopped at an incomplete or a corrupt record.
func isBadRecord(err error) bool {
	return errors.Is(err, errTornRecord) || errors.Is(err, InvalidNodeType) || errors.Is(err, InvalidNodeCRC) ||
		errors.Is(err, errCorruptPayload)
}

// scanFile reads the records of the file up to size like scanRecords. A crash may end the file with a bad record: an
// incomplete one or, after a power loss, one which zeros fill up to the end. Then, it returns the offset of that
// record together with errTornRecord.
//
// But scanRecords also stops at a bad record, which valid records follow. If skip is set, scanFile skips the bad
// record, if the valid records start within the size, which its header declares, or behind it with nothing but zeros
// in between: former versions have appended behind an incomplete record after a restart. It returns the skipped
// bytes. Any other bad record is corruption, which may hide a change, and fails the scan.
func scanFile(f *os.File, size int64, skip bool, fn func(rec record, off int64, raw []byte) error) (end, skipped int64, err error) {
	var lastTx uint64
	track := func(rec record, off int64, raw []byte) error {
		lastTx = rec.tx
		return fn(rec, off, raw)
	}

	from := int64(0)
	for {
		end, err = scanRecords(f, from, size, track)
		if err == nil || !isBadRecord(err) {
			return end, skipped, err
		}

		next, declared, rerr := resync(f, end, size, lastTx)
		if rerr != nil {
			return end, skipped, rerr
		}

		if next >= 0 && skip {
			skipped += next - end
			from = next
			continue
		}

		if next >= 0 {
			return end, skipped, fmt.Errorf("record at %d of %s: %w: %v", end, f.Name(), errHiddenRecords, err)
		}

		if errors.Is(err, errTornRecord) {
			return end, skipped, err
		}

		// a complete record, which zeros fill up, as a power loss may leave the end of a file
		if declared <= size {
			zeros, zerr := zerosAt(f, declared-1, size)
			if zerr != nil {
				return end, skipped, zerr
			}

			if zeros {
				return end, skipped, errTornRecord
			}
		}

		return end, skipped, err
	}
}

// resync searches valid records behind the bad record at x, see validAt. They must start within the size, which the
// header of the bad record declares, or behind it with nothing but zeros in between. It returns their offset or -1,
// and the declared end of the bad record or -1, if the file ends within its header.
func resync(f *os.File, x, size int64, minTx uint64) (next, declared int64, err error) {
	if size-x < headerSize {
		return -1, -1, nil
	}

	var hdr [headerSize]byte
	if _, err := f.ReadAt(hdr[:], x); err != nil {
		return 0, 0, err
	}

	declared = x + headerSize + int64(binary.LittleEndian.Uint32(hdr[1:]))
	r := bufio.NewReaderSize(io.NewSectionReader(f, x+1, size-x-1), 64*1024)
	var buf []byte
	prev := hdr[0] // the byte before y
	for y := x + 1; y+headerSize <= size; y++ {
		if y > declared && prev != 0 {
			break
		}

		h, err := r.Peek(headerSize)
		if err != nil {
			return 0, 0, err
		}

		// valid records right behind a complete one are no hint at a crash, but at a corrupt record
		if y != declared && (h[0] == byte(setKeyValue) || h[0] == byte(removeKeyValue)) &&
			int64(binary.LittleEndian.Uint32(h[1:])) <= size-y-headerSize {
			var ok bool
			if ok, buf, err = validAt(f, y, size, minTx, buf); err != nil {
				return 0, 0, err
			}

			if ok {
				return y, declared, nil
			}
		}

		if prev, err = r.ReadByte(); err != nil {
			return 0, 0, err
		}
	}

	return -1, declared, nil
}

// validAt tells whether valid records start at off: two in a row, or one, which the end of the file, an incomplete
// record or zeros follow, each with a transaction of at least minTx. Thus, a valid record within the bytes of another
// one is unlikely to pass. It reuses buf.
func validAt(f *os.File, off, size int64, minTx uint64, buf []byte) (bool, []byte, error) {
	for n := range 2 {
		if size-off < headerSize {
			return n > 0, buf, nil
		}

		buf = slices.Grow(buf[:0], headerSize)[:headerSize]
		if _, err := f.ReadAt(buf, off); err != nil {
			return false, buf, err
		}

		kind := nodeKind(buf[0])
		plen := int64(binary.LittleEndian.Uint32(buf[1:]))
		crc := binary.LittleEndian.Uint32(buf[5:])
		if plen > size-off-headerSize || n > 0 && kind == 0 && plen == 0 && crc == 0 {
			return n > 0, buf, nil
		}

		if kind != setKeyValue && kind != removeKeyValue {
			return false, buf, nil
		}

		buf = slices.Grow(buf, int(plen))[:headerSize+int(plen)]
		if _, err := f.ReadAt(buf[headerSize:], off+headerSize); err != nil {
			return false, buf, err
		}

		if crc32.ChecksumIEEE(buf[headerSize:]) != crc {
			return false, buf, nil
		}

		rec, err := decodePayload(kind, buf[headerSize:])
		if err != nil || rec.tx < minTx {
			return false, buf, nil
		}

		minTx = rec.tx
		off += headerSize + plen
	}

	return true, buf, nil
}

// zerosAt tells whether the file holds nothing but zeros between from and to.
func zerosAt(f *os.File, from, to int64) (bool, error) {
	r := bufio.NewReaderSize(io.NewSectionReader(f, from, to-from), 64*1024)
	for {
		b, err := r.ReadByte()
		if err == io.EOF {
			return true, nil
		}

		if err != nil {
			return false, err
		}

		if b != 0 {
			return false, nil
		}
	}
}

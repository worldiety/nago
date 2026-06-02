package msgstore

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
)

// segHeaderSize is Magic(4) + FormatMagic(4) + Version(1) = 9 bytes.
const segHeaderSize = 9

// syncMarker is written before every message to allow resynchronisation
// after corruption (bitrot, partial writes). The 8-byte pattern is chosen
// to be unlikely to occur by chance in compressed or binary payloads.
var syncMarker = [8]byte{0xDE, 0xAD, 0x4E, 0x45, 0x56, 0x53, 0xBE, 0xEF}

// msgFrameOverhead is the per-message framing added before the message body:
// SyncMarker(8) + TotalLen(4) = 12 bytes.
const msgFrameOverhead = 8 + 4

// msgFixedSize is the fixed portion of the inner message body (without frame):
// SequenceID(8) + Timestamp(8) + TraceID(16) + Encoding(1) + PayloadLen(4) + UncompressedLen(4) + CRC32(4) = 45 bytes.
const msgFixedSize = 8 + 8 + 16 + 1 + 4 + 4 + 4

var (
	magicNAGO = [4]byte{'N', 'A', 'G', 'O'}
	magicNEVS = [4]byte{'N', 'E', 'V', 'S'}
)

const formatVersion uint8 = 1

// Message represents a single event stored in a segment file.
//
// When obtained from an iterator (Replay, readMessages), the Payload slice may
// be a view into a shared read buffer. It is only valid until the iterator
// advances to the next message. Callers that need to retain the payload beyond
// the current iteration step must copy it:
//
//	kept := slices.Clone(msg.Payload)
type Message struct {
	SequenceID       uint64
	Timestamp        int64 // unix nanoseconds
	TraceID          [16]byte
	Encoding         Encoding
	PayloadLen       uint32 // byte length of (possibly compressed) payload
	UncompressedLen  uint32 // original uncompressed length
	Payload          []byte
}

// IsTombstone returns true if the message has been soft-deleted.
func (m *Message) IsTombstone() bool {
	return m.SequenceID == 0
}

// MarshalBinary encodes the message into wire format including trailing CRC32.
// A new slice is allocated on every call. For high-throughput paths use
// MarshalInto with a reusable buffer instead.
func (m *Message) MarshalBinary() []byte {
	return m.MarshalInto(nil)
}

// MarshalInto encodes the message into dst, growing it if necessary.
// The returned slice (which may differ from dst if growth was needed) contains
// the complete wire-format message including the leading sync marker and
// TotalLen frame header, ready for WriteAt.
// Passing the same slice across calls avoids per-message heap allocations.
func (m *Message) MarshalInto(dst []byte) []byte {
	innerSize := msgFixedSize + len(m.Payload)
	totalSize := msgFrameOverhead + innerSize
	if cap(dst) < totalSize {
		dst = make([]byte, totalSize)
	} else {
		dst = dst[:totalSize]
	}

	// frame header: sync marker + inner message length
	copy(dst[0:8], syncMarker[:])
	binary.BigEndian.PutUint32(dst[8:12], uint32(innerSize))

	// inner message body starts after frame header
	b := dst[msgFrameOverhead:]
	binary.BigEndian.PutUint64(b[0:8], m.SequenceID)
	binary.BigEndian.PutUint64(b[8:16], uint64(m.Timestamp))
	copy(b[16:32], m.TraceID[:])
	b[32] = byte(m.Encoding)
	binary.BigEndian.PutUint32(b[33:37], m.PayloadLen)
	binary.BigEndian.PutUint32(b[37:41], m.UncompressedLen)
	copy(b[41:41+len(m.Payload)], m.Payload)

	crc := crc32.ChecksumIEEE(b[:innerSize-4])
	binary.BigEndian.PutUint32(b[innerSize-4:], crc)

	return dst
}

var (
	ErrCorruptCRC        = errors.New("msgstore: CRC32 mismatch")
	ErrTruncated         = errors.New("msgstore: truncated message")
	ErrPayloadTooLarge   = errors.New("msgstore: payload exceeds maximum message size")
	ErrInvalidSyncMarker = errors.New("msgstore: invalid sync marker")
	ErrNotFound          = errors.New("msgstore: message not found")
)

// UnmarshalMessage decodes a single framed message starting at buf.
// It returns the message and the total number of bytes consumed (including frame header).
// The returned Message.Payload is an independent copy of the payload bytes.
// If the buffer is too short, ErrTruncated is returned.
// If the sync marker does not match, ErrInvalidSyncMarker is returned.
// If the CRC does not match, ErrCorruptCRC is returned.
func UnmarshalMessage(buf []byte, maxMessageSize int64) (Message, int, error) {
	m, n, err := UnmarshalMessageNoCopy(buf, maxMessageSize)
	if err != nil {
		return m, n, err
	}
	// make an independent copy so the caller can hold onto it
	p := make([]byte, len(m.Payload))
	copy(p, m.Payload)
	m.Payload = p
	return m, n, nil
}

// UnmarshalMessageNoCopy decodes a single framed message starting at buf without
// copying the payload. The returned Message.Payload is a sub-slice of buf
// and shares the same backing array. It is only valid as long as buf is not
// modified or reused.
//
// The buf must start with the 8-byte sync marker followed by a 4-byte TotalLen,
// then the inner message body. The returned byte count n includes the frame
// overhead.
//
// This is the zero-allocation fast path used by iterators where the message
// is consumed before the buffer is reused for the next message.
func UnmarshalMessageNoCopy(buf []byte, maxMessageSize int64) (Message, int, error) {
	// need at least the frame header
	if len(buf) < msgFrameOverhead {
		return Message{}, 0, ErrTruncated
	}

	// validate sync marker
	if [8]byte(buf[0:8]) != syncMarker {
		return Message{}, 0, ErrInvalidSyncMarker
	}

	innerLen := binary.BigEndian.Uint32(buf[8:12])
	framedTotal := msgFrameOverhead + int(innerLen)

	if int64(innerLen) < int64(msgFixedSize) {
		return Message{}, 0, fmt.Errorf("%w: inner length %d too small", ErrTruncated, innerLen)
	}

	if len(buf) < framedTotal {
		return Message{}, 0, ErrTruncated
	}

	// inner body starts after frame header
	inner := buf[msgFrameOverhead : msgFrameOverhead+int(innerLen)]

	var m Message
	m.SequenceID = binary.BigEndian.Uint64(inner[0:8])
	m.Timestamp = int64(binary.BigEndian.Uint64(inner[8:16]))
	copy(m.TraceID[:], inner[16:32])
	m.Encoding = Encoding(inner[32])
	m.PayloadLen = binary.BigEndian.Uint32(inner[33:37])
	m.UncompressedLen = binary.BigEndian.Uint32(inner[37:41])

	expectedInner := msgFixedSize + int(m.PayloadLen)
	if expectedInner != int(innerLen) {
		return Message{}, 0, fmt.Errorf("%w: TotalLen/PayloadLen mismatch", ErrCorruptCRC)
	}

	if int64(m.PayloadLen) > maxMessageSize {
		return Message{}, 0, fmt.Errorf("%w: %d bytes", ErrPayloadTooLarge, m.PayloadLen)
	}

	// zero-copy: payload references the input buffer directly
	m.Payload = inner[41 : 41+m.PayloadLen]

	// verify CRC over everything before the CRC field
	wantCRC := binary.BigEndian.Uint32(inner[int(innerLen)-4:])
	gotCRC := crc32.ChecksumIEEE(inner[:int(innerLen)-4])
	if gotCRC != wantCRC {
		return Message{}, 0, ErrCorruptCRC
	}

	return m, framedTotal, nil
}

// marshalSegHeader writes the 9-byte segment file header.
func marshalSegHeader() []byte {
	buf := make([]byte, segHeaderSize)
	copy(buf[0:4], magicNAGO[:])
	copy(buf[4:8], magicNEVS[:])
	buf[8] = formatVersion
	return buf
}

// validateSegHeader checks the first 9 bytes of a segment file.
func validateSegHeader(buf []byte) error {
	if len(buf) < segHeaderSize {
		return fmt.Errorf("msgstore: segment header too short: %d bytes", len(buf))
	}
	if [4]byte(buf[0:4]) != magicNAGO {
		return fmt.Errorf("msgstore: invalid magic: %x", buf[0:4])
	}
	if [4]byte(buf[4:8]) != magicNEVS {
		return fmt.Errorf("msgstore: invalid format magic: %x", buf[4:8])
	}
	if buf[8] != formatVersion {
		return fmt.Errorf("msgstore: unsupported version: %d", buf[8])
	}
	return nil
}


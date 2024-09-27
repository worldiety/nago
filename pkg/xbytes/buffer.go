package xbytes

import (
	"encoding/binary"
	"errors"
	"io"
)

var ReadLimitExceeded = errors.New("read limit exceeded")

// Buffer provides a bunch of slice helper functions to read and write dynamic length data types. If applicable,
// uses little endian encoding. For slice and string types the stdlib uvarint encoding is used, because it has a huge
// space benefit for small slices. Some people say, that varint encoding comes with a heavy branch-penalty. I don't
// believe that for today's CPU and branch prediction mechanics.
//
// It does never fail on write. The signatures return only errors to be compatible with common interfaces.
// However, reading may always fail, if not enough data is available. The latest returned error is held and
// any other method will short-circuit to the first error until Reset.
type Buffer struct {
	Buf       []byte
	Pos       int
	Err       error
	ReadLimit int // ReadLimit is considered for all variable type reads, like ReadSlice or ReadString
}

func (b *Buffer) WriteString(s string) (n int, err error) {
	// not sure if the go compiler is clever enough to inline this down to the copy func, which would be gc-free by definition
	return b.WriteSlice([]byte(s))
}

func (b *Buffer) WriteSlice(s []byte) (n int, err error) {
	if b.Err != nil {
		return 0, b.Err
	}

	n, err = b.WriteUvarint(uint64(len(s)))
	if err != nil {
		return n, err
	}

	n2, err := b.Write(s)
	return n2 + n, err
}

func (b *Buffer) ReadSlice() (s []byte, err error) {
	if b.Err != nil {
		return s, b.Err
	}

	l, err := b.ReadUvarint()
	if err != nil {
		return s, err
	}

	// optional protection against corruption or oom-attacks
	if b.ReadLimit > 0 && int(l) > b.ReadLimit {
		return s, ReadLimitExceeded
	}

	if int(l) > len(b.Buf)-b.Pos {
		return nil, io.EOF
	}

	s = make([]byte, l)
	copy(s, b.Buf[b.Pos:b.Pos+int(l)])
	b.Pos += int(l)

	return s, nil
}

func (b *Buffer) ReadInto(buf []byte) ([]byte, error) {
	if b.Err != nil {
		return buf, b.Err
	}

	l, err := b.ReadUvarint()
	if err != nil {
		return buf, err
	}

	// optional protection against corruption or oom-attacks
	if b.ReadLimit > 0 && int(l) > b.ReadLimit {
		return buf, ReadLimitExceeded
	}

	if int(l) > len(b.Buf)-b.Pos {
		b.Err = io.EOF
		return buf, b.Err
	}

	if cap(buf) < int(l) {
		buf = make([]byte, l)
	} else {
		buf = buf[:l]
	}

	copy(buf, b.Buf[b.Pos:b.Pos+int(l)])
	b.Pos += int(l)

	return buf, nil
}

func (b *Buffer) ReadString() (s string, err error) {
	if b.Err != nil {
		return s, b.Err
	}

	l, err := b.ReadUvarint()
	if err != nil {
		return s, err
	}

	// optional protection against corruption or (fuzzy) oom-attacks
	if b.ReadLimit > 0 && int(l) > b.ReadLimit {
		return s, ReadLimitExceeded
	}

	if int(l) > len(b.Buf)-b.Pos {
		b.Err = io.EOF
		return "", b.Err
	}

	s = string(b.Buf[b.Pos : b.Pos+int(l)])
	b.Pos += int(l)

	return s, nil
}

func (b *Buffer) WriteByte(c byte) error {
	if b.Err != nil {
		return b.Err
	}

	if b.Pos >= len(b.Buf) {
		b.Buf = append(b.Buf, c)
	} else {
		b.Buf[b.Pos] = c
	}
	b.Pos++
	return nil
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	required := len(p) - (len(b.Buf) - b.Pos)
	if len(p) >= required {
		for range required {
			b.Buf = append(b.Buf, 0)
		}
	}

	copy(b.Buf[b.Pos:], p)
	b.Pos += len(p)
	return len(p), nil
}

func (b *Buffer) Reset() {
	b.Pos = 0
	b.Err = nil
}

func (b *Buffer) ReadUint64() (uint64, error) {
	if b.Err != nil {
		return 0, b.Err
	}

	var tmp [8]byte
	if _, err := b.Read(tmp[:]); err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(tmp[:]), nil
}

func (b *Buffer) ReadUint32() (uint32, error) {
	if b.Err != nil {
		return 0, b.Err
	}

	var tmp [4]byte
	if _, err := b.Read(tmp[:]); err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint32(tmp[:]), nil
}

func (b *Buffer) WriteUint32(v uint32) error {
	var tmp [4]byte

	tmp[0] = byte(v)
	tmp[1] = byte(v >> 8)
	tmp[2] = byte(v >> 16)
	tmp[3] = byte(v >> 24)
	_, err := b.Write(tmp[:])
	return err
}

func (b *Buffer) ReadUvarint() (uint64, error) {
	return binary.ReadUvarint(b)
}

func (b *Buffer) WriteUvarint(v uint64) (int, error) {
	var tmp [binary.MaxVarintLen64]byte
	l := binary.PutUvarint(tmp[:], v)
	return b.Write(tmp[:l])
}

func (b *Buffer) Read(p []byte) (n int, err error) {
	if b.Err != nil {
		return 0, b.Err
	}

	if len(p) > len(b.Buf)-b.Pos {
		b.Err = io.EOF
		return 0, b.Err
	}

	copy(p, b.Buf[b.Pos:b.Pos+len(p)])
	b.Pos += len(p)
	return len(p), nil
}

func (b *Buffer) ReadByte() (byte, error) {
	if b.Err != nil {
		return 0, b.Err
	}

	if b.Pos >= len(b.Buf) {
		b.Err = io.EOF
		return 0, b.Err
	}

	c := b.Buf[b.Pos]
	b.Pos++
	return c, nil
}

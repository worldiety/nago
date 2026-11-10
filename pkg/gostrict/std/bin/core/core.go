// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package core implements package bin for a host in Go. The native code and the virtual machine share it, so that both
// behave the same. A program can not import it.
package core

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
	"unsafe"
)

// MaxReserve is the greatest number of bytes, which [NewBuffer] reserves in advance. A capacity is only a hint, so a
// larger one must not take the memory at once, and the buffer grows beyond it as needed.
const MaxReserve = 64 << 20

// Text returns the bytes as UTF-8 and replaces invalid sequences by U+FFFD. Valid bytes share their memory with the
// string, because neither of them ever changes.
func Text(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	if utf8.Valid(b) {
		return unsafe.String(&b[0], len(b))
	}
	return strings.ToValidUTF8(string(b), "�")
}

// FromText returns the UTF-8 of s, which shares its memory with s, because neither of them ever changes.
func FromText(s string) []byte {
	if s == "" {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func isASCII(b []byte) bool {
	for _, c := range b {
		if c >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

// Latin1 returns the bytes as ISO 8859-1, which maps each byte to the code point of its value.
func Latin1(b []byte) string {
	if isASCII(b) {
		return Text(b)
	}
	var sb strings.Builder
	sb.Grow(len(b) * 2)
	for _, c := range b {
		sb.WriteRune(rune(c))
	}
	return sb.String()
}

// FromLatin1 returns the bytes of s as ISO 8859-1 or false, if s has a code point beyond U+00FF.
func FromLatin1(s string) ([]byte, bool) {
	raw := FromText(s)
	if isASCII(raw) {
		return raw, true
	}
	res := make([]byte, 0, len(s))
	for _, c := range s {
		if c > 0xff {
			return nil, false
		}
		res = append(res, byte(c))
	}
	return res, true
}

// Hex returns the bytes in lower case hexadecimal.
func Hex(b []byte) string {
	return hex.EncodeToString(b)
}

// FromHex returns the bytes of the hexadecimal s or false.
func FromHex(s string) ([]byte, bool) {
	b, err := hex.DecodeString(s)
	return b, err == nil
}

// Base64 returns the bytes in the standard encoding of Base64 with padding.
func Base64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

// FromBase64 returns the bytes of s in the standard encoding of Base64 with padding or false.
func FromBase64(s string) ([]byte, bool) {
	b, err := base64.StdEncoding.DecodeString(s)
	return b, err == nil
}

// Slice returns the bytes in from..to, shared and clamped into the range.
func Slice(b []byte, from, to int64) []byte {
	n := int64(len(b))
	from, to = max(from, 0), min(to, n)
	if from >= to {
		return nil
	}
	return b[from:to:to]
}

// Index returns the index of sub in b or -1.
func Index(b, sub []byte) int64 {
	return int64(bytes.Index(b, sub))
}

// Reader reads a sequence of values, see package bin.
type Reader struct {
	data   []byte
	pos    int
	bit    uint
	reason string
	closed bool
}

// NewReader returns a reader at the beginning of data.
func NewReader(data []byte) *Reader {
	return &Reader{data: data}
}

// Close ends the reader, which reads nothing afterwards, and returns the reason of its failure or "".
func (r *Reader) Close() string {
	r.closed = true
	return r.reason
}

func (r *Reader) failed() bool {
	return r.reason != "" || r.closed
}

// Fail ends the reading with the reason, unless it has already failed.
func (r *Reader) Fail(reason string) {
	if r.reason == "" {
		if reason == "" {
			reason = "invalid data"
		}
		r.reason = "bin: " + reason + " at offset " + fmt.Sprint(r.pos)
	}
}

// Failed tells, whether the reading has failed.
func (r *Reader) Failed() bool {
	return r.failed()
}

// take returns the next n whole bytes or nil, if they do not exist.
func (r *Reader) take(n int64) []byte {
	r.Align()
	if r.failed() {
		return nil
	}
	if n < 0 || n > int64(len(r.data)-r.pos) {
		r.Fail(fmt.Sprintf("reading %d bytes exceeds the length %d", n, len(r.data)))
		return nil
	}
	end := r.pos + int(n)
	b := r.data[r.pos:end:end]
	r.pos = end
	return b
}

func (r *Reader) U8() uint32 {
	if b := r.take(1); b != nil {
		return uint32(b[0])
	}
	return 0
}

func (r *Reader) I8() int32 { return int32(int8(r.U8())) }

func (r *Reader) U16BE() uint32 {
	if b := r.take(2); b != nil {
		return uint32(binary.BigEndian.Uint16(b))
	}
	return 0
}

func (r *Reader) U16LE() uint32 {
	if b := r.take(2); b != nil {
		return uint32(binary.LittleEndian.Uint16(b))
	}
	return 0
}

func (r *Reader) I16BE() int32 { return int32(int16(r.U16BE())) }
func (r *Reader) I16LE() int32 { return int32(int16(r.U16LE())) }

func (r *Reader) U24BE() uint32 {
	if b := r.take(3); b != nil {
		return uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2])
	}
	return 0
}

func (r *Reader) U24LE() uint32 {
	if b := r.take(3); b != nil {
		return uint32(b[2])<<16 | uint32(b[1])<<8 | uint32(b[0])
	}
	return 0
}

func (r *Reader) U32BE() uint32 {
	if b := r.take(4); b != nil {
		return binary.BigEndian.Uint32(b)
	}
	return 0
}

func (r *Reader) U32LE() uint32 {
	if b := r.take(4); b != nil {
		return binary.LittleEndian.Uint32(b)
	}
	return 0
}

func (r *Reader) I32BE() int32 { return int32(r.U32BE()) }
func (r *Reader) I32LE() int32 { return int32(r.U32LE()) }

func (r *Reader) U64BE() uint64 {
	if b := r.take(8); b != nil {
		return binary.BigEndian.Uint64(b)
	}
	return 0
}

func (r *Reader) U64LE() uint64 {
	if b := r.take(8); b != nil {
		return binary.LittleEndian.Uint64(b)
	}
	return 0
}

func (r *Reader) I64BE() int64 { return int64(r.U64BE()) }
func (r *Reader) I64LE() int64 { return int64(r.U64LE()) }

func (r *Reader) F32BE() float64 { return float64(math.Float32frombits(r.U32BE())) }
func (r *Reader) F32LE() float64 { return float64(math.Float32frombits(r.U32LE())) }
func (r *Reader) F64BE() float64 { return math.Float64frombits(r.U64BE()) }
func (r *Reader) F64LE() float64 { return math.Float64frombits(r.U64LE()) }

func (r *Reader) remainingBits() int {
	return (len(r.data)-r.pos)*8 - int(r.bit)
}

func (r *Reader) checkBits(n uint32) bool {
	if r.failed() {
		return false
	}
	if n > 64 || int(n) > r.remainingBits() {
		r.Fail(fmt.Sprintf("reading %d bits exceeds the data", n))
		return false
	}
	return true
}

func (r *Reader) advanceBits(n uint) {
	r.bit += n
	if r.bit == 8 {
		r.bit = 0
		r.pos++
	}
}

// Bits reads n bits with the highest bit of each byte first.
func (r *Reader) Bits(n uint32) uint64 {
	if !r.checkBits(n) {
		return 0
	}
	var v uint64
	for left := uint(n); left > 0; {
		avail := 8 - r.bit
		take := min(left, avail)
		part := uint64(r.data[r.pos]) >> (avail - take) & (1<<take - 1)
		v = v<<take | part
		r.advanceBits(take)
		left -= take
	}
	return v
}

// BitsLE reads n bits with the lowest bit of each byte first.
func (r *Reader) BitsLE(n uint32) uint64 {
	if !r.checkBits(n) {
		return 0
	}
	var v uint64
	var shift uint
	for left := uint(n); left > 0; {
		take := min(left, 8-r.bit)
		part := uint64(r.data[r.pos]) >> r.bit & (1<<take - 1)
		v |= part << shift
		shift += take
		r.advanceBits(take)
		left -= take
	}
	return v
}

// Align skips the rest of a partly read byte.
func (r *Reader) Align() {
	if r.bit != 0 {
		r.bit = 0
		r.pos++
	}
}

// Skip skips n whole bytes.
func (r *Reader) Skip(n int64) {
	r.take(n)
}

// SeekTo continues at the byte offset. It is not called Seek, because go vet takes a Seek with an int64 for the one of
// io.Seeker.
func (r *Reader) SeekTo(offset int64) {
	if r.failed() {
		return
	}
	if offset < 0 || offset > int64(len(r.data)) {
		r.Fail(fmt.Sprintf("seeking to %d exceeds the length %d", offset, len(r.data)))
		return
	}
	r.pos, r.bit = int(offset), 0
}

// Pos returns the offset of the next whole byte.
func (r *Reader) Pos() int64 {
	if r.bit != 0 {
		return int64(r.pos + 1)
	}
	return int64(r.pos)
}

// Remaining returns the number of whole bytes from [Reader.Pos] on, 0 after a failure.
func (r *Reader) Remaining() int64 {
	if r.failed() {
		return 0
	}
	return int64(len(r.data)) - r.Pos()
}

// Bytes returns the next n bytes without copying them.
func (r *Reader) Bytes(n int64) []byte {
	return r.take(n)
}

// Rest returns all remaining bytes without copying them.
func (r *Reader) Rest() []byte {
	return r.take(r.Remaining())
}

func trimZ(b []byte) []byte {
	if i := bytes.IndexByte(b, 0); i >= 0 {
		return b[:i]
	}
	return b
}

// Text reads n bytes of UTF-8.
func (r *Reader) Text(n int64) string { return Text(r.take(n)) }

// TextZ reads a field of n bytes of UTF-8, which ends at its first zero.
func (r *Reader) TextZ(n int64) string { return Text(trimZ(r.take(n))) }

// Latin1 reads n bytes of ISO 8859-1, which ends at its first zero, if z is set.
func (r *Reader) Latin1(n int64, z bool) string {
	b := r.take(n)
	if z {
		b = trimZ(b)
	}
	return Latin1(b)
}

// Buffer collects bytes, see package bin.
type Buffer struct {
	data []byte
	// limit is the greatest number of bytes or 0 for no limit
	limit int
	over  bool
	bit   uint
}

// NewBuffer returns an empty buffer with room for capacity bytes, at most for the limit and for [MaxReserve]. A
// limit of 0 or less is no limit: the buffer grows, as long as there is memory.
func NewBuffer(capacity int64, limit int) *Buffer {
	limit = max(limit, 0)
	reserve := int64(MaxReserve)
	if limit > 0 {
		reserve = min(reserve, int64(limit))
	}
	return &Buffer{data: make([]byte, 0, min(max(capacity, 0), reserve)), limit: limit}
}

// Len returns the number of bytes.
func (b *Buffer) Len() int64 {
	return int64(len(b.data))
}

// Bytes returns the bytes for reading. The caller must not keep them.
func (b *Buffer) Bytes() []byte {
	return b.data
}

// reserve appends n bytes and returns them or nil beyond the limit.
func (b *Buffer) reserve(n int) []byte {
	b.bit = 0
	if b.over || (b.limit > 0 && n > b.limit-len(b.data)) {
		b.over = true
		return nil
	}
	at := len(b.data)
	if at+n <= cap(b.data) {
		b.data = b.data[:at+n]
	} else {
		b.data = append(b.data, make([]byte, n)...)
	}
	return b.data[at:]
}

func (b *Buffer) AppendU8(v uint32) {
	if p := b.reserve(1); p != nil {
		p[0] = byte(v)
	}
}

func (b *Buffer) AppendU16BE(v uint32) {
	if p := b.reserve(2); p != nil {
		binary.BigEndian.PutUint16(p, uint16(v))
	}
}

func (b *Buffer) AppendU16LE(v uint32) {
	if p := b.reserve(2); p != nil {
		binary.LittleEndian.PutUint16(p, uint16(v))
	}
}

func (b *Buffer) AppendU24BE(v uint32) {
	if p := b.reserve(3); p != nil {
		p[0], p[1], p[2] = byte(v>>16), byte(v>>8), byte(v)
	}
}

func (b *Buffer) AppendU24LE(v uint32) {
	if p := b.reserve(3); p != nil {
		p[0], p[1], p[2] = byte(v), byte(v>>8), byte(v>>16)
	}
}

func (b *Buffer) AppendU32BE(v uint32) {
	if p := b.reserve(4); p != nil {
		binary.BigEndian.PutUint32(p, v)
	}
}

func (b *Buffer) AppendU32LE(v uint32) {
	if p := b.reserve(4); p != nil {
		binary.LittleEndian.PutUint32(p, v)
	}
}

func (b *Buffer) AppendU64BE(v uint64) {
	if p := b.reserve(8); p != nil {
		binary.BigEndian.PutUint64(p, v)
	}
}

func (b *Buffer) AppendU64LE(v uint64) {
	if p := b.reserve(8); p != nil {
		binary.LittleEndian.PutUint64(p, v)
	}
}

func (b *Buffer) AppendF32BE(v float64) { b.AppendU32BE(math.Float32bits(float32(v))) }
func (b *Buffer) AppendF32LE(v float64) { b.AppendU32LE(math.Float32bits(float32(v))) }
func (b *Buffer) AppendF64BE(v float64) { b.AppendU64BE(math.Float64bits(v)) }
func (b *Buffer) AppendF64LE(v float64) { b.AppendU64LE(math.Float64bits(v)) }

// AppendBytes appends a copy of p.
func (b *Buffer) AppendBytes(p []byte) {
	if t := b.reserve(len(p)); t != nil {
		copy(t, p)
	}
}

// appendBit appends one bit into the partly written byte, whose next position is the highest or the lowest free one.
func (b *Buffer) appendBit(v uint64, msb bool) {
	if b.bit == 0 {
		bit := b.bit
		if b.reserve(1) == nil {
			return
		}
		b.bit = bit
	}
	last := &b.data[len(b.data)-1]
	if v&1 != 0 {
		if msb {
			*last |= 1 << (7 - b.bit)
		} else {
			*last |= 1 << b.bit
		}
	}
	b.bit = (b.bit + 1) % 8
}

// AppendBits appends the lowest n bits of v, the highest of them first, and fills each byte from its highest bit on.
func (b *Buffer) AppendBits(v uint64, n uint32) {
	n = min(n, 64)
	for i := uint(n); i > 0; i-- {
		b.appendBit(v>>(i-1), true)
	}
}

// AppendBitsLE appends the lowest n bits of v, the lowest of them first, and fills each byte from its lowest bit on.
func (b *Buffer) AppendBitsLE(v uint64, n uint32) {
	n = min(n, 64)
	for i := uint(0); i < uint(n); i++ {
		b.appendBit(v>>i, false)
	}
}

func (b *Buffer) at(i int64, n int) []byte {
	if i < 0 || i > int64(len(b.data)-n) {
		return nil
	}
	return b.data[i : int(i)+n]
}

// Set overwrites the byte at i and tells, whether i is in range.
func (b *Buffer) Set(i int64, v uint32) bool {
	p := b.at(i, 1)
	if p == nil {
		return false
	}
	p[0] = byte(v)
	return true
}

func (b *Buffer) SetU16BE(i int64, v uint32) bool {
	p := b.at(i, 2)
	if p == nil {
		return false
	}
	binary.BigEndian.PutUint16(p, uint16(v))
	return true
}

func (b *Buffer) SetU16LE(i int64, v uint32) bool {
	p := b.at(i, 2)
	if p == nil {
		return false
	}
	binary.LittleEndian.PutUint16(p, uint16(v))
	return true
}

func (b *Buffer) SetU32BE(i int64, v uint32) bool {
	p := b.at(i, 4)
	if p == nil {
		return false
	}
	binary.BigEndian.PutUint32(p, v)
	return true
}

func (b *Buffer) SetU32LE(i int64, v uint32) bool {
	p := b.at(i, 4)
	if p == nil {
		return false
	}
	binary.LittleEndian.PutUint32(p, v)
	return true
}

func (b *Buffer) SetU64BE(i int64, v uint64) bool {
	p := b.at(i, 8)
	if p == nil {
		return false
	}
	binary.BigEndian.PutUint64(p, v)
	return true
}

func (b *Buffer) SetU64LE(i int64, v uint64) bool {
	p := b.at(i, 8)
	if p == nil {
		return false
	}
	binary.LittleEndian.PutUint64(p, v)
	return true
}

// Reset empties the buffer and forgets a broken limit.
func (b *Buffer) Reset() {
	b.data, b.over, b.bit = b.data[:0], false, 0
}

// Snapshot returns a copy of the bytes.
func (b *Buffer) Snapshot() []byte {
	return append([]byte(nil), b.data...)
}

// Freeze hands the storage over without copying it and empties the buffer, so that no two views of the same memory
// exist. It returns the reason of the failure or "", if the buffer has no limit or stayed within it.
func (b *Buffer) Freeze() ([]byte, string) {
	res, over := b.data, b.over
	b.data, b.over, b.bit = nil, false, 0
	if over {
		return nil, fmt.Sprintf("bin: the buffer exceeds its limit of %d bytes", b.limit)
	}
	return res, ""
}

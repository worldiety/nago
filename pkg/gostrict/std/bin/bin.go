// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package bin reads and writes binary data, e.g. the special formats of industrial plants. It separates what can
// change from what can not, so that the compiler rules out races without copying:
//
//   - [Bytes] never changes, so a function of [Bytes.Slice] and a package variable share its memory, and it may enter a
//     function, which runs concurrently (GS046).
//
//   - [Buffer] collects bytes and is rejected there. [Buffer.Freeze] hands its memory over to a [Bytes] without
//     copying and empties the buffer, so that no two views of the same memory ever exist.
//
//   - [Decode] is the only way to a [Reader] and reports a read beyond the end as its error, which the compiler forces
//     the caller to check (GS033), so the single reads need no check.
//
//     h, err := bin.Decode(data, func(r bin.Reader) Header {
//     return Header{Magic: r.U32BE(), Version: r.U16LE(), Flags: uint32(r.Bits(3)), Name: r.TextZ(16)}
//     })
//     if err != nil {
//     return Header{}, err
//     }
//
// A byte is a uint32 from 0 to 255: there is no type byte. Lengths, indices, offsets and capacities are int64, so that
// a file beyond 2 GiB keeps its length: write int64(0) for an index. A number of bits is a uint32. A variable shift
// count is a uint32 or uint64.
//
//gostrict:header
package bin

import (
	"errors"
	"math"
	"math/bits"

	"go.wdy.de/nago/pkg/gostrict/std/bin/core"
	"go.wdy.de/nago/pkg/gostrict/std/js"
)

// Bytes is an immutable sequence of bytes. Create it with [Empty], [Of], [FromText] or [Buffer.Freeze]. Compare two of
// them with [Bytes.Equal], because == is not allowed for them.
type Bytes struct {
	ref js.Ref
}

func wrap(b []byte) Bytes { return Bytes{ref: js.NewRef(b)} }

func raw(b Bytes) []byte {
	v, _ := b.ref.Value().([]byte)
	return v
}

// Empty returns no bytes.
//
//gostrict:js () => new Uint8Array(0)
func Empty() Bytes { return wrap(nil) }

// Of returns the bytes of the low 8 bits of the values, e.g. the magic number of a format: bin.Of(0x89, 0x50).
//
//gostrict:js (v) => Uint8Array.from(v)
func Of(values ...uint32) Bytes {
	b := make([]byte, len(values))
	for i, v := range values {
		b[i] = byte(v)
	}
	return wrap(b)
}

// FromText returns the UTF-8 of s. It shares the memory of s, except in JavaScript.
//
//gostrict:js (s) => $bin_fromText(s)
func FromText(s string) Bytes { return wrap(core.FromText(s)) }

// FromLatin1 returns the bytes of s as ISO 8859-1 or false, if s has a code point beyond U+00FF.
//
//gostrict:zero
//gostrict:js (s) => $bin_fromLatin1(s)
func FromLatin1(s string) (Bytes, bool) {
	b, ok := core.FromLatin1(s)
	return wrap(b), ok
}

// FromHex returns the bytes of the hexadecimal s or false.
//
//gostrict:zero
//gostrict:js (s) => $bin_fromHex(s)
func FromHex(s string) (Bytes, bool) {
	b, ok := core.FromHex(s)
	return wrap(b), ok
}

// FromBase64 returns the bytes of s in the standard encoding of Base64 with padding or false.
//
//gostrict:zero
//gostrict:js (s) => $bin_fromBase64(s)
func FromBase64(s string) (Bytes, bool) {
	b, ok := core.FromBase64(s)
	return wrap(b), ok
}

// Len returns the number of bytes.
//
//gostrict:js (b) => $B(b.length)
func (b Bytes) Len() int64 { return int64(len(raw(b))) }

// Get returns the byte at i and true, or 0 and false, if i is out of range.
//
//gostrict:zero
//gostrict:js (b, i) => ($bin_in(i, b.length) ? [b[$N(i)], true] : [0, false])
func (b Bytes) Get(i int64) (uint32, bool) {
	s := raw(b)
	if i < 0 || i >= int64(len(s)) {
		return 0, false
	}
	return uint32(s[i]), true
}

// GetOr returns the byte at i or def, if i is out of range.
//
//gostrict:js (b, i, d) => ($bin_in(i, b.length) ? b[$N(i)] : d)
func (b Bytes) GetOr(i int64, def uint32) uint32 {
	if v, ok := b.Get(i); ok {
		return v
	}
	return def
}

// Slice returns the bytes in from..to, clamped into the range, without copying them.
//
//gostrict:js (b, from, to) => $bin_slice(b, from, to)
func (b Bytes) Slice(from, to int64) Bytes { return wrap(core.Slice(raw(b), from, to)) }

// Index returns the index of the first occurrence of sub or -1.
//
//gostrict:js (b, sub) => $B($bin_index(b, sub))
func (b Bytes) Index(sub Bytes) int64 { return core.Index(raw(b), raw(sub)) }

// Equal tells, whether both have the same bytes.
//
//gostrict:js (a, b) => $bin_equal(a, b)
func (b Bytes) Equal(o Bytes) bool { return string(raw(b)) == string(raw(o)) }

// HasPrefix tells, whether the bytes begin with p.
//
//gostrict:js (b, p) => p.length <= b.length && $bin_equal(b.subarray(0, p.length), p)
func (b Bytes) HasPrefix(p Bytes) bool {
	s, t := raw(b), raw(p)
	return len(t) <= len(s) && string(s[:len(t)]) == string(t)
}

// HasSuffix tells, whether the bytes end with p.
//
//gostrict:js (b, p) => p.length <= b.length && $bin_equal(b.subarray(b.length - p.length), p)
func (b Bytes) HasSuffix(p Bytes) bool {
	s, t := raw(b), raw(p)
	return len(t) <= len(s) && string(s[len(s)-len(t):]) == string(t)
}

// Text returns the bytes as UTF-8 and replaces invalid sequences by U+FFFD. It shares the memory of the bytes, except in
// JavaScript.
//
//gostrict:js (b) => $bin_text(b)
func (b Bytes) Text() string { return core.Text(raw(b)) }

// Latin1 returns the bytes as ISO 8859-1.
//
//gostrict:js (b) => $bin_latin1(b)
func (b Bytes) Latin1() string { return core.Latin1(raw(b)) }

// Hex returns the bytes in lower case hexadecimal.
//
//gostrict:js (b) => $bin_hex(b)
func (b Bytes) Hex() string { return core.Hex(raw(b)) }

// Base64 returns the bytes in the standard encoding of Base64 with padding.
//
//gostrict:js (b) => $bin_base64(b)
func (b Bytes) Base64() string { return core.Base64(raw(b)) }

// Reader reads a sequence of values, see [Decode]. A read beyond the end returns 0 and the failure is the error of
// Decode. After the first failure all reads return 0 or an empty value. A read of whole bytes begins at the next whole
// byte after [Reader.Bits].
type Reader struct {
	ref js.Ref
}

// Decode calls f with a reader of data and returns its result. The error tells about a read beyond the end and
// [Reader.Fail], and the result is whatever f returned then. The reader ends with f.
//
//gostrict:calls now
func Decode[T any](data Bytes, f func(r Reader) T) (T, error) {
	c := core.NewReader(raw(data))
	v := f(Reader{ref: js.NewRef(c)})
	if reason := c.Close(); reason != "" {
		return v, errors.New(reason)
	}
	return v, nil
}

// U8 reads an unsigned byte.
//
//gostrict:js (r) => $bin_u8(r)
func (r Reader) U8() uint32 { return r.c().U8() }

// I8 reads a signed byte.
//
//gostrict:js (r) => $bin_u8(r) << 24 >> 24
func (r Reader) I8() int32 { return r.c().I8() }

// U16BE reads an unsigned big endian integer of 16 bits.
//
//gostrict:js (r) => $bin_un(r, 2, false)
func (r Reader) U16BE() uint32 { return r.c().U16BE() }

// U16LE reads an unsigned little endian integer of 16 bits.
//
//gostrict:js (r) => $bin_un(r, 2, true)
func (r Reader) U16LE() uint32 { return r.c().U16LE() }

// I16BE reads a signed big endian integer of 16 bits.
//
//gostrict:js (r) => $bin_un(r, 2, false) << 16 >> 16
func (r Reader) I16BE() int32 { return r.c().I16BE() }

// I16LE reads a signed little endian integer of 16 bits.
//
//gostrict:js (r) => $bin_un(r, 2, true) << 16 >> 16
func (r Reader) I16LE() int32 { return r.c().I16LE() }

// U24BE reads an unsigned big endian integer of 24 bits.
//
//gostrict:js (r) => $bin_un(r, 3, false)
func (r Reader) U24BE() uint32 { return r.c().U24BE() }

// U24LE reads an unsigned little endian integer of 24 bits.
//
//gostrict:js (r) => $bin_un(r, 3, true)
func (r Reader) U24LE() uint32 { return r.c().U24LE() }

// U32BE reads an unsigned big endian integer of 32 bits.
//
//gostrict:js (r) => $bin_un(r, 4, false)
func (r Reader) U32BE() uint32 { return r.c().U32BE() }

// U32LE reads an unsigned little endian integer of 32 bits.
//
//gostrict:js (r) => $bin_un(r, 4, true)
func (r Reader) U32LE() uint32 { return r.c().U32LE() }

// I32BE reads a signed big endian integer of 32 bits.
//
//gostrict:js (r) => $bin_un(r, 4, false) | 0
func (r Reader) I32BE() int32 { return r.c().I32BE() }

// I32LE reads a signed little endian integer of 32 bits.
//
//gostrict:js (r) => $bin_un(r, 4, true) | 0
func (r Reader) I32LE() int32 { return r.c().I32LE() }

// U64BE reads an unsigned big endian integer of 64 bits.
//
//gostrict:js (r) => $bin_u64(r, false)
func (r Reader) U64BE() uint64 { return r.c().U64BE() }

// U64LE reads an unsigned little endian integer of 64 bits.
//
//gostrict:js (r) => $bin_u64(r, true)
func (r Reader) U64LE() uint64 { return r.c().U64LE() }

// I64BE reads a signed big endian integer of 64 bits.
//
//gostrict:js (r) => $i64($bin_u64(r, false))
func (r Reader) I64BE() int64 { return r.c().I64BE() }

// I64LE reads a signed little endian integer of 64 bits.
//
//gostrict:js (r) => $i64($bin_u64(r, true))
func (r Reader) I64LE() int64 { return r.c().I64LE() }

// F32BE reads a big endian float of 32 bits.
//
//gostrict:js (r) => $bin_f32($bin_un(r, 4, false))
func (r Reader) F32BE() float64 { return r.c().F32BE() }

// F32LE reads a little endian float of 32 bits.
//
//gostrict:js (r) => $bin_f32($bin_un(r, 4, true))
func (r Reader) F32LE() float64 { return r.c().F32LE() }

// F64BE reads a big endian float of 64 bits.
//
//gostrict:js (r) => $bin_f64($bin_u64(r, false))
func (r Reader) F64BE() float64 { return r.c().F64BE() }

// F64LE reads a little endian float of 64 bits.
//
//gostrict:js (r) => $bin_f64($bin_u64(r, true))
func (r Reader) F64LE() float64 { return r.c().F64LE() }

// Bits reads up to 64 bits with the highest bit of each byte first.
//
//gostrict:js (r, n) => $bin_bits(r, n)
func (r Reader) Bits(n uint32) uint64 { return r.c().Bits(n) }

// BitsLE reads up to 64 bits with the lowest bit of each byte first.
//
//gostrict:js (r, n) => $bin_bitsLE(r, n)
func (r Reader) BitsLE(n uint32) uint64 { return r.c().BitsLE(n) }

// Align skips the rest of a partly read byte.
//
//gostrict:js (r) => $bin_align(r)
func (r Reader) Align() { r.c().Align() }

// Skip skips n whole bytes.
//
//gostrict:js (r, n) => void $bin_take(r, n)
func (r Reader) Skip(n int64) { r.c().Skip(n) }

// Pos returns the offset of the next whole byte.
//
//gostrict:js (r) => $B($bin_pos(r))
func (r Reader) Pos() int64 { return r.c().Pos() }

// Remaining returns the number of whole bytes from [Reader.Pos] on, 0 after a failure.
//
//gostrict:js (r) => $B($bin_remaining(r))
func (r Reader) Remaining() int64 { return r.c().Remaining() }

// Bytes returns the next n bytes without copying them.
//
//gostrict:js (r, n) => $bin_bytes(r, n)
func (r Reader) Bytes(n int64) Bytes { return wrap(r.c().Bytes(n)) }

// Rest returns all remaining bytes without copying them.
//
//gostrict:js (r) => $bin_bytes(r, $bin_remaining(r))
func (r Reader) Rest() Bytes { return wrap(r.c().Rest()) }

// Text reads n bytes of UTF-8.
//
//gostrict:js (r, n) => $bin_text($bin_bytes(r, n))
func (r Reader) Text(n int64) string { return r.c().Text(n) }

// TextZ reads a field of n bytes of UTF-8, which ends at its first zero.
//
//gostrict:js (r, n) => $bin_text($bin_trimZ($bin_bytes(r, n)))
func (r Reader) TextZ(n int64) string { return r.c().TextZ(n) }

// Latin1 reads n bytes of ISO 8859-1.
//
//gostrict:js (r, n) => $bin_latin1($bin_bytes(r, n))
func (r Reader) Latin1(n int64) string { return r.c().Latin1(n, false) }

// Latin1Z reads a field of n bytes of ISO 8859-1, which ends at its first zero.
//
//gostrict:js (r, n) => $bin_latin1($bin_trimZ($bin_bytes(r, n)))
func (r Reader) Latin1Z(n int64) string { return r.c().Latin1(n, true) }

// Fail ends the reading with the reason, which is the error of [Decode], e.g. for a wrong magic number.
//
//gostrict:js (r, s) => $bin_fail(r, s)
func (r Reader) Fail(reason string) { r.c().Fail(reason) }

// Failed tells, whether the reading has failed, e.g. to end a loop early.
//
//gostrict:js (r) => $bin_failed(r)
func (r Reader) Failed() bool { return r.c().Failed() }

// Buffer collects bytes. A buffer has no limit, unless the host sets one, e.g. vm.Engine.MaxBuffer, and grows, as long
// as there is memory. It never fails while it grows: beyond a limit of the host it ignores all further bytes and
// [Buffer.Freeze] reports it. The compiler keeps a buffer out of package variables and of functions, which run
// concurrently.
type Buffer struct {
	ref js.Ref
}

// NewBuffer returns an empty buffer with room for capacity bytes, which it can outgrow. The capacity is a hint: the
// buffer reserves at most 64 MiB in advance.
//
//gostrict:js (c) => $bin_newBuffer(c)
func NewBuffer(capacity int64) Buffer { return Buffer{ref: js.NewRef(core.NewBuffer(capacity, 0))} }

// Len returns the number of bytes.
//
//gostrict:js (b) => $B(b.n)
func (b Buffer) Len() int64 { return b.c().Len() }

// Get returns the byte at i and true, or 0 and false, if i is out of range.
//
//gostrict:zero
//gostrict:js (b, i) => ($bin_in(i, b.n) ? [b.a[$N(i)], true] : [0, false])
func (b Buffer) Get(i int64) (uint32, bool) {
	s := b.c().Bytes()
	if i < 0 || i >= int64(len(s)) {
		return 0, false
	}
	return uint32(s[i]), true
}

// GetOr returns the byte at i or def, if i is out of range.
//
//gostrict:js (b, i, d) => ($bin_in(i, b.n) ? b.a[$N(i)] : d)
func (b Buffer) GetOr(i int64, def uint32) uint32 {
	if v, ok := b.Get(i); ok {
		return v
	}
	return def
}

// Set overwrites the byte at i with the low 8 bits of v and tells, whether i is in range.
//
//gostrict:js (b, i, v) => $bin_set(b, i, 1, false, v)
func (b Buffer) Set(i int64, v uint32) bool { return b.c().Set(i, v) }

// SetU16BE overwrites 2 bytes at i, e.g. a length field, and tells, whether they are in range.
//
//gostrict:js (b, i, v) => $bin_set(b, i, 2, false, v)
func (b Buffer) SetU16BE(i int64, v uint32) bool { return b.c().SetU16BE(i, v) }

// SetU16LE overwrites 2 bytes at i and tells, whether they are in range.
//
//gostrict:js (b, i, v) => $bin_set(b, i, 2, true, v)
func (b Buffer) SetU16LE(i int64, v uint32) bool { return b.c().SetU16LE(i, v) }

// SetU32BE overwrites 4 bytes at i and tells, whether they are in range.
//
//gostrict:js (b, i, v) => $bin_set(b, i, 4, false, v)
func (b Buffer) SetU32BE(i int64, v uint32) bool { return b.c().SetU32BE(i, v) }

// SetU32LE overwrites 4 bytes at i and tells, whether they are in range.
//
//gostrict:js (b, i, v) => $bin_set(b, i, 4, true, v)
func (b Buffer) SetU32LE(i int64, v uint32) bool { return b.c().SetU32LE(i, v) }

// SetU64BE overwrites 8 bytes at i and tells, whether they are in range.
//
//gostrict:js (b, i, v) => $bin_set64(b, i, false, v)
func (b Buffer) SetU64BE(i int64, v uint64) bool { return b.c().SetU64BE(i, v) }

// SetU64LE overwrites 8 bytes at i and tells, whether they are in range.
//
//gostrict:js (b, i, v) => $bin_set64(b, i, true, v)
func (b Buffer) SetU64LE(i int64, v uint64) bool { return b.c().SetU64LE(i, v) }

// AppendU8 appends the low 8 bits of v.
//
//gostrict:js (b, v) => $bin_append(b, 1, false, v)
func (b Buffer) AppendU8(v uint32) { b.c().AppendU8(v) }

// AppendU16BE appends the low 16 bits of v in big endian.
//
//gostrict:js (b, v) => $bin_append(b, 2, false, v)
func (b Buffer) AppendU16BE(v uint32) { b.c().AppendU16BE(v) }

// AppendU16LE appends the low 16 bits of v in little endian.
//
//gostrict:js (b, v) => $bin_append(b, 2, true, v)
func (b Buffer) AppendU16LE(v uint32) { b.c().AppendU16LE(v) }

// AppendU24BE appends the low 24 bits of v in big endian.
//
//gostrict:js (b, v) => $bin_append(b, 3, false, v)
func (b Buffer) AppendU24BE(v uint32) { b.c().AppendU24BE(v) }

// AppendU24LE appends the low 24 bits of v in little endian.
//
//gostrict:js (b, v) => $bin_append(b, 3, true, v)
func (b Buffer) AppendU24LE(v uint32) { b.c().AppendU24LE(v) }

// AppendU32BE appends v in big endian.
//
//gostrict:js (b, v) => $bin_append(b, 4, false, v)
func (b Buffer) AppendU32BE(v uint32) { b.c().AppendU32BE(v) }

// AppendU32LE appends v in little endian.
//
//gostrict:js (b, v) => $bin_append(b, 4, true, v)
func (b Buffer) AppendU32LE(v uint32) { b.c().AppendU32LE(v) }

// AppendU64BE appends v in big endian.
//
//gostrict:js (b, v) => $bin_append64(b, false, v)
func (b Buffer) AppendU64BE(v uint64) { b.c().AppendU64BE(v) }

// AppendU64LE appends v in little endian.
//
//gostrict:js (b, v) => $bin_append64(b, true, v)
func (b Buffer) AppendU64LE(v uint64) { b.c().AppendU64LE(v) }

// AppendF32BE appends v as a float of 32 bits in big endian.
//
//gostrict:js (b, v) => $bin_append(b, 4, false, $bin_f32bits(v))
func (b Buffer) AppendF32BE(v float64) { b.c().AppendF32BE(v) }

// AppendF32LE appends v as a float of 32 bits in little endian.
//
//gostrict:js (b, v) => $bin_append(b, 4, true, $bin_f32bits(v))
func (b Buffer) AppendF32LE(v float64) { b.c().AppendF32LE(v) }

// AppendF64BE appends v as a float of 64 bits in big endian.
//
//gostrict:js (b, v) => $bin_append64(b, false, $bin_f64bits(v))
func (b Buffer) AppendF64BE(v float64) { b.c().AppendF64BE(v) }

// AppendF64LE appends v as a float of 64 bits in little endian.
//
//gostrict:js (b, v) => $bin_append64(b, true, $bin_f64bits(v))
func (b Buffer) AppendF64LE(v float64) { b.c().AppendF64LE(v) }

// AppendBytes appends p.
//
//gostrict:js (b, p) => $bin_appendBytes(b, p)
func (b Buffer) AppendBytes(p Bytes) { b.c().AppendBytes(raw(p)) }

// AppendText appends the UTF-8 of s.
//
//gostrict:js (b, s) => $bin_appendBytes(b, $bin_fromText(s))
func (b Buffer) AppendText(s string) { b.c().AppendBytes(core.FromText(s)) }

// AppendBits appends the lowest n bits of v, the highest of them first, and fills each byte from its highest bit on.
//
//gostrict:js (b, v, n) => $bin_appendBits(b, v, n, true)
func (b Buffer) AppendBits(v uint64, n uint32) { b.c().AppendBits(v, n) }

// AppendBitsLE appends the lowest n bits of v, the lowest of them first, and fills each byte from its lowest bit on.
//
//gostrict:js (b, v, n) => $bin_appendBits(b, v, n, false)
func (b Buffer) AppendBitsLE(v uint64, n uint32) { b.c().AppendBitsLE(v, n) }

// Reset empties the buffer and forgets a broken limit of the host.
//
//gostrict:js (b) => $bin_reset(b)
func (b Buffer) Reset() { b.c().Reset() }

// Snapshot returns a copy of the bytes and leaves the buffer as it is.
//
//gostrict:js (b) => b.a.slice(0, b.n)
func (b Buffer) Snapshot() Bytes { return wrap(b.c().Snapshot()) }

// Freeze hands the bytes over without copying them and empties the buffer. It fails only, if the buffer has exceeded a
// limit of the host, and returns empty bytes then.
//
//gostrict:zero
//gostrict:js (b) => $bin_freeze(b)
func (b Buffer) Freeze() (Bytes, error) {
	s, reason := b.c().Freeze()
	if reason != "" {
		return Empty(), errors.New(reason)
	}
	return wrap(s), nil
}

// RotL32 rotates x to the left by k bits, by any k.
//
//gostrict:js (x, k) => ((k &= 31), k === 0 ? x : ((x << k) | (x >>> (32 - k))) >>> 0)
func RotL32(x, k uint32) uint32 { return bits.RotateLeft32(x, int(k&31)) }

// RotR32 rotates x to the right by k bits, by any k.
//
//gostrict:js (x, k) => ((k &= 31), k === 0 ? x : ((x >>> k) | (x << (32 - k))) >>> 0)
func RotR32(x, k uint32) uint32 { return bits.RotateLeft32(x, -int(k&31)) }

// RotL64 rotates x to the left by k bits, by any k.
//
//gostrict:js (x, k) => $bin_rot64(x, k & 63)
func RotL64(x uint64, k uint32) uint64 { return bits.RotateLeft64(x, int(k&63)) }

// RotR64 rotates x to the right by k bits, by any k.
//
//gostrict:js (x, k) => $bin_rot64(x, (64 - (k & 63)) & 63)
func RotR64(x uint64, k uint32) uint64 { return bits.RotateLeft64(x, -int(k&63)) }

// PopCount32 returns the number of bits, which are set.
//
//gostrict:js (x) => $bin_pop32(x)
func PopCount32(x uint32) uint32 { return uint32(bits.OnesCount32(x)) }

// PopCount64 returns the number of bits, which are set.
//
//gostrict:js (x) => $bin_pop32($N(x & 0xffffffffn)) + $bin_pop32($N(x >> 32n))
func PopCount64(x uint64) uint32 { return uint32(bits.OnesCount64(x)) }

// LeadingZeros32 returns the number of zeros above the highest set bit, 32 for 0.
//
//gostrict:js (x) => $M.clz32(x)
func LeadingZeros32(x uint32) uint32 { return uint32(bits.LeadingZeros32(x)) }

// LeadingZeros64 returns the number of zeros above the highest set bit, 64 for 0.
//
//gostrict:js (x) => ((x >> 32n) !== 0n ? $M.clz32($N(x >> 32n)) : 32 + $M.clz32($N(x)))
func LeadingZeros64(x uint64) uint32 { return uint32(bits.LeadingZeros64(x)) }

// TrailingZeros32 returns the number of zeros below the lowest set bit, 32 for 0.
//
//gostrict:js (x) => (x === 0 ? 32 : 31 - $M.clz32(x & -x))
func TrailingZeros32(x uint32) uint32 { return uint32(bits.TrailingZeros32(x)) }

// TrailingZeros64 returns the number of zeros below the lowest set bit, 64 for 0.
//
//gostrict:js (x) => $bin_tz64(x)
func TrailingZeros64(x uint64) uint32 { return uint32(bits.TrailingZeros64(x)) }

// Swap16 reverses the low 2 bytes of x and clears the others.
//
//gostrict:js (x) => (((x & 0xff) << 8) | ((x >>> 8) & 0xff)) >>> 0
func Swap16(x uint32) uint32 { return uint32(bits.ReverseBytes16(uint16(x))) }

// Swap32 reverses the 4 bytes of x.
//
//gostrict:js (x) => $bin_swap32(x)
func Swap32(x uint32) uint32 { return bits.ReverseBytes32(x) }

// Swap64 reverses the 8 bytes of x.
//
//gostrict:js (x) => $bin_swap64(x)
func Swap64(x uint64) uint64 { return bits.ReverseBytes64(x) }

// SignExtend32 returns the lowest n bits of v as a signed integer, e.g. a measured value of 12 bits.
//
//gostrict:js (v, n) => $bin_sext32(v, n)
func SignExtend32(v, n uint32) int32 {
	if n == 0 {
		return 0
	}
	if n >= 32 {
		return int32(v)
	}
	s := 32 - n
	return int32(v<<s) >> s
}

// SignExtend64 returns the lowest n bits of v as a signed integer.
//
//gostrict:js (v, n) => (n === 0 ? 0n : $B.asIntN($M.min(n, 64), v))
func SignExtend64(v uint64, n uint32) int64 {
	if n == 0 {
		return 0
	}
	if n >= 64 {
		return int64(v)
	}
	s := 64 - n
	return int64(v<<s) >> s
}

// Float32Bits returns the bits of f as a float of 32 bits.
//
//gostrict:js (f) => $bin_f32bits(f)
func Float32Bits(f float64) uint32 { return math.Float32bits(float32(f)) }

// Float32FromBits returns the float of 32 bits with the bits u.
//
//gostrict:js (u) => $bin_f32(u)
func Float32FromBits(u uint32) float64 { return float64(math.Float32frombits(u)) }

// Float64Bits returns the bits of f.
//
//gostrict:js (f) => $bin_f64bits(f)
func Float64Bits(f float64) uint64 { return math.Float64bits(f) }

// Float64FromBits returns the float with the bits u.
//
//gostrict:js (u) => $bin_f64(u)
func Float64FromBits(u uint64) float64 { return math.Float64frombits(u) }

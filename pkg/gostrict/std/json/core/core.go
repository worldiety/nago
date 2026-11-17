// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package core reads JSON for package json, natively and in the register machine, so that both accept the same
// documents and report the same errors at the same offsets. The runtime of JavaScript implements the same rules, see
// $json_Decode in runtime.js.
//
// The rules are those of RFC 8259 without extensions. A string, which is not valid UTF-8, has each run of invalid bytes
// replaced by U+FFFD like [bin.Bytes.Text], and so has an escaped lone surrogate. An integer takes no fraction and no
// exponent. A number, which exceeds float64, is out of range.
package core

import (
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
	"unsafe"

	bincore "go.wdy.de/nago/pkg/gostrict/std/bin/core"
)

// MaxDepth is the deepest nesting of arrays and objects.
const MaxDepth = 512

// Decoder reads one value from data. After the first error, every method returns a zero value, so that a caller only
// checks [Decoder.Err] at the end.
type Decoder struct {
	data  []byte
	pos   int
	depth int
	err   string
	buf   []byte
}

// New returns a decoder of data, which must not change while it is read.
func New(data []byte) *Decoder {
	return &Decoder{data: data}
}

// Err returns the first error, e.g. "json: expected a string at offset 12", or "".
func (d *Decoder) Err() string {
	return d.err
}

// Failed tells, whether an error occurred.
func (d *Decoder) Failed() bool {
	return d.err != ""
}

func (d *Decoder) failAt(pos int, msg string) {
	if d.err == "" {
		d.err = "json: " + msg + " at offset " + strconv.Itoa(pos)
	}
}

func (d *Decoder) ws() {
	for d.pos < len(d.data) {
		switch d.data[d.pos] {
		case ' ', '\t', '\n', '\r':
			d.pos++
		default:
			return
		}
	}
}

// unexpected reports the byte at the position or the end of the input.
func (d *Decoder) unexpected() {
	if d.pos >= len(d.data) {
		d.failAt(d.pos, "unexpected end of input")
		return
	}
	c := d.data[d.pos]
	if c >= 0x20 && c < 0x7f {
		d.failAt(d.pos, "unexpected character '"+string(rune(c))+"'")
		return
	}
	d.failAt(d.pos, "unexpected byte 0x"+hex2(c))
}

func hex2(c byte) string {
	const digits = "0123456789abcdef"
	return string([]byte{digits[c>>4], digits[c&15]})
}

// startsValue tells, whether c can start a value.
func startsValue(c byte) bool {
	switch c {
	case '"', '{', '[', 't', 'f', 'n', '-':
		return true
	}
	return c >= '0' && c <= '9'
}

// expect reports a value of another type or an invalid byte, before a value of the type is read.
func (d *Decoder) expect(what string) {
	if d.pos < len(d.data) && startsValue(d.data[d.pos]) {
		d.failAt(d.pos, "expected "+what)
		return
	}
	d.unexpected()
}

// Null reads null and tells, whether the next value was null. Every type takes null as its zero value.
func (d *Decoder) Null() bool {
	if d.err != "" {
		return true
	}
	d.ws()
	if d.pos < len(d.data) && d.data[d.pos] == 'n' {
		d.literal("null")
		return true
	}
	return false
}

func (d *Decoder) literal(lit string) {
	for i := 0; i < len(lit); i++ {
		if d.pos >= len(d.data) || d.data[d.pos] != lit[i] {
			d.unexpected()
			return
		}
		d.pos++
	}
}

// Bool reads true or false.
func (d *Decoder) Bool() bool {
	if d.err != "" {
		return false
	}
	d.ws()
	switch {
	case d.pos < len(d.data) && d.data[d.pos] == 't':
		d.literal("true")
		return d.err == ""
	case d.pos < len(d.data) && d.data[d.pos] == 'f':
		d.literal("false")
	default:
		d.expect("a boolean")
	}
	return false
}

// number reads a number and returns its text and whether it is an integer.
func (d *Decoder) number(what string) (text []byte, start int, integer bool) {
	d.ws()
	start = d.pos
	if d.pos >= len(d.data) || (d.data[d.pos] != '-' && (d.data[d.pos] < '0' || d.data[d.pos] > '9')) {
		d.expect(what)
		return nil, start, false
	}
	if d.data[d.pos] == '-' {
		d.pos++
	}
	switch {
	case d.pos < len(d.data) && d.data[d.pos] == '0':
		d.pos++
	case d.pos < len(d.data) && d.data[d.pos] >= '1' && d.data[d.pos] <= '9':
		d.digits()
	default:
		d.unexpected()
		return nil, start, false
	}
	integer = true
	if d.pos < len(d.data) && d.data[d.pos] == '.' {
		integer = false
		d.pos++
		if !d.digit() {
			d.unexpected()
			return nil, start, false
		}
		d.digits()
	}
	if d.pos < len(d.data) && (d.data[d.pos] == 'e' || d.data[d.pos] == 'E') {
		integer = false
		d.pos++
		if d.pos < len(d.data) && (d.data[d.pos] == '+' || d.data[d.pos] == '-') {
			d.pos++
		}
		if !d.digit() {
			d.unexpected()
			return nil, start, false
		}
		d.digits()
	}
	return d.data[start:d.pos], start, integer
}

func (d *Decoder) digit() bool {
	return d.pos < len(d.data) && d.data[d.pos] >= '0' && d.data[d.pos] <= '9'
}

func (d *Decoder) digits() {
	for d.digit() {
		d.pos++
	}
}

// integer reads an integer, whose magnitude does not exceed max, or max+1 for a negative one, if signed.
func (d *Decoder) integer(signed bool, max uint64) (neg bool, mag uint64) {
	if d.err != "" {
		return false, 0
	}
	text, start, isInt := d.number("an integer")
	if d.err != "" {
		return false, 0
	}
	if !isInt {
		d.failAt(start, "expected an integer")
		return false, 0
	}
	if text[0] == '-' {
		neg, text = true, text[1:]
	}
	limit := max
	if neg {
		limit = 0
		if signed {
			limit = max + 1
		}
	}
	for _, c := range text {
		v := uint64(c - '0')
		if v > limit || mag > (limit-v)/10 {
			d.failAt(start, "number out of range")
			return false, 0
		}
		mag = mag*10 + v
	}
	return neg && mag != 0, mag
}

// Int32 reads an integer of int32.
func (d *Decoder) Int32() int32 {
	neg, mag := d.integer(true, 1<<31-1)
	if neg {
		return int32(-int64(mag))
	}
	return int32(mag)
}

// Uint32 reads an integer of uint32.
func (d *Decoder) Uint32() uint32 {
	_, mag := d.integer(false, 1<<32-1)
	return uint32(mag)
}

// Int64 reads an integer of int64.
func (d *Decoder) Int64() int64 {
	neg, mag := d.integer(true, 1<<63-1)
	if neg {
		return int64(-mag)
	}
	return int64(mag)
}

// Uint64 reads an integer of uint64.
func (d *Decoder) Uint64() uint64 {
	_, mag := d.integer(false, 1<<64-1)
	return mag
}

// Float64 reads a number.
func (d *Decoder) Float64() float64 {
	if d.err != "" {
		return 0
	}
	text, start, _ := d.number("a number")
	if d.err != "" {
		return 0
	}
	// the text is not kept, it only avoids a copy
	f, err := strconv.ParseFloat(unsafe.String(&text[0], len(text)), 64)
	if err != nil && math.IsInf(f, 0) {
		d.failAt(start, "number out of range")
		return 0
	}
	return f
}

// BigInt reads an integer of any size and returns its decimal text without a negative zero.
func (d *Decoder) BigInt() string {
	if d.err != "" {
		return "0"
	}
	text, start, isInt := d.number("an integer")
	if d.err != "" {
		return "0"
	}
	if !isInt {
		d.failAt(start, "expected an integer")
		return "0"
	}
	if string(text) == "-0" {
		return "0"
	}
	return string(text)
}

// MaxDecimalScale is the greatest scale of a decimal, the same as decimal.MaxScale.
const MaxDecimalScale = 1000

// maxDecimalExponent bounds the exponent of a decimal far beyond any scale, so that reading it can not overflow.
const maxDecimalExponent = 1 << 20

// Decimal reads a decimal number exactly, without a float in between: a JSON number or a string, which contains one,
// e.g. 12.30, 1.5e3 or "12.30". It returns the digits of the integer M with its sign, without a negative zero, and the
// scale S, so that the number is M * 10^-S, e.g. "1230" and 2. The scale is the number of digits of the fraction less
// the exponent and at least 0, e.g. 1.5e3 is "1500" and 0. A scale above [MaxDecimalScale] and an exponent, which
// would append more zeros, are out of range.
func (d *Decoder) Decimal() (digits string, scale int32) {
	if d.err != "" {
		return "0", 0
	}
	d.ws()
	start := d.pos
	var text []byte
	if d.pos < len(d.data) && d.data[d.pos] == '"' {
		b, ok := d.str("a decimal")
		if !ok {
			return "0", 0
		}
		// the string holds a JSON number, nothing around it, not even white space
		sub := Decoder{data: b}
		if len(b) > 0 && (b[0] == '-' || (b[0] >= '0' && b[0] <= '9')) {
			text, _, _ = sub.number("a decimal")
		}
		if text == nil || sub.err != "" || sub.pos != len(b) {
			d.failAt(start, "expected a decimal")
			return "0", 0
		}
	} else {
		text, start, _ = d.number("a decimal")
		if d.err != "" {
			return "0", 0
		}
	}

	digits, scale, ok := decimalOf(text)
	if !ok {
		d.failAt(start, "number out of range")
		return "0", 0
	}
	return digits, scale
}

// decimalOf returns the digits and the scale of a valid JSON number, see [Decoder.Decimal], or false, if it is out of
// range.
func decimalOf(text []byte) (string, int32, bool) {
	neg := text[0] == '-'
	if neg {
		text = text[1:]
	}

	i := 0
	for i < len(text) && text[i] >= '0' && text[i] <= '9' {
		i++
	}
	m := append([]byte(nil), text[:i]...)
	frac := 0
	if i < len(text) && text[i] == '.' {
		i++
		for i < len(text) && text[i] >= '0' && text[i] <= '9' {
			m = append(m, text[i])
			frac++
			i++
		}
	}

	exp := 0
	if i < len(text) {
		// e or E
		i++
		expNeg := false
		if text[i] == '+' || text[i] == '-' {
			expNeg = text[i] == '-'
			i++
		}
		for ; i < len(text); i++ {
			exp = exp*10 + int(text[i]-'0')
			if exp > maxDecimalExponent {
				return "", 0, false
			}
		}
		if expNeg {
			exp = -exp
		}
	}

	scale := frac - exp
	if scale < 0 {
		if -scale > MaxDecimalScale {
			return "", 0, false
		}
		m = append(m, strings.Repeat("0", -scale)...)
		scale = 0
	}
	if scale > MaxDecimalScale {
		return "", 0, false
	}

	k := 0
	for k < len(m)-1 && m[k] == '0' {
		k++
	}
	m = m[k:]
	if neg && (len(m) != 1 || m[0] != '0') {
		return "-" + string(m), int32(scale), true
	}
	return string(m), int32(scale), true
}

// String reads a string.
func (d *Decoder) String() string {
	b, ok := d.str("a string")
	if !ok {
		return ""
	}
	return string(b)
}

// Base64 reads a string of standard Base64 with padding.
func (d *Decoder) Base64() []byte {
	d.ws()
	start := d.pos
	b, ok := d.str("a string")
	if !ok {
		return nil
	}
	res, ok := bincore.FromBase64(unsafe.String(unsafe.SliceData(b), len(b)))
	if !ok {
		d.failAt(start, "invalid base64")
		return nil
	}
	return res
}

// str reads a string. The result refers to the data or to a buffer, which the next string reuses.
func (d *Decoder) str(what string) ([]byte, bool) {
	if d.err != "" {
		return nil, false
	}
	d.ws()
	if d.pos >= len(d.data) || d.data[d.pos] != '"' {
		d.expect(what)
		return nil, false
	}
	d.pos++
	start := d.pos
	// the common case without escapes and with valid UTF-8 refers to the data
	for d.pos < len(d.data) {
		c := d.data[d.pos]
		if c == '"' {
			raw := d.data[start:d.pos]
			d.pos++
			if utf8.Valid(raw) {
				return raw, true
			}
			d.buf = append(d.buf[:0], bincore.Text(raw)...)
			return d.buf, true
		}
		if c == '\\' || c < 0x20 {
			break
		}
		d.pos++
	}

	d.buf = d.buf[:0]
	seg := start
	flush := func() {
		if seg < d.pos {
			raw := d.data[seg:d.pos]
			if utf8.Valid(raw) {
				d.buf = append(d.buf, raw...)
			} else {
				d.buf = append(d.buf, bincore.Text(raw)...)
			}
		}
	}
	for {
		if d.pos >= len(d.data) {
			d.failAt(d.pos, "unexpected end of input")
			return nil, false
		}
		c := d.data[d.pos]
		switch {
		case c == '"':
			flush()
			d.pos++
			return d.buf, true
		case c < 0x20:
			d.failAt(d.pos, "invalid character in string")
			return nil, false
		case c == '\\':
			flush()
			if !d.escape() {
				return nil, false
			}
			seg = d.pos
		default:
			d.pos++
		}
	}
}

// escape appends the escape at the position to the buffer.
func (d *Decoder) escape() bool {
	at := d.pos
	d.pos++
	if d.pos >= len(d.data) {
		d.failAt(d.pos, "unexpected end of input")
		return false
	}
	c := d.data[d.pos]
	d.pos++
	switch c {
	case '"', '\\', '/':
		d.buf = append(d.buf, c)
	case 'b':
		d.buf = append(d.buf, '\b')
	case 'f':
		d.buf = append(d.buf, '\f')
	case 'n':
		d.buf = append(d.buf, '\n')
	case 'r':
		d.buf = append(d.buf, '\r')
	case 't':
		d.buf = append(d.buf, '\t')
	case 'u':
		r, ok := d.hex4()
		if !ok {
			d.failAt(at, "invalid escape")
			return false
		}
		if r >= 0xd800 && r < 0xdc00 {
			// a high surrogate needs its low surrogate in the next escape, otherwise it is invalid on its own
			if d.pos+1 < len(d.data) && d.data[d.pos] == '\\' && d.data[d.pos+1] == 'u' {
				save := d.pos
				d.pos += 2
				lo, ok := d.hex4()
				if ok && lo >= 0xdc00 && lo < 0xe000 {
					d.buf = utf8.AppendRune(d.buf, 0x10000+(r-0xd800)<<10+(lo-0xdc00))
					return true
				}
				d.pos = save
			}
			r = utf8.RuneError
		} else if r >= 0xdc00 && r < 0xe000 {
			r = utf8.RuneError
		}
		d.buf = utf8.AppendRune(d.buf, r)
	default:
		d.failAt(at, "invalid escape")
		return false
	}
	return true
}

func (d *Decoder) hex4() (rune, bool) {
	if d.pos+4 > len(d.data) {
		return 0, false
	}
	var r rune
	for _, c := range d.data[d.pos : d.pos+4] {
		var v byte
		switch {
		case c >= '0' && c <= '9':
			v = c - '0'
		case c >= 'a' && c <= 'f':
			v = c - 'a' + 10
		case c >= 'A' && c <= 'F':
			v = c - 'A' + 10
		default:
			return 0, false
		}
		r = r<<4 | rune(v)
	}
	d.pos += 4
	return r, true
}

// Array reads the start of an array.
func (d *Decoder) Array() bool {
	return d.open('[', "an array")
}

// Object reads the start of an object.
func (d *Decoder) Object() bool {
	return d.open('{', "an object")
}

func (d *Decoder) open(c byte, what string) bool {
	if d.err != "" {
		return false
	}
	d.ws()
	if d.pos >= len(d.data) || d.data[d.pos] != c {
		d.expect(what)
		return false
	}
	if d.depth == MaxDepth {
		d.failAt(d.pos, "nesting too deep")
		return false
	}
	d.depth++
	d.pos++
	return true
}

// Next tells, whether the array or object, whose start the caller has read, has its element i, and reads the comma
// before it or the end after the last one. The caller reads the element, e.g.:
//
//	if d.Array() {
//		for i := 0; d.Next(i, ']'); i++ { ... }
//	}
func (d *Decoder) Next(i int, end byte) bool {
	if d.err != "" {
		return false
	}
	d.ws()
	if d.pos < len(d.data) && d.data[d.pos] == end && i == 0 {
		d.pos++
		d.depth--
		return false
	}
	if i > 0 {
		switch {
		case d.pos < len(d.data) && d.data[d.pos] == ',':
			d.pos++
		case d.pos < len(d.data) && d.data[d.pos] == end:
			d.pos++
			d.depth--
			return false
		default:
			d.unexpected()
			return false
		}
	}
	return true
}

// Key reads the key of an object and its colon. The result refers to the data or to the buffer, which the next string
// reuses.
func (d *Decoder) Key() ([]byte, bool) {
	k, ok := d.str("a string")
	if !ok {
		return nil, false
	}
	d.ws()
	if d.pos >= len(d.data) || d.data[d.pos] != ':' {
		d.unexpected()
		return nil, false
	}
	d.pos++
	return k, true
}

// Skip reads a value of any type, e.g. of a key, which the type does not know.
func (d *Decoder) Skip() {
	if d.err != "" {
		return
	}
	d.ws()
	if d.pos >= len(d.data) {
		d.unexpected()
		return
	}
	switch c := d.data[d.pos]; {
	case c == '"':
		d.str("a string")
	case c == '{':
		if d.Object() {
			for i := 0; d.Next(i, '}'); i++ {
				if _, ok := d.Key(); !ok {
					return
				}
				d.Skip()
			}
		}
	case c == '[':
		if d.Array() {
			for i := 0; d.Next(i, ']'); i++ {
				d.Skip()
			}
		}
	case c == 't':
		d.literal("true")
	case c == 'f':
		d.literal("false")
	case c == 'n':
		d.literal("null")
	default:
		d.number("a value")
	}
}

// End reads the whitespace after the value, which must be the last one.
func (d *Decoder) End() {
	if d.err != "" {
		return
	}
	d.ws()
	if d.pos < len(d.data) {
		d.failAt(d.pos, "unexpected data after the value")
	}
}

// Key returns the key of a field with the struct tag or false, if the tag json:"-" leaves the field out. Every engine
// uses it, so that they agree.
func Key(name, tag string) (string, bool) {
	t, ok := reflect.StructTag(tag).Lookup("json")
	if !ok {
		return name, true
	}
	if t == "-" {
		return "", false
	}
	if k, _, _ := strings.Cut(t, ","); k != "" {
		return k, true
	}
	return name, true
}

// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"encoding/base64"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Encoder writes compact JSON, natively and in the register machine, so that both write the same bytes as the runtime
// of JavaScript, see $json_Encode in runtime.js:
//
//   - a float64 as the shortest text, which reads back the same, with an exponent below 1e-6 and from 1e21 on, like
//     JavaScript, and -0 as -0. NaN and the infinities have no form in JSON and are an error.
//   - a string escapes only the quote, the backslash and the control characters, invalid UTF-8 becomes U+FFFD.
//   - the fields of a struct in their order, the entries of a map in the order of insertion.
type Encoder struct {
	buf []byte
	err string
}

// Bytes returns what has been written and the first error or "".
func (e *Encoder) Bytes() ([]byte, string) {
	return e.buf, e.err
}

// Len returns the number of bytes written so far, e.g. for [Encoder.Truncate].
func (e *Encoder) Len() int {
	return len(e.buf)
}

// Truncate removes what has been written after the length n.
func (e *Encoder) Truncate(n int) {
	e.buf = e.buf[:n]
}

// Empty tells, whether what has been written after the length n is null, "", [] or {}, which a field with the option
// omitempty leaves out.
func (e *Encoder) Empty(n int) bool {
	switch string(e.buf[n:]) {
	case "null", `""`, "[]", "{}":
		return true
	}
	return false
}

// Byte writes a delimiter, e.g. '[' or ','.
func (e *Encoder) Byte(c byte) {
	e.buf = append(e.buf, c)
}

// Null writes null.
func (e *Encoder) Null() {
	e.buf = append(e.buf, "null"...)
}

// Bool writes true or false.
func (e *Encoder) Bool(b bool) {
	e.buf = strconv.AppendBool(e.buf, b)
}

// Int writes an integer.
func (e *Encoder) Int(v int64) {
	e.buf = strconv.AppendInt(e.buf, v, 10)
}

// Uint writes an unsigned integer.
func (e *Encoder) Uint(v uint64) {
	e.buf = strconv.AppendUint(e.buf, v, 10)
}

// Digits writes the decimal text of a big.Int.
func (e *Encoder) Digits(s string) {
	e.buf = append(e.buf, s...)
}

// Decimal writes the integer of digits, with an optional sign, scaled by 10^-scale as a string like decimal.String
// does, with its scale clamped into 0..[MaxDecimalScale], e.g. "12.30" for "1230" and 2. A string keeps the digits and
// the scale for every reader, also for one, which reads a number as a float, like JavaScript does.
func (e *Encoder) Decimal(digits string, scale int32) {
	scale = min(max(scale, 0), MaxDecimalScale)
	e.buf = append(e.buf, '"')
	if strings.HasPrefix(digits, "-") {
		e.buf = append(e.buf, '-')
		digits = digits[1:]
	}
	if scale > 0 {
		// the whole part has at least one digit
		if missing := int(scale) + 1 - len(digits); missing > 0 {
			digits = strings.Repeat("0", missing) + digits
		}
		split := len(digits) - int(scale)
		e.buf = append(e.buf, digits[:split]...)
		e.buf = append(e.buf, '.')
		digits = digits[split:]
	}
	e.buf = append(e.buf, digits...)
	e.buf = append(e.buf, '"')
}

// Float writes a number like Number.prototype.toString of JavaScript, but -0 as -0.
func (e *Encoder) Float(f float64) {
	switch {
	case math.IsNaN(f):
		e.fail("NaN")
		return
	case math.IsInf(f, 1):
		e.fail("+Inf")
		return
	case math.IsInf(f, -1):
		e.fail("-Inf")
		return
	case f == 0:
		if math.Signbit(f) {
			e.buf = append(e.buf, "-0"...)
		} else {
			e.buf = append(e.buf, '0')
		}
		return
	}
	format := byte('f')
	if a := math.Abs(f); a < 1e-6 || a >= 1e21 {
		format = 'e'
	}
	e.buf = strconv.AppendFloat(e.buf, f, format, -1, 64)
	if format == 'e' {
		// JavaScript writes e-7 instead of e-07
		n := len(e.buf)
		if n >= 4 && e.buf[n-4] == 'e' && e.buf[n-3] == '-' && e.buf[n-2] == '0' {
			e.buf[n-2] = e.buf[n-1]
			e.buf = e.buf[:n-1]
		}
	}
}

func (e *Encoder) fail(value string) {
	if e.err == "" {
		e.err = "json: unsupported value " + value
	}
	e.Null()
}

// String writes a string.
func (e *Encoder) String(s string) {
	if !utf8.ValidString(s) {
		s = strings.ToValidUTF8(s, "�")
	}
	e.buf = append(e.buf, '"')
	start := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 0x20 && c != '"' && c != '\\' {
			continue
		}
		e.buf = append(e.buf, s[start:i]...)
		switch c {
		case '"', '\\':
			e.buf = append(e.buf, '\\', c)
		case '\b':
			e.buf = append(e.buf, '\\', 'b')
		case '\f':
			e.buf = append(e.buf, '\\', 'f')
		case '\n':
			e.buf = append(e.buf, '\\', 'n')
		case '\r':
			e.buf = append(e.buf, '\\', 'r')
		case '\t':
			e.buf = append(e.buf, '\\', 't')
		default:
			e.buf = append(e.buf, '\\', 'u', '0', '0', hexDigit(c>>4), hexDigit(c&15))
		}
		start = i + 1
	}
	e.buf = append(e.buf, s[start:]...)
	e.buf = append(e.buf, '"')
}

func hexDigit(c byte) byte {
	return "0123456789abcdef"[c]
}

// Base64 writes bytes as a string of standard Base64 with padding.
func (e *Encoder) Base64(b []byte) {
	e.buf = append(e.buf, '"')
	e.buf = base64.StdEncoding.AppendEncode(e.buf, b)
	e.buf = append(e.buf, '"')
}

// Key writes the key of a field or an entry and its colon.
func (e *Encoder) Key(k string) {
	e.String(k)
	e.buf = append(e.buf, ':')
}

// OmitEmpty tells, whether the struct tag has the option omitempty, see [Encoder.Empty].
func OmitEmpty(tag string) bool {
	t, ok := reflect.StructTag(tag).Lookup("json")
	if !ok {
		return false
	}
	_, opts, _ := strings.Cut(t, ",")
	for opts != "" {
		var opt string
		opt, opts, _ = strings.Cut(opts, ",")
		if opt == "omitempty" {
			return true
		}
	}
	return false
}

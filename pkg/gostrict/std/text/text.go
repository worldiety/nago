// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package text provides the string functions of the gostrict subset, which behave the same in Go and in
// JavaScript. Go counts the bytes of UTF-8 and JavaScript the units of UTF-16, so the functions count characters,
// which are the code points of Unicode, unless they say otherwise. No function panics.
//
//gostrict:header
package text

import (
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"

	"go.wdy.de/nago/pkg/gostrict/std/list"
)

// RuneCount returns the number of characters, e.g. 1 for "ä" and for "😀".
func RuneCount(s string) int32 {
	return int32(utf8.RuneCountInString(s))
}

// ByteLen returns the number of bytes of s encoded as UTF-8, e.g. 2 for "ä".
func ByteLen(s string) int32 {
	return int32(len(s))
}

// Chars returns the characters of s, each as a string.
func Chars(s string) list.List[string] {
	l := list.New[string]()
	for _, r := range s {
		l.Append(string(r))
	}
	return l
}

// CodePoints returns the code points of the characters of s.
func CodePoints(s string) list.List[rune] {
	l := list.New[rune]()
	for _, r := range s {
		l.Append(r)
	}
	return l
}

// FromCodePoint returns the character of the code point, or "�" if it is not valid.
func FromCodePoint(r rune) string {
	return string(r)
}

// Compare compares a and b by their code points and returns -1, 0 or +1.
func Compare(a, b string) int32 {
	return int32(strings.Compare(a, b))
}

// Count returns the number of non-overlapping occurrences of sub in s. For an empty sub, it returns the number
// of characters plus one.
func Count(s, sub string) int32 {
	return int32(strings.Count(s, sub))
}

// Repeat returns n copies of s, or "" if n is not positive.
func Repeat(s string, n int32) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(s, int(n))
}

// Split splits s at each sep like strings.Split. An empty sep splits s into its characters.
func Split(s, sep string) list.List[string] {
	return list.Of(strings.Split(s, sep)...)
}

// Join concatenates the elements of l with sep between them.
func Join(l list.List[string], sep string) string {
	var sb strings.Builder
	for i, s := range l.All() {
		if i > 0 {
			sb.WriteString(sep)
		}
		sb.WriteString(s)
	}
	return sb.String()
}

// Index returns the index of the character, at which sub starts in s, or -1.
func Index(s, sub string) int32 {
	i := strings.Index(s, sub)
	if i < 0 {
		return -1
	}
	return int32(utf8.RuneCountInString(s[:i]))
}

// Slice returns the characters of s from the index from up to, but not including, the index to, joined like the
// elements of [Chars], e.g. "rü" for Slice("Grüße", 1, 3). A from below 0 counts as 0 and a to beyond the number
// of characters as that number. If from is not less than to, it returns "".
func Slice(s string, from, to int32) string {
	from = max(from, 0)
	if from >= to {
		return ""
	}
	var sb strings.Builder
	i := int32(0)
	for _, r := range s {
		if i >= to {
			break
		}
		if i >= from {
			// an invalid byte is utf8.RuneError, which Chars also returns as "�"
			sb.WriteRune(r)
		}
		i++
	}
	return sb.String()
}

// FromInt formats n in decimal, e.g. "-42".
func FromInt(n int32) string {
	return strconv.Itoa(int(n))
}

// FromInt64 formats n in decimal, e.g. "-42".
func FromInt64(n int64) string {
	return strconv.FormatInt(n, 10)
}

// FromUint64 formats n in decimal, e.g. "42".
func FromUint64(n uint64) string {
	return strconv.FormatUint(n, 10)
}

// FromFloat formats f with the number of decimals from 0 to 100, e.g. "3.14" for FromFloat(3.14159, 2). It rounds
// the exact binary value of f half to even, e.g. "0.12" for 0.125, which is exact, and "0.13" for 0.135, which is
// slightly more in binary. It writes no exponent, NaN as "NaN" and the infinities as "+Inf" and "-Inf".
func FromFloat(f float64, decimals int32) string {
	return strconv.FormatFloat(f, 'f', int(min(max(decimals, 0), 100)), 64)
}

// GroupDigits writes a decimal number like "-1234567.89", which has an optional sign, at least one digit and
// optional decimals after a point, with decimalMark instead of the point and groupMark between the groups of three
// digits before it, counted from the right, e.g. "-1.234.567,89" for GroupDigits("-1234567.89", ",", "."). An empty
// groupMark writes no groups. It returns anything else unchanged, e.g. "NaN", "+Inf" or "1e5". It suits the results
// of [FromFloat] and of decimal.String after decimal.Round, so that an amount is never formatted through a float64.
func GroupDigits(s, decimalMark, groupMark string) string {
	i := 0
	if len(s) > 0 && (s[0] == '+' || s[0] == '-') {
		i = 1
	}
	j := i
	for j < len(s) && isDigit(s[j]) {
		j++
	}
	if j == i {
		return s
	}
	if j < len(s) {
		if s[j] != '.' || j+1 == len(s) {
			return s
		}
		for k := j + 1; k < len(s); k++ {
			if !isDigit(s[k]) {
				return s
			}
		}
	}
	var sb strings.Builder
	sb.WriteString(s[:i])
	for k := i; k < j; k++ {
		if k > i && (j-k)%3 == 0 {
			sb.WriteString(groupMark)
		}
		sb.WriteByte(s[k])
	}
	if j < len(s) {
		sb.WriteString(decimalMark)
		sb.WriteString(s[j+1:])
	}
	return sb.String()
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// ParseInt parses a decimal integer with an optional sign, e.g. "-42". It fails for anything else and for numbers
// beyond the range of int32 and returns 0 then.
//
//gostrict:zero
func ParseInt(s string) (int32, error) {
	digits := s
	neg := false
	if len(digits) > 0 && (digits[0] == '+' || digits[0] == '-') {
		neg = digits[0] == '-'
		digits = digits[1:]
	}
	if digits == "" {
		return 0, errors.New("text.ParseInt: invalid syntax: " + s)
	}
	n := 0
	for i := 0; i < len(digits); i++ {
		c := digits[i]
		if c < '0' || c > '9' {
			return 0, errors.New("text.ParseInt: invalid syntax: " + s)
		}
		n = n*10 + int(c-'0')
		if n > 1<<31 {
			return 0, errors.New("text.ParseInt: value out of range: " + s)
		}
	}
	if neg {
		n = -n
	}
	if n > 1<<31-1 {
		return 0, errors.New("text.ParseInt: value out of range: " + s)
	}
	return int32(n), nil
}

// ParseInt64 parses a decimal integer with an optional sign, like [ParseInt], for the range of int64.
//
//gostrict:zero
func ParseInt64(s string) (int64, error) {
	if !isDecimalInt(s) {
		return 0, errors.New("text.ParseInt64: invalid syntax: " + s)
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, errors.New("text.ParseInt64: value out of range: " + s)
	}
	return v, nil
}

// ParseUint64 parses a decimal integer with an optional +, for the range of uint64.
//
//gostrict:zero
func ParseUint64(s string) (uint64, error) {
	if !isDecimalInt(s) || strings.HasPrefix(s, "-") {
		return 0, errors.New("text.ParseUint64: invalid syntax: " + s)
	}
	v, err := strconv.ParseUint(strings.TrimPrefix(s, "+"), 10, 64)
	if err != nil {
		return 0, errors.New("text.ParseUint64: value out of range: " + s)
	}
	return v, nil
}

func isDecimalInt(s string) bool {
	if len(s) > 0 && (s[0] == '+' || s[0] == '-') {
		s = s[1:]
	}
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// ParseFloat parses a decimal number with an optional sign, digits with an optional point and an optional exponent
// with e or E, e.g. "-1.5", ".5", "5." or "2.5E-3", into the nearest float64 and rounds a tie to even, for any
// number of digits and any exponent. The sign holds also for zero, so "-0" and "-1e-400" are -0. It fails for
// anything else, e.g. a comma, a space, an underscore, "Inf", "NaN" or a hexadecimal number, and for a number, which
// would round to an infinity, but not for one, which rounds to 0 or to a subnormal number. On failure it returns 0.
//
//gostrict:zero
func ParseFloat(s string) (float64, error) {
	neg, digits, exp, ok := scanFloat(s)
	if !ok {
		return 0, errors.New("text.ParseFloat: invalid syntax: " + s)
	}
	var f float64
	switch {
	case digits == "" || exp < -323:
		// 0.digits * 10^exp is less than 10^-324, which is less than half of the least subnormal number
	case exp > 309:
		// at least 10^309, which is beyond the greatest float64
		return 0, errors.New("text.ParseFloat: value out of range: " + s)
	default:
		// strconv rounds exactly for any number of digits, but caps the exponent of its input, which may then
		// differ from the exponent of s with many zeros, e.g. "0.00…01e100005", so it gets the small exponent here
		var err error
		f, err = strconv.ParseFloat("0."+digits+"e"+strconv.FormatInt(exp, 10), 64)
		if err != nil {
			return 0, errors.New("text.ParseFloat: value out of range: " + s)
		}
	}
	if neg {
		f = -f
	}
	return f, nil
}

// scanFloat splits a number of the form of [ParseFloat] into its sign, its digits without the zeros at both ends
// and the exponent, so that its magnitude is 0.digits * 10^exp. It reports false for any other form.
func scanFloat(s string) (neg bool, digits string, exp int64, ok bool) {
	i := 0
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		neg = s[i] == '-'
		i++
	}
	var sb strings.Builder
	seen, point := false, false
	for ; i < len(s); i++ {
		c := s[i]
		if c == '.' && !point {
			point = true
			continue
		}
		if !isDigit(c) {
			break
		}
		seen = true
		switch {
		case c == '0' && sb.Len() == 0:
			// a leading zero after the point moves the digits to the right
			if point {
				exp--
			}
		case point:
			sb.WriteByte(c)
		default:
			sb.WriteByte(c)
			exp++
		}
	}
	if !seen {
		return false, "", 0, false
	}
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		i++
		eneg := false
		if i < len(s) && (s[i] == '+' || s[i] == '-') {
			eneg = s[i] == '-'
			i++
		}
		// the exponent stops growing at maxExp: any number of digits is far less, so a number with a greater
		// exponent is out of range or 0 anyway
		const maxExp = 1e12
		start := i
		var e int64
		for ; i < len(s) && isDigit(s[i]); i++ {
			if e < maxExp {
				e = e*10 + int64(s[i]-'0')
			}
		}
		if i == start {
			return false, "", 0, false
		}
		if eneg {
			e = -e
		}
		exp += e
	}
	if i < len(s) {
		return false, "", 0, false
	}
	return neg, strings.TrimRight(sb.String(), "0"), exp, true
}

// UpperASCII maps the letters a to z to upper case and keeps any other character.
func UpperASCII(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' {
			return r - 'a' + 'A'
		}
		return r
	}, s)
}

// LowerASCII maps the letters A to Z to lower case and keeps any other character.
func LowerASCII(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r - 'A' + 'a'
		}
		return r
	}, s)
}

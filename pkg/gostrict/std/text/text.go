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

// FromInt formats n in decimal, e.g. "-42".
func FromInt(n int32) string {
	return strconv.Itoa(int(n))
}

// FromFloat formats f with the number of decimals from 0 to 100, e.g. "3.14" for FromFloat(3.14159, 2). It rounds
// the exact binary value of f half to even, e.g. "0.12" for 0.125, which is exact, and "0.13" for 0.135, which is
// slightly more in binary. It writes no exponent, NaN as "NaN" and the infinities as "+Inf" and "-Inf".
func FromFloat(f float64, decimals int32) string {
	return strconv.FormatFloat(f, 'f', int(min(max(decimals, 0), 100)), 64)
}

// ParseInt parses a decimal integer with an optional sign, e.g. "-42". It fails for anything else and for numbers
// beyond the range of int32.
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

// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"strings"
	"testing"
)

func TestDecoderDecimal(t *testing.T) {
	cases := []struct {
		in     string
		digits string
		scale  int32
		err    string
	}{
		{in: `12.30`, digits: "1230", scale: 2},
		{in: `"12.30"`, digits: "1230", scale: 2},
		{in: ` -0.50`, digits: "-50", scale: 2},
		{in: `"-0.00"`, digits: "0", scale: 2},
		{in: `-0`, digits: "0", scale: 0},
		{in: `007`, err: "json: unexpected data after the value at offset 1"},
		{in: `""`, err: "json: expected a decimal at offset 0"},
		{in: `"1 "`, err: "json: expected a decimal at offset 0"},
		{in: `"007"`, err: "json: expected a decimal at offset 0"},
		{in: `1.5e3`, digits: "1500", scale: 0},
		{in: `1.5E-3`, digits: "15", scale: 4},
		{in: `"1e2"`, digits: "100", scale: 0},
		{in: `0e5`, digits: "0", scale: 0},
		{in: `123456789012345678901234567890.123456789`, digits: "123456789012345678901234567890123456789", scale: 9},
		{in: `"0." `, err: "json: expected a decimal at offset 0"},
		{in: `" 1"`, err: "json: expected a decimal at offset 0"},
		{in: `"1,5"`, err: "json: expected a decimal at offset 0"},
		{in: `true`, err: "json: expected a decimal at offset 0"},
		{in: `x`, err: "json: unexpected character 'x' at offset 0"},
		{in: `1e1001`, err: "json: number out of range at offset 0"},
		{in: `1e1000`, digits: "1" + strings.Repeat("0", 1000), scale: 0},
		{in: `1e-1000`, digits: "1", scale: 1000},
		{in: `1.5e-1000`, err: "json: number out of range at offset 0"},
		{in: `1e99999999999999999999`, err: "json: number out of range at offset 0"},
		{in: `"0.` + strings.Repeat("1", 1001) + `"`, err: "json: number out of range at offset 0"},
	}
	for _, c := range cases {
		d := New([]byte(c.in))
		digits, scale := d.Decimal()
		d.End()
		if c.err != "" {
			if d.Err() != c.err {
				t.Errorf("%s: error %q, want %q", c.in, d.Err(), c.err)
			}
			continue
		}
		if d.Err() != "" || digits != c.digits || scale != c.scale {
			t.Errorf("%s: %q, %d, %q, want %q, %d", c.in, digits, scale, d.Err(), c.digits, c.scale)
		}
	}
}

func TestEncoderDecimal(t *testing.T) {
	cases := []struct {
		digits string
		scale  int32
		want   string
	}{
		{"1230", 2, `"12.30"`},
		{"-1230", 2, `"-12.30"`},
		{"5", 3, `"0.005"`},
		{"-5", 3, `"-0.005"`},
		{"0", 2, `"0.00"`},
		{"42", 0, `"42"`},
		{"123", 3, `"0.123"`},
		// a scale out of range, which only a literal creates, is clamped like decimal.String does
		{"5", -2, `"5"`},
		{"1", 1001, `"0.` + strings.Repeat("0", 999) + `1"`},
	}
	for _, c := range cases {
		var e Encoder
		e.Decimal(c.digits, c.scale)
		got, _ := e.Bytes()
		if string(got) != c.want {
			t.Errorf("%s, %d: %s, want %s", c.digits, c.scale, got, c.want)
		}
	}
}

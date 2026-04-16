// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rest

import (
	"testing"
)

func TestEncodeZipKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want string
	}{
		{"uuid", "550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440000"},
		{"hex", "deadbeef1234abcd", "deadbeef1234abcd"},
		{"simple name", "hello_world.json", "hello_world.json"},
		{"empty", "", ""},
		{"slash", "users/123/profile", "users%2F123%2Fprofile"},
		{"colon", "bucket:key", "bucket%3Akey"},
		{"space", "hello world", "hello%20world"},
		{"backslash", `win\path`, "win%5Cpath"},
		{"unicode", "über/café", "%C3%BCber%2Fcaf%C3%A9"},
		{"path traversal", "../etc/passwd", "..%2Fetc%2Fpasswd"},
		{"windows reserved", "NUL:COM1", "NUL%3ACOM1"},
		{"percent literal", "100%done", "100%25done"},
		{"mixed", "org.example/user:42 data", "org.example%2Fuser%3A42%20data"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodeZipKey(tt.key)
			if got != tt.want {
				t.Errorf("EncodeZipKey(%q)\n got  %q\n want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestDecodeZipKey_Roundtrip(t *testing.T) {
	keys := []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"deadbeef1234abcd",
		"hello_world.json",
		"",
		"users/123/profile",
		"bucket:key",
		"hello world",
		`win\path`,
		"über/café",
		"../etc/passwd",
		"NUL:COM1",
		"100%done",
		"org.example/user:42 data",
		"日本語キー",
		string([]byte{0x00, 0x01, 0x1f, 0x7f, 0xff}), // control / high bytes
	}

	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			encoded := EncodeZipKey(key)
			decoded, err := DecodeZipKey(encoded)
			if err != nil {
				t.Fatalf("DecodeZipKey(%q) unexpected error: %v", encoded, err)
			}
			if decoded != key {
				t.Errorf("roundtrip failed:\n got  %q\n want %q", decoded, key)
			}
		})
	}
}

func TestDecodeZipKey_Errors(t *testing.T) {
	cases := []struct {
		name    string
		encoded string
	}{
		{"bare percent", "%"},
		{"truncated one char", "%2"},
		{"invalid hex upper", "%GG"},
		{"invalid hex lower", "%zz"},
		{"mixed invalid", "abc%FZ"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := DecodeZipKey(c.encoded)
			if err == nil {
				t.Errorf("DecodeZipKey(%q) expected error, got nil", c.encoded)
			}
		})
	}
}


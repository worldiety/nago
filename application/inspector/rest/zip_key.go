// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rest

import (
	"fmt"
	"strings"
)

// EncodeZipKey encodes a blob key for use as a filename inside a ZIP archive.
// Characters safe for ZIP filenames on all common operating systems — alphanumeric,
// '-', '_' and '.' — are kept verbatim so that simple keys like UUIDs or hex strings
// remain human-readable in the archive.
//
// Every other byte (including '/', ':', '\', space and arbitrary UTF-8 sequences) is
// percent-encoded as %XX (uppercase hex).  The encoding is fully reversible via
// [DecodeZipKey], which must be used during import to recover the original blob key.
func EncodeZipKey(key string) string {
	// fast path: check whether encoding is needed at all
	needsEncoding := false
	for i := 0; i < len(key); i++ {
		if !isSafeZipByte(key[i]) {
			needsEncoding = true
			break
		}
	}

	if !needsEncoding {
		return key
	}

	var buf strings.Builder
	buf.Grow(len(key) * 3 / 2) // rough upper bound

	for i := 0; i < len(key); i++ {
		b := key[i]
		if isSafeZipByte(b) {
			buf.WriteByte(b)
		} else {
			buf.WriteByte('%')
			buf.WriteByte(hexNibble(b >> 4))
			buf.WriteByte(hexNibble(b & 0x0f))
		}
	}

	return buf.String()
}

// DecodeZipKey reverses [EncodeZipKey] and returns the original blob key.
// It returns an error when the input contains a malformed percent-encoding sequence.
func DecodeZipKey(encoded string) (string, error) {
	// fast path: no percent sign, nothing to decode
	if !strings.ContainsRune(encoded, '%') {
		return encoded, nil
	}

	var buf strings.Builder
	buf.Grow(len(encoded))

	for i := 0; i < len(encoded); {
		b := encoded[i]
		if b != '%' {
			buf.WriteByte(b)
			i++
			continue
		}

		if i+2 >= len(encoded) {
			return "", fmt.Errorf("zip key: truncated percent-encoding at position %d", i)
		}

		hi, ok1 := fromHexNibble(encoded[i+1])
		lo, ok2 := fromHexNibble(encoded[i+2])
		if !ok1 || !ok2 {
			return "", fmt.Errorf("zip key: invalid hex digits %q at position %d", encoded[i:i+3], i)
		}

		buf.WriteByte(hi<<4 | lo)
		i += 3
	}

	return buf.String(), nil
}

// isSafeZipByte reports whether b can appear verbatim in a ZIP filename
// without causing issues on any common operating system or ZIP tool.
func isSafeZipByte(b byte) bool {
	return (b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9') ||
		b == '-' || b == '_' || b == '.'
}

func hexNibble(v byte) byte {
	const digits = "0123456789ABCDEF"
	return digits[v&0x0f]
}

func fromHexNibble(b byte) (byte, bool) {
	switch {
	case b >= '0' && b <= '9':
		return b - '0', true
	case b >= 'A' && b <= 'F':
		return b - 'A' + 10, true
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10, true
	default:
		return 0, false
	}
}


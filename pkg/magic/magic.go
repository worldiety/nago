// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package magic

import (
	"bytes"
	"go.wdy.de/nago/pkg/blob/crypto"
	"net/http"
	"unicode"
)

var zipMagicBytes = [][]byte{
	{0x50, 0x4B, 0x03, 0x04}, // Normale ZIP-Datei
	{0x50, 0x4B, 0x05, 0x06}, // Leere ZIP-Datei (End of Central Directory)
	{0x50, 0x4B, 0x07, 0x08}, // Spanned ZIP-Datei
}

// Detect returns the estimated mimetype of the given buffer.
func Detect(buf []byte) string {
	if bytes.HasPrefix(buf, []byte("%PDF-")) {
		return "application/pdf"
	}

	for _, m := range zipMagicBytes {
		if bytes.HasPrefix(buf, m) {
			return "application/zip"
		}
	}

	if crypto.IsEncrypted(buf) {
		return "application/x-nago-encrypted"
	}

	if bytes.HasPrefix(buf, []byte("{{")) {
		return "text/html"
	}

	if maybeJson(buf) {
		return "application/json"
	}

	return http.DetectContentType(buf)
}

// Ext returns the typical estimated filename extensions
func Ext(buf []byte) string {
	switch Detect(buf) {
	case "application/pdf":
		return ".pdf"
	case "application/zip":
		return ".zip"
	default:
		return ".bin"
	}
}

func maybeJson(buf []byte) bool {
	for _, b := range buf {
		// inspect without producing any garbage
		if unicode.IsSpace(rune(b)) {
			continue
		}

		if b == '{' || b == '[' {
			// looks like a start of an object or array. Thus, a simple text starting with this, will be falsely detected as json
			return true
		} else {
			// anything else is definitely not JSON
			return false
		}
	}

	return false
}

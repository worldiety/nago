// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package magic

import (
	"bytes"
	"net/http"
	"unicode"

	"go.wdy.de/nago/pkg/blob/crypto"
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

	if maybeSVG(buf) {
		return "image/svg+xml"
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

func maybeSVG(buf []byte) bool {
	// skip a possible UTF-8 BOM
	buf = bytes.TrimPrefix(buf, []byte{0xEF, 0xBB, 0xBF})

	// only inspect the beginning of the document to keep detection cheap
	const window = 1024
	if len(buf) > window {
		buf = buf[:window]
	}

	for {
		buf = bytes.TrimLeftFunc(buf, unicode.IsSpace)
		if len(buf) == 0 || buf[0] != '<' {
			return false
		}

		// Skip xml declaration, comments, doctype or other markup declarations
		// Only look for elements with <svg* as root
		switch {
		case bytes.HasPrefix(buf, []byte("<?")):
			end := bytes.Index(buf, []byte("?>"))
			if end < 0 {
				return false
			}
			buf = buf[end+2:]

		case bytes.HasPrefix(buf, []byte("<!--")):
			end := bytes.Index(buf[4:], []byte("-->"))
			if end < 0 {
				return false
			}
			buf = buf[4+end+3:]

		case bytes.HasPrefix(buf, []byte("<!")):
			end := bytes.IndexByte(buf, '>')
			if end < 0 {
				return false
			}
			buf = buf[end+1:]

		default:
			lower := bytes.ToLower(buf)
			return bytes.HasPrefix(lower, []byte("<svg>")) || bytes.HasPrefix(lower, []byte("<svg ")) ||
				bytes.HasPrefix(lower, []byte("<svg/")) || bytes.HasPrefix(lower, []byte("<svg\t")) ||
				bytes.HasPrefix(lower, []byte("<svg\n")) || bytes.HasPrefix(lower, []byte("<svg\r"))
		}
	}
}

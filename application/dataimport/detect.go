// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataimport

import "unicode"

func maybeJsonObj(buf []byte) bool {
	for _, b := range buf {
		// inspect without producing any garbage
		if unicode.IsSpace(rune(b)) {
			continue
		}

		return b == '{'
	}

	return false
}

func maybeJsonArray(buf []byte) bool {
	for _, b := range buf {
		// inspect without producing any garbage
		if unicode.IsSpace(rune(b)) {
			continue
		}

		return b == '['
	}

	return false
}

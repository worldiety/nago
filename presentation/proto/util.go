// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package proto

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
)

var validComponentIdRegex = regexp.MustCompile(`[A-Za-z0-9_\-{/}]+[/*]?`)

// NewScopeID returns a hex encoded and filename safe 32 byte entropy.
func NewScopeID() ScopeID {
	var tmp [32]byte
	if _, err := rand.Read(tmp[:]); err != nil {
		panic(err)
	}

	return ScopeID(hex.EncodeToString(tmp[:]))
}

// Valid identifies a unique constructor for a specific ComponentType.
// Such an addressable Component is likely a page and instantiated and rendered.
// In return, a ComponentInvalidated event will be sent in the future.
// For details, see the [NewComponentRequested] event.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
func (c RootViewID) Valid() bool {
	if c == "." {
		return true
	}

	if strings.HasPrefix(string(c), "/") || strings.HasSuffix(string(c), "/") {
		return false
	}

	return validComponentIdRegex.FindString(string(c)) == string(c)
}

func (c RootViewID) IsWildcard() bool {
	return strings.HasSuffix(string(c), "/*")
}

func (c RootViewID) Matches(other RootViewID) bool {
	if c == other {
		return true
	}

	if c.IsWildcard() {
		return strings.HasPrefix(string(other), string(c[:len(c)-1]))
	}

	return false
}

func NewStrings(str []string) Strings {
	if str == nil {
		return nil
	}

	tmp := make([]Str, len(str))
	for i, str := range str {
		tmp[i] = Str(str)
	}

	return tmp
}

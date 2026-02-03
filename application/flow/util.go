// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"fmt"
	"go/token"
	"unicode"

	"go.wdy.de/nago/presentation/ui"
)

func validateIdentifier(s string) error {
	if s == "" {
		return fmt.Errorf("identifier cannot be empty")
	}

	for i, r := range s {
		if i == 0 {
			if !unicode.IsLetter(r) && r != '_' {
				return fmt.Errorf("must start with letter or underscore")
			}
		} else {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				return fmt.Errorf("invalid character %q", r)
			}
		}
	}

	if token.Lookup(s).IsKeyword() {
		return fmt.Errorf("%q is a reserved keyword", s)
	}

	return nil
}

func NumberOnly(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func FixLength(s ui.Length) ui.Length {
	if s == "" {
		return ""
	}

	if NumberOnly(string(s)) {
		return s + "rem"
	}

	return s
}

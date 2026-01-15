// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"errors"
	"unicode"
)

type Ident string

func (i Ident) Validate() error {
	if len(i) == 0 {
		return errors.New("identifier must not be empty")
	}

	for idx, r := range i {
		if idx == 0 {
			if !isIdentStart(r) {
				return errors.New("identifier must start with letter or '_'")
			}
		} else {
			if !isIdentPart(r) {
				return errors.New("identifier contains invalid character")
			}
		}
	}

	return nil
}

func (i Ident) IsPublic() bool {
	if len(i) == 0 {
		return false
	}
	r := []rune(i)[0]
	return unicode.IsUpper(r)
}

func (i Ident) IsPrivate() bool {
	if len(i) == 0 {
		return false
	}
	r := []rune(i)[0]
	return unicode.IsLower(r) || r == '_'
}

func isIdentStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isIdentPart(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

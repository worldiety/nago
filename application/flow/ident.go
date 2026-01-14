// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"fmt"
	"unicode"
)

type Ident string

func (i Ident) Validate() error {
	if len(i) == 0 {
		return fmt.Errorf("identifier must not be empty")
	}

	first := true
	pos := 0
	for _, r := range i {
		if first {
			if !unicode.IsUpper(r) || !unicode.IsLetter(r) {
				return fmt.Errorf("identifier must start with an uppercase letter")
			}
			first = false
		} else {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				return fmt.Errorf("invalid character %q at position %d", r, pos)
			}
		}
		pos++
	}

	return nil
}

// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	nui "go.wdy.de/nago/presentation/ui"
)

// TKeyboardOptions are the options of the virtual keyboard, e.g. the capitalization, see KeyboardOptions. It is a
// building block of a text field and has no view of its own.
type TKeyboardOptions struct{ ref js.Ref }

// KeyboardOptions returns the options of the keyboard with the defaults.
//
//gostrict:host
func KeyboardOptions() TKeyboardOptions {
	return TKeyboardOptions{ref: wrap(nui.KeyboardOptions())}
}

// Capitalization enables or disables the automatic capitalization.
//
//gostrict:host
func (c TKeyboardOptions) Capitalization(capitalization bool) TKeyboardOptions {
	return TKeyboardOptions{ref: wrap(get[nui.TKeyboardOptions](c.ref).Capitalization(capitalization))}
}

// AutoCorrectEnabled enables or disables the auto-correction.
//
//gostrict:host
func (c TKeyboardOptions) AutoCorrectEnabled(autoCorrectEnabled bool) TKeyboardOptions {
	return TKeyboardOptions{ref: wrap(get[nui.TKeyboardOptions](c.ref).AutoCorrectEnabled(autoCorrectEnabled))}
}

// KeyboardType sets the type of the keyboard. It is a hint to the frontend. Technically, it is impossible to actually
// guarantee anything, and you always have to consider bugs and hacks: a malicious user may send you anything, the
// keyboards of Android are never guaranteed, and a user may inject anything using autocompletion or the clipboard.
//
//gostrict:host
func (c TKeyboardOptions) KeyboardType(keyboardType KeyboardType) TKeyboardOptions {
	return TKeyboardOptions{ref: wrap(get[nui.TKeyboardOptions](c.ref).KeyboardType(nui.KeyboardType(keyboardType)))}
}

// KeyboardType is the type of the virtual keyboard, which is a hint to the frontend, see TKeyboardOptions.KeyboardType.
type KeyboardType int32

const (
	KeyboardDefault KeyboardType = 0
	KeyboardAscii   KeyboardType = 1
	// KeyboardInteger shows a numeric keypad and formats the input as an integer, which drops leading zeros. Use
	// KeyboardPhone for digit sequences like codes or postal codes.
	KeyboardInteger KeyboardType = 2
	KeyboardFloat   KeyboardType = 3
	KeyboardEMail   KeyboardType = 4
	KeyboardPhone   KeyboardType = 5
	KeyboardSearch  KeyboardType = 6
	KeyboardURL     KeyboardType = 7
)

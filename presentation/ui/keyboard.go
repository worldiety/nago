// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import "go.wdy.de/nago/presentation/proto"

// TKeyboardOptions is a utility component(Keyboard Options).
// Keyboard Options defines configuration options for virtual keyboard behavior.
// It allows customization of capitalization, auto-correction, and keyboard type hints.
// These options are primarily used in text input components to enhance user experience.
type TKeyboardOptions struct {
	capitalization     bool
	autoCorrectEnabled bool
	keyboardType       KeyboardType
}

func KeyboardOptions() TKeyboardOptions {
	return TKeyboardOptions{}
}

func (opts TKeyboardOptions) ora() proto.KeyboardOptions {
	return proto.KeyboardOptions{
		Capitalization:     proto.Bool(opts.capitalization),
		AutoCorrectEnabled: proto.Bool(opts.autoCorrectEnabled),
		KeyboardType:       opts.keyboardType.ora(),
	}
}

// Capitalization enables or disables automatic capitalization.
func (opts TKeyboardOptions) Capitalization(capitalization bool) TKeyboardOptions {
	opts.capitalization = capitalization
	return opts
}

// AutoCorrectEnabled enables or disables auto-correction.
func (opts TKeyboardOptions) AutoCorrectEnabled(autoCorrectEnabled bool) TKeyboardOptions {
	opts.autoCorrectEnabled = autoCorrectEnabled
	return opts
}

// KeyboardType is a hint to the frontend. Technically, it is impossible to actually
// guarantee anything, and you have always to considers bugs and hacks:
//   - a malicious user may send you anything, which would otherwise not be possible (e.g. text instead of numbers)
//   - Android IME hints or keyboard types are never guaranteed. A user may install third-party keyboards which just ignore anything
//   - a user may inject anything using wrong autocompletion or the clipboard
func (opts TKeyboardOptions) KeyboardType(keyboardType KeyboardType) TKeyboardOptions {
	opts.keyboardType = keyboardType
	return opts
}

type KeyboardType int

func (k KeyboardType) ora() proto.KeyboardType {
	switch k {
	case KeyboardAscii:
		return proto.KeyboardAscii
	case KeyboardInteger:
		return proto.KeyboardInteger
	case KeyboardFloat:
		return proto.KeyboardFloat
	case KeyboardEMail:
		return proto.KeyboardEMail
	case KeyboardPhone:
		return proto.KeyboardPhone
	case KeyboardSearch:
		return proto.KeyboardSearch
	case KeyboardURL:
		return proto.KeyboardURL
	default:
		return proto.KeyboardDefault
	}
}

const (
	KeyboardDefault KeyboardType = iota
	KeyboardAscii
	KeyboardInteger
	KeyboardFloat
	KeyboardEMail
	KeyboardPhone
	KeyboardSearch
	KeyboardURL
)

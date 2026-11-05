// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package stdvalues

import "go.wdy.de/nago/pkg/gostrict/std/text"

// TextRepeat repeats a text through text.Repeat as a value, a function without type parameters.
func TextRepeat(a, b int32) int32 {
	f := text.Repeat
	return text.RuneCount(f("ab", a)) + b
}

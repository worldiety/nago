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

// TRichText is a block of rich text, see RichText. The text may contain formatting, e.g. bold, italic or links.
type TRichText struct{ ref js.Ref }

func (TRichText) isView() {}

// RichText returns a block of rich text with the value.
//
//gostrict:host
func RichText(value string) TRichText {
	return TRichText{ref: wrap(nui.RichText(value))}
}

// Value sets the rich text to be shown.
//
//gostrict:host
func (c TRichText) Value(value string) TRichText {
	return TRichText{ref: wrap(get[nui.TRichText](c.ref).Value(value))}
}

// Frame sets the layout frame of the rich text, e.g. its size.
//
//gostrict:host
func (c TRichText) Frame(frame Frame) TRichText {
	return TRichText{ref: wrap(get[nui.TRichText](c.ref).Frame(frame.nago()))}
}

// FullWidth expands the rich text to the full available width.
//
//gostrict:host
func (c TRichText) FullWidth() TRichText {
	return TRichText{ref: wrap(get[nui.TRichText](c.ref).FullWidth())}
}

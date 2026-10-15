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
	"go.wdy.de/nago/presentation/ui/markdown"
)

// MarkdownRichText returns a rich text of the markdown value, like markdown.RichText of nago.
//
//gostrict:host
func MarkdownRichText(value string) TRichText {
	return TRichText{ref: wrap(markdown.RichText(value))}
}

// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package ui

import (
	nui "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/document"
)

// The conversions of batch c5, see native.go.

func (s DocumentSize) nago() document.Size {
	return document.Size{Width: nui.Length(s.Width), Height: nui.Length(s.Height)}
}

func documentSizeFrom(v document.Size) DocumentSize {
	return DocumentSize{Width: Length(v.Width), Height: Length(v.Height)}
}

func documentCommentsOf(comments []TDocumentComment) []document.TComment {
	res := make([]document.TComment, 0, len(comments))
	for _, c := range comments {
		res = append(res, get[document.TComment](c.ref))
	}
	return res
}

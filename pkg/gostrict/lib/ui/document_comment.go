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
	"go.wdy.de/nago/presentation/ui/document"
)

// TDocumentComment is a comment, which overlays a view, like document.TComment of nago. See DocumentComment.
type TDocumentComment struct{ ref js.Ref }

func (TDocumentComment) isView() {}

// DocumentComment wraps the view as a comment, like document.Comment of nago.
//
//gostrict:host
func DocumentComment(v View) TDocumentComment {
	return TDocumentComment{ref: wrap(document.Comment(viewOf(v)))}
}

// InputSelectedValue binds the selected state of the comment.
//
//gostrict:host
func (c TDocumentComment) InputSelectedValue(selected BoolState) TDocumentComment {
	return TDocumentComment{ref: wrap(get[document.TComment](c.ref).InputSelectedValue(selected.nago()))}
}

// InputValue binds the text state of the comment.
//
//gostrict:host
func (c TDocumentComment) InputValue(comment StringState) TDocumentComment {
	return TDocumentComment{ref: wrap(get[document.TComment](c.ref).InputValue(comment.nago()))}
}

// Resolve sets the action, which marks the comment thread as resolved.
//
//gostrict:host
func (c TDocumentComment) Resolve(action func()) TDocumentComment {
	return TDocumentComment{ref: wrap(get[document.TComment](c.ref).Resolve(action))}
}

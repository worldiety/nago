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
	"go.wdy.de/nago/presentation/core"
	nui "go.wdy.de/nago/presentation/ui"
)

// TPDF is a viewer of a PDF file, which is loaded from a source URL, see PDF.
type TPDF struct{ ref js.Ref }

func (TPDF) isView() {}

// PDF returns a viewer of the PDF file, which is loaded from src.
//
//gostrict:host
func PDF(src URI) TPDF {
	return TPDF{ref: wrap(nui.PDF(core.URI(src)))}
}

// Src sets the source URL of the PDF file.
//
//gostrict:host
func (c TPDF) Src(src URI) TPDF {
	return TPDF{ref: wrap(get[nui.TPDF](c.ref).Src(core.URI(src)))}
}

// Frame sets the size of the viewer.
//
//gostrict:host
func (c TPDF) Frame(frame Frame) TPDF {
	return TPDF{ref: wrap(get[nui.TPDF](c.ref).Frame(frame.nago()))}
}

// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// TPDF is a component to view PDF files from a source URL.
type TPDF struct {
	src   core.URI // the source URL of the PDF file
	frame Frame    // layout frame for sizing
}

// PDF creates a new PDF viewer component with the given source URL.
func PDF(src core.URI) TPDF {
	return TPDF{src: src}
}

// Src sets the source URL of the PDF viewer.
func (c TPDF) Src(src core.URI) TPDF {
	c.src = src
	return c
}

// Frame sets the viewer's frame for sizing purposes
func (c TPDF) Frame(frame Frame) TPDF {
	c.frame = frame
	return c
}

func (c TPDF) Render(_ core.RenderContext) core.RenderNode {
	return &proto.PDF{
		Src:   proto.Str(c.src),
		Frame: c.frame.ora(),
	}
}

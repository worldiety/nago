// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"net/url"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// TPDF is a component to view PDF files from a source URL.
type TPDF struct {
	src   url.URL // the source URL of the PDF file
	frame Frame   // layout frame for sizing
}

// PDF creates a new PDF viewer component with the given source URL.
func PDF(src url.URL) TPDF {
	return TPDF{src: src}
}

// Src sets the source URL of the PDF viewer.
func (c TPDF) Src(src url.URL) TPDF {
	c.src = src
	return c
}

// SrcFromString parses the given string as a URL and sets it as the source URL of the PDF viewer.
func (c TPDF) SrcFromString(src string) (TPDF, error) {
	parsed, err := url.Parse(src)
	if err != nil {
		return c, err
	}

	c.src = *parsed

	return c, nil
}

// Frame sets the viewer's frame for sizing purposes
func (c TPDF) Frame(frame Frame) TPDF {
	c.frame = frame
	return c
}

func (c TPDF) Render(_ core.RenderContext) core.RenderNode {
	return &proto.PDF{
		Src:   proto.Str(c.src.String()),
		Frame: c.frame.ora(),
	}
}

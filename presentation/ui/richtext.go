// Copyright (c) 2025 worldiety GmbH
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

// TRichText is a basic component (Rich Text).
// It renders a block of rich text content, which may include formatting
// such as bold, italic, links, or other markup depending on the renderer.
// The component supports layout control via frame settings.
type TRichText struct {
	value string // raw rich text value
	frame Frame  // layout frame
}

// RichText creates a new rich text component with the given value.
func RichText(value string) TRichText {
	return TRichText{
		value: value,
	}
}

// Frame sets the layout frame for the rich text component.
func (c TRichText) Frame(frame Frame) TRichText {
	c.frame = frame
	return c
}

// FullWidth expands the rich text component to take the full available width.
func (c TRichText) FullWidth() TRichText {
	c.frame = c.frame.FullWidth()
	return c
}

// Render builds and returns the protocol representation of the rich text component,
// including its value and layout frame.
func (c TRichText) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.RichText{
		Value: proto.Str(c.value),
		Frame: c.frame.ora(),
	}
}

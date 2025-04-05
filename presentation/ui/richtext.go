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

type TRichText struct {
	value string
	frame Frame
}

func RichText(value string) TRichText {
	return TRichText{
		value: value,
	}
}

func (c TRichText) Frame(frame Frame) TRichText {
	c.frame = frame
	return c
}

func (c TRichText) FullWidth() TRichText {
	c.frame = c.frame.FullWidth()
	return c
}

func (c TRichText) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.RichText{
		Value: proto.Str(c.value),
		Frame: c.frame.ora(),
	}
}

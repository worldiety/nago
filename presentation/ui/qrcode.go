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

type TQrCode struct {
	value              string
	accessibilityLabel string
	frame              Frame
}

func QrCode(value string) TQrCode {
	return TQrCode{
		value: value,
	}
}

func (c TQrCode) Frame(frame Frame) TQrCode {
	c.frame = frame
	return c
}

// AccessibilityLabel sets a label for screen readers. See also https://www.w3.org/WAI/tutorials/images/decision-tree/.
func (c TQrCode) AccessibilityLabel(label string) TQrCode {
	c.accessibilityLabel = label
	return c
}

func (c TQrCode) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.QrCode{
		Value:              proto.Str(c.value),
		AccessibilityLabel: proto.Str(c.accessibilityLabel),
		Frame:              c.frame.ora(),
	}
}

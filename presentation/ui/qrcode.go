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

// TQrCode is a basic component (QR Code).
// It generates and displays a QR code based on a given value (string).
// The component supports accessibility labeling and can be styled using a frame.
type TQrCode struct {
	value              string // value encoded into the QR code
	accessibilityLabel string // label for screen readers
	frame              Frame  // layout frame for size and positioning
}

// QrCode creates a new QR code with the provided value.
func QrCode(value string) TQrCode {
	return TQrCode{
		value: value,
	}
}

// Frame sets the layout frame for the QR code.
func (c TQrCode) Frame(frame Frame) TQrCode {
	c.frame = frame
	return c
}

// AccessibilityLabel sets a label for screen readers, improving accessibility.
// See: https://www.w3.org/WAI/tutorials/images/decision-tree/
func (c TQrCode) AccessibilityLabel(label string) TQrCode {
	c.accessibilityLabel = label
	return c
}

// Render builds and returns the protocol representation of the QR code,
// including the encoded value, accessibility label, and frame.
func (c TQrCode) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.QrCode{
		Value:              proto.Str(c.value),
		AccessibilityLabel: proto.Str(c.accessibilityLabel),
		Frame:              c.frame.ora(),
	}
}

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
	nui "go.wdy.de/nago/presentation/ui"
)

// TQrCode is a QR code, which encodes a value, see QrCode.
type TQrCode struct{ ref js.Ref }

func (TQrCode) isView() {}

// QrCode returns a QR code, which encodes the value.
//
//gostrict:host
func QrCode(value string) TQrCode {
	return TQrCode{ref: wrap(nui.QrCode(value))}
}

// Frame sets the size and the position of the QR code.
//
//gostrict:host
func (c TQrCode) Frame(frame Frame) TQrCode {
	return TQrCode{ref: wrap(get[nui.TQrCode](c.ref).Frame(frame.nago()))}
}

// AccessibilityLabel sets the label for a screen reader.
//
//gostrict:host
func (c TQrCode) AccessibilityLabel(label string) TQrCode {
	return TQrCode{ref: wrap(get[nui.TQrCode](c.ref).AccessibilityLabel(label))}
}

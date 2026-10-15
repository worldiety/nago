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

// TModal is a modal, which shows its content above all other views, see Modal.
type TModal struct{ ref js.Ref }

func (TModal) isView() {}

// Modal places the content and blocks the controls of the background. See also Overlay.
//
//gostrict:host
func Modal(content View) TModal {
	return TModal{ref: wrap(nui.Modal(viewOf(content)))}
}

// Overlay places the content above the other views with absolute positions, and it does not block the background.
// See also Modal.
//
//gostrict:host
func Overlay(content View) TModal {
	return TModal{ref: wrap(nui.Overlay(viewOf(content)))}
}

// NotificationsOverlay is like Overlay, but it is placed in front of other overlays.
//
//gostrict:host
func NotificationsOverlay(content View) TModal {
	return TModal{ref: wrap(nui.NotificationsOverlay(viewOf(content)))}
}

// Top sets the top offset of the modal.
//
//gostrict:host
func (c TModal) Top(top Length) TModal {
	return TModal{ref: wrap(get[nui.TModal](c.ref).Top(nui.Length(top)))}
}

// Left sets the left offset of the modal.
//
//gostrict:host
func (c TModal) Left(left Length) TModal {
	return TModal{ref: wrap(get[nui.TModal](c.ref).Left(nui.Length(left)))}
}

// Right sets the right offset of the modal.
//
//gostrict:host
func (c TModal) Right(right Length) TModal {
	return TModal{ref: wrap(get[nui.TModal](c.ref).Right(nui.Length(right)))}
}

// Bottom sets the bottom offset of the modal.
//
//gostrict:host
func (c TModal) Bottom(bottom Length) TModal {
	return TModal{ref: wrap(get[nui.TModal](c.ref).Bottom(nui.Length(bottom)))}
}

// AllowBackgroundScrolling configures, whether the background can scroll while the modal is open.
//
//gostrict:host
func (c TModal) AllowBackgroundScrolling(allowBackgroundScrolling bool) TModal {
	return TModal{ref: wrap(get[nui.TModal](c.ref).AllowBackgroundScrolling(allowBackgroundScrolling))}
}

// OnDismissRequest calls fn, when the user requests to dismiss the modal.
//
//gostrict:host
func (c TModal) OnDismissRequest(fn func()) TModal {
	return TModal{ref: wrap(get[nui.TModal](c.ref).OnDismissRequest(fn))}
}

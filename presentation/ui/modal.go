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

// TModal is a composite component (Modal).
// It represents a modal or overlay container that displays content above all other views.
// A modal blocks interaction with the background unless configured as an overlay.
// It supports dismiss callbacks, positioning, and background scrolling behavior.
type TModal struct {
	content                  core.View       // content displayed inside the modal
	onDismissRequest         func()          // callback invoked when the modal is dismissed
	mtype                    proto.ModalType // modal type (default modal or overlay)
	top, left, right, bottom Length          // optional positioning values
	allowBackgroundScrolling bool            // whether background content can scroll while modal is open
}

// Modal places the content and blocks the background controls. See also [Overlay] for a different behavior.
func Modal(content core.View) TModal {
	return TModal{content: content}
}

// Overlay supports absolut positions, which are indeed required to ensure that there is no transparent blocking
// view. The html frontend renderer has a problem to distinguish between capturing events for conventional modals
// and not capturing them as in [Modal] mode.
func Overlay(content core.View) TModal {
	return TModal{content: content, mtype: proto.ModalTypeOverlay, allowBackgroundScrolling: true}
}

// Top sets the top offset of the modal.
func (c TModal) Top(top Length) TModal {
	c.top = top
	return c
}

// AllowBackgroundScrolling configures whether background content can scroll
// while the modal is open.
func (c TModal) AllowBackgroundScrolling(allowBackgroundScrolling bool) TModal {
	c.allowBackgroundScrolling = allowBackgroundScrolling
	return c
}

// Left sets the left offset of the modal.
func (c TModal) Left(left Length) TModal {
	c.left = left
	return c
}

// Right sets the right offset of the modal.
func (c TModal) Right(right Length) TModal {
	c.right = right
	return c
}

// Bottom sets the bottom offset of the modal.
func (c TModal) Bottom(bottom Length) TModal {
	c.bottom = bottom
	return c
}

// Render builds and returns the protocol representation of the modal.
func (c TModal) Render(context core.RenderContext) core.RenderNode {
	return &proto.Modal{
		Content:                  render(context, c.content),
		ModalType:                c.mtype,
		OnDismissRequest:         context.MountCallback(c.onDismissRequest),
		Top:                      c.top.ora(),
		Left:                     c.left.ora(),
		Right:                    c.right.ora(),
		Bottom:                   c.bottom.ora(),
		AllowBackgroundScrolling: proto.Bool(c.allowBackgroundScrolling),
	}
}

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
	"go.wdy.de/nago/presentation/ui/avatar"
)

// AvatarStyle is the shape of an avatar, see AvatarCircle and AvatarRounded. Like avatar.Style of nago.
type AvatarStyle int32

const (
	// AvatarCircle shows the avatar as a circle, which is the default.
	AvatarCircle AvatarStyle = 0
	// AvatarRounded shows the avatar with rounded corners.
	AvatarRounded AvatarStyle = 1
)

// TAvatar is a user or entity representation, which shows an image or the initials of a text, see AvatarText.
type TAvatar struct{ ref js.Ref }

func (TAvatar) isView() {}

// AvatarText returns an avatar with the initials of the text, e.g. "JD" for "John Doe". Like avatar.Text of nago.
//
//gostrict:host
func AvatarText(paraphe string) TAvatar {
	return TAvatar{ref: wrap(avatar.Text(paraphe))}
}

// AvatarURI returns an avatar, which shows the image of the URI. Like avatar.URI of nago.
//
//gostrict:host
func AvatarURI(uri URI) TAvatar {
	return TAvatar{ref: wrap(avatar.URI(core.URI(uri)))}
}

// Border sets the border of the avatar.
//
//gostrict:host
func (c TAvatar) Border(border Border) TAvatar {
	return TAvatar{ref: wrap(get[avatar.TAvatar](c.ref).Border(border.nago()))}
}

// Action calls fn, when the user clicks the avatar.
//
//gostrict:host
func (c TAvatar) Action(fn func()) TAvatar {
	return TAvatar{ref: wrap(get[avatar.TAvatar](c.ref).Action(fn))}
}

// Size sets the width and the height of the avatar, the text and the image adapt to it.
//
//gostrict:host
func (c TAvatar) Size(widthAndHeight Length) TAvatar {
	return TAvatar{ref: wrap(get[avatar.TAvatar](c.ref).Size(nui.Length(widthAndHeight)))}
}

// Style sets the shape of the avatar.
//
//gostrict:host
func (c TAvatar) Style(style AvatarStyle) TAvatar {
	return TAvatar{ref: wrap(get[avatar.TAvatar](c.ref).Style(avatar.Style(style)))}
}

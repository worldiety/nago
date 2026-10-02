// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"errors"
	"io"

	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/form"
)

// The gallery runs without image management. A real app passes the use cases of cfg.ImageManagement()
// or nil, which resolves them from the window context.
var (
	createSrcSet image.CreateSrcSet = func(permission.Auditable, image.Options, image.File) (image.SrcSet, error) {
		return image.SrcSet{}, errors.ErrUnsupported
	}
	loadSrcSet image.LoadSrcSet = func(permission.Auditable, image.ID) (std.Option[image.SrcSet], error) {
		return std.None[image.SrcSet](), nil
	}
	loadBestFit image.LoadBestFit = func(permission.Auditable, image.ID, image.ObjectFit, int, int) (std.Option[io.ReadCloser], error) {
		return std.None[io.ReadCloser](), nil
	}
)

func init() {
	app.Register("avatar-picker", func(wnd core.Window) core.View {
		img := core.AutoState[image.ID](wnd)

		return HStack(
			form.AvatarPicker(wnd, createSrcSet, "profile-image", img.Get(), img, "Ada Lovelace", avatar.Circle),
			form.AvatarPicker(wnd, createSrcSet, "logo", img.Get(), img, "Acme Corp", avatar.Rounded),
		).Gap(L32)
	})
}

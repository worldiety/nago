// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/form"
)

// createSrcSet, loadSrcSet and loadBestFit are the image use cases, see avatar_picker.go.
func init() {
	app.Register("single-image-picker", func(wnd core.Window) core.View {
		img := core.AutoState[image.ID](wnd)

		return VStack(
			Text("Cover image"),
			form.SingleImagePicker(wnd, createSrcSet, loadSrcSet, loadBestFit, "cover-image", img.Get(), img),
		).Alignment(Leading).Gap(L8).Frame(Frame{Width: L400})
	})
}

// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
)

func init() {
	app.Register("font", func(wnd core.Window) core.View {
		return VStack(
			Text("DisplaySmall").Font(DisplaySmall),
			Text("HeadlineMedium").Font(HeadlineMedium),
			Text("TitleLarge").Font(TitleLarge),
			Text("BodyMedium").Font(BodyMedium),
			Text("LabelSmall").Font(LabelSmall),
			Text("MonoMedium").Font(MonoMedium),
			Text("Font{Size: L20, Style: ItalicFontStyle}").Font(Font{Size: L20, Style: ItalicFontStyle}),
		).Alignment(Leading).Gap(L8)
	})
}

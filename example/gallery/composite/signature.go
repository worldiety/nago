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
	"go.wdy.de/nago/presentation/ui/esignature"
)

func init() {
	app.Register("signature", func(wnd core.Window) core.View {
		return esignature.Signature().
			TopText("Signed on 12 March 2026").
			Body(Text("Ada Lovelace").Font(Font{Size: L32, Style: ItalicFontStyle})).
			BottomText("Ada Lovelace, Head of Engineering")
	})
}

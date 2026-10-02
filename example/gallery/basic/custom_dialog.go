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
	icons "go.wdy.de/nago/presentation/icons/hero/solid"
	. "go.wdy.de/nago/presentation/ui"
)

func init() {
	app.Register("custom-dialog", func(wnd core.Window) core.View {
		presented := core.AutoState[bool](wnd)
		closeDlg := func() { presented.Set(false) }

		return VStack(
			If(presented.Get(), Modal(
				Dialog(Text("Your profile is complete. You can now invite your team.")).
					Title(Text("Welcome")).
					TitleX(TertiaryButton(closeDlg).PreIcon(icons.XMark)).
					Footer(HStack(
						SecondaryButton(closeDlg).Title("Later"),
						PrimaryButton(closeDlg).Title("Invite team"),
					).Gap(L8)),
			).OnDismissRequest(closeDlg)),
			PrimaryButton(func() { presented.Set(true) }).Title("Open dialog"),
		)
	})
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uidrive

import (
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/drive"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"golang.org/x/text/language"
)

var (
	StrNoRoot = i18n.MustString("nago.drive.no_root", i18n.Values{language.English: "No Drive Root to show", language.German: "Es ist kein Drive verfügbar."})
)

func PageDrive(wnd core.Window, uc drive.UseCases) core.View {
	drives, err := uc.ReadDrives(wnd.Subject(), wnd.Subject().ID())
	if err != nil {
		return alert.BannerError(err)
	}

	root := drive.FID(wnd.Values()["fid"])
	if root == "" {
		if id, ok := drives.Global[drive.FSDrive]; ok {
			root = id
		} else if id, ok := drives.Private[drive.FSDrive]; ok {
			root = id
		}
	}

	if root == "" {
		return ui.Text(StrNoRoot.Get(wnd))
	}

	rootState := core.AutoState[drive.FID](wnd).Init(func() drive.FID {
		return root
	})

	return Drive(rootState).
		Frame(ui.Frame{}.FullWidth())
}

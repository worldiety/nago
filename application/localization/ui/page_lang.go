// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uilocalization

import (
	"fmt"

	"go.wdy.de/nago/application/localization"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/cardlayout"
	"golang.org/x/text/language"
)

func PageLanguage(wnd core.Window, uc localization.UseCases) core.View {
	res, err := uc.FindResources(wnd.Subject())
	if err != nil {
		return alert.BannerError(err)
	}

	addLangPresented := core.AutoState[bool](wnd)

	order := 0
	return ui.VStack(
		addLangDialog(wnd, uc, addLangPresented),
		cardlayout.Card(StrPriorities.Get(wnd)).
			Footer(ui.PrimaryButton(func() {
				addLangPresented.Set(true)
			}).Title(rstring.ActionAdd.Get(wnd))).
			Body(
				ui.VStack(
					ui.ForEach(res.Tags(), func(t language.Tag) core.View {
						defer func() {
							order++
						}()
						if order == 0 {
							return ui.HStack(
								ui.Text(languageName(t)),
								ui.Text(" ("),
								ui.Text(StrFallback.Get(wnd)).Font(ui.MonoSmall),
								ui.Text(")"),
							)
						}

						return ui.HStack(
							ui.Text(fmt.Sprintf("%d. ", order)),
							ui.Text(languageName(t)),
						)
					})...,
				).Alignment(ui.Leading),
			).
			Frame(ui.Frame{MaxWidth: ui.L560}.FullWidth()),
	).FullWidth()
}

func addLangDialog(wnd core.Window, uc localization.UseCases, presented *core.State[bool]) core.View {
	if !presented.Get() {
		return nil
	}

	tag := core.AutoState[string](wnd)

	return alert.Dialog(
		StrAddLanguage.Get(wnd),
		ui.TextField(rstring.LabelLanguage.Get(wnd), tag.String()).
			InputValue(tag).
			SupportingText(StrAddLanguageSupportingText.Get(wnd)),
		presented,
		alert.Cancel(nil),
		alert.Save(func() (close bool) {
			tag, err := language.Parse(tag.Get())
			if err != nil {
				alert.ShowBannerError(wnd, err)
				return
			}

			if err := uc.AddLanguage(wnd.Subject(), tag); err != nil {
				alert.ShowBannerError(wnd, err)
				return
			}

			return true
		}),
	)
}

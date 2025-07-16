// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uisignature

import (
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/application/signature"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

func PageMySignature(wnd core.Window, ucSig signature.UseCases, createSrcSet image.CreateSrcSet) core.View {
	settings, err := ucSig.LoadUserSettings(wnd.Subject(), wnd.Subject().ID())
	if err != nil {
		return alert.BannerError(err)
	}

	sigImg := core.AutoState[image.ID](wnd).Init(func() image.ID {
		return settings.ImageSignature
	}).Observe(func(newValue image.ID) {
		if err := ucSig.UpdateUserSettings(wnd.Subject(), wnd.Subject().ID(), signature.UpdateUserSettingsData{
			ImageSignature: newValue,
		}); err != nil {
			alert.ShowBannerError(wnd, err)
			return
		}
	})

	return ui.VStack(
		ui.H1("Meine Unterschrift"),

		ui.Text("Die hier hinterlegte Unterschrift erhöht nicht den rechtlichen Stellenwert der elektronischen Signatur. Bei den elektronischen Nago-Signaturen handelt es sich um unqualifizierte elektronische Signaturen. Die Unterschriften werden jedoch in einem kryptografisch auditierbaren Verlauf notiert und können genau dem aktuell angemeldeten Konto zugeordnet werden.\n\nEine einmal erteilte Unterschrift kann nicht mehr rückgängig gemacht werden, da die unterzeichneten Dokumente bereits zum nächsten Empfänger gelangt sind. Kontaktiere in diesem Fall den Dokumentenersteller, um die im jeweiligen Prozess möglichen Schritte zu prüfen."),
		ui.VStack(
			ui.VStack(
				ui.Image().
					ObjectFit(ui.FitContain).
					Adaptive(sigImg.Get(), "").
					Border(ui.Border{}.Radius(ui.L16)).
					Frame(ui.Frame{Width: ui.Full, Height: ui.Full}),

				ui.VStack(
					ui.PrimaryButton(func() {
						wnd.ImportFiles(core.ImportFilesOptions{
							AllowedMimeTypes: []string{"image/png", "image/jpeg"},
							OnCompletion: func(files []core.File) {
								for _, file := range files {
									srcSet, err := createSrcSet(wnd.Subject(), image.Options{}, file)
									if err != nil {
										alert.ShowBannerError(wnd, err)
										return
									}

									sigImg.Set(srcSet.ID)
									sigImg.Notify()
								}
							},
						})
					}).PreIcon(flowbiteOutline.Pen).AccessibilityLabel("Unterschrift aktualisieren"),
				).Position(ui.Position{Type: ui.PositionAbsolute, Right: "0px", Top: "0px"}).
					Padding(ui.Padding{}.All(ui.L4)),
			).
				Position(ui.Position{Type: ui.PositionOffset}).
				BackgroundColor(ui.ColorCardFooter).
				Frame(ui.Frame{Height: ui.L96, Width: ui.L320}).
				Padding(ui.Padding{}.All(ui.L8)).
				Border(ui.Border{}.Radius(ui.L16)),
		).Gap(ui.L4),
	).
		BackgroundColor(ui.ColorCardBody).
		Gap(ui.L8).
		Alignment(ui.Leading).
		Border(ui.Border{}.Radius(ui.L16)).
		Padding(ui.Padding{}.All(ui.L16)).
		Frame(ui.Frame{Width: ui.Full, MaxWidth: ui.L560})
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uisignature

import (
	"fmt"
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/application/signature"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/esignature"
	"go.wdy.de/nago/presentation/ui/tabs"
)

// UserSignature creates a field for an electronic signature. It can only be used if the given user id equals
// the current user subject. If you want to specify the role, use [TSignature.TopText].
func UserSignature(wnd core.Window, uid user.ID, resource user.Resource) esignature.TSignature {
	ucSig, ok := core.FromContext[signature.UseCases](wnd.Context(), "")
	if !ok {
		return esignature.Signature().Body(alert.BannerError(fmt.Errorf("signature use cases not initialized")))
	}

	var usrSig signature.Signature
	for sig, err := range ucSig.FindSignaturesByResource(wnd.Subject(), resource) {
		if err != nil {
			return esignature.Signature().Body(alert.BannerError(err))
		}

		if sig.User == uid {
			usrSig = sig
			break
		}
	}

	if !usrSig.IsZero() {
		// already has signature, just show it
		name := xstrings.Join2(" ", usrSig.Firstname, usrSig.Lastname)
		var signing core.View
		if usrSig.Image == "" {
			signing = sigText(name)
		} else {
			signing = sigImg(usrSig.Image)
		}

		return esignature.Signature().
			TopText("Nago Signatur").
			BottomText(name).
			Body(signing)
	}

	displayName, _ := core.FromContext[user.DisplayName](wnd.Context(), "")
	name := displayName(uid).Displayname

	signDlgPresented := core.AutoState[bool](wnd)

	var signIt core.View
	if uid == wnd.Subject().ID() {
		signIt = ui.VStack(
			pickSignatureDlg(wnd, ucSig, signDlgPresented, resource),
			ui.PrimaryButton(func() {
				signDlgPresented.Set(true)
			}).Title("Unterzeichnen"),
		)
	}

	return esignature.Signature().
		TopText("Nago Signatur").
		BottomText(name).
		Body(signIt)
}

func sigImg(img image.ID) core.View {
	return ui.Image().
		Adaptive(img, "").
		ObjectFit(ui.FitContain).
		Border(ui.Border{}.Radius(ui.L16)).
		Frame(ui.Frame{Height: ui.L48, Width: ""})
}

func sigText(text string) core.View {
	return ui.VStack(ui.Text(text).Font(ui.Font{Name: "Caveat", Size: ui.L32})).Padding(ui.Padding{}.All(ui.L4))
}

func pickSignatureDlg(wnd core.Window, ucSig signature.UseCases, presented *core.State[bool], res user.Resource) core.View {
	if !presented.Get() {
		return nil
	}

	name := xstrings.Join2(" ", wnd.Subject().Firstname(), wnd.Subject().Lastname())
	settings, err := ucSig.LoadUserSettings(wnd.Subject(), wnd.Subject().ID())
	if err != nil {
		return alert.BannerError(err)
	}

	activePage := core.AutoState[int](wnd).Init(func() int {
		if settings.ImageSignature != "" {
			return 1
		}

		return 0
	})
	firstname := core.AutoState[string](wnd).Init(func() string {
		return wnd.Subject().Firstname()
	})

	lastname := core.AutoState[string](wnd).Init(func() string {
		return wnd.Subject().Lastname()
	})

	return alert.Dialog(
		"Signatur auswählen",
		tabs.Tabs(
			tabs.Page("als Text", func() core.View {
				return ui.VStack(
					ui.HStack(
						ui.TextField("Vorname", firstname.Get()).InputValue(firstname).FullWidth(),
						ui.TextField("Nachname", lastname.Get()).InputValue(lastname).FullWidth(),
					).FullWidth().Gap(ui.L8),

					esignature.Signature().
						TopText("Nago Signatur").
						BottomText(xstrings.Join2(" ", firstname.Get(), lastname.Get())).
						Body(sigText(xstrings.Join2(" ", firstname.Get(), lastname.Get()))),
				).Gap(ui.L8).FullWidth()
			}),
			tabs.Page("als Unterschrift", func() core.View {
				return ui.VStack(
					esignature.Signature().
						TopText("Nago Signatur").
						BottomText(name).
						Body(sigImg(settings.ImageSignature)),
				)
			}).Disabled(settings.ImageSignature == ""),
			tabs.Page("per Hand zeichnen", func() core.View {
				return ui.Text("bild")
			}).Disabled(true),
		).InputValue(activePage).FullWidth(),
		presented,
		alert.Large(),
		alert.Cancel(nil),
		alert.Custom(func(close func(closeDlg bool)) core.View {
			return ui.PrimaryButton(func() {
				switch activePage.Get() {
				case 0:
					_, err := ucSig.SignUnqualified(signature.AnonSignData{
						Firstname: firstname.Get(),
						Lastname:  lastname.Get(),
						Email:     user.Email(wnd.Subject().Email()),
						SignData: signature.SignData{
							Location:  string(wnd.Path()),
							Resources: []user.Resource{res},
						},
					})

					if err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}
				case 1:
					_, err := ucSig.SignUnqualifiedWithSubject(wnd.Subject(), signature.SignData{
						Location:       string(wnd.Path()),
						Resources:      []user.Resource{res},
						SignatureImage: settings.ImageSignature,
					})

					if err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}
				default:
					alert.ShowBannerError(wnd, fmt.Errorf("not implemented"))
				}
				close(true)

			}).Title("Annehmen und Unterzeichnen")
		}),
	)
}

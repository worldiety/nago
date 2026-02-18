// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uitoken

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/token"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/form"
	"go.wdy.de/nago/presentation/ui/list"
)

func PageCrud(wnd core.Window, uc token.UseCases) core.View {
	if !wnd.Subject().Valid() {
		return alert.BannerError(user.InvalidSubjectErr)
	}

	selectedToken := core.AutoState[token.Token](wnd)
	accessRightsPresented := core.AutoState[bool](wnd)
	rotatePresented := core.AutoState[bool](wnd)
	deletePresented := core.AutoState[bool](wnd)

	var rows []ui.TTableRow
	for tok, err := range uc.FindAll(wnd.Subject()) {
		if err != nil {
			return alert.BannerError(err)
		}

		rows = append(rows, ui.TableRow(
			ui.TableCell(ui.Text(tok.Name)),
			ui.TableCell(ui.Text(tok.Description)),
			ui.TableCell(ui.Text(tok.CreatedAt.Format(xtime.GermanDate))),
			ui.TableCell(ui.Text(formatValidUntil(tok.ValidUntil))),
			ui.TableCell(ui.HStack(
				ui.SecondaryButton(func() {
					selectedToken.Set(tok)
					deletePresented.Set(true)
				}).PreIcon(flowbiteOutline.TrashBin).AccessibilityLabel(tok.Name+" löschen"),

				ui.SecondaryButton(func() {
					selectedToken.Set(tok)
					rotatePresented.Set(true)
				}).PreIcon(flowbiteOutline.Refresh).AccessibilityLabel(tok.Name+" rotieren"),

				ui.SecondaryButton(func() {
					selectedToken.Set(tok)
					accessRightsPresented.Set(true)
				}).PreIcon(flowbiteOutline.ShieldCheck).AccessibilityLabel(tok.Name+" Rechte einsehen"),
			).Gap(ui.L8)).Alignment(ui.Trailing),
		))
	}

	createPresented := core.AutoState[bool](wnd)

	plainTokenPresented := core.AutoState[bool](wnd)
	plainToken := core.AutoState[string](wnd).Observe(func(newValue string) {
		plainTokenPresented.Set(newValue != "")
	})

	return ui.VStack(
		ui.H1("Access Tokens"),
		ui.TextLayout(
			ui.Text("API Access Tokens werden konfiguriert, um anderen Anwendungen Zugriff auf diese Anwendung über eine http-basierte Programmierschnittstelle zu gewähren.\n\n"),
		),
		createDialog(wnd, createPresented, plainToken, uc),
		deleteDialog(wnd, deletePresented, selectedToken.Get(), uc),
		viewAccessRightsDialog(wnd, accessRightsPresented, selectedToken.Get(), uc),
		rotateDialog(wnd, rotatePresented, selectedToken.Get(), plainToken, uc),
		PlainTokenDialog(wnd, plainTokenPresented, plainToken),
		ui.HStack(
			ui.SecondaryButton(func() {
				core.HTTPOpen(wnd.Navigation(), "/api/doc", "_blank")
			}).PreIcon(flowbiteOutline.BookOpen).AccessibilityLabel("Open API Dokumentation anzeigen"),
			ui.PrimaryButton(func() {
				createPresented.Set(true)
			}).Title("Access Token hinzufügen"),
		).FullWidth().Alignment(ui.Trailing).Gap(ui.L8),
		ui.Table(
			ui.TableColumn(ui.Text("Name")),
			ui.TableColumn(ui.Text("Beschreibung")),
			ui.TableColumn(ui.Text("Erstellt am")),
			ui.TableColumn(ui.Text("Gültig bis")),
			ui.TableColumn(ui.Text("Optionen")),
		).Rows(rows...),
		ui.If(len(rows) == 0, ui.Text("Noch keine Tokens vorhanden")),
	).Alignment(ui.Leading).FullWidth().Gap(ui.L16)
}

func viewAccessRightsDialog(wnd core.Window, presented *core.State[bool], token token.Token, uc token.UseCases) core.View {
	if !presented.Get() {
		return nil
	}

	resolved, err := uc.ResolveTokenRights(wnd.Subject(), token.ID)
	if err != nil {
		return alert.BannerError(err)
	}

	return alert.Dialog(
		"Zugriffsrechte für "+token.Name,
		ui.VStack(
			listRoles(resolved),
			listGroups(resolved),
			listPermissions(resolved),
		).Gap(ui.L16).FullWidth(),
		presented,
		alert.Closeable(),
		alert.Ok(),
	)
}

func listRoles(t token.ResolvedTokenRights) core.View {
	if len(t.Roles) == 0 {
		return ui.Text("Keine Rollen zugewiesen")
	}

	return list.List(ui.ForEach(t.Roles, func(t role.Role) core.View {
		return list.Entry().
			Headline(t.Name).
			SupportingText(t.Description).
			Leading(ui.ImageIcon(flowbiteOutline.UserSettings))
	})...).Caption(ui.Text("Rollen")).Frame(ui.Frame{}.FullWidth())
}

func listGroups(t token.ResolvedTokenRights) core.View {
	if len(t.Groups) == 0 {
		return ui.Text("Keine Gruppen zugewiesen")
	}

	return list.List(ui.ForEach(t.Groups, func(t group.Group) core.View {
		return list.Entry().
			Headline(t.Name).
			SupportingText(t.Description).
			Leading(ui.ImageIcon(flowbiteOutline.UsersGroup))
	})...).Caption(ui.Text("Gruppen")).Frame(ui.Frame{}.FullWidth())
}

func listPermissions(t token.ResolvedTokenRights) core.View {
	if len(t.Roles) == 0 {
		return ui.Text("Keine Berechtigungen zugewiesen")
	}

	return list.List(ui.ForEach(t.Permissions, func(t permission.Permission) core.View {
		return list.Entry().
			Headline(t.Name).
			SupportingText(t.Description).
			Leading(ui.ImageIcon(flowbiteOutline.Shield))
	})...).Caption(ui.Text("Berechtigungen")).Frame(ui.Frame{}.FullWidth()).Footer(ui.Text("Diese Berechtigungen sind die Gesamtmenge von Einzelzuweisungen und vererbt über Rollenzugehörigkeiten."))
}

func deleteDialog(wnd core.Window, presented *core.State[bool], token token.Token, uc token.UseCases) core.View {
	if !presented.Get() {
		return nil
	}

	return alert.Dialog(
		token.Name+" rotieren",
		ui.Text("Soll der Access Token '"+token.Name+"' gelöscht werden?"),
		presented,
		alert.Cancel(nil),
		alert.Delete(func() {
			if err := uc.Delete(wnd.Subject(), token.ID); err != nil {
				alert.ShowBannerError(wnd, err)
			}
		}),
	)
}

func rotateDialog(wnd core.Window, presented *core.State[bool], token token.Token, plainToken *core.State[string], uc token.UseCases) core.View {
	if !presented.Get() {
		return nil
	}

	return alert.Dialog(
		token.Name+" rotieren",
		ui.Text("Soll der Token '"+token.Name+"' rotiert werden? Die Berechtigungen ändern sich dadurch nicht, aber sämtliche Zugriffe über den alten Token sind dann nicht mehr möglich."),
		presented,
		alert.Closeable(),
		alert.Cancel(nil),
		alert.Save(func() (close bool) {
			plain, err := uc.Rotate(wnd.Subject(), token.ID)
			if err != nil {
				alert.ShowBannerError(wnd, err)
				return false
			}

			plainToken.Set(string(plain))
			plainToken.Notify()

			return true
		}),
	)
}

func PlainTokenDialog(wnd core.Window, presented *core.State[bool], plainToken *core.State[string]) core.View {
	if !presented.Get() {
		return nil
	}

	return alert.Dialog(
		"Access Token",
		ui.VStack(
			ui.Text("Kopieren Sie den folgenden Access Token und verwahren Sie ihn sicher auf. Ein Passwortmanager oder ein Secret Vault sind dafür geeignet. Der Token wird nicht gespeichert und kann nicht wieder eingesehen werden."),
			ui.HStack(
				ui.Text(plainToken.Get()),
				ui.TertiaryButton(func() {
					if err := wnd.Clipboard().SetText(plainToken.Get()); err != nil {
						alert.ShowBannerError(wnd, err)
					}
				}).PreIcon(flowbiteOutline.Clipboard).AccessibilityLabel("Access Token in die Zwischenablage kopieren"),
			),
		),
		presented,
		alert.Closeable(),
		alert.Ok(),
		alert.Large(),
	)
}

func createDialog(wnd core.Window, presented *core.State[bool], plainToken *core.State[string], uc token.UseCases) core.View {
	if !presented.Get() {
		return nil
	}

	tokenState := core.AutoState[token.CreationData](wnd)

	return alert.Dialog(
		"Neuen API Access Token erstellen",
		form.Auto(form.AutoOptions{Window: wnd}, tokenState),
		presented,
		alert.Width(ui.L560),
		alert.Cancel(nil),
		alert.Save(func() (close bool) {
			_, plain, err := uc.Create(wnd.Subject(), tokenState.Get())
			if err != nil {
				alert.ShowBannerError(wnd, err)
				return false
			}

			plainToken.Set(string(plain))
			plainToken.Notify()

			return true
		}),
	)
}

func formatValidUntil(t xtime.Date) string {
	if t.IsZero() {
		return "unbefristet"
	}

	return t.Format(xtime.GermanDate)
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uisecret

import (
	"reflect"
	"slices"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/cardlayout"
	"go.wdy.de/nago/presentation/ui/form"
	"go.wdy.de/nago/presentation/ui/picker"
)

// EditSecretPage edits an existing secret. The stored values of write-only fields like passwords are never sent
// to the browser again, see [secret.WithoutWriteOnly].
func EditSecretPage(wnd core.Window, pages Pages, uc UseCases) core.View {
	id := secret.ID(wnd.Values()["id"])
	optScr, err := uc.Secrets.FindMySecretByID(wnd.Subject(), id)
	if err != nil {
		return alert.BannerError(err)
	}

	if optScr.IsNone() {
		return alert.Banner("Nicht gefunden", "Das Secret ist nicht mehr verfügbar.")
	}

	return secretEditor(wnd, pages, uc, optScr.Unwrap(), false)
}

// secretEditor edits the credentials and the sharing of a secret. A new secret is a draft, which is stored only by
// saving it.
func secretEditor(wnd core.Window, pages Pages, uc UseCases, scr secret.Secret, create bool) core.View {
	initialCredentials := secret.WithoutWriteOnly(scr.Credentials)
	credentials := core.AutoState[secret.Credentials](wnd).Init(func() secret.Credentials {
		return initialCredentials
	})

	// groups of the subject, foreign groups of the secret are kept by the use case and only listed
	var availGroups []group.Group
	var initialGroups []group.Group
	for grp, err := range uc.FindMyGroups(wnd.Subject()) {
		if err != nil {
			return alert.BannerError(err)
		}

		availGroups = append(availGroups, grp)
		if slices.Contains(scr.Groups, grp.ID) {
			initialGroups = append(initialGroups, grp)
		}
	}

	var foreignGroups []group.ID
	for _, gid := range scr.Groups {
		if !slices.ContainsFunc(availGroups, func(g group.Group) bool { return g.ID == gid }) {
			foreignGroups = append(foreignGroups, gid)
		}
	}

	selectedGroups := core.AutoState[[]group.Group](wnd).Init(func() []group.Group {
		return initialGroups
	})

	var availUsers []user.User
	var initialUsers []user.User
	for usr, err := range uc.FindAllUsers(user.SU()) {
		if err != nil {
			return alert.BannerError(err)
		}

		availUsers = append(availUsers, usr)
		if slices.Contains(scr.Owners, usr.ID) {
			initialUsers = append(initialUsers, usr)
		}
	}

	selectedUsers := core.AutoState[[]user.User](wnd).Init(func() []user.User {
		return initialUsers
	})

	dirty := create ||
		!reflect.DeepEqual(credentials.Get(), initialCredentials) ||
		!sameIDs(groupIDs(selectedGroups.Get()), groupIDs(initialGroups)) ||
		!sameIDs(userIDs(selectedUsers.Get()), userIDs(initialUsers))

	sharedWith := append(groupNames(wnd.Subject(), uc.FindGroupByID, groupIDs(selectedGroups.Get())), groupNames(wnd.Subject(), uc.FindGroupByID, foreignGroups)...)
	access := accessSentence(wnd, selectedUsers.Get(), sharedWith)

	title := "Secret bearbeiten"
	if create {
		title = "Neues Secret"
	}

	deletePresented := core.AutoState[bool](wnd)

	return ui.VStack(
		ui.If(!create, alert.Dialog(
			"Secret löschen?",
			ui.Text("Das Secret wird für alle entfernt, nicht nur für Sie. "+access),
			deletePresented,
			alert.Cancel(nil),
			alert.Delete(func() {
				if err := uc.Secrets.DeleteMySecretByID(wnd.Subject(), scr.ID); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				wnd.Navigation().ResetTo(pages.Vault, nil)
			}),
		)),
		ui.H1(title),
		ui.If(!create, ui.HStack(ui.TertiaryButton(func() {
			deletePresented.Set(true)
		}).PreIcon(heroSolid.Trash).Title("Secret löschen")).
			FullWidth().
			Alignment(ui.Trailing)),

		typeCard(wnd, scr),

		cardlayout.Card("Zugangsdaten").
			Style(cardlayout.TitleCompact).
			Body(form.Auto[secret.Credentials](credentialsFormOptions(scr, create), credentials)).
			Frame(ui.Frame{}.FullWidth()),

		accessCard(availGroups, selectedGroups, foreignGroups, availUsers, selectedUsers, access),

		ui.HLineWithColor(ui.ColorAccent),
		ui.HStack(
			ui.SecondaryButton(func() {
				wnd.Navigation().ResetTo(pages.Vault, nil)
			}).Title(rstring.ActionCancel.Get(wnd)),
			ui.PrimaryButton(func() {
				if create {
					_, err := uc.Secrets.CreateSecret(wnd.Subject(), credentials.Get(), secret.CreateOptions{
						Groups: groupIDs(selectedGroups.Get()),
						Owners: userIDs(selectedUsers.Get()),
					})
					if err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}

					wnd.Navigation().ResetTo(pages.Vault, nil)
					return
				}

				if err := uc.Secrets.UpdateMyCredentials(wnd.Subject(), scr.ID, credentials.Get()); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				if err := uc.Secrets.UpdateMySecretGroups(wnd.Subject(), scr.ID, groupIDs(selectedGroups.Get())); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				if err := uc.Secrets.UpdateMySecretOwners(wnd.Subject(), scr.ID, userIDs(selectedUsers.Get())); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				wnd.Navigation().ResetTo(pages.Vault, nil)
			}).Title(rstring.ActionSave.Get(wnd)).Enabled(dirty),
		).
			FullWidth().
			Gap(ui.L8).
			Alignment(ui.Trailing),
	).Gap(ui.L16).Alignment(ui.Leading).Frame(ui.Frame{Width: ui.Full, MaxWidth: ui.L560})
}

func typeCard(wnd core.Window, scr secret.Secret) core.View {
	spec := newCredentialTypeSpec(reflect.TypeOf(scr.Credentials))
	logo := spec.LogoView()
	avtLogo, ok := logo.(avatar.TAvatar)
	if !ok {
		return logo
	}

	return cardlayout.Card(wnd.Bundle().Resolve(spec.name)).
		Style(cardlayout.TitleCompact).
		Body(
			ui.VStack(
				avtLogo.Size(ui.L80),
				ui.Text(wnd.Bundle().Resolve(spec.description)).TextAlignment(ui.TextAlignCenter),
			).Gap(ui.L16).
				FullWidth(),
		).
		Frame(ui.Frame{}.FullWidth())
}

// credentialsFormOptions renders the write-only fields of an existing secret with a hint, that an empty field keeps
// the stored value. The other fields are rendered by the default renderers of [form.Auto].
func credentialsFormOptions(scr secret.Secret, create bool) form.AutoOptions {
	if create {
		return form.AutoOptions{}
	}

	stored := secret.WriteOnlyFields(scr.Credentials)
	if len(stored) == 0 {
		return form.AutoOptions{}
	}

	return form.AutoOptions{
		Renderers: func(yield func(form.Renderer) bool) {
			if !yield(writeOnlyRenderer(stored)) {
				return
			}

			for r := range form.Renderers {
				if !yield(r) {
					return
				}
			}
		},
	}
}

// writeOnlyRenderer renders a write-only field like [form.RenderText] renders a secret field, but explains below
// the field, whether a value is stored and that an empty field keeps it.
func writeOnlyRenderer(stored map[string]bool) form.Renderer {
	return func(ctx form.FieldContext) core.View {
		isStored, ok := stored[ctx.Field().Name]
		if !ok {
			return nil
		}

		strState := core.DerivedState[string](ctx.State(), ctx.Path()).Init(func() string {
			return ctx.FieldValue().String()
		})

		strState.Observe(func(newValue string) {
			ctx.SetValue(newValue)
		})

		hint := "Es ist noch kein Wert gespeichert."
		if isStored {
			hint = "Ein Wert ist gespeichert und wird aus Sicherheitsgründen nicht angezeigt. Bleibt das Feld leer, wird er nicht geändert."
		}

		if text := ctx.SupportingText(); text != "" {
			hint = text + " " + hint
		}

		return ui.PasswordField(ctx.Label(), strState.Get()).
			InputValue(strState).
			ID(ctx.ID()).
			SupportingText(hint).
			ErrorText(ctx.ErrorText()).
			Disabled(ctx.Disabled()).
			AutoComplete(false).
			FullWidth()
	}
}

func accessCard(
	availGroups []group.Group,
	selectedGroups *core.State[[]group.Group],
	foreignGroups []group.ID,
	availUsers []user.User,
	selectedUsers *core.State[[]user.User],
	access string,
) core.View {
	var systemHint core.View
	if slices.ContainsFunc(selectedGroups.Get(), func(g group.Group) bool { return g.ID == group.System }) {
		systemHint = ui.Text("Secrets in der Systemgruppe nutzt nago selbst, z. B. für den Mailversand. Alle Mitglieder der Systemgruppe erhalten Zugriff.").
			Font(ui.Font{Size: ui.L14}).
			Color(ui.SW0)
	}

	var foreignHint core.View
	if len(foreignGroups) > 0 {
		foreignHint = ui.Text("Außerdem in Gruppen geteilt, denen Sie nicht angehören. Diese Freigaben bleiben erhalten.").
			Font(ui.Font{Size: ui.L14}).
			Color(ui.ST0)
	}

	return cardlayout.Card("Wer hat Zugriff?").
		Body(ui.VStack(
			ui.Text("Es gibt genau ein Secret. Alle hier genannten Personen und alle Mitglieder der ausgewählten Gruppen können es lesen, ändern und löschen."),
			picker.Picker[group.Group]("Gruppen", availGroups, selectedGroups).
				MultiSelect(true).
				Frame(ui.Frame{}.FullWidth()),
			systemHint,
			foreignHint,
			picker.Picker[user.User]("Personen", availUsers, selectedUsers).
				MultiSelect(true).
				Frame(ui.Frame{}.FullWidth()),
			ui.Text(access).Font(ui.Font{Weight: ui.DisplayAndLabelFontWeight}),
		).Gap(ui.L16).Alignment(ui.Leading).FullWidth()).
		Style(cardlayout.TitleCompact).
		Frame(ui.Frame{}.FullWidth())
}

func groupIDs(groups []group.Group) []group.ID {
	res := make([]group.ID, 0, len(groups))
	for _, g := range groups {
		res = append(res, g.ID)
	}
	return res
}

func userIDs(users []user.User) []user.ID {
	res := make([]user.ID, 0, len(users))
	for _, u := range users {
		res = append(res, u.ID)
	}
	return res
}

func sameIDs[T ~string](a, b []T) bool {
	a, b = slices.Clone(a), slices.Clone(b)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}

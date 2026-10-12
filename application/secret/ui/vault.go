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
	"strings"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/list"
	"go.wdy.de/nago/presentation/ui/picker"
)

// VaultPage lists every secret, which the subject can access, exactly once. Where a secret is shared is shown in
// its row, and a filter narrows the list to a group.
func VaultPage(wnd core.Window, pages Pages, uc UseCases) core.View {
	if !wnd.Subject().Valid() {
		return alert.BannerError(user.InvalidSubjectErr)
	}

	var secrets []secret.Secret
	for scr, err := range uc.Secrets.FindMySecrets(wnd.Subject()) {
		if err != nil {
			return alert.BannerError(err)
		}

		secrets = append(secrets, scr)
	}

	slices.SortFunc(secrets, func(a, b secret.Secret) int {
		return strings.Compare(strings.ToLower(secretName(a)), strings.ToLower(secretName(b)))
	})

	query := core.AutoState[string](wnd)
	selectedGroups := core.AutoState[[]group.Group](wnd)

	var groupID group.ID
	if len(selectedGroups.Get()) > 0 {
		groupID = selectedGroups.Get()[0].ID
	}

	var items []core.View
	for _, scr := range secrets {
		if groupID != "" && !slices.Contains(scr.Groups, groupID) {
			continue
		}

		spec := newCredentialTypeSpec(reflect.TypeOf(scr.Credentials))
		if q := strings.ToLower(strings.TrimSpace(query.Get())); q != "" &&
			!strings.Contains(strings.ToLower(secretName(scr)), q) &&
			!strings.Contains(strings.ToLower(wnd.Bundle().Resolve(spec.name)), q) {
			continue
		}

		items = append(items, itemEntry(wnd, pages, uc, scr, spec))
	}

	var content core.View
	switch {
	case len(secrets) == 0:
		content = ui.Text("Es sind noch keine Secrets vorhanden.")
	case len(items) == 0:
		content = ui.Text("Keine Secrets gefunden.")
	default:
		content = list.List(items...).Frame(ui.Frame{}.FullWidth())
	}

	return ui.VStack(
		ui.H1("Tresor"),
		ui.HStack(
			ui.TextField("", query.Get()).
				InputValue(query).
				Style(ui.TextFieldReduced).
				Leading(ui.ImageIcon(heroOutline.MagnifyingGlass)).
				Frame(ui.Frame{MinWidth: ui.L200}),
			groupFilter(wnd, uc, secrets, selectedGroups),
			ui.Spacer(),
			addButton(wnd, pages),
		).Gap(ui.L8).Wrap(true).FullWidth().Padding(ui.Padding{Bottom: ui.L16}),
		content,
	).FullWidth().Alignment(ui.Leading)
}

// groupFilter offers the groups, into which any of the secrets is shared.
func groupFilter(wnd core.Window, uc UseCases, secrets []secret.Secret, selected *core.State[[]group.Group]) core.View {
	var gids []group.ID
	for _, scr := range secrets {
		for _, gid := range scr.Groups {
			if !slices.Contains(gids, gid) {
				gids = append(gids, gid)
			}
		}
	}

	if len(gids) == 0 {
		return nil
	}

	names := groupNames(wnd.Subject(), uc.FindGroupByID, gids)
	groups := make([]group.Group, 0, len(gids))
	for i, gid := range gids {
		groups = append(groups, group.Group{ID: gid, Name: names[i]})
	}

	slices.SortFunc(groups, func(a, b group.Group) int {
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})

	return picker.Picker[group.Group]("", groups, selected).
		Title("Gruppe").
		Stringer(func(g group.Group) string { return g.Name }).
		Frame(ui.Frame{MinWidth: ui.L200})
}

func addButton(wnd core.Window, pages Pages) core.View {
	if !wnd.Subject().HasPermission(secret.PermCreateSecret) {
		return nil
	}

	return ui.PrimaryButton(func() {
		wnd.Navigation().ForwardTo(pages.CreateSecret, nil)
	}).PreIcon(heroSolid.Plus).Title("Secret hinzufügen")
}

func itemEntry(wnd core.Window, pages Pages, uc UseCases, scr secret.Secret, spec credentialTypeSpec) list.TEntry {
	return list.Entry().
		Leading(spec.LogoView()).
		Headline(wnd.Bundle().Resolve(secretName(scr))).
		SupportingText(wnd.Bundle().Resolve(spec.name) + " · " + accessSummary(wnd.Subject(), uc.FindGroupByID, scr)).
		Trailing(ui.ImageIcon(heroOutline.ChevronRight)).
		Action(func() {
			wnd.Navigation().ForwardTo(pages.EditSecret, core.Values{"id": string(scr.ID)})
		})
}

// secretName returns the name of the secret or the default name of its type.
func secretName(scr secret.Secret) string {
	if scr.Credentials == nil {
		return string(scr.ID)
	}

	if name := scr.Credentials.GetName(); name != "" {
		return name
	}

	if field, ok := reflect.TypeOf(scr.Credentials).FieldByName("Name"); ok {
		if name := field.Tag.Get("value"); name != "" {
			return name
		}
	}

	return newCredentialTypeSpec(reflect.TypeOf(scr.Credentials)).name
}

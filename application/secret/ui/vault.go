// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uisecret

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/list"
	"maps"
	"reflect"
	"slices"
	"strings"
)

func VaultPage(wnd core.Window, pages Pages, findMySecrets secret.FindMySecrets, findGrpById group.FindByID) core.View {
	if !wnd.Subject().Valid() {
		return alert.BannerError(user.InvalidSubjectErr)
	}

	return ui.VStack(
		ui.H1("Tresor"),
		addButton(wnd, pages),
		notInGroupSecrets(wnd, pages, findMySecrets),
		groupViews(wnd, pages, findMySecrets, findGrpById),
	).FullWidth().Alignment(ui.Leading)
}

func addButton(wnd core.Window, pages Pages) core.View {
	if !wnd.Subject().HasPermission(secret.PermCreateSecret) {
		return nil
	}

	return ui.HStack(
		ui.PrimaryButton(func() {
			wnd.Navigation().ForwardTo(pages.CreateSecret, nil)
		}).PreIcon(heroSolid.Plus).Title("Secret hinzufügen"),
	).FullWidth().
		Alignment(ui.Trailing).
		Padding(ui.Padding{Bottom: ui.L16})
}

func notInGroupSecrets(wnd core.Window, pages Pages, findMySecrets secret.FindMySecrets) core.View {
	var items []core.View
	var total int
	for scr, err := range findMySecrets(wnd.Subject()) {
		if err != nil {
			return alert.BannerError(err)
		}

		total++

		if len(scr.Groups) == 0 {
			items = append(items, itemEntry(wnd, pages, scr))
		}
	}

	if total == 0 {
		return ui.Text("Es sind noch keine Einträge vorhanden.")
	}

	if len(items) == 0 {
		return nil
	}

	return list.List(items...).
		Caption(ui.Text("nur persönlich")).
		Frame(ui.Frame{}.FullWidth())
}

func groupViews(wnd core.Window, pages Pages, findMySecrets secret.FindMySecrets, findGrpById group.FindByID) core.View {
	tmp := map[group.Group][]secret.Secret{}

	for scr, err := range findMySecrets(wnd.Subject()) {
		if err != nil {
			return alert.BannerError(err)
		}

		for _, gid := range scr.Groups {
			var grp group.Group
			if findGrpById != nil {
				g, err := findGrpById(wnd.Subject(), gid)
				if err != nil {
					// silently ignored, probably a permission problem
				} else if g.IsSome() {
					grp = g.Unwrap()
				}
			}

			if grp.ID == "" {
				// either the group is missing or we don't have the permission to look up the group data
				grp.ID = gid
			}

			tmp[grp] = append(tmp[grp], scr)
		}
	}

	groups := slices.Collect(maps.Keys(tmp))
	slices.SortFunc(groups, func(a, b group.Group) int {
		return strings.Compare(string(a.ID), string(b.ID))
	})

	var groupLists []core.View
	if len(groups) > 0 {

		groupLists = append(groupLists, ui.VStack(ui.H2("in Gruppen geteilt")).Padding(ui.Padding{Top: ui.L80}))
	}

	for _, g := range groups {

		secrets := tmp[g]
		slices.SortFunc(secrets, func(a, b secret.Secret) int {
			return strings.Compare(a.Credentials.GetName(), b.Credentials.GetName())
		})

		var items []core.View
		for _, scr := range secrets {
			items = append(items, itemEntry(wnd, pages, scr))
		}

		name := g.Name
		if name == "" {
			name = string(g.ID)
		}

		groupLists = append(groupLists, list.List(items...).Caption(ui.Text(name)).Frame(ui.Frame{}.FullWidth()))
	}

	return ui.VStack(groupLists...).Gap(ui.L16).FullWidth().Alignment(ui.Leading)
}

func itemEntry(wnd core.Window, pages Pages, scr secret.Secret) list.TEntry {
	spec := newCredentialTypeSpec(reflect.TypeOf(scr.Credentials))
	name := scr.Credentials.GetName()
	if name == "" {
		field, ok := spec.refType.FieldByName("Name")
		if ok {
			name = field.Tag.Get("value")
		}
	}

	return list.Entry().
		Leading(spec.LogoView()).
		Headline(name).
		Trailing(ui.ImageIcon(heroOutline.ChevronRight)).
		SupportingText(spec.description).
		Action(func() {
			wnd.Navigation().ForwardTo(pages.EditSecret, core.Values{"id": string(scr.ID)})
		})
}

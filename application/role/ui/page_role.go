// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uirole

import (
	"fmt"
	"os"
	"slices"

	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/form"
	"go.wdy.de/nago/presentation/ui/picker"
)

func PageRole(wnd core.Window, pages Pages, uc role.UseCases) core.View {
	optRole, err := uc.FindByID(wnd.Subject(), role.ID(wnd.Values()["role"]))
	if err != nil {
		return alert.BannerError(err)
	}

	if optRole.IsNone() {
		return alert.BannerError(fmt.Errorf("role not found: %s: %w", wnd.Values()["role"], os.ErrNotExist))
	}

	uRole := optRole.Unwrap()

	stateName := core.AutoState[string](wnd).Init(func() string {
		return uRole.Name
	})

	stateDesc := core.AutoState[string](wnd).Init(func() string {
		return uRole.Description
	})

	statePerms := core.AutoState[[]permission.Permission](wnd).Init(func() []permission.Permission {
		var tmp []permission.Permission

		for pid, err := range uc.ListPermissions(wnd.Subject(), uRole.ID) {
			if err != nil {
				alert.ShowBannerError(wnd, err)
				return nil
			}

			if perm, ok := permission.Find(pid); ok {
				tmp = append(tmp, perm)
			}
		}

		return tmp
	})

	return ui.VStack(
		ui.H1(uRole.Name),
		form.Card(
			ui.TextField(rstring.LabelIdentifier.Get(wnd), string(uRole.ID)).
				FullWidth().
				Disabled(true),
			ui.TextField(rstring.LabelName.Get(wnd), stateName.Get()).
				FullWidth().
				InputValue(stateName),
			ui.TextField(rstring.LabelDescription.Get(wnd), stateDesc.Get()).
				FullWidth().
				Lines(3).
				InputValue(stateDesc),
			picker.Picker[permission.Permission](rstring.LabelPermission.Get(wnd), slices.Collect(permission.All()), statePerms).
				MultiSelect(true).
				FullWidth(),
		).Gap(ui.L8),
		ui.HLine(),
		ui.HStack(
			ui.SecondaryButton(func() {
				wnd.Navigation().Back()
			}).Title(rstring.ActionBack.Get(wnd)),
			ui.PrimaryButton(func() {
				uRole.Name = stateName.Get()
				uRole.Description = stateDesc.Get()

				if err := uc.Update(wnd.Subject(), uRole); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				var pids []permission.ID
				for _, perm := range statePerms.Get() {
					pids = append(pids, perm.ID)
				}
				if err := uc.UpdatePermissions(wnd.Subject(), uRole.ID, pids); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				wnd.Navigation().BackwardTo(pages.Roles, wnd.Values())
			}).Title(rstring.ActionSave.Get(wnd)),
		).FullWidth().Alignment(ui.Trailing).Gap(ui.L8),
	).Alignment(ui.Leading).
		Frame(ui.Frame{}.Larger())
}

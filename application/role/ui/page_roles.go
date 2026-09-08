// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uirole

import (
	"fmt"
	"strings"

	"github.com/worldiety/i18n"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/dataview"
	"go.wdy.de/nago/presentation/ui/form"
	"golang.org/x/text/language"
)

var (
	StrCreateRole             = i18n.MustString("nago.role.create", i18n.Values{language.German: "Rolle erstellen", language.English: "Create Role"})
	StrCreateRoleWithCustomID = i18n.MustString("nago.role.create_with_custom_id", i18n.Values{language.German: "Rolle mit ID erstellen", language.English: "Create Role with ID"})

	StrSystemRoleHint = i18n.MustString("nago.role.system_role_hint", i18n.Values{
		language.German:  "Diese Rolle wird von der Anwendung verwaltet. Ihre Berechtigungen gehören zu der Funktion, die sie bereitstellt, und lassen sich hier nicht ändern. Weisen Sie die Rolle einfach den Nutzern zu, die die Funktion verwenden sollen.",
		language.English: "This role is maintained by the application. Its permissions belong to the feature providing it and cannot be changed here. Simply assign the role to the users who should use that feature.",
	})
)

func PageRoles(wnd core.Window, pages Pages, useCases role.UseCases) core.View {
	createPresented := core.AutoState[bool](wnd)
	createWithCustomID := core.AutoState[bool](wnd)

	return ui.VStack(
		dialogCreate(wnd, useCases, createWithCustomID.Get(), createPresented),
		ui.H1(rstring.LabelRoles.Get(wnd)),
		dataview.FromData(wnd, dataview.Data[role.Role, role.ID]{
			FindAll: func(yield func(role.ID, error) bool) {
				for r, err := range useCases.FindAll(wnd.Subject()) {
					if !yield(r.ID, err) {
						return
					}
				}
			},
			FindByID: func(id role.ID) (option.Opt[role.Role], error) {
				return useCases.FindByID(wnd.Subject(), id)
			},
			Fields: []dataview.Field[role.Role]{
				{
					ID:   "name",
					Name: rstring.LabelName.Get(wnd),
					Map: func(obj role.Role) core.View {
						return ui.Text(obj.Name).AccessibilityLabel(string(obj.ID))
					},
					Comparator: func(a, b role.Role) int {
						return strings.Compare(a.Name, b.Name)
					},
				},
				{
					ID:   "description",
					Name: rstring.LabelDescription.Get(wnd),
					Map: func(obj role.Role) core.View {
						return ui.Text(obj.Description)
					},
				},
				{
					ID:   "system",
					Name: role.StrLabelSystem.Get(wnd),
					Map: func(obj role.Role) core.View {
						// Shown as a column rather than hidden, so it is obvious up front why a role cannot
						// be selected for deletion.
						if !obj.IsSystem() {
							return nil
						}

						return ui.Text(role.StrLabelSystem.Get(wnd))
					},
				},
			},
		}).Search(true).
			NextActionIndicator(true).
			Action(func(e role.Role) {
				wnd.Navigation().ForwardTo(pages.Role, core.Values{"role": string(e.ID)})
			}).CreateOptions(
			dataview.CreateOption{
				Name: StrCreateRoleWithCustomID.Get(wnd),
				Action: func() error {
					createWithCustomID.Set(true)
					createPresented.Set(true)
					return nil
				},
			},
			dataview.CreateOption{
				Name: StrCreateRole.Get(wnd),
				Action: func() error {
					createWithCustomID.Set(false)
					createPresented.Set(true)
					return nil
				},
			},
		).SelectOptions(
			dataview.NewSelectOptionDelete[role.ID](wnd, func(selected []role.ID) error {
				// The use case refuses system roles anyway; checking here as well turns a bulk delete that
				// aborts halfway into one that reports the problem before removing anything.
				for _, id := range selected {
					optRole, err := useCases.FindByID(wnd.Subject(), id)
					if err != nil {
						return err
					}

					if optRole.IsSome() && optRole.Unwrap().IsSystem() {
						return fmt.Errorf("%s: %s", optRole.Unwrap().Name, StrSystemRoleHint.Get(wnd))
					}
				}

				for _, id := range selected {
					if err := useCases.Delete(wnd.Subject(), id); err != nil {
						return err
					}
				}

				return nil
			}),
		),
	).FullWidth().
		Alignment(ui.Leading)
}

func dialogCreate(wnd core.Window, useCases role.UseCases, customID bool, presented *core.State[bool]) core.View {
	if !presented.Get() {
		return nil
	}

	state := core.AutoState[role.Role](wnd)
	errState := core.AutoState[error](wnd)

	// The System flag is set by [Configurator.DeclareSystemRole], never by hand, so it must not appear as a
	// checkbox in the create dialog.
	ignore := []string{"System"}
	if !customID {
		ignore = append(ignore, "ID")
	}

	return alert.Dialog(
		StrCreateRole.Get(wnd),
		form.Auto[role.Role](form.AutoOptions{IgnoreFields: ignore, Errors: errState.Get()}, state),
		presented,
		alert.Closeable(),
		alert.Cancel(nil),
		alert.Create(func() (close bool) {
			_, err := useCases.Create(wnd.Subject(), state.Get())
			errState.Set(err)
			return err == nil
		}),
	)
}

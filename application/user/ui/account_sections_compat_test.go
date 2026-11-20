// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser_test

import (
	"slices"
	"testing"

	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

// TestAccountViewWithoutRegistration is T1: without a registered function, the account view shows the sections,
// entries, texts and order nago always had, in the dialog of the scaffold (click on the avatar) and wherever
// uiuser.AccountView is built: „Konto“ with „Profil verwalten“ and „Farbschema wechseln“, „Einstellungen“ with
// „Verwalten“ for whom admin.HasEntries holds, then „Abmelden“. Nothing else stands between name and mail above and
// „Abmelden“ below. The counter row is a person, for whom admin.HasEntries is false: the same view without
// „Einstellungen“. The entries keep their actions, and the scaffold keeps the placement of its dialog. Written at
// the state before the sections became data, where it passes as well.
func TestAccountViewWithoutRegistration(t *testing.T) {
	w := newACWorld(t, nil, nil)
	ada, otto := w.adminAndPlain(t)

	withSettings := []string{acAccount, acProfile, acTheme, acSettings, acAdminCntr, acSignOut}
	withoutSettings := []string{acAccount, acProfile, acTheme, acSignOut}

	rows := 0
	for _, path := range []string{"scaffold dialog", "account view"} {
		for _, c := range []struct {
			name string
			s    user.Subject
			want []string
		}{
			{"admin.HasEntries true", ada, withSettings},
			{"counter row: admin.HasEntries false", otto, withoutSettings},
		} {
			t.Run(path+"/"+c.name, func(t *testing.T) {
				rows++
				var root proto.Component
				if path == "scaffold dialog" {
					_, root = w.acScaffoldDialog(t, c.s)
				} else {
					_, root = w.acView(t, c.s)
				}

				texts := acTexts(root)
				if got := acSectionTexts(t, texts); !slices.Equal(got, c.want) {
					t.Errorf("sections: got %q, expected %q", got, c.want)
				}

				// name and mail stand above the sections
				acOrder(t, "name and mail", texts, c.s.Name(), c.s.Email(), acAccount)
				if !slices.Contains(c.want, acSettings) {
					acAbsent(t, "without entries of the admin center", texts, acSettings, acAdminCntr)
				}
			})
		}
	}

	if rows != 4 {
		t.Fatalf("premise: %d rows ran", rows)
	}

	// the entries keep their actions
	um, err := w.cfg.UserManagement()
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range []struct {
		entry string
		route string
	}{
		{acProfile, string(um.Pages.MyProfile)},
		{acAdminCntr, string(w.mgmt.Pages.AdminCenter)},
	} {
		t.Run("action/"+c.entry, func(t *testing.T) {
			win, _ := w.acView(t, ada)
			if string(win.Route().Path) == c.route {
				t.Fatalf("premise: the window is at %s before the click", c.route)
			}

			win.Click(win.Find(nagotest.Text(c.entry)))
			if got := string(win.Route().Path); got != c.route {
				t.Errorf("click on %q: route %s, expected %s", c.entry, got, c.route)
			}
		})
	}

	// the scaffold places its dialog at the upper right in a window higher than 600 dp, otherwise in the center
	for _, c := range []struct {
		name    string
		height  int
		align   ui.Alignment
		padding bool
	}{
		{"higher than 600", 800, ui.TopTrailing, true},
		{"counter row: 600 or less", 560, ui.Center, false},
	} {
		t.Run("placement/"+c.name, func(t *testing.T) {
			_, modal := w.acScaffoldDialog(t, ada, nagotest.Size(1280, c.height))
			box, ok := modal.(*proto.Modal).Content.(*proto.Box)
			if !ok || len(box.Children) != 1 {
				t.Fatalf("premise: the dialog is no box with one child: %T", modal.(*proto.Modal).Content)
			}

			if got := box.Children[0].Alignment; got != proto.Alignment(c.align) {
				t.Errorf("alignment %v, expected %v", got, c.align)
			}

			if got := box.Padding.Top != ""; got != c.padding {
				t.Errorf("padding of the modal %+v, expected set %v", box.Padding, c.padding)
			}
		})
	}
}

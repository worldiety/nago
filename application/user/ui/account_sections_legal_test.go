// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser_test

import (
	"reflect"
	"slices"
	"testing"

	"go.wdy.de/nago/application/theme"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/proto"
)

// TestAccountViewLegalTextsLine is T9: without any legal text, nothing follows „Abmelden“ in the account view, in
// particular no line and no empty row, which would leave room below it. The counter row sets the imprint: then a
// line follows „Abmelden“ and below it the link „Impressum“. Both rows in the dialog of the scaffold and in the
// account view of a page. Written at the state before the sections became data, where the line stands also without
// legal texts.
func TestAccountViewLegalTextsLine(t *testing.T) {
	const impress = "https://example.org/impressum"
	w := newACWorld(t, nil, nil)
	ada := w.person(t, "Ada", "Admin")

	open := func(t *testing.T, path string) proto.Component {
		t.Helper()
		if path == "scaffold dialog" {
			_, modal := w.acScaffoldDialog(t, ada)
			return modal
		}

		_, root := w.acView(t, ada)
		return root
	}

	// after returns the components a person sees after the button „Abmelden“ in its stack, and their texts.
	after := func(t *testing.T, root proto.Component) (dividers int, siblings int, texts []string) {
		t.Helper()
		btn := acSignOutButton(t, root)
		if len(btn.Parents) == 0 {
			t.Fatalf("premise: the button %q has no stack", acSignOut)
		}

		children := acChildren(btn.Parents[len(btn.Parents)-1])
		at := slices.Index(children, btn.Component)
		if at < 0 {
			t.Fatalf("premise: the button %q is no child of its stack", acSignOut)
		}

		for _, c := range children[at+1:] {
			siblings++
			for _, n := range acNodes(c) {
				if _, ok := n.Component.(*proto.Divider); ok {
					dividers++
				}
			}

			texts = append(texts, acTexts(c)...)
		}

		return dividers, siblings, texts
	}

	paths := []string{"scaffold dialog", "account view"}
	for _, path := range paths {
		t.Run("without legal texts/"+path, func(t *testing.T) {
			root := open(t, path)
			dividers, siblings, texts := after(t, root)
			if dividers != 0 {
				t.Errorf("%d lines below %q without legal texts", dividers, acSignOut)
			}

			if siblings != 0 {
				t.Errorf("%d components below %q without legal texts (room below it): %s", siblings, acSignOut, acJoin(texts))
			}
		})
	}

	sm := std.Must(w.cfg.SettingsManagement())
	cur, err := sm.UseCases.LoadGlobal(user.SU(), reflect.TypeFor[theme.Settings]())
	if err != nil {
		t.Fatal(err)
	}

	s := cur.(theme.Settings)
	s.Impress = impress
	if err := sm.UseCases.StoreGlobal(user.SU(), s); err != nil {
		t.Fatal(err)
	}

	for _, path := range paths {
		t.Run("counter row: imprint set/"+path, func(t *testing.T) {
			root := open(t, path)
			dividers, _, texts := after(t, root)
			if dividers != 1 {
				t.Errorf("%d lines below %q with the imprint, expected one", dividers, acSignOut)
			}

			if !slices.Equal(texts, []string{"Impressum"}) {
				t.Errorf("below %q: %s, expected the link Impressum only", acSignOut, acJoin(texts))
			}

			var links []string
			for _, n := range acNodes(root) {
				if tv, ok := n.Component.(*proto.TextView); ok && tv.Value == "Impressum" {
					links = append(links, string(tv.Link.Url))
				}
			}

			if !slices.Equal(links, []string{impress}) {
				t.Errorf("link Impressum to %q, expected %q", links, impress)
			}

			// the line stands between „Abmelden“ and the link
			var order []string
			for _, n := range acNodes(root) {
				switch c := n.Component.(type) {
				case *proto.Divider:
					order = append(order, "line")
				case *proto.TextView:
					if c.Value == acSignOut || c.Value == "Impressum" {
						order = append(order, string(c.Value))
					}
				}
			}

			if i := slices.Index(order, acSignOut); i < 0 || !slices.Equal(order[i:], []string{acSignOut, "line", "Impressum"}) {
				t.Errorf("order below the sections %q, expected %q", order, []string{acSignOut, "line", "Impressum"})
			}
		})
	}
}

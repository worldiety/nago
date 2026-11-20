// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser_test

// Acceptance tests of the sections of the account view as data (T2 to T8, T7b and T10 of the concept, see
// account_sections_helpers_test.go): uiuser.AccountSectionModel, AccountSections, InsertAccountSection,
// RemoveAccountEntry, AccountDialog, TAccountSection.ID and application.Configurator.AddAccountSections.

import (
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/user"
	uiuser "go.wdy.de/nago/application/user/ui"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

// Sections of the tests and pages with an own frame.
const (
	acX       = uiuser.AccountSectionID("test-section-x")
	acY       = uiuser.AccountSectionID("test-section-y")
	acZ       = uiuser.AccountSectionID("test-section-z")
	acXTitle  = "Abschnitt X"
	acXEntry  = "Eintrag X"
	acYTitle  = "Abschnitt Y"
	acYEntry  = "Eintrag Y"
	acUnknown = uiuser.AccountSectionID("test-section-unknown")

	acRouteOwn       = "ac-own"        // a page with an own frame and AccountDialog without options
	acRouteOwnPlaced = "ac-own-placed" // the same with alert.Alignment and alert.ModalPadding
	acRouteReplace   = "ac-replace"    // AccountView(wnd).Sections(own)
	acRouteAppend    = "ac-append"     // AccountView(wnd).WithSections(own)
	acOpenDialog     = "Konto öffnen"
	acOwnTitle       = "Eigener Abschnitt"
	acOwnEntry       = "Eigener Eintrag"
	acOwnID          = "test-own-section"
	acCloseLabel     = "Dialog schließen" // rstring.ActionDialogClose in German, the locale of nagotest
)

// acIDsKnown are the ids of the sections, which the tests look for.
var acIDsKnown = []string{string(uiuser.SectionAccount), string(uiuser.SectionSettings), string(acX), string(acY), string(acZ), string(acUnknown), acOwnID}

// The section texts of nago, without and with the admin center.
var (
	acNago      = []string{acAccount, acProfile, acTheme}
	acNagoAdmin = []string{acSettings, acAdminCntr}
)

func acCat(parts ...[]string) []string {
	var res []string
	for _, p := range parts {
		res = append(res, p...)
	}

	return append(res, acSignOut)
}

// acSection is a section of an application with one text as entry, or without entries if entry is empty.
func acSection(id uiuser.AccountSectionID, title, entry string) uiuser.AccountSectionModel {
	m := uiuser.AccountSectionModel{ID: id, Title: title}
	if entry != "" {
		m.Entries = []uiuser.AccountEntryModel{{ID: uiuser.AccountEntryID(string(id) + "-entry"), View: ui.Text(entry)}}
	}

	return m
}

func acInsert(after uiuser.AccountSectionID, s uiuser.AccountSectionModel) uiuser.AccountSections {
	return func(wnd core.Window, sections []uiuser.AccountSectionModel) []uiuser.AccountSectionModel {
		return uiuser.InsertAccountSection(sections, after, s)
	}
}

// acOwnPage is a page with an own frame, which shows AccountDialog with opts after a click on acOpenDialog and
// tells the state it passes.
func acOwnPage(opts ...alert.Option) func(wnd core.Window) core.View {
	return func(wnd core.Window) core.View {
		presented := core.StateOf[bool](wnd, "test.account.dialog.presented")
		return ui.VStack(
			ui.Text(fmt.Sprintf("presented=%v", presented.Get())),
			ui.PrimaryButton(func() { presented.Set(true) }).Title(acOpenDialog),
			uiuser.AccountDialog(wnd, presented, opts...),
		)
	}
}

// newACAPIWorld is newACWorld with the pages of this file. fns are registered in this order, before the standard
// systems if early, otherwise at the end of the configuration.
func newACAPIWorld(t *testing.T, early bool, fns ...uiuser.AccountSections) *acWorld {
	t.Helper()
	register := func(cfg *application.Configurator) {
		for _, fn := range fns {
			cfg.AddAccountSections(fn)
		}
	}

	own := func(cfg *application.Configurator) {
		cfg.RootView(acRouteOwn, acOwnPage())
		cfg.RootView(acRouteOwnPlaced, acOwnPage(alert.Alignment(ui.BottomLeading), alert.ModalPadding(ui.Padding{}.All(ui.L24))))
		cfg.RootView(acRouteReplace, func(wnd core.Window) core.View {
			return uiuser.AccountView(wnd).Sections(uiuser.AccountSection(acOwnTitle).ID(acOwnID).Entries(ui.Text(acOwnEntry)))
		})
		cfg.RootView(acRouteAppend, func(wnd core.Window) core.View {
			return uiuser.AccountView(wnd).WithSections(uiuser.AccountSection(acOwnTitle).ID(acOwnID).Entries(ui.Text(acOwnEntry)))
		})
	}

	if early {
		return newACWorld(t, register, own)
	}

	return newACWorld(t, nil, func(cfg *application.Configurator) {
		own(cfg)
		register(cfg)
	})
}

// acOwnDialog opens route as s and clicks acOpenDialog. Before the click, there is no dialog and the page tells
// presented=false.
func (w *acWorld) acOwnDialog(t *testing.T, s user.Subject, route core.NavigationPath) (*nagotest.Window, proto.Component) {
	t.Helper()
	win := w.app.Open(t, s, route)
	win.Find(nagotest.Text("presented=false"))
	if n := win.FindAll(nagotest.Type[*proto.Modal]()).Len(); n != 0 {
		t.Fatalf("AccountDialog with presented false: %d dialogs in the tree, expected none", n)
	}

	win.Click(win.Find(nagotest.Text(acOpenDialog)))
	win.Find(nagotest.Text("presented=true"))
	modal := win.Find(nagotest.Type[*proto.Modal]()).Node().Component
	if texts := acTexts(modal); len(texts) == 0 || texts[0] != acDialogTitle {
		t.Fatalf("AccountDialog with presented true: no dialog %q: %s", acDialogTitle, acJoin(texts))
	}

	return win, modal
}

// acRecorder is an AccountSections, which keeps what it received last and returns it unchanged.
type acRecorder struct {
	mutex    sync.Mutex
	sections []string // per section: id, title and the ids of its entries
}

func (r *acRecorder) fn(wnd core.Window, sections []uiuser.AccountSectionModel) []uiuser.AccountSectionModel {
	var res []string
	for _, s := range sections {
		x := fmt.Sprintf("%s %q:", s.ID, s.Title)
		for _, e := range s.Entries {
			x += " " + string(e.ID)
		}

		res = append(res, x)
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.sections = res
	return sections
}

func (r *acRecorder) get() []string {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return slices.Clone(r.sections)
}

// TestAccountSectionsInsertAfterAccount is T2: a section X registered after nago's „Konto“ stands between „Konto“ and
// „Einstellungen“, in the order of the section ids as well, with its title and its entry inside its section; nago's
// entries stay in nago's sections. A function registered before the standard systems applies as well. The function
// receives nago's sections with their titles and the promised ids of sections and entries. Counter row: without
// registration only „Konto“ and „Einstellungen“. Each row also for a person without entries of the admin center.
func TestAccountSectionsInsertAfterAccount(t *testing.T) {
	x := acSection(acX, acXTitle, acXEntry)
	withX := []string{acXTitle, acXEntry}
	for _, c := range []struct {
		name    string
		early   bool
		insert  bool
		ada     []string
		otto    []string
		adaIDs  []string
		ottoIDs []string
	}{
		{
			name: "registered at the end of the configuration", insert: true,
			ada: acCat(acNago, withX, acNagoAdmin), otto: acCat(acNago, withX),
			adaIDs:  []string{string(uiuser.SectionAccount), string(acX), string(uiuser.SectionSettings)},
			ottoIDs: []string{string(uiuser.SectionAccount), string(acX)},
		},
		{
			name: "registered before the standard systems", early: true, insert: true,
			ada: acCat(acNago, withX, acNagoAdmin), otto: acCat(acNago, withX),
			adaIDs:  []string{string(uiuser.SectionAccount), string(acX), string(uiuser.SectionSettings)},
			ottoIDs: []string{string(uiuser.SectionAccount), string(acX)},
		},
		{
			name: "counter row: no registration",
			ada:  acCat(acNago, acNagoAdmin), otto: acCat(acNago),
			adaIDs:  []string{string(uiuser.SectionAccount), string(uiuser.SectionSettings)},
			ottoIDs: []string{string(uiuser.SectionAccount)},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			// the recorder comes first and changes nothing; the counter row registers nothing at all
			rec := &acRecorder{}
			var fns []uiuser.AccountSections
			if c.insert {
				fns = append(fns, rec.fn, acInsert(uiuser.SectionAccount, x))
			}

			w := newACAPIWorld(t, c.early, fns...)
			ada, otto := w.adminAndPlain(t)
			for _, p := range []struct {
				name     string
				s        user.Subject
				want     []string
				ids      []string
				received []string
			}{
				{"admin.HasEntries true", ada, c.ada, c.adaIDs, []string{
					`nago-account "Konto": nago-account-profile nago-account-theme`,
					`nago-settings "Einstellungen": nago-settings-admin-center`,
				}},
				{"admin.HasEntries false", otto, c.otto, c.ottoIDs, []string{
					`nago-account "Konto": nago-account-profile nago-account-theme`,
				}},
			} {
				t.Run(p.name, func(t *testing.T) {
					_, root := w.acView(t, p.s)
					texts := acTexts(root)
					if got := acSectionTexts(t, texts); !slices.Equal(got, p.want) {
						t.Errorf("sections: got %q, expected %q", got, p.want)
					}

					if got := acIDs(root, acIDsKnown...); !slices.Equal(got, p.ids) {
						t.Errorf("ids of the sections: got %q, expected %q", got, p.ids)
					}

					for _, in := range []struct{ text, id string }{
						{acAccount, string(uiuser.SectionAccount)},
						{acProfile, string(uiuser.SectionAccount)},
						{acTheme, string(uiuser.SectionAccount)},
						{acSettings, string(uiuser.SectionSettings)},
						{acAdminCntr, string(uiuser.SectionSettings)},
						{acXTitle, string(acX)},
						{acXEntry, string(acX)},
					} {
						if slices.Contains(p.want, in.text) && !acTextWithin(root, in.text, in.id) {
							t.Errorf("%q is not within the section %s", in.text, in.id)
						}
					}

					if c.insert {
						if got := rec.get(); !slices.Equal(got, p.received) {
							t.Errorf("the function received the sections %q, expected nago's %q", got, p.received)
						}
					}
				})
			}
		})
	}
}

// TestAccountSectionsRemoveEntryPerWindow is T3: a function removes „Verwalten“ (uiuser.EntryAdminCenter) for the
// person A of the window only; then A sees neither „Verwalten“ nor the section „Einstellungen“ left without entries,
// in the account view and in the dialog of the scaffold. Counter row: the person B, for whom the function changes
// nothing, sees „Einstellungen“ with „Verwalten“. Premise: admin.HasEntries holds for both.
func TestAccountSectionsRemoveEntryPerWindow(t *testing.T) {
	var hide atomic.Value
	hide.Store(user.ID(""))
	w := newACAPIWorld(t, false, func(wnd core.Window, sections []uiuser.AccountSectionModel) []uiuser.AccountSectionModel {
		if wnd.Subject().ID() == hide.Load().(user.ID) {
			return uiuser.RemoveAccountEntry(sections, uiuser.EntryAdminCenter)
		}

		return sections
	})

	a := w.person(t, "Anna", "Ausgeblendet", user.PermFindAll)
	b := w.person(t, "Bert", "Behalten", user.PermFindAll)
	for _, s := range []user.Subject{a, b} {
		if !w.mgmt.HasEntries(s) {
			t.Fatalf("premise: admin.HasEntries is false for %s", s.Name())
		}
	}

	hide.Store(a.ID())
	for _, path := range []string{"account view", "scaffold dialog"} {
		for _, c := range []struct {
			name string
			s    user.Subject
			want []string
			ids  []string
		}{
			{"A, the function removes the entry", a, acCat(acNago), []string{string(uiuser.SectionAccount)}},
			{"counter row: B, the function keeps it", b, acCat(acNago, acNagoAdmin), []string{string(uiuser.SectionAccount), string(uiuser.SectionSettings)}},
		} {
			t.Run(path+"/"+c.name, func(t *testing.T) {
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

				if got := acIDs(root, acIDsKnown...); !slices.Equal(got, c.ids) {
					t.Errorf("ids of the sections: got %q, expected %q", got, c.ids)
				}
			})
		}
	}
}

// TestAccountSectionsRegistrations is T4: of two registrations, the second receives what the first returned, in
// the order of registration: the first inserts X after „Konto“, the second Y after X, which gives Konto, X, Y,
// Einstellungen. Counter rows: the first only shows X without Y; the second registered first does not see X, its
// anchor is unknown then, and Y comes last.
func TestAccountSectionsRegistrations(t *testing.T) {
	x := acSection(acX, acXTitle, acXEntry)
	y := acSection(acY, acYTitle, acYEntry)
	withX := []string{acXTitle, acXEntry}
	withY := []string{acYTitle, acYEntry}
	for _, c := range []struct {
		name   string
		order  []string // "x" for the first, "y" for the second
		want   []string
		second []string // what the second received, if registered
	}{
		{"first, then second", []string{"x", "y"}, acCat(acNago, withX, withY, acNagoAdmin), []string{
			`nago-account "Konto": nago-account-profile nago-account-theme`,
			`test-section-x "Abschnitt X": test-section-x-entry`,
			`nago-settings "Einstellungen": nago-settings-admin-center`,
		}},
		{"counter row: the first only", []string{"x"}, acCat(acNago, withX, acNagoAdmin), nil},
		{"counter row: second, then first", []string{"y", "x"}, acCat(acNago, withX, acNagoAdmin, withY), []string{
			`nago-account "Konto": nago-account-profile nago-account-theme`,
			`nago-settings "Einstellungen": nago-settings-admin-center`,
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			rec := &acRecorder{}
			var fns []uiuser.AccountSections
			for _, o := range c.order {
				switch o {
				case "x":
					fns = append(fns, acInsert(uiuser.SectionAccount, x))
				case "y":
					fns = append(fns, func(wnd core.Window, sections []uiuser.AccountSectionModel) []uiuser.AccountSectionModel {
						return uiuser.InsertAccountSection(rec.fn(wnd, sections), acX, y)
					})
				}
			}

			w := newACAPIWorld(t, false, fns...)
			ada, _ := w.adminAndPlain(t)
			_, root := w.acView(t, ada)
			if got := acSectionTexts(t, acTexts(root)); !slices.Equal(got, c.want) {
				t.Errorf("sections: got %q, expected %q", got, c.want)
			}

			if got := rec.get(); !slices.Equal(got, c.second) {
				t.Errorf("the second received %q, expected %q", got, c.second)
			}
		})
	}
}

// TestAccountSectionsSameInEveryDialog is T5: with a section X registered, the dialog of the scaffold (click on the
// avatar) and AccountDialog on a page with an own frame show the same person the same sections in the same order,
// X at the same place. Counter row: without registration, X is in neither of them, and both are still equal.
func TestAccountSectionsSameInEveryDialog(t *testing.T) {
	x := acSection(acX, acXTitle, acXEntry)
	for _, c := range []struct {
		name string
		fns  []uiuser.AccountSections
		want []string
	}{
		{"X registered", []uiuser.AccountSections{acInsert(uiuser.SectionAccount, x)}, acCat(acNago, []string{acXTitle, acXEntry}, acNagoAdmin)},
		{"counter row: no registration", nil, acCat(acNago, acNagoAdmin)},
	} {
		t.Run(c.name, func(t *testing.T) {
			w := newACAPIWorld(t, false, c.fns...)
			ada, _ := w.adminAndPlain(t)
			_, scaffold := w.acScaffoldDialog(t, ada)
			_, own := w.acOwnDialog(t, ada, acRouteOwn)

			a, b := acSectionTexts(t, acTexts(scaffold)), acSectionTexts(t, acTexts(own))
			if !slices.Equal(a, c.want) {
				t.Errorf("dialog of the scaffold: %q, expected %q", a, c.want)
			}

			if !slices.Equal(b, c.want) {
				t.Errorf("AccountDialog on an own page: %q, expected %q", b, c.want)
			}

			if ia, ib := acIDs(scaffold, acIDsKnown...), acIDs(own, acIDsKnown...); !slices.Equal(ia, ib) {
				t.Errorf("ids of the sections differ: scaffold %q, own page %q", ia, ib)
			}
		})
	}
}

// TestAccountViewSectionsReplaceAppend is T6: AccountView(wnd).Sections(...) shows only the given sections, also
// none of the registered ones; the counter row WithSections(...) shows the registered sections and appends the given
// one. The given section carries the id set by TAccountSection.ID.
func TestAccountViewSectionsReplaceAppend(t *testing.T) {
	w := newACAPIWorld(t, false, acInsert(uiuser.SectionAccount, acSection(acX, acXTitle, acXEntry)))
	ada, _ := w.adminAndPlain(t)

	t.Run("Sections replaces everything", func(t *testing.T) {
		win := w.app.Open(t, ada, acRouteReplace)
		texts := acTexts(win.Tree())
		acOrder(t, "Sections", texts, acOwnTitle, acOwnEntry, acSignOut)
		acAbsent(t, "Sections", texts, acAccount, acProfile, acTheme, acXTitle, acXEntry, acSettings, acAdminCntr)
		if got := acIDs(win.Tree(), acIDsKnown...); !slices.Equal(got, []string{acOwnID}) {
			t.Errorf("ids of the sections %q, expected %q", got, []string{acOwnID})
		}
	})

	t.Run("counter row: WithSections appends", func(t *testing.T) {
		win := w.app.Open(t, ada, acRouteAppend)
		want := acCat(acNago, []string{acXTitle, acXEntry}, acNagoAdmin, []string{acOwnTitle, acOwnEntry})
		if got := acSectionTexts(t, acTexts(win.Tree())); !slices.Equal(got, want) {
			t.Errorf("sections: got %q, expected %q", got, want)
		}

		wantIDs := []string{string(uiuser.SectionAccount), string(acX), string(uiuser.SectionSettings), acOwnID}
		if got := acIDs(win.Tree(), acIDsKnown...); !slices.Equal(got, wantIDs) {
			t.Errorf("ids of the sections %q, expected %q", got, wantIDs)
		}
	})
}

// TestAccountSectionsUnknownAnchor is T7: inserted after an unknown section, X comes last, also after a section of
// nago, which the person does not see. Counter row: after a known section, X stands at its place.
func TestAccountSectionsUnknownAnchor(t *testing.T) {
	withX := []string{acXTitle, acXEntry}
	for _, c := range []struct {
		name  string
		after uiuser.AccountSectionID
		admin bool
		want  []string
	}{
		{"unknown anchor", acUnknown, true, acCat(acNago, acNagoAdmin, withX)},
		{"anchor „Einstellungen“, which the person does not see", uiuser.SectionSettings, false, acCat(acNago, withX)},
		{"counter row: anchor „Konto“", uiuser.SectionAccount, true, acCat(acNago, withX, acNagoAdmin)},
	} {
		t.Run(c.name, func(t *testing.T) {
			w := newACAPIWorld(t, false, acInsert(c.after, acSection(acX, acXTitle, acXEntry)))
			ada, otto := w.adminAndPlain(t)
			s := otto
			if c.admin {
				s = ada
			}

			_, root := w.acView(t, s)
			if got := acSectionTexts(t, acTexts(root)); !slices.Equal(got, c.want) {
				t.Errorf("sections: got %q, expected %q", got, c.want)
			}
		})
	}
}

// TestAccountSectionsSameAnchorOrder is T7b: X, then Y inserted after „Konto“ give Konto, X, Y, Einstellungen, within
// one function and in two registrations. Counter row: Y alone gives Konto, Y, Einstellungen.
func TestAccountSectionsSameAnchorOrder(t *testing.T) {
	x := acSection(acX, acXTitle, acXEntry)
	y := acSection(acY, acYTitle, acYEntry)
	withX := []string{acXTitle, acXEntry}
	withY := []string{acYTitle, acYEntry}
	for _, c := range []struct {
		name string
		fns  []uiuser.AccountSections
		want []string
	}{
		{"one function", []uiuser.AccountSections{func(wnd core.Window, sections []uiuser.AccountSectionModel) []uiuser.AccountSectionModel {
			return uiuser.InsertAccountSection(uiuser.InsertAccountSection(sections, uiuser.SectionAccount, x), uiuser.SectionAccount, y)
		}}, acCat(acNago, withX, withY, acNagoAdmin)},
		{"two registrations", []uiuser.AccountSections{acInsert(uiuser.SectionAccount, x), acInsert(uiuser.SectionAccount, y)}, acCat(acNago, withX, withY, acNagoAdmin)},
		{"counter row: Y alone", []uiuser.AccountSections{acInsert(uiuser.SectionAccount, y)}, acCat(acNago, withY, acNagoAdmin)},
	} {
		t.Run(c.name, func(t *testing.T) {
			w := newACAPIWorld(t, false, c.fns...)
			ada, _ := w.adminAndPlain(t)
			_, root := w.acView(t, ada)
			if got := acSectionTexts(t, acTexts(root)); !slices.Equal(got, c.want) {
				t.Errorf("sections: got %q, expected %q", got, c.want)
			}
		})
	}
}

// acSectionCount returns the visible children of the nearest common ancestor of nago's two sections, i.e. the
// amount of sections the stack of sections lays out with its gap.
func acSectionCount(t *testing.T, root proto.Component) int {
	t.Helper()
	var chains [][]proto.Component
	for _, n := range acNodes(root) {
		if id := acID(n.Component); id == string(uiuser.SectionAccount) || id == string(uiuser.SectionSettings) {
			chains = append(chains, n.Parents)
		}
	}

	if len(chains) != 2 {
		t.Fatalf("premise: %d of nago's sections, expected „Konto“ and „Einstellungen“", len(chains))
	}

	var lca proto.Component
	for i := 0; i < len(chains[0]) && i < len(chains[1]) && chains[0][i] == chains[1][i]; i++ {
		lca = chains[0][i]
	}

	if lca == nil {
		t.Fatalf("premise: nago's sections have no common ancestor")
	}

	return len(acChildren(lca))
}

// TestAccountSectionsEmpty is T8: a section without entries is not shown, neither its title nor its id nor an
// element in the stack of the sections, which would add the gap. Counter row: the same section with one entry shows
// title and entry and adds one element.
func TestAccountSectionsEmpty(t *testing.T) {
	base := func() int {
		w := newACAPIWorld(t, false)
		ada, _ := w.adminAndPlain(t)
		_, root := w.acView(t, ada)
		return acSectionCount(t, root)
	}()

	if base != 2 {
		t.Fatalf("premise: the stack of the sections holds %d elements without registration, expected 2", base)
	}

	empty := uiuser.AccountSectionModel{ID: acX, Title: acXTitle, Entries: []uiuser.AccountEntryModel{}}
	for _, c := range []struct {
		name    string
		section uiuser.AccountSectionModel
		shown   bool
	}{
		{"no entries (nil)", acSection(acX, acXTitle, ""), false},
		{"no entries (empty)", empty, false},
		{"counter row: one entry", acSection(acX, acXTitle, acXEntry), true},
	} {
		t.Run(c.name, func(t *testing.T) {
			w := newACAPIWorld(t, false, acInsert(uiuser.SectionAccount, c.section))
			ada, _ := w.adminAndPlain(t)
			_, root := w.acView(t, ada)
			texts := acTexts(root)
			want, count, ids := acCat(acNago, acNagoAdmin), base, []string{string(uiuser.SectionAccount), string(uiuser.SectionSettings)}
			if c.shown {
				want, count, ids = acCat(acNago, []string{acXTitle, acXEntry}, acNagoAdmin), base+1, []string{string(uiuser.SectionAccount), string(acX), string(uiuser.SectionSettings)}
			}

			if got := acSectionTexts(t, texts); !slices.Equal(got, want) {
				t.Errorf("sections: got %q, expected %q", got, want)
			}

			if got := acIDs(root, acIDsKnown...); !slices.Equal(got, ids) {
				t.Errorf("ids of the sections %q, expected %q", got, ids)
			}

			if got := acSectionCount(t, root); got != count {
				t.Errorf("the stack of the sections holds %d elements, expected %d", got, count)
			}
		})
	}
}

// TestAccountDialog is T10: AccountDialog with presented false puts no dialog into the tree; with presented true it
// is the dialog „Nutzerkonto“ with one button to close it, also with options, which place it; a click on the button
// sets presented false and the dialog is gone. Counter row for the options: without them, the dialog is not placed
// at the lower left.
func TestAccountDialog(t *testing.T) {
	w := newACAPIWorld(t, false)
	ada, _ := w.adminAndPlain(t)
	for _, c := range []struct {
		name    string
		route   core.NavigationPath
		placed  bool
		padding proto.Length
	}{
		{"with alert.Alignment and alert.ModalPadding", acRouteOwnPlaced, true, proto.Length(ui.L24)},
		{"counter row: without options", acRouteOwn, false, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			win, modal := w.acOwnDialog(t, ada, c.route)
			acOrder(t, "dialog", acTexts(modal), acDialogTitle, acAccount, acSignOut)

			box, ok := modal.(*proto.Modal).Content.(*proto.Box)
			if !ok || len(box.Children) != 1 {
				t.Fatalf("premise: the dialog is no box with one child: %T", modal.(*proto.Modal).Content)
			}

			if got := box.Children[0].Alignment == proto.Alignment(ui.BottomLeading); got != c.placed {
				t.Errorf("placed at the lower left %v, expected %v (alignment %v)", got, c.placed, box.Children[0].Alignment)
			}

			if box.Padding.Top != c.padding {
				t.Errorf("padding of the modal %+v, expected %q", box.Padding, c.padding)
			}

			closers := win.Find(nagotest.Type[*proto.Modal]()).FindAll(nagotest.Label(acCloseLabel))
			if closers.Len() != 1 {
				t.Fatalf("%d buttons %q in the dialog, expected one", closers.Len(), acCloseLabel)
			}

			win.Click(closers.Find(nagotest.Where("with action", func(n nagotest.Node) bool { return acPtr(n.Component, "Action") != 0 })))
			win.Find(nagotest.Text("presented=false"))
			if n := win.FindAll(nagotest.Type[*proto.Modal]()).Len(); n != 0 {
				t.Errorf("after the click on %q: %d dialogs, expected none", acCloseLabel, n)
			}
		})
	}
}

// acModels builds sections with entries; each spec is the id of a section followed by the ids of its entries.
func acModels(specs ...[]string) []uiuser.AccountSectionModel {
	var res []uiuser.AccountSectionModel
	for _, s := range specs {
		m := uiuser.AccountSectionModel{ID: uiuser.AccountSectionID(s[0]), Title: "title " + s[0]}
		for _, e := range s[1:] {
			m.Entries = append(m.Entries, uiuser.AccountEntryModel{ID: uiuser.AccountEntryID(e), View: ui.Text(e)})
		}

		res = append(res, m)
	}

	return res
}

// acShape returns per section with entries its id and the ids of its entries.
func acShape(sections []uiuser.AccountSectionModel) []string {
	var res []string
	for _, s := range sections {
		if len(s.Entries) == 0 {
			continue
		}

		x := string(s.ID) + ":"
		for _, e := range s.Entries {
			x += " " + string(e.ID)
		}

		res = append(res, x)
	}

	return res
}

// TestInsertAccountSection: the helper of T2, T7 and T7b as a unit. A section inserted after an anchor stands after
// it and after the sections which follow it up to the next of nago's sections; an unknown anchor puts it last. The
// inserted section keeps title and entries.
func TestInsertAccountSection(t *testing.T) {
	a, s := string(uiuser.SectionAccount), string(uiuser.SectionSettings)
	for _, c := range []struct {
		name  string
		in    []string
		after uiuser.AccountSectionID
		want  []string
	}{
		{"after Konto", []string{a, s}, uiuser.SectionAccount, []string{a, "z", s}},
		{"counter row: after Einstellungen", []string{a, s}, uiuser.SectionSettings, []string{a, s, "z"}},
		{"after Konto behind one inserted there before", []string{a, "x", s}, uiuser.SectionAccount, []string{a, "x", "z", s}},
		{"after Konto behind two inserted there before", []string{a, "x", "y", s}, uiuser.SectionAccount, []string{a, "x", "y", "z", s}},
		{"after Konto without a further section of nago", []string{a, "x"}, uiuser.SectionAccount, []string{a, "x", "z"}},
		{"after Einstellungen behind one inserted there before", []string{a, s, "x"}, uiuser.SectionSettings, []string{a, s, "x", "z"}},
		{"after an inserted section", []string{a, "x", s}, "x", []string{a, "x", "z", s}},
		{"after an inserted section behind the one following it", []string{a, "x", "y", s}, "x", []string{a, "x", "y", "z", s}},
		{"unknown anchor", []string{a, s}, acUnknown, []string{a, s, "z"}},
		{"no sections", nil, uiuser.SectionAccount, []string{"z"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			var specs [][]string
			for _, id := range c.in {
				specs = append(specs, []string{id, id + "-e"})
			}

			z := uiuser.AccountSectionModel{ID: "z", Title: "Z", Entries: []uiuser.AccountEntryModel{{ID: "z-e", View: ui.Text("z")}}}
			got := uiuser.InsertAccountSection(acModels(specs...), c.after, z)
			var ids []string
			for _, m := range got {
				ids = append(ids, string(m.ID))
				if m.ID == "z" && (m.Title != "Z" || len(m.Entries) != 1 || m.Entries[0].ID != "z-e") {
					t.Errorf("the inserted section changed: %+v", m)
				}
			}

			if !slices.Equal(ids, c.want) {
				t.Errorf("got %q, expected %q", ids, c.want)
			}
		})
	}
}

// TestRemoveAccountEntry: the helper of T3 as a unit. The entry is gone wherever it was, every other entry stays
// in its order, and a section left without entries is not shown (the shape lists sections with entries only).
// Counter row: an unknown id changes nothing.
func TestRemoveAccountEntry(t *testing.T) {
	a, s := string(uiuser.SectionAccount), string(uiuser.SectionSettings)
	nago := func() []uiuser.AccountSectionModel {
		return acModels(
			[]string{a, string(uiuser.EntryProfile), string(uiuser.EntryTheme)},
			[]string{"x", "x-1", "x-2"},
			[]string{s, string(uiuser.EntryAdminCenter)},
		)
	}

	for _, c := range []struct {
		name string
		id   uiuser.AccountEntryID
		want []string
	}{
		{"Verwalten", uiuser.EntryAdminCenter, []string{a + ": nago-account-profile nago-account-theme", "x: x-1 x-2"}},
		{"Farbschema wechseln", uiuser.EntryTheme, []string{a + ": nago-account-profile", "x: x-1 x-2", s + ": nago-settings-admin-center"}},
		{"an entry of an application", "x-1", []string{a + ": nago-account-profile nago-account-theme", "x: x-2", s + ": nago-settings-admin-center"}},
		{"counter row: unknown entry", "unknown", []string{a + ": nago-account-profile nago-account-theme", "x: x-1 x-2", s + ": nago-settings-admin-center"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := acShape(uiuser.RemoveAccountEntry(nago(), c.id)); !slices.Equal(got, c.want) {
				t.Errorf("got %q, expected %q", got, c.want)
			}
		})
	}
}

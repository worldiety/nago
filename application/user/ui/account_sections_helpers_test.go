// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser_test

// The world and the instruments of the acceptance tests of the account view and its sections (T1 to T10 and T7b of
// the concept „Eigene Abschnitte im Nutzerkonto von nago“, wokoda requirements/konzept-nago-nutzerkonto.md §5).
// This file and the tests T1 and T9 use only what nago had before the sections became data, so that they compile
// and run at the former state as well.
//
// ASSUMPTIONS of the tests (only here):
//
//	AC-1 The avatar of nago's scaffold is the only entry of the scaffold menu without a title, but with an action
//	     (application/cfg_scaffold.go, Decorator). A click on it is a FunctionCallRequested with its action, as the
//	     frontend sends it (TopLevelMenuEntry.vue).
//	AC-2 A person sees the texts of the rendered tree, except those of invisible components and of the choices of a
//	     menu, which only open on a further click (e.g. the color schemes below „Farbschema wechseln“).

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/user"
	uiuser "go.wdy.de/nago/application/user/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

// Routes and texts of the world.
const (
	acRouteScaffold = "ac-scaffold" // a page in nago's scaffold
	acRouteView     = "ac-view"     // a page, which is only uiuser.AccountView, without a frame
	acScaffoldText  = "SEITE IM SCAFFOLD"
	acDialogTitle   = "Nutzerkonto"
	acSignOut       = "Abmelden"
)

// The texts of nago's sections and entries, as nago shows them in German.
const (
	acAccount   = "Konto"
	acProfile   = "Profil verwalten"
	acTheme     = "Farbschema wechseln"
	acSettings  = "Einstellungen"
	acAdminCntr = "Verwalten"
)

// acWorld is a nago application with the standard systems, nago's scaffold as decorator and the pages of the tests.
type acWorld struct {
	app   *nagotest.App
	cfg   *application.Configurator
	users user.UseCases
	mgmt  application.AdminManagement
	seq   int
}

// newACWorld configures the world; before runs first, after at the end of the configuration, e.g. to register
// sections or further pages. Both may be nil.
func newACWorld(t *testing.T, before, after func(cfg *application.Configurator)) *acWorld {
	t.Helper()
	w := &acWorld{}
	w.app = nagotest.New(t, func(cfg *application.Configurator) {
		w.cfg = cfg
		cfg.SetApplicationID("de.worldiety.accountsectionstest")
		if before != nil {
			before(cfg)
		}

		if err := cfg.StandardSystems(); err != nil {
			t.Fatal(err)
		}

		if _, err := cfg.SettingsManagement(); err != nil {
			t.Fatal(err)
		}

		cfg.SetDecorator(cfg.NewScaffold().Decorator())
		cfg.RootView(acRouteScaffold, cfg.DecorateRootView(func(wnd core.Window) core.View {
			return ui.Text(acScaffoldText)
		}))
		cfg.RootView(acRouteView, func(wnd core.Window) core.View {
			return uiuser.AccountView(wnd)
		})

		if after != nil {
			after(cfg)
		}
	})

	w.users = std.Must(w.cfg.UserManagement()).UseCases
	w.mgmt = std.Must(w.cfg.AdminManagement())
	return w
}

// person creates a verified account with exactly the given permissions and returns its subject.
func (w *acWorld) person(t *testing.T, first, last string, perms ...permission.ID) user.Subject {
	t.Helper()
	w.seq++
	mail := fmt.Sprintf("%s.%s.%d@example.com", strings.ToLower(first), strings.ToLower(last), w.seq)
	u, err := w.users.Create(user.SU(), user.ShortRegistrationUser{Firstname: first, Lastname: last, Email: user.Email(mail), Password: "Sup3r-Geheim!2026", PasswordRepeated: "Sup3r-Geheim!2026", Verified: true})
	if err != nil {
		t.Fatal(err)
	}

	if len(perms) > 0 {
		if err := w.users.UpdateOtherPermissions(user.SU(), u.ID, perms); err != nil {
			t.Fatal(err)
		}
	}

	return std.Must(w.users.SubjectFromUser(user.SU(), u.ID)).Unwrap()
}

// adminAndPlain returns a person, for whom nago's admin center has entries (admin.HasEntries), and one without.
// Both premises are checked.
func (w *acWorld) adminAndPlain(t *testing.T) (ada, otto user.Subject) {
	t.Helper()
	ada = w.person(t, "Ada", "Admin", user.PermFindAll)
	otto = w.person(t, "Otto", "Ohne")
	if !w.mgmt.HasEntries(ada) {
		t.Fatalf("premise: admin.HasEntries is false for a person with %s", user.PermFindAll)
	}

	if w.mgmt.HasEntries(otto) {
		t.Fatalf("premise: admin.HasEntries is true for a person without permissions")
	}

	return ada, otto
}

var acRID atomic.Int64

// acDispatch clicks the callback ptr like the frontend and waits until the window is idle.
func acDispatch(t *testing.T, win *nagotest.Window, ptr proto.Ptr) {
	t.Helper()
	if ptr == 0 {
		t.Fatalf("premise: a click without callback")
	}

	if err := win.Scope().Dispatch(&proto.FunctionCallRequested{Ptr: ptr, RID: proto.RID(1_000_000 + acRID.Add(1))}); err != nil {
		t.Fatalf("click: %v", err)
	}

	win.Settle()
}

// acAvatar returns the action of the avatar of nago's scaffold (AC-1).
func acAvatar(t *testing.T, win *nagotest.Window) proto.Ptr {
	t.Helper()
	scaffold := win.Find(nagotest.Type[*proto.Scaffold]()).Node().Component.(*proto.Scaffold)
	var ptrs []proto.Ptr
	for _, e := range scaffold.Menu {
		if strings.TrimSpace(string(e.Title)) == "" && e.Action != 0 && e.Icon != nil && !e.Icon.IsZero() {
			ptrs = append(ptrs, e.Action)
		}
	}

	if len(ptrs) != 1 {
		t.Fatalf("premise: %d entries of the scaffold menu without title, but with icon and action, expected the avatar only", len(ptrs))
	}

	return ptrs[0]
}

// acScaffoldDialog opens the page in nago's scaffold as s and clicks the avatar. It returns the window and the dialog.
func (w *acWorld) acScaffoldDialog(t *testing.T, s auth.Subject, opts ...nagotest.OpenOption) (*nagotest.Window, proto.Component) {
	t.Helper()
	win := w.app.Open(t, s, acRouteScaffold, opts...)
	win.Find(nagotest.Text(acScaffoldText))
	if n := win.FindAll(nagotest.Type[*proto.Modal]()).Len(); n != 0 {
		t.Fatalf("premise: %d dialogs before the click on the avatar", n)
	}

	acDispatch(t, win, acAvatar(t, win))
	modal := win.Find(nagotest.Type[*proto.Modal]()).Node().Component
	if texts := acTexts(modal); len(texts) == 0 || texts[0] != acDialogTitle {
		t.Fatalf("premise: the avatar opens no dialog %q: %s", acDialogTitle, acJoin(texts))
	}

	return win, modal
}

// acView opens the page, which is only the account view, as s and returns the window and its tree.
func (w *acWorld) acView(t *testing.T, s auth.Subject) (*nagotest.Window, proto.Component) {
	t.Helper()
	win := w.app.Open(t, s, acRouteView)
	win.Find(nagotest.Text(acSignOut))
	return win, win.Tree()
}

// acNodes returns the components a person sees below root in document order (AC-2), together with their ancestors.
func acNodes(root proto.Component) []nagotest.Node {
	var res []nagotest.Node
	var visit func(c proto.Component, parents []proto.Component)
	var walk func(v reflect.Value, parents []proto.Component)
	visit = func(c proto.Component, parents []proto.Component) {
		if acIsNil(c) || acBool(c, "Invisible") {
			return
		}

		res = append(res, nagotest.Node{Component: c, Parents: parents})
		ps := append(parents[:len(parents):len(parents)], c)
		if m, ok := c.(*proto.Menu); ok {
			// the choices of a menu open on a further click
			visit(m.Anchor, ps)
			return
		}

		walk(reflect.ValueOf(c), ps)
	}

	walk = func(v reflect.Value, parents []proto.Component) {
		switch v.Kind() {
		case reflect.Interface, reflect.Pointer:
			if v.IsNil() {
				return
			}

			walk(v.Elem(), parents)
		case reflect.Struct:
			for i := range v.NumField() {
				f := v.Field(i)
				if !f.CanInterface() {
					continue
				}

				if (f.Kind() == reflect.Interface || f.Kind() == reflect.Pointer) && !f.IsNil() {
					if c, ok := f.Interface().(proto.Component); ok {
						visit(c, parents)
						continue
					}
				}

				walk(f, parents)
			}
		case reflect.Slice, reflect.Array:
			if v.Type().Elem().Kind() == reflect.Uint8 {
				return
			}

			for i := range v.Len() {
				e := v.Index(i)
				if e.Kind() == reflect.Interface && !e.IsNil() {
					if c, ok := e.Interface().(proto.Component); ok {
						visit(c, parents)
						continue
					}
				}

				walk(e, parents)
			}
		default:
			// scalars, strings and maps contain no components
		}
	}

	visit(root, nil)
	return res
}

// acTexts returns the texts a person sees below root in document order (AC-2).
func acTexts(root proto.Component) []string {
	var res []string
	for _, n := range acNodes(root) {
		if tv, ok := n.Component.(*proto.TextView); ok && tv.Value != "" {
			res = append(res, string(tv.Value))
		}
	}

	return res
}

// acSectionTexts returns the texts from nago's „Konto“ up to „Abmelden“, both included, i.e. what the sections
// show between name and mail above and the sign-out below.
func acSectionTexts(t *testing.T, texts []string) []string {
	t.Helper()
	from := slices.Index(texts, acAccount)
	to := slices.Index(texts, acSignOut)
	if from < 0 || to < from {
		t.Fatalf("premise: no %q before %q: %s", acAccount, acSignOut, acJoin(texts))
	}

	return slices.Clone(texts[from : to+1])
}

// acOrder asserts, that each of want is exactly one of texts, and that they occur in this order.
func acOrder(t *testing.T, what string, texts []string, want ...string) {
	t.Helper()
	last := -1
	for _, x := range want {
		var at []int
		for i, s := range texts {
			if s == x {
				at = append(at, i)
			}
		}

		if len(at) != 1 {
			t.Errorf("%s: %q occurs %d times, expected once: %s", what, x, len(at), acJoin(texts))
			return
		}

		if at[0] < last {
			t.Errorf("%s: %q is not in the order %q: %s", what, x, want, acJoin(texts))
			return
		}

		last = at[0]
	}
}

// acAbsent asserts, that none of xs is one of texts.
func acAbsent(t *testing.T, what string, texts []string, xs ...string) {
	t.Helper()
	for _, x := range xs {
		if slices.Contains(texts, x) {
			t.Errorf("%s: %q is shown: %s", what, x, acJoin(texts))
		}
	}
}

func acJoin(texts []string) string { return strings.Join(texts, " | ") }

// acID returns the field Id of c, if any.
func acID(c proto.Component) string {
	v := reflect.Indirect(reflect.ValueOf(c))
	if v.Kind() != reflect.Struct {
		return ""
	}

	f := v.FieldByName("Id")
	if !f.IsValid() || f.Kind() != reflect.String {
		return ""
	}

	return f.String()
}

// acIDs returns the ids of the visible components below root, which are one of known, in document order.
func acIDs(root proto.Component, known ...string) []string {
	var res []string
	for _, n := range acNodes(root) {
		if id := acID(n.Component); id != "" && slices.Contains(known, id) {
			res = append(res, id)
		}
	}

	return res
}

// acTextWithin reports, whether the text x below root has an ancestor with the id.
func acTextWithin(root proto.Component, x, id string) bool {
	for _, n := range acNodes(root) {
		if tv, ok := n.Component.(*proto.TextView); ok && string(tv.Value) == x {
			if slices.ContainsFunc(n.Parents, func(c proto.Component) bool { return acID(c) == id }) {
				return true
			}
		}
	}

	return false
}

func acIsNil(c proto.Component) bool {
	if c == nil {
		return true
	}

	v := reflect.ValueOf(c)
	return v.Kind() == reflect.Pointer && v.IsNil()
}

func acBool(c proto.Component, name string) bool {
	v := reflect.Indirect(reflect.ValueOf(c))
	if v.Kind() != reflect.Struct {
		return false
	}

	f := v.FieldByName(name)
	return f.IsValid() && f.Kind() == reflect.Bool && f.Bool()
}

// acChildren returns the visible children of the field Children of c.
func acChildren(c proto.Component) []proto.Component {
	v := reflect.Indirect(reflect.ValueOf(c))
	if v.Kind() != reflect.Struct {
		return nil
	}

	f := v.FieldByName("Children")
	if !f.IsValid() || f.Kind() != reflect.Slice {
		return nil
	}

	var res []proto.Component
	for i := range f.Len() {
		e := f.Index(i)
		if e.Kind() == reflect.Struct {
			// e.g. proto.AlignedComponent
			e = e.FieldByName("Component")
		}

		if !e.IsValid() || (e.Kind() == reflect.Interface && e.IsNil()) {
			continue
		}

		if child, ok := e.Interface().(proto.Component); ok && !acIsNil(child) && !child.IsZero() && !acBool(child, "Invisible") {
			res = append(res, child)
		}
	}

	return res
}

// acSignOutButton returns the node of the button „Abmelden“ below root: the nearest component with an action,
// which carries the text, and its ancestors.
func acSignOutButton(t *testing.T, root proto.Component) nagotest.Node {
	t.Helper()
	var res []nagotest.Node
	for _, n := range acNodes(root) {
		tv, ok := n.Component.(*proto.TextView)
		if !ok || string(tv.Value) != acSignOut {
			continue
		}

		for i := len(n.Parents) - 1; i >= 0; i-- {
			if acPtr(n.Parents[i], "Action") != 0 {
				res = append(res, nagotest.Node{Component: n.Parents[i], Parents: n.Parents[:i]})
				break
			}
		}
	}

	if len(res) != 1 {
		t.Fatalf("premise: %d buttons %q, expected one", len(res), acSignOut)
	}

	return res[0]
}

func acPtr(c proto.Component, name string) proto.Ptr {
	v := reflect.Indirect(reflect.ValueOf(c))
	if v.Kind() != reflect.Struct {
		return 0
	}

	f := v.FieldByName(name)
	if !f.IsValid() || f.Type() != reflect.TypeFor[proto.Ptr]() {
		return 0
	}

	return proto.Ptr(f.Uint())
}

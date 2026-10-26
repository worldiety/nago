// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application_test

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/admin"
	cfgcms "go.wdy.de/nago/application/cms/cfg"
	"go.wdy.de/nago/application/ent"
	cfgent "go.wdy.de/nago/application/ent/cfg"
	"go.wdy.de/nago/application/group"
	cfginspector "go.wdy.de/nago/application/inspector/cfg"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	cfgscheduler "go.wdy.de/nago/application/scheduler/cfg"
	cfgtheme "go.wdy.de/nago/application/theme/cfg"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"golang.org/x/text/language"
)

// visibilityEntity is distinct for each phantom type P, because the ent module registers itself per Go type.
type visibilityEntity[P any] struct {
	ID string `json:"id"`
}

func (e visibilityEntity[P]) Identity() string {
	return e.ID
}

func (e visibilityEntity[P]) WithIdentity(id string) visibilityEntity[P] {
	e.ID = id
	return e
}

func enableVisibilityEntity[P any](cfg *application.Configurator, n int) (ent.Permissions, error) {
	prefix := permission.ID(fmt.Sprintf("de.worldiety.adminvisibility.entity%02d", n))
	mod, err := cfgent.Enable[visibilityEntity[P], string](cfg, prefix, fmt.Sprintf("Entity %02d", n), cfgent.Options[visibilityEntity[P], string]{})
	return mod.Permissions, err
}

var visibilityEntities = []func(cfg *application.Configurator, n int) (ent.Permissions, error){
	enableVisibilityEntity[[1]byte], enableVisibilityEntity[[2]byte], enableVisibilityEntity[[3]byte], enableVisibilityEntity[[4]byte],
	enableVisibilityEntity[[5]byte], enableVisibilityEntity[[6]byte], enableVisibilityEntity[[7]byte], enableVisibilityEntity[[8]byte],
	enableVisibilityEntity[[9]byte], enableVisibilityEntity[[10]byte], enableVisibilityEntity[[11]byte], enableVisibilityEntity[[12]byte],
	enableVisibilityEntity[[13]byte], enableVisibilityEntity[[14]byte], enableVisibilityEntity[[15]byte], enableVisibilityEntity[[16]byte],
	enableVisibilityEntity[[17]byte], enableVisibilityEntity[[18]byte], enableVisibilityEntity[[19]byte], enableVisibilityEntity[[20]byte],
	enableVisibilityEntity[[21]byte], enableVisibilityEntity[[22]byte], enableVisibilityEntity[[23]byte], enableVisibilityEntity[[24]byte],
	enableVisibilityEntity[[25]byte], enableVisibilityEntity[[26]byte], enableVisibilityEntity[[27]byte], enableVisibilityEntity[[28]byte],
	enableVisibilityEntity[[29]byte], enableVisibilityEntity[[30]byte],
}

const (
	adminVisibilityRoleAdmin  role.ID = "de.worldiety.adminvisibility.admin"
	adminVisibilityRoleEditor role.ID = "de.worldiety.adminvisibility.editor"
	adminVisibilityRoleUser   role.ID = "de.worldiety.adminvisibility.user"
)

type adminVisibilityProfile struct {
	name    string
	subject auth.Subject
	visible bool
}

type adminVisibilitySubject struct {
	name   string
	roles  []role.ID
	groups []group.ID
	perms  []permission.ID
	// unverified users are not valid, even if they have permissions
	unverified bool
}

type adminVisibilityEnv struct {
	cfg        *application.Configurator
	query      admin.QueryGroups
	profiles   []adminVisibilityProfile
	newSubject func(s adminVisibilitySubject) (auth.Subject, error)
}

// The permissions are declared globally and cannot be declared twice, thus all tests share a single configuration,
// also across -count runs.
var adminVisibility struct {
	once sync.Once
	env  *adminVisibilityEnv
	err  error
}

func adminVisibilitySetup(tb testing.TB) *adminVisibilityEnv {
	tb.Helper()

	adminVisibility.once.Do(func() {
		adminVisibility.env, adminVisibility.err = newAdminVisibilityEnv()
	})

	if adminVisibility.err != nil {
		tb.Fatal(adminVisibility.err)
	}

	return adminVisibility.env
}

// newAdminVisibilityEnv configures an application with the standard systems, a few modules and 30 entities.
func newAdminVisibilityEnv() (*adminVisibilityEnv, error) {
	dir, err := os.MkdirTemp("", "nago-adminvisibility-*")
	if err != nil {
		return nil, err
	}

	cfg := application.NewConfigurator()
	cfg.SetDataDir(dir)
	cfg.SetApplicationID("de.worldiety.adminvisibility")

	if err := cfg.StandardSystems(); err != nil {
		return nil, err
	}

	for _, enable := range []func() error{
		func() error { _, err := cfg.RoleManagement(); return err },
		func() error { _, err := cfg.GroupManagement(); return err },
		func() error { _, err := cfg.PermissionManagement(); return err },
		func() error { _, err := cfg.SettingsManagement(); return err },
		func() error { _, err := cfgscheduler.Enable(cfg); return err },
		func() error { _, err := cfginspector.Enable(cfg); return err },
		func() error { _, err := cfgtheme.Enable(cfg); return err },
		func() error { _, err := cfgcms.Enable(cfg); return err },
	} {
		if err := enable(); err != nil {
			return nil, err
		}
	}

	var entityPerms []ent.Permissions
	for i, enable := range visibilityEntities {
		perms, err := enable(cfg, i+1)
		if err != nil {
			return nil, err
		}

		entityPerms = append(entityPerms, perms)
	}

	// like the user circles, a group may show cards depending on the data of the subject instead of its permissions
	var circleMember user.ID
	cfg.AddAdminCenterGroup(func(subject auth.Subject) admin.Group {
		if subject.ID() != circleMember {
			return admin.Group{}
		}

		return admin.Group{Title: "Mein Kreis", Entries: []admin.Card{{Title: "Kreis", Target: "circle"}}}
	})

	var allPerms []permission.ID
	for perm := range permission.All() {
		allPerms = append(allPerms, perm.ID)
	}

	// a domain user may work with all entities, but none of these use cases unlocks a card in the admin center
	var domainPerms []permission.ID
	for _, perms := range entityPerms {
		domainPerms = append(domainPerms, perms.Create, perms.FindByID, perms.Update)
	}

	// an editor may manage exactly one entity, which is the last card an early exit finds
	last := entityPerms[len(entityPerms)-1]

	for id, perms := range map[role.ID][]permission.ID{
		adminVisibilityRoleAdmin:  allPerms,
		adminVisibilityRoleEditor: {last.FindAll, last.FindByID, last.Update},
		adminVisibilityRoleUser:   domainPerms,
	} {
		if err := cfg.DeclareSystemRole(role.Role{ID: id, Name: string(id)}, perms...); err != nil {
			return nil, err
		}
	}

	groups, err := cfg.GroupManagement()
	if err != nil {
		return nil, err
	}

	team, err := groups.UseCases.Create(user.SU(), group.Group{Name: "Team"})
	if err != nil {
		return nil, err
	}

	ops, err := groups.UseCases.Create(user.SU(), group.Group{Name: "Ops"})
	if err != nil {
		return nil, err
	}

	// there is no use case to grant a global permission to a group, thus write it like the rebac editor does
	rdb, err := cfg.RDB()
	if err != nil {
		return nil, err
	}

	rdb.ValidateStaticRules(false)
	err = rdb.Put(rebac.Triple{
		Source:   rebac.Entity{Namespace: group.Namespace, Instance: rebac.Instance(ops)},
		Relation: rebac.Relation(role.PermFindAll),
		Target:   rebac.Entity{Namespace: rebac.Global, Instance: rebac.AllInstances},
	})
	rdb.ValidateStaticRules(true)
	if err != nil {
		return nil, err
	}

	users, err := cfg.UserManagement()
	if err != nil {
		return nil, err
	}

	bundle, ok := i18n.Default.MatchBundle(language.German)
	if !ok {
		return nil, fmt.Errorf("no german bundle")
	}

	env := &adminVisibilityEnv{cfg: cfg}
	env.newSubject = func(s adminVisibilitySubject) (auth.Subject, error) {
		usr, err := users.UseCases.Create(user.SU(), user.ShortRegistrationUser{
			Firstname:        s.name,
			Lastname:         "Visibility",
			Email:            user.Email(s.name + "@example.com"),
			Password:         "Sup3r-Geheim!2026",
			PasswordRepeated: "Sup3r-Geheim!2026",
			Verified:         !s.unverified,
		})
		if err != nil {
			return nil, err
		}

		if err := users.UseCases.UpdateOtherRoles(user.SU(), usr.ID, s.roles); err != nil {
			return nil, err
		}

		if err := users.UseCases.UpdateOtherGroups(user.SU(), usr.ID, s.groups); err != nil {
			return nil, err
		}

		if err := users.UseCases.UpdateOtherPermissions(user.SU(), usr.ID, s.perms); err != nil {
			return nil, err
		}

		optSubject, err := users.UseCases.SubjectFromUser(user.SU(), usr.ID)
		if err != nil {
			return nil, err
		}

		if optSubject.IsNone() {
			return nil, fmt.Errorf("no subject for %s", usr.ID)
		}

		subject := optSubject.Unwrap()
		if setter, ok := subject.(interface{ SetBundle(*i18n.Bundle) }); ok {
			setter.SetBundle(bundle)
		}

		return subject, nil
	}

	for _, p := range []struct {
		subject adminVisibilitySubject
		visible bool
	}{
		{adminVisibilitySubject{name: "admin", roles: []role.ID{adminVisibilityRoleAdmin}, groups: []group.ID{team}}, true},
		{adminVisibilitySubject{name: "editor", roles: []role.ID{adminVisibilityRoleEditor}, groups: []group.ID{team}}, true},
		{adminVisibilitySubject{name: "user", roles: []role.ID{adminVisibilityRoleUser}, groups: []group.ID{team}}, false},
		{adminVisibilitySubject{name: "direct", perms: []permission.ID{user.PermFindAll}}, true},
		{adminVisibilitySubject{name: "grouped", roles: []role.ID{adminVisibilityRoleUser}, groups: []group.ID{ops}}, true},
		{adminVisibilitySubject{name: "circle", roles: []role.ID{adminVisibilityRoleUser}}, true},
		{adminVisibilitySubject{name: "unverified", roles: []role.ID{adminVisibilityRoleAdmin}, unverified: true}, false},
	} {
		subject, err := env.newSubject(p.subject)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p.subject.name, err)
		}

		if p.subject.name == "circle" {
			circleMember = subject.ID()
		}

		env.profiles = append(env.profiles, adminVisibilityProfile{name: p.subject.name, subject: subject, visible: p.visible})
	}

	mgmt, err := cfg.AdminManagement()
	if err != nil {
		return nil, err
	}

	env.query = mgmt.QueryGroups

	return env, nil
}

type adminVisibilityVariant struct {
	name    string
	visible func(subject auth.Subject) bool
}

func adminVisibilityVariants(env *adminVisibilityEnv) []adminVisibilityVariant {
	mgmt := std.Must(env.cfg.AdminManagement())

	return []adminVisibilityVariant{
		{"query-groups", func(subject auth.Subject) bool { return len(env.query(subject, "")) > 0 }},
		{"has-entries", mgmt.HasEntries},
	}
}

// TestAdminCenterVisibilityVariantsAgree ensures that HasEntries decides like QueryGroups.
func TestAdminCenterVisibilityVariantsAgree(t *testing.T) {
	env := adminVisibilitySetup(t)
	for _, profile := range env.profiles {
		for _, variant := range adminVisibilityVariants(env) {
			if got := variant.visible(profile.subject); got != profile.visible {
				t.Errorf("%s/%s: got %v, want %v", profile.name, variant.name, got, profile.visible)
			}
		}
	}
}

// BenchmarkAdminCenterVisibility measures the decision, whether the link to the admin center is shown, for the
// typical profiles: an admin sees the first card, an editor the last one and a domain user none.
func BenchmarkAdminCenterVisibility(b *testing.B) {
	env := adminVisibilitySetup(b)
	for _, variant := range adminVisibilityVariants(env) {
		for _, profile := range env.profiles {
			if profile.name != "admin" && profile.name != "editor" && profile.name != "user" {
				continue
			}

			b.Run(variant.name+"/"+profile.name, func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					if variant.visible(profile.subject) != profile.visible {
						b.Fatal("unexpected visibility")
					}
				}
			})
		}
	}
}

// TestHasEntriesFollowsRuntimeRoleChanges ensures that a role, which a user gets while being logged in, shows the
// link to the admin center immediately.
func TestHasEntriesFollowsRuntimeRoleChanges(t *testing.T) {
	env := adminVisibilitySetup(t)
	subject, err := env.newSubject(adminVisibilitySubject{name: "promoted", roles: []role.ID{adminVisibilityRoleUser}})
	if err != nil {
		t.Fatal(err)
	}

	assertAll := func(want bool) {
		t.Helper()
		for _, variant := range adminVisibilityVariants(env) {
			if got := variant.visible(subject); got != want {
				t.Errorf("%s: got %v, want %v", variant.name, got, want)
			}
		}
	}

	assertAll(false)

	users := std.Must(env.cfg.UserManagement())
	if err := users.UseCases.UpdateOtherRoles(user.SU(), subject.ID(), []role.ID{adminVisibilityRoleUser, adminVisibilityRoleEditor}); err != nil {
		t.Fatal(err)
	}

	assertAll(true)

	if err := users.UseCases.UpdateOtherRoles(user.SU(), subject.ID(), []role.ID{adminVisibilityRoleUser}); err != nil {
		t.Fatal(err)
	}

	assertAll(false)
}

// TestHasEntriesRespectsReplacedFindAll ensures that groups, which a developer injects by replacing FindAll, are
// considered.
func TestHasEntriesRespectsReplacedFindAll(t *testing.T) {
	cfg := application.NewConfigurator()
	cfg.SetDataDir(t.TempDir())
	cfg.SetApplicationID("de.worldiety.adminmutatortest")

	if err := cfg.StandardSystems(); err != nil {
		t.Fatal(err)
	}

	cfg.WithAdminManagement(func(m *application.AdminManagement) {
		m.FindAll = func(subject auth.Subject) []admin.Group {
			return []admin.Group{{Title: "Hilfe", Entries: []admin.Card{{Title: "Für alle", Target: "help"}}}}
		}
	})

	users := std.Must(cfg.UserManagement())
	usr, err := users.UseCases.Create(user.SU(), user.ShortRegistrationUser{
		Firstname:        "Nobody",
		Lastname:         "Special",
		Email:            "nobody@example.com",
		Password:         "Sup3r-Geheim!2026",
		PasswordRepeated: "Sup3r-Geheim!2026",
		Verified:         true,
	})
	if err != nil {
		t.Fatal(err)
	}

	subject := std.Must(users.UseCases.SubjectFromUser(user.SU(), usr.ID)).Unwrap()

	if !std.Must(cfg.AdminManagement()).HasEntries(subject) {
		t.Error("the injected card is visible to everybody")
	}
}

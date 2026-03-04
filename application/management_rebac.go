// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"context"
	"iter"

	"github.com/worldiety/i18n"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"golang.org/x/text/language"
)

// RDB returns the nago ReBAC (relation-based access control) database. Even though there is a separate module,
// the rebac system is always available, and the module is only required if you want the admin user interface for it.
// The default resolvers are
//   - users which are members of a role resolve to the assigned role relations
func (c *Configurator) RDB() (*rebac.DB, error) {
	if c.rdb == nil {
		store, err := c.EntityStore("nago.rebac")
		if err != nil {
			return nil, err
		}

		db, err := rebac.NewDB(store)
		if err != nil {
			return nil, err
		}

		// automatically resolve role relations by user memberships
		db.AddResolver(rebac.NewSourceMemberResolver(user.Namespace, role.Namespace))
		db.RegisterResources(globalResources{})
		db.RegisterResources(relationResources{})

		c.rdb = db

	}

	return c.rdb, nil
}

var _ rebac.Resources = (*globalResources)(nil)

type globalResources struct {
}

func (g globalResources) Identity() rebac.Namespace {
	return rebac.Global
}

func (g globalResources) Info(bundler i18n.Bundler) rebac.NamespaceInfo {
	return rebac.NamespaceInfo{
		ID:          rebac.Global,
		Name:        "global",
		Description: "",
	}
}

func (g globalResources) All(ctx context.Context) iter.Seq2[rebac.InfoID, error] {
	return func(yield func(rebac.InfoID, error) bool) {
		yield(rebac.NewInfoID(rebac.Global, "*"), nil)
	}
}

func (g globalResources) FindByID(ctx context.Context, iid rebac.InfoID) (option.Opt[rebac.InstanceInfo], error) {
	ns, id, err := iid.Parse()
	if err != nil {
		return option.None[rebac.InstanceInfo](), err
	}

	if ns != rebac.Global {
		return option.None[rebac.InstanceInfo](), nil
	}

	if id != "*" {
		return option.None[rebac.InstanceInfo](), nil
	}

	return option.Some[rebac.InstanceInfo](rebac.InstanceInfo{
		ID:          "*",
		Namespace:   rebac.Global,
		Name:        "all (*)",
		Description: "",
	}), nil
}

func (g globalResources) Visible() bool {
	return false
}

var _ rebac.Resources = (*relationResources)(nil)

type relationResources struct {
}

func (r relationResources) Identity() rebac.Namespace {
	return rebac.Relations
}

func (r relationResources) Info(bundler i18n.Bundler) rebac.NamespaceInfo {
	return rebac.NamespaceInfo{
		ID:   rebac.Relations,
		Name: "relations",
	}
}

func (r relationResources) All(ctx context.Context) iter.Seq2[rebac.InfoID, error] {
	return func(yield func(rebac.InfoID, error) bool) {
		for id := range rebac.AllRelations {
			if !yield(rebac.NewInfoID(rebac.Relations, rebac.Instance(id)), nil) {
				return
			}
		}

		for p := range permission.All() {
			if !yield(rebac.NewInfoID(rebac.Relations, rebac.Instance(p.ID)), nil) {
				return
			}
		}
	}
}

func (r relationResources) FindByID(ctx context.Context, iid rebac.InfoID) (option.Opt[rebac.InstanceInfo], error) {
	ns, id, err := iid.Parse()
	if err != nil {
		return option.None[rebac.InstanceInfo](), err
	}

	if ns != rebac.Relations {
		return option.None[rebac.InstanceInfo](), nil
	}

	bnd, _ := i18n.BundleFrom(ctx)

	perm, ok := permission.Find(permission.ID(id))
	if ok {
		name := perm.Name
		desc := perm.Description

		if bnd != nil {
			name = bnd.Resolve(name)
			desc = bnd.Resolve(desc)
		}
		return option.Some[rebac.InstanceInfo](rebac.InstanceInfo{
			ID:          id,
			Namespace:   rebac.Relations,
			Name:        name,
			Description: desc,
		}), nil
	}

	if _, ok := relationNames[rebac.Relation(id)]; !ok {
		return option.None[rebac.InstanceInfo](), nil
	}

	var name string
	var desc string
	strName := relationNames[rebac.Relation(id)]
	strDesc := relationDescriptions[rebac.Relation(id)]
	if bnd != nil {
		name = strName.Get(bnd)
		desc = strDesc.Get(bnd)
	} else {
		name = strName.String()
		desc = strDesc.String()
	}

	return option.Some(rebac.InstanceInfo{
		Namespace:   rebac.Relations,
		ID:          id,
		Name:        name,
		Description: desc,
	}), nil
}

func (r relationResources) Visible() bool {
	return false
}

var relationNames = map[rebac.Relation]i18n.StrHnd{
	rebac.Owner:   i18n.MustString("nago.relation.owner.name", i18n.Values{language.German: "Besitzer", language.English: "Owner"}),
	rebac.Writer:  i18n.MustString("nago.relation.writer.name", i18n.Values{language.German: "Schreiber", language.English: "Writer"}),
	rebac.Deleter: i18n.MustString("nago.relation.deleter.name", i18n.Values{language.German: "Löschender", language.English: "Deleter"}),
	rebac.Member:  i18n.MustString("nago.relation.member.name", i18n.Values{language.German: "Mitglied", language.English: "Member"}),
	rebac.Viewer:  i18n.MustString("nago.relation.viewer.name", i18n.Values{language.German: "Viewer", language.English: "Viewer"}),
	rebac.Parent:  i18n.MustString("nago.relation.parent.name", i18n.Values{language.German: "Parent", language.English: "Parent"}),
}

var relationDescriptions = map[rebac.Relation]i18n.StrHnd{
	rebac.Owner:   i18n.MustString("nago.relation.owner.description", i18n.Values{language.German: "Die Quelle ist der Besitzer des zugeordneten Objektes.", language.English: "The subject is the owner of the associated object."}),
	rebac.Writer:  i18n.MustString("nago.relation.writer.description", i18n.Values{language.German: "Die Quelle kann das zugeordnete Objekt schreiben.", language.English: "The subject can write the associated object."}),
	rebac.Deleter: i18n.MustString("nago.relation.deleter.description", i18n.Values{language.German: "Die Quelle kann das zugeordnete Objekt löschen.", language.English: "The subject can delete the associated object."}),
	rebac.Member:  i18n.MustString("nago.relation.member.description", i18n.Values{language.German: "Die Quelle ist Mitglied des zugeordneten Objektes.", language.English: "The subject is a member of the associated object."}),
	rebac.Viewer:  i18n.MustString("nago.relation.viewer.description", i18n.Values{language.German: "Die Quelle kann das zugeordnete Objekt lesen.", language.English: "The subject can read the associated object."}),
	rebac.Parent:  i18n.MustString("nago.relation.parent.description", i18n.Values{language.German: "Die Quelle ist das Elternelement des zugeordneten Objekts.", language.English: "The subject is the parent of the associated object."}),
}

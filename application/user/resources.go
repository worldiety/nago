// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"context"
	"iter"

	"github.com/worldiety/i18n"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/pkg/xstrings"
	"golang.org/x/text/language"
)

var (
	StrResUsers = i18n.MustString("nago.user.resources.name", i18n.Values{language.German: "Nutzer", language.English: "Users"})
	StrResDesc  = i18n.MustString("nago.user.resources.desc", i18n.Values{language.German: "Authentifizierte und autorisierte Benutzer bzw. Konten im System."})
)

type Resources struct {
	findAll  FindAllIdentifiers
	findByID FindByID
}

func NewResources(findAll FindAllIdentifiers, findByID FindByID) Resources {
	return Resources{
		findAll:  findAll,
		findByID: findByID,
	}
}

func (r Resources) Identity() rebac.Namespace {
	return Namespace
}

func (r Resources) Info(bundler i18n.Bundler) rebac.NamespaceInfo {
	return rebac.NamespaceInfo{
		ID:          Namespace,
		Name:        StrResUsers.Get(bundler),
		Description: StrResDesc.Get(bundler),
	}
}

func (r Resources) All(ctx context.Context) iter.Seq2[rebac.Instance, error] {
	return func(yield func(rebac.Instance, error) bool) {
		for id, err := range r.findAll(SU()) {
			if !yield(rebac.Instance(id), err) {
				return
			}
		}
	}
}

func (r Resources) FindByID(ctx context.Context, id rebac.Instance) (option.Opt[rebac.InstanceInfo], error) {
	optUsr, err := r.findByID(SU(), ID(id))
	if err != nil {
		return option.None[rebac.InstanceInfo](), err
	}

	if optUsr.IsNone() {
		return option.None[rebac.InstanceInfo](), nil
	}

	usr := optUsr.Unwrap()
	return option.Some(rebac.InstanceInfo{
		ID:          id,
		Name:        usr.String(),
		Description: xstrings.Join2(",", usr.Contact.CompanyName, usr.Contact.City),
	}), nil
}

func (r Resources) Relations(ctx context.Context, id rebac.Instance) iter.Seq[rebac.Triple] {
	return func(yield func(rebac.Triple) bool) {
		for _, permission := range Permissions {
			if !yield(rebac.Triple{
				Source:   rebac.Entity{Namespace: Namespace, Instance: id},
				Relation: rebac.Relation(permission),
				Target:   rebac.Entity{Namespace: rebac.Global, Instance: rebac.AllInstances},
			}) {
				return
			}
		}
	}
}

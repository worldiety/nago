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

func (r Resources) All(ctx context.Context) iter.Seq2[rebac.InfoID, error] {
	return func(yield func(rebac.InfoID, error) bool) {
		for id, err := range r.findAll(SU()) {
			if !yield(rebac.NewInfoID(Namespace, rebac.Instance(id)), err) {
				return
			}
		}
	}
}

func (r Resources) FindByID(ctx context.Context, iid rebac.InfoID) (option.Opt[rebac.InstanceInfo], error) {
	ns, id, err := iid.Parse()
	if err != nil {
		return option.None[rebac.InstanceInfo](), err
	}

	if ns != Namespace {
		return option.None[rebac.InstanceInfo](), nil
	}

	optUsr, err := r.findByID(SU(), ID(id))
	if err != nil {
		return option.None[rebac.InstanceInfo](), err
	}

	if optUsr.IsNone() {
		return option.None[rebac.InstanceInfo](), nil
	}

	usr := optUsr.Unwrap()
	return option.Some(rebac.InstanceInfo{
		Namespace:   Namespace,
		ID:          id,
		Name:        usr.String(),
		Description: xstrings.Join2(", ", usr.Contact.CompanyName, usr.Contact.City),
	}), nil
}

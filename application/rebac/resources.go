// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rebac

import (
	"context"
	"fmt"
	"iter"

	"github.com/worldiety/i18n"
	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/data"
)

// RepositoryResources is a default implementation of the Resources interface, which just wraps entities from
// a given repository. The InstanceInfo can be customized by providing a mapper function.
// Inspect
type RepositoryResources[T data.Aggregate[ID], ID ~string] struct {
	repo       data.Repository[T, ID]
	name, desc i18n.StrHnd
	mapper     func(T) InstanceInfo
	ns         Namespace
}

func NewRepositoryResources[T data.Aggregate[ID], ID ~string](name, desc i18n.StrHnd, repo data.Repository[T, ID]) RepositoryResources[T, ID] {
	ns := Namespace(repo.Name())
	return RepositoryResources[T, ID]{
		ns:   Namespace(repo.Name()),
		repo: repo,
		name: name,
		desc: desc,
		mapper: func(t T) InstanceInfo {
			return InstanceInfo{
				Namespace: ns,
				ID:        Instance(t.Identity()),
				Name:      fmt.Sprintf("%v", t),
			}
		},
	}
}

func (r RepositoryResources[T, ID]) Map(fn func(T) InstanceInfo) RepositoryResources[T, ID] {
	r.mapper = fn
	return r
}

func (r RepositoryResources[T, ID]) Identity() Namespace {
	return r.ns
}

func (r RepositoryResources[T, ID]) Info(bundler i18n.Bundler) NamespaceInfo {
	return NamespaceInfo{
		ID:          r.Identity(),
		Name:        r.name.Get(bundler),
		Description: r.desc.Get(bundler),
	}
}

func (r RepositoryResources[T, ID]) All(ctx context.Context) iter.Seq2[InfoID, error] {
	return func(yield func(InfoID, error) bool) {
		ns := r.Identity()
		for id, err := range r.repo.Identifiers() {
			if !yield(NewInfoID(ns, Instance(id)), err) {
				return
			}
		}
	}
}

func (r RepositoryResources[T, ID]) FindByID(ctx context.Context, iid InfoID) (option.Opt[InstanceInfo], error) {
	ns, id, err := iid.Parse()
	if err != nil {
		return option.None[InstanceInfo](), err
	}

	if ns != r.ns {
		return option.None[InstanceInfo](), nil
	}

	optUsr, err := r.repo.FindByID(ID(id))
	if err != nil {
		return option.None[InstanceInfo](), err
	}

	if optUsr.IsNone() {
		return option.None[InstanceInfo](), nil
	}

	usr := optUsr.Unwrap()
	return option.Some(r.mapper(usr)), nil
}

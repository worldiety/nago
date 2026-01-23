// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"iter"
	"maps"
	"slices"
	"strings"

	"go.wdy.de/nago/pkg/xmaps"
)

type Repositories struct {
	repos map[RepositoryID]*Repository
}

func NewRepositories() *Repositories {
	return &Repositories{
		repos: make(map[RepositoryID]*Repository),
	}
}

func (r *Repositories) ByID(id RepositoryID) (*Repository, bool) {
	v, ok := r.repos[id]
	return v, ok
}

func (r *Repositories) Add(repo *Repository) {
	r.repos[repo.Identity()] = repo
}

func (r *Repositories) All() iter.Seq[*Repository] {
	return slices.Values(slices.SortedFunc(maps.Values(r.repos), func(a, b *Repository) int {
		return strings.Compare(string(a.Identity()), string(b.Identity()))
	}))
}

func (r *Repositories) Clone() *Repositories {
	return &Repositories{repos: xmaps.Clone(r.repos)}
}

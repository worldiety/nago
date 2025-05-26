// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package grant

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"iter"
	"strings"
)

// Grant sets the exact permission set for the given user. The user is removed automatically from the member
// list if permissions are empty.
type Grant func(subject auth.Subject, id ID, permissions ...permission.ID) error

// ListGrants returns the set of permissions which have been granted to the given user.
type ListGrants func(subject auth.Subject, id ID) ([]permission.ID, error)

// ListGranted returns the set of all known users who have at least a single granted permission to the given resource.
type ListGranted func(subject auth.Subject, res user.Resource) iter.Seq2[user.ID, error]

// ID is a triple composite key of <repo-name>/<resource-id>/<user-id>.
type ID string

func NewID(resource user.Resource, uid user.ID) ID {
	return ID(resource.Name + "/" + resource.ID + "/" + string(uid)) // this is faster than str buffer or sprintf
}

func (id ID) Valid() bool {
	// no heap
	return strings.Count(string(id), "/") == 2
}

func (id ID) Split() (user.Resource, user.ID) {
	// no heap
	a := strings.Index(string(id), "/")
	b := strings.LastIndex(string(id), "/")
	if a == b {
		return user.Resource{}, ""
	}

	return user.Resource{
		Name: string(id[:a]),
		ID:   string(id[a+1 : b]),
	}, user.ID(id[b+1:])
}

type Granting struct {
	ID ID `json:"id"`
}

func (g Granting) Identity() ID {
	return g.ID
}

type Repository data.Repository[Granting, ID]

type UseCases struct {
	Grant       Grant
	ListGrants  ListGrants
	ListGranted ListGranted
}

func NewUseCases(
	repo Repository,
	findUserByID user.FindByID,
	setUsrPerm user.SetResourcePermissions,
) UseCases {
	return UseCases{
		Grant:       NewGrant(repo, findUserByID, setUsrPerm),
		ListGranted: NewListGranted(repo),
		ListGrants:  NewListGrants(repo, findUserByID),
	}
}

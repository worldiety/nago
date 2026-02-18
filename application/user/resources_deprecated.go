// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"fmt"
	"iter"
	"strconv"
	"strings"

	"go.wdy.de/nago/application/permission"
)

// deprecated: use rebac api
// AddResourcePermissions appends those permissions to the users' resource lookup table which have not been already
// defined. Adding a wildcard does not automatically substitute fine-grained permissions. A wildcard takes
// logically a higher precedence.
type AddResourcePermissions func(subject AuditableUser, uid ID, resource Resource, permission ...permission.ID) error

// deprecated: use rebac api
// RemoveResourcePermissions removes those permissions from the users' resource lookup table which have been already
// defined. Removing a wildcard does not automatically remove fine-grained permissions.
type RemoveResourcePermissions func(subject AuditableUser, uid ID, resource Resource, permission ...permission.ID) error

// deprecated: use rebac api
// ListResourcePermissions returns an iterator over all defined resource permissions. Note that the
// returned order of resources is implementation-dependent and may be even random for subsequent calls.
// See also [ListGrantedPermissions] and [ListGrantedUsers].
type ListResourcePermissions func(subject AuditableUser, uid ID) iter.Seq2[ResourceWithPermissions, error]

// deprecated: use rebac api
// GrantPermissions sets the exact permission set for the given user. The user is removed automatically from the
// [GrantingIndexRepository] if permissions are empty.
type GrantPermissions func(subject AuditableUser, id GrantingKey, permissions ...permission.ID) error

// deprecated: use rebac api
// ListGrantedPermissions returns the set of permissions which have been granted to the given user and resource
// combination. It is way more efficient when iterating over all users, because the inverse index
// [GrantingIndexRepository] is used.
type ListGrantedPermissions func(subject AuditableUser, id GrantingKey) ([]permission.ID, error)

// deprecated: use rebac api
// ListGrantedUsers returns the set of all known users who have at least a single granted permission to the
// given resource. Note that due to eventual consistency, the returned sequence of user ids must not match
// the actual set of available users. This is efficient, because the inverse index
// [GrantingIndexRepository] is used.
type ListGrantedUsers func(subject AuditableUser, res Resource) iter.Seq2[ID, error]

// deprecated: the GrantingKey does not perform proper escaping and is not generally safe for serialization. Historically this was just for storage names and [data.RandIdent] which both needs no special escaping.
// GrantingKey is a triple composite key of <repo-name>/<resource-id>/<user-id>.
type GrantingKey string

// deprecated
func NewGrantingKey(resource Resource, uid ID) GrantingKey {
	return GrantingKey(resource.Name + "/" + resource.ID + "/" + string(uid)) // this is faster than str buffer or sprintf
}

func (id GrantingKey) Valid() bool {
	// no heap
	return strings.Count(string(id), "/") == 2
}

func (id GrantingKey) Split() (Resource, ID) {
	// no heap
	a := strings.Index(string(id), "/")
	b := strings.LastIndex(string(id), "/")
	if a == b {
		return Resource{}, ""
	}

	return Resource{
		Name: string(id[:a]),
		ID:   string(id[a+1 : b]),
	}, ID(id[b+1:])
}

// deprecated: use rebac api
type Granting struct {
	ID GrantingKey `json:"id"`
}

func (g Granting) Identity() GrantingKey {
	return g.ID
}

// deprecated: encoding is not quotation safe
type Resource struct {
	// ID is the string version of the root aggregate or entity key used in the named Store or Repository.
	// If ID is empty, all values from the Named Store or Repository are applicable.
	ID string

	// Name of the Store or Repository
	Name string
}

func (r Resource) MarshalText() ([]byte, error) {
	// TODO fix me: this looks awful in json and is usually totally unnecessary
	return []byte(strconv.Quote(r.Name + "/" + r.ID)), nil
}

func (r *Resource) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		r.Name = ""
		r.ID = ""
	}

	str, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}

	tokens := strings.SplitN(str, "/", 2)
	if len(tokens) != 2 {
		return fmt.Errorf("invalid json format for resource: %s", str)
	}

	r.Name = tokens[0]
	r.ID = tokens[1]
	return nil
}

// deprecated: use rebac api
type ResourceWithPermissions struct {
	Permissions iter.Seq[permission.ID]
	Resource
}

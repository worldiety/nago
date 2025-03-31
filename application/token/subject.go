// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package token

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"golang.org/x/text/language"
	"iter"
)

type subject struct {
	token Token
}

func (s *subject) HasResourcePermission(name string, id string, p permission.ID) bool {
	//TODO implement me
	panic("implement me")
}

func newSubject(token Token) *subject {
	return &subject{token: token}
}

func (s *subject) Audit(permission permission.ID) error {
	//TODO implement me
	panic("implement me")
}

func (s *subject) HasPermission(permission permission.ID) bool {
	//TODO implement me
	panic("implement me")
}

func (s *subject) Permissions() iter.Seq[permission.ID] {
	//TODO implement me
	panic("implement me")
}

func (s *subject) AuditResource(name string, id string, p permission.ID) error {
	//TODO implement me
	panic("implement me")
}

func (s *subject) ID() user.ID {
	//TODO implement me
	panic("implement me")
}

func (s *subject) Name() string {
	//TODO implement me
	panic("implement me")
}

func (s *subject) Firstname() string {
	//TODO implement me
	panic("implement me")
}

func (s *subject) Lastname() string {
	//TODO implement me
	panic("implement me")
}

func (s *subject) Email() string {
	//TODO implement me
	panic("implement me")
}

func (s *subject) Avatar() string {
	//TODO implement me
	panic("implement me")
}

func (s *subject) Roles() iter.Seq[role.ID] {
	//TODO implement me
	panic("implement me")
}

func (s *subject) HasRole(id role.ID) bool {
	//TODO implement me
	panic("implement me")
}

func (s *subject) Groups() iter.Seq[group.ID] {
	//TODO implement me
	panic("implement me")
}

func (s *subject) HasGroup(id group.ID) bool {
	//TODO implement me
	panic("implement me")
}

func (s *subject) HasLicense(id license.ID) bool {
	//TODO implement me
	panic("implement me")
}

func (s *subject) Licenses() iter.Seq[license.ID] {
	//TODO implement me
	panic("implement me")
}

func (s *subject) Valid() bool {
	//TODO implement me
	panic("implement me")
}

func (s *subject) Language() language.Tag {
	//TODO implement me
	panic("implement me")
}

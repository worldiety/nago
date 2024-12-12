package user

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"golang.org/x/text/language"
	"iter"
)

func NewSystem() System {
	return func() Subject {
		return sysUser{}
	}
}

type sysUser struct {
}

func (s sysUser) Permissions() iter.Seq[permission.ID] {
	return func(yield func(permission.ID) bool) {

	}
}

func (s sysUser) ID() ID {
	return ""
}

func (s sysUser) Name() string {
	return "SYSTEM"
}

func (s sysUser) Firstname() string {
	return ""
}

func (s sysUser) Lastname() string {
	return ""
}

func (s sysUser) Email() string {
	return "system@localhost"
}

func (s sysUser) Roles() iter.Seq[role.ID] {
	return func(yield func(id role.ID) bool) {}
}

func (s sysUser) HasRole(rid role.ID) bool {
	return true
}

func (s sysUser) Groups() iter.Seq[group.ID] {
	return func(yield func(group.ID) bool) {}
}

func (s sysUser) HasGroup(gid group.ID) bool {
	return true
}

func (s sysUser) Audit(permission permission.ID) error {
	return nil
}

func (s sysUser) HasPermission(permission permission.ID) bool {
	return true
}

func (s sysUser) HasLicense(id license.ID) bool {
	return true
}

func (s sysUser) Licenses() iter.Seq[license.ID] {
	return func(yield func(license.ID) bool) {}
}

func (s sysUser) Valid() bool {
	return true
}

func (s sysUser) Language() language.Tag {
	return language.English
}

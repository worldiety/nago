package auth

import (
	"fmt"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/xiter"
	"golang.org/x/text/language"
	"iter"
)

var _ Subject = InvalidSubject{}

type InvalidSubject struct {
	DeniedLog func(permission string)
}

func (i InvalidSubject) AuditResource(name string, id string, p permission.ID) error {
	return NotLoggedIn("invalid subject")
}

func (i InvalidSubject) Avatar() string {
	return ""
}

func (i InvalidSubject) Permissions() iter.Seq[permission.ID] {
	return func(yield func(permission.ID) bool) {

	}
}

func (i InvalidSubject) Roles() iter.Seq[role.ID] {
	return xiter.Empty[role.ID]()
}

func (i InvalidSubject) Groups() iter.Seq[group.ID] {
	return xiter.Empty[group.ID]()
}

func (i InvalidSubject) HasLicense(id license.ID) bool {
	return false
}

func (i InvalidSubject) Licenses() iter.Seq[license.ID] {
	return xiter.Empty[license.ID]()
}

func (i InvalidSubject) Firstname() string {
	return ""
}

func (i InvalidSubject) Lastname() string {
	return ""
}

func (i InvalidSubject) Email() string {
	return ""
}

func (i InvalidSubject) HasRole(rid role.ID) bool {
	return false
}

func (i InvalidSubject) HasGroup(gid group.ID) bool {
	return false
}

func (i InvalidSubject) Language() language.Tag {
	return language.German // TODO
}

func (i InvalidSubject) HasPermission(permission permission.ID) bool {
	return false
}

func (i InvalidSubject) Valid() bool {
	return false
}

func (i InvalidSubject) ID() user.ID {
	return ""
}

func (i InvalidSubject) Name() string {
	return ""
}

func (i InvalidSubject) Audit(id permission.ID) error {
	// TODO where to put audit logs and when to decide what is important? this is just insanely verbose
	if i.DeniedLog == nil {
		//slog.Info("audit", "usecase", id, "granted", false, "invalid", true)
	} else {
		//i.DeniedLog(id)
	}

	return NotLoggedIn("invalid subject")
}

type NotLoggedIn string

func (e NotLoggedIn) Error() string {
	return fmt.Sprintf("permission denied: %s", string(e))
}

func (e NotLoggedIn) PermissionDenied() bool {
	return true
}

func (e NotLoggedIn) NotLoggedIn() bool {
	return true
}

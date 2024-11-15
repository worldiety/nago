package auth

import (
	"fmt"
	"golang.org/x/text/language"
)

var _ Subject = InvalidSubject{}

type InvalidSubject struct {
	DeniedLog func(permission string)
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

func (i InvalidSubject) HasRole(rid RID) bool {
	return false
}

func (i InvalidSubject) HasGroup(gid GID) bool {
	return false
}

func (i InvalidSubject) Language() language.Tag {
	return language.German // TODO
}

func (i InvalidSubject) HasPermission(permission string) bool {
	return false
}

func (i InvalidSubject) Valid() bool {
	return false
}

func (i InvalidSubject) Roles(yield func(RID) bool) {
}

func (i InvalidSubject) Groups(yield func(GID) bool) {
}

func (i InvalidSubject) ID() UID {
	return ""
}

func (i InvalidSubject) Name() string {
	return ""
}

func (i InvalidSubject) Audit(id string) error {
	// TODO where to put audit logs and when to decide what is important? this is just insanely verbose
	if i.DeniedLog == nil {
		//slog.Info("audit", "usecase", id, "granted", false, "invalid", true)
	} else {
		//i.DeniedLog(id)
	}

	return notLoggedIn("invalid subject")
}

type notLoggedIn string

func (e notLoggedIn) Error() string {
	return fmt.Sprintf("permission denied: %s", string(e))
}

func (e notLoggedIn) PermissionDenied() bool {
	return true
}

func (e notLoggedIn) NotLoggedIn() bool {
	return true
}

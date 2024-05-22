package iam

import (
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	"strings"
)

type PermissionDeniedError string

func (e PermissionDeniedError) Error() string {
	return fmt.Sprintf("permission denied: %s", string(e))
}

var _ auth.Subject = nagoSubject{}

type nagoSubject struct {
	session     core.SessionID
	user        User
	permissions map[string]struct{}
	audit       func(session core.SessionID, usr User, permission string, granted bool)
}

func newNagoSubject(session core.SessionID, user User, permissions map[string]struct{}, audit func(session core.SessionID, usr User, permission string, granted bool)) *nagoSubject {
	return &nagoSubject{session: session, user: user, permissions: permissions, audit: audit}
}

func (n nagoSubject) HasPermission(permission string) bool {
	_, ok := n.permissions[permission]
	return ok
}

func (n nagoSubject) Valid() bool {
	return true
}

func (n nagoSubject) ID() auth.UID {
	return n.user.ID
}

func (n nagoSubject) Name() string {
	return strings.TrimSpace(fmt.Sprintf("%s %s", n.user.Firstname, n.user.Lastname))
}

func (n nagoSubject) Roles(yield func(auth.RID) bool) {
	for _, role := range n.user.Roles {
		if !yield(role) {
			return
		}
	}
}

func (n nagoSubject) EMail() Email {
	return n.user.Email
}

func (n nagoSubject) Groups(yield func(auth.GID) bool) {
	for _, role := range n.user.Groups {
		if !yield(role) {
			return
		}
	}
}

func (n nagoSubject) Audit(permission string) error {
	_, granted := n.permissions[permission]
	n.audit(n.session, n.user, permission, granted)
	if !granted {
		return PermissionDeniedError(permission)
	}

	return nil
}

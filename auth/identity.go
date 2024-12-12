package auth

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/user"
)

type Subject = user.Subject

func OneOf(subject Subject, permissions ...permission.ID) bool {
	for _, permission := range permissions {
		if subject.HasPermission(permission) {
			return true
		}
	}

	return false
}

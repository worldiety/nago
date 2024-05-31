package auth

import (
	"fmt"
	"golang.org/x/text/language"
	"log/slog"
)

var _ Subject = InvalidSubject{}

type InvalidSubject struct {
	DeniedLog func(permission string)
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
	if i.DeniedLog == nil {
		slog.Info("audit", "usecase", id, "granted", false, "invalid", true)
	} else {
		i.DeniedLog(id)
	}
	return fmt.Errorf("invalid subject: permission denied: %s", id)
}

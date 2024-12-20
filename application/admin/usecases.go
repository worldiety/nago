package admin

import (
	uibackup "go.wdy.de/nago/application/backup/ui"
	uibilling "go.wdy.de/nago/application/billing/ui"
	uigroup "go.wdy.de/nago/application/group/ui"
	uilicense "go.wdy.de/nago/application/license/ui"
	uimail "go.wdy.de/nago/application/mail/ui"
	"go.wdy.de/nago/application/permission"
	uipermission "go.wdy.de/nago/application/permission/ui"
	"go.wdy.de/nago/application/role"
	uirole "go.wdy.de/nago/application/role/ui"
	uisession "go.wdy.de/nago/application/session/ui"
	uiuser "go.wdy.de/nago/application/user/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
)

type FindAllGroups func() []Group
type QueryGroups func(subject auth.Subject, filterText string) []Group

type Pages struct {
	Mail       uimail.Pages
	Billing    uibilling.Pages
	Session    uisession.Pages
	User       uiuser.Pages
	Role       uirole.Pages
	Group      uigroup.Pages
	Permission uipermission.Pages
	License    uilicense.Pages
	Dashboard  core.NavigationPath
	Backup     uibackup.Pages
}

type Card struct {
	Title      string
	Text       string
	Target     core.NavigationPath
	Role       role.ID
	Permission permission.ID
}

type Group struct {
	Title   string
	Entries []Card
}

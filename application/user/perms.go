package user

import (
	"go.wdy.de/nago/application/permission"
)

var (
	PermCreate              = permission.Declare[Create]("nago.user.create", "Nutzer anlegen", "Träger dieser Berechtigung können neue Nutzer anlegen.")
	PermFindByID            = permission.Declare[FindByID]("nago.user.find_by_id", "Einen Nutzer per EID finden", "Träger dieser Berechtigung können die Eigenschaften anderer Nutzer anzeigen.")
	PermFindByMail          = permission.Declare[FindByMail]("nago.user.find_by_mail", "Einen Nutzer per Mail finden", "Träger dieser Berechtigung können die Eigenschaften anderer Nutzer anzeigen.")
	PermFindAll             = permission.Declare[FindAll]("nago.user.find_all", "Alle Nutzer finden", "Träger dieser Berechtigung können vorhandene Nutzer und ihre Eigenschaften anzeigen.")
	PermChangeOtherPassword = permission.Declare[FindAll]("nago.user.change_other_password", "Kennwort ändern", "Träger dieser Berechtigung können vorhandene Nutzer und ihre Eigenschaften anzeigen.")
	PermDelete              = permission.Declare[FindAll]("nago.user.delete", "Nutzer Löschen", "Träger dieser Berechtigung können vorhandene Nutzer und ihre Eigenschaften anzeigen.")
)

package usercircle

import "go.wdy.de/nago/application/permission"

var (
	PermCreate     = permission.Declare[Create]("nago.usercircle.create", "Nutzerkreis anlegen", "Träger dieser Berechtigung können neue Nutzerkreise anlegen.")
	PermUpdate     = permission.Declare[Update]("nago.usercircle.update", "Nutzerkreis aktualisieren", "Träger dieser Berechtigung können einen existierenden Nutzerkreise bearbeiten.")
	PermFindByID   = permission.Declare[FindByID]("nago.usercircle.find_by_id", "Nutzerkreis laden", "Träger dieser Berechtigung können existierende Nutzerkreise per ID laden.")
	PermFindAll    = permission.Declare[FindAll]("nago.usercircle.find_all", "Nutzerkreise laden", "Träger dieser Berechtigung können alle Nutzerkreise laden.")
	PermDeleteByID = permission.Declare[DeleteByID]("nago.usercircle.delete", "Nutzerkreis löschen", "Träger dieser Berechtigung können existierende Nutzerkreise entfernen.")
)
